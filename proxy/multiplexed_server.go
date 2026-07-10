package proxy

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"goproxy/config"
	"goproxy/storage"
)

const defaultProtocolDetectTimeout = 5 * time.Second

// MultiplexedServer 在同一个 TCP 端口上同时提供 HTTP 和 SOCKS5 代理服务。
type MultiplexedServer struct {
	http          *Server
	socks5        *SOCKS5Server
	port          string
	detectTimeout time.Duration
}

// NewMultiplexed 创建随机或最低延迟模式的双协议代理服务器。
func NewMultiplexed(store *storage.Storage, cfg *config.Config, mode, port string) *MultiplexedServer {
	return &MultiplexedServer{
		http:          New(store, cfg, mode, port),
		socks5:        NewSOCKS5(store, cfg, mode, port),
		port:          port,
		detectTimeout: defaultProtocolDetectTimeout,
	}
}

// NewMultiplexedFixed 创建绑定到指定上游节点的双协议代理服务器。
func NewMultiplexedFixed(store *storage.Storage, cfg *config.Config, port, proxyAddr string) *MultiplexedServer {
	return &MultiplexedServer{
		http:          NewFixed(store, cfg, port, proxyAddr),
		socks5:        NewSOCKS5Fixed(store, cfg, port, proxyAddr),
		port:          port,
		detectTimeout: defaultProtocolDetectTimeout,
	}
}

// Start 启动双协议代理服务器。
func (s *MultiplexedServer) Start() error {
	listener, err := net.Listen("tcp", s.port)
	if err != nil {
		return err
	}
	defer listener.Close()

	log.Printf("[proxy] HTTP+SOCKS5 server listening on %s [%s]", s.port, s.modeDescription())
	return s.Serve(listener)
}

func (s *MultiplexedServer) modeDescription() string {
	switch s.http.mode {
	case "lowest-latency":
		return "最低延迟"
	case "fixed":
		return fmt.Sprintf("固定节点 %s", s.http.fixedProxyAddr)
	default:
		return "随机轮换"
	}
}

// Serve 在已创建的监听器上运行双协议代理服务。
func (s *MultiplexedServer) Serve(listener net.Listener) error {
	httpListener := newConnectionListener(listener.Addr())
	httpErr := make(chan error, 1)
	go func() {
		err := s.http.Serve(httpListener)
		if err != nil && !errors.Is(err, net.ErrClosed) && !errors.Is(err, http.ErrServerClosed) {
			httpErr <- err
			_ = listener.Close()
		}
	}()

	var pendingMu sync.Mutex
	pending := make(map[net.Conn]struct{})
	var detectWG sync.WaitGroup

	closePending := func() {
		pendingMu.Lock()
		for conn := range pending {
			_ = conn.Close()
		}
		pendingMu.Unlock()
	}

	defer func() {
		_ = httpListener.Close()
		closePending()
		detectWG.Wait()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case serveErr := <-httpErr:
				return serveErr
			default:
				return err
			}
		}

		pendingMu.Lock()
		pending[conn] = struct{}{}
		pendingMu.Unlock()
		detectWG.Add(1)
		go func() {
			defer detectWG.Done()
			s.dispatchConnection(conn, httpListener, &pendingMu, pending)
		}()
	}
}

func (s *MultiplexedServer) dispatchConnection(conn net.Conn, httpListener *connectionListener, pendingMu *sync.Mutex, pending map[net.Conn]struct{}) {
	reader := bufio.NewReader(conn)
	if s.detectTimeout > 0 {
		_ = conn.SetReadDeadline(time.Now().Add(s.detectTimeout))
	}
	first, err := reader.Peek(1)
	_ = conn.SetReadDeadline(time.Time{})

	pendingMu.Lock()
	delete(pending, conn)
	pendingMu.Unlock()

	if err != nil {
		_ = conn.Close()
		return
	}

	wrapped := &bufferedConn{Conn: conn, reader: reader}
	if first[0] == 0x05 {
		go s.socks5.handleConnection(wrapped)
		return
	}
	if !httpListener.deliver(wrapped) {
		_ = wrapped.Close()
	}
}

type bufferedConn struct {
	net.Conn
	reader *bufio.Reader
}

func (c *bufferedConn) Read(p []byte) (int, error) {
	return c.reader.Read(p)
}

type connectionListener struct {
	addr      net.Addr
	conns     chan net.Conn
	done      chan struct{}
	closeOnce sync.Once
}

func newConnectionListener(addr net.Addr) *connectionListener {
	return &connectionListener{
		addr:  addr,
		conns: make(chan net.Conn),
		done:  make(chan struct{}),
	}
}

func (l *connectionListener) Accept() (net.Conn, error) {
	select {
	case conn := <-l.conns:
		return conn, nil
	case <-l.done:
		return nil, net.ErrClosed
	}
}

func (l *connectionListener) Close() error {
	l.closeOnce.Do(func() {
		close(l.done)
	})
	return nil
}

func (l *connectionListener) Addr() net.Addr {
	return l.addr
}

func (l *connectionListener) deliver(conn net.Conn) bool {
	select {
	case l.conns <- conn:
		return true
	case <-l.done:
		return false
	}
}
