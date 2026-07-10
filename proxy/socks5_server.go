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
	buf := make([]byte, 257)

	// 读取客户端问候: [VER(1), NMETHODS(1), METHODS(1-255)]
	n, err := io.ReadAtLeast(conn, buf, 2)
	if err != nil {
		return err
	}

	version := buf[0]
	if version != 0x05 {
		return fmt.Errorf("unsupported SOCKS version: %d", version)
	}

	nmethods := int(buf[1])
	if n < 2+nmethods {
		if _, err := io.ReadFull(conn, buf[n:2+nmethods]); err != nil {
			return err
		}
	}

	// 检查是否需要认证
	needAuth := s.cfg.ProxyAuthEnabled
	methods := buf[2 : 2+nmethods]

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
	buf := make([]byte, 513)

	// 读取认证请求: [VER(1), ULEN(1), UNAME(1-255), PLEN(1), PASSWD(1-255)]
	n, err := io.ReadAtLeast(conn, buf, 2)
	if err != nil {
		return err
	}

	if buf[0] != 0x01 {
		return fmt.Errorf("unsupported auth version: %d", buf[0])
	}

	ulen := int(buf[1])
	if n < 2+ulen {
		if _, err := io.ReadFull(conn, buf[n:2+ulen]); err != nil {
			return err
		}
		n = 2 + ulen
	}

	username := string(buf[2 : 2+ulen])

	// 读取密码长度和密码
	if n < 2+ulen+1 {
		if _, err := io.ReadFull(conn, buf[n:2+ulen+1]); err != nil {
			return err
		}
		n = 2 + ulen + 1
	}

	plen := int(buf[2+ulen])
	if n < 2+ulen+1+plen {
		if _, err := io.ReadFull(conn, buf[n:2+ulen+1+plen]); err != nil {
			return err
		}
	}

	password := string(buf[2+ulen+1 : 2+ulen+1+plen])

	// 验证用户名和密码
	if username != s.cfg.ProxyAuthUsername || password != s.cfg.ProxyAuthPassword {
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
	buf := make([]byte, 262)

	// 读取请求: [VER(1), CMD(1), RSV(1), ATYP(1), DST.ADDR(variable), DST.PORT(2)]
	n, err := io.ReadAtLeast(conn, buf, 4)
	if err != nil {
		return "", err
	}

	if buf[0] != 0x05 {
		return "", fmt.Errorf("invalid version: %d", buf[0])
	}

	cmd := buf[1]
	if cmd != 0x01 { // 只支持 CONNECT
		s.sendSOCKS5Reply(conn, 0x07) // Command not supported
		return "", fmt.Errorf("unsupported command: %d", cmd)
	}

	atyp := buf[3]
	var host string
	var addrLen int

	switch atyp {
	case 0x01: // IPv4
		addrLen = 4
		if n < 4+addrLen+2 {
			if _, err := io.ReadFull(conn, buf[n:4+addrLen+2]); err != nil {
				return "", err
			}
		}
		host = fmt.Sprintf("%d.%d.%d.%d", buf[4], buf[5], buf[6], buf[7])
	case 0x03: // Domain name
		addrLen = int(buf[4])
		if n < 4+1+addrLen+2 {
			if _, err := io.ReadFull(conn, buf[n:4+1+addrLen+2]); err != nil {
				return "", err
			}
		}
		host = string(buf[5 : 5+addrLen])
	case 0x04: // IPv6
		addrLen = 16
		if n < 4+addrLen+2 {
			if _, err := io.ReadFull(conn, buf[n:4+addrLen+2]); err != nil {
				return "", err
			}
		}
		// 简化处理，直接转换
		host = net.IP(buf[4 : 4+addrLen]).String()
	default:
		s.sendSOCKS5Reply(conn, 0x08) // Address type not supported
		return "", fmt.Errorf("unsupported address type: %d", atyp)
	}

	// 读取端口
	portOffset := 4
	if atyp == 0x03 {
		portOffset = 5 + addrLen
	} else {
		portOffset = 4 + addrLen
	}
	port := binary.BigEndian.Uint16(buf[portOffset : portOffset+2])

	return fmt.Sprintf("%s:%d", host, port), nil
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
