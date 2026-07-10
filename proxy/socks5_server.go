package proxy

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"goproxy/config"
	"goproxy/proxyutil"
	"goproxy/storage"
)

// SOCKS5Server SOCKS5 协议服务器
type SOCKS5Server struct {
	storage        *storage.Storage
	cfg            *config.Config
	mode           string // "random"、"lowest-latency" 或 "fixed"
	port           string
	fixedProxyAddr string
}

// NewSOCKS5 创建 SOCKS5 服务器
func NewSOCKS5(s *storage.Storage, cfg *config.Config, mode string, port string) *SOCKS5Server {
	return &SOCKS5Server{
		storage: s,
		cfg:     cfg,
		mode:    mode,
		port:    port,
	}
}

// NewSOCKS5Fixed 创建绑定到指定上游节点的 SOCKS5 代理服务器
func NewSOCKS5Fixed(s *storage.Storage, cfg *config.Config, port, proxyAddr string) *SOCKS5Server {
	return &SOCKS5Server{
		storage:        s,
		cfg:            cfg,
		mode:           "fixed",
		port:           port,
		fixedProxyAddr: proxyAddr,
	}
}

// Start 启动 SOCKS5 服务器
func (s *SOCKS5Server) Start() error {
	modeDesc := "随机轮换"
	if s.mode == "lowest-latency" {
		modeDesc = "最低延迟"
	} else if s.mode == "fixed" {
		modeDesc = fmt.Sprintf("固定节点 %s", s.fixedProxyAddr)
	}
	authStatus := "无认证"
	if s.cfg.ProxyAuthEnabled {
		authStatus = fmt.Sprintf("需认证 (用户: %s)", s.cfg.ProxyAuthUsername)
	}
	log.Printf("socks5 server listening on %s [%s] [%s]", s.port, modeDesc, authStatus)

	listener, err := net.Listen("tcp", s.port)
	if err != nil {
		return err
	}
	defer listener.Close()

	return s.Serve(listener)
}

// Serve 在已创建的监听器上运行 SOCKS5 代理，供固定端口管理器动态启停
func (s *SOCKS5Server) Serve(listener net.Listener) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go s.handleConnection(conn)
	}
}

// handleConnection 处理 SOCKS5 连接
func (s *SOCKS5Server) handleConnection(clientConn net.Conn) {
	defer clientConn.Close()

	// 握手与请求阶段使用有限超时，成功建立上游后清除，不影响长连接。
	handshakeTimeout := time.Duration(s.cfg.ValidateTimeout) * time.Second
	if handshakeTimeout > 0 {
		_ = clientConn.SetDeadline(time.Now().Add(handshakeTimeout))
	}

	// SOCKS5 握手
	if err := s.socks5Handshake(clientConn); err != nil {
		log.Printf("[socks5] handshake failed: %v", err)
		return
	}

	// 读取请求
	target, err := s.readSOCKS5Request(clientConn)
	if err != nil {
		log.Printf("[socks5] read request failed: %v", err)
		return
	}

	// 固定端口仅尝试绑定节点一次，其他模式增加重试次数以应对质量差的代理
	tried := []string{}
	maxRetries := s.cfg.MaxRetry + 2
	if s.mode == "fixed" {
		maxRetries = 0
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		p, err := s.selectSOCKS5Proxy(tried)
		if err != nil {
			log.Printf("[socks5] no available socks5 upstream proxy: %v", err)
			s.sendSOCKS5Reply(clientConn, 0x01) // General failure
			return
		}

		tried = append(tried, p.Address)

		// 连接上游代理
		upstreamConn, err := s.dialViaProxy(p, target)
		if err != nil {
			log.Printf("[socks5] dial %s via %s (%s) failed: %v, removing", target, p.Address, p.Protocol, err)
			s.storage.RecordProxyUse(p.Address, false)
			removeOrDisableProxy(s.storage, p)
			continue
		}

		// 发送成功响应
		if err := s.sendSOCKS5Reply(clientConn, 0x00); err != nil {
			upstreamConn.Close()
			return
		}
		_ = clientConn.SetDeadline(time.Time{})

		s.storage.RecordProxyUse(p.Address, true)
		log.Printf("[socks5] %s via %s established", target, p.Address)

		// 双向转发数据
		go io.Copy(upstreamConn, clientConn)
		io.Copy(clientConn, upstreamConn)

		// 转发完成，关闭连接
		upstreamConn.Close()
		return
	}

	// 所有重试都失败
	s.sendSOCKS5Reply(clientConn, 0x01) // General failure
	log.Printf("[socks5] all proxies failed for %s", target)
}

// selectSOCKS5Proxy 根据使用模式选择 SOCKS5 上游代理
func (s *SOCKS5Server) selectSOCKS5Proxy(tried []string) (*storage.Proxy, error) {
	if s.mode == "fixed" {
		return s.storage.GetByAddress(s.fixedProxyAddr)
	}

	cfg := config.Get()
	sourceFilter := sourceFilterFromMode(cfg.CustomProxyMode)

	// 混用 + 优先模式
	if cfg.CustomProxyMode == "mixed" && (cfg.CustomPriority || cfg.CustomFreePriority) {
		preferSource := "custom"
		if cfg.CustomFreePriority {
			preferSource = "free"
		}
		var p *storage.Proxy
		var err error
		if s.mode == "lowest-latency" {
			p, err = s.storage.GetLowestLatencyByProtocolExcludeFiltered("socks5", tried, preferSource)
		} else {
			p, err = s.storage.GetRandomByProtocolExcludeFiltered("socks5", tried, preferSource)
		}
		if err == nil {
			return p, nil
		}
		// fallback
		if s.mode == "lowest-latency" {
			return s.storage.GetLowestLatencyByProtocolExcludeFiltered("socks5", tried, "")
		}
		return s.storage.GetRandomByProtocolExcludeFiltered("socks5", tried, "")
	}

	if s.mode == "lowest-latency" {
		return s.storage.GetLowestLatencyByProtocolExcludeFiltered("socks5", tried, sourceFilter)
	}
	return s.storage.GetRandomByProtocolExcludeFiltered("socks5", tried, sourceFilter)
}

// socks5Handshake 处理 SOCKS5 握手
func (s *SOCKS5Server) socks5Handshake(conn net.Conn) error {
	// 读取客户端问候: [VER(1), NMETHODS(1), METHODS(1-255)]
	header := make([]byte, 2)
	if _, err := io.ReadFull(conn, header); err != nil {
		return err
	}
	if header[0] != 0x05 {
		return fmt.Errorf("unsupported SOCKS version: %d", header[0])
	}

	methods := make([]byte, int(header[1]))
	if _, err := io.ReadFull(conn, methods); err != nil {
		return err
	}

	// 检查是否需要认证
	needAuth := s.cfg.ProxyAuthEnabled

	// 选择认证方式
	var selectedMethod byte = 0xFF // No acceptable methods
	if needAuth {
		// 需要用户名/密码认证 (0x02)
		for _, method := range methods {
			if method == 0x02 {
				selectedMethod = 0x02
				break
			}
		}
	} else {
		// 无需认证 (0x00)
		for _, method := range methods {
			if method == 0x00 {
				selectedMethod = 0x00
				break
			}
		}
	}

	// 发送方法选择: [VER(1), METHOD(1)]
	if _, err := conn.Write([]byte{0x05, selectedMethod}); err != nil {
		return err
	}

	if selectedMethod == 0xFF {
		return fmt.Errorf("no acceptable authentication method")
	}

	// 如果需要认证，进行用户名/密码认证
	if selectedMethod == 0x02 {
		if err := s.socks5Auth(conn); err != nil {
			return err
		}
	}

	return nil
}

// socks5Auth 处理 SOCKS5 用户名/密码认证
func (s *SOCKS5Server) socks5Auth(conn net.Conn) error {
	// 读取认证请求: [VER(1), ULEN(1), UNAME(1-255), PLEN(1), PASSWD(1-255)]
	header := make([]byte, 2)
	if _, err := io.ReadFull(conn, header); err != nil {
		return err
	}
	if header[0] != 0x01 {
		return fmt.Errorf("unsupported auth version: %d", header[0])
	}

	usernameBytes := make([]byte, int(header[1]))
	if _, err := io.ReadFull(conn, usernameBytes); err != nil {
		return err
	}

	passwordLength := make([]byte, 1)
	if _, err := io.ReadFull(conn, passwordLength); err != nil {
		return err
	}
	passwordBytes := make([]byte, int(passwordLength[0]))
	if _, err := io.ReadFull(conn, passwordBytes); err != nil {
		return err
	}

	// 验证用户名和密码
	if string(usernameBytes) != s.cfg.ProxyAuthUsername || string(passwordBytes) != s.cfg.ProxyAuthPassword {
		// 认证失败: [VER(1), STATUS(1)]
		conn.Write([]byte{0x01, 0x01})
		return fmt.Errorf("authentication failed")
	}

	// 认证成功: [VER(1), STATUS(1)]
	if _, err := conn.Write([]byte{0x01, 0x00}); err != nil {
		return err
	}

	return nil
}

// readSOCKS5Request 读取 SOCKS5 请求
func (s *SOCKS5Server) readSOCKS5Request(conn net.Conn) (string, error) {
	// 读取请求头: [VER(1), CMD(1), RSV(1), ATYP(1)]
	header := make([]byte, 4)
	if _, err := io.ReadFull(conn, header); err != nil {
		return "", err
	}
	if header[0] != 0x05 {
		return "", fmt.Errorf("invalid version: %d", header[0])
	}
	if header[1] != 0x01 { // 只支持 CONNECT
		s.sendSOCKS5Reply(conn, 0x07) // Command not supported
		return "", fmt.Errorf("unsupported command: %d", header[1])
	}

	var host string
	switch header[3] {
	case 0x01: // IPv4
		address := make([]byte, net.IPv4len)
		if _, err := io.ReadFull(conn, address); err != nil {
			return "", err
		}
		host = net.IP(address).String()
	case 0x03: // Domain name
		length := make([]byte, 1)
		if _, err := io.ReadFull(conn, length); err != nil {
			return "", err
		}
		address := make([]byte, int(length[0]))
		if _, err := io.ReadFull(conn, address); err != nil {
			return "", err
		}
		host = string(address)
	case 0x04: // IPv6
		address := make([]byte, net.IPv6len)
		if _, err := io.ReadFull(conn, address); err != nil {
			return "", err
		}
		host = net.IP(address).String()
	default:
		s.sendSOCKS5Reply(conn, 0x08) // Address type not supported
		return "", fmt.Errorf("unsupported address type: %d", header[3])
	}

	portBytes := make([]byte, 2)
	if _, err := io.ReadFull(conn, portBytes); err != nil {
		return "", err
	}
	port := binary.BigEndian.Uint16(portBytes)
	return net.JoinHostPort(host, fmt.Sprintf("%d", port)), nil
}

// sendSOCKS5Reply 发送 SOCKS5 响应
func (s *SOCKS5Server) sendSOCKS5Reply(conn net.Conn, rep byte) error {
	// [VER(1), REP(1), RSV(1), ATYP(1), BND.ADDR(variable), BND.PORT(2)]
	// 简化：使用 0.0.0.0:0
	reply := []byte{
		0x05,       // VER
		rep,        // REP: 0x00=成功, 0x01=一般失败, 0x07=命令不支持, 0x08=地址类型不支持
		0x00,       // RSV
		0x01,       // ATYP: IPv4
		0, 0, 0, 0, // BND.ADDR: 0.0.0.0
		0, 0, // BND.PORT: 0
	}
	_, err := conn.Write(reply)
	return err
}

// dialViaProxy 通过上游代理连接目标
func (s *SOCKS5Server) dialViaProxy(p *storage.Proxy, target string) (net.Conn, error) {
	timeout := time.Duration(s.cfg.ValidateTimeout) * time.Second

	switch p.Protocol {
	case "http":
		return proxyutil.DialHTTPProxyConnect(p.Address, target, timeout)

	case "socks5":
		dialer, err := proxyutil.SOCKS5Dialer(p.Address)
		if err != nil {
			return nil, err
		}
		return dialer.Dial("tcp", target)

	default:
		return nil, fmt.Errorf("unsupported protocol: %s", p.Protocol)
	}
}
