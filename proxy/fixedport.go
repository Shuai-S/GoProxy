package proxy

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"reflect"
	"sync"

	"goproxy/config"
	"goproxy/storage"
)

// FixedPortManager 动态管理固定端口监听器
// 配置变化时只重启新增或发生变化的端口，不影响未变化的连接。
type FixedPortManager struct {
	mu        sync.Mutex
	storage   *storage.Storage
	cfg       *config.Config
	listeners map[int]*fixedListener
}

type fixedListener struct {
	binding  config.FixedPortBinding
	listener net.Listener
}

// NewFixedPortManager 创建固定端口管理器
func NewFixedPortManager(store *storage.Storage, cfg *config.Config) *FixedPortManager {
	return &FixedPortManager{
		storage:   store,
		cfg:       cfg,
		listeners: make(map[int]*fixedListener),
	}
}

// Apply 应用固定端口配置，动态停止、启动或重启对应监听器
func (m *FixedPortManager) Apply(bindings []config.FixedPortBinding) {
	m.mu.Lock()
	defer m.mu.Unlock()

	wanted := make(map[int]config.FixedPortBinding, len(bindings))
	for _, binding := range bindings {
		wanted[binding.Port] = binding
	}

	for port, current := range m.listeners {
		binding, exists := wanted[port]
		if exists && reflect.DeepEqual(current.binding, binding) {
			continue
		}
		if err := current.listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			log.Printf("[fixed] 关闭端口 %d 失败: %v", port, err)
		}
		delete(m.listeners, port)
		log.Printf("[fixed] 已停止固定端口 %d", port)
	}

	for port, binding := range wanted {
		if _, exists := m.listeners[port]; exists {
			continue
		}
		m.startLocked(binding)
	}
}

func (m *FixedPortManager) startLocked(binding config.FixedPortBinding) {
	addr := fmt.Sprintf(":%d", binding.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("[fixed] 启动端口 %d 失败: %v", binding.Port, err)
		return
	}

	m.listeners[binding.Port] = &fixedListener{
		binding:  binding,
		listener: listener,
	}

	log.Printf("[fixed] 端口 %d 已绑定 %s [%s]", binding.Port, binding.ProxyAddress, binding.Protocol)
	go func() {
		var serveErr error
		if binding.Protocol == "socks5" {
			serveErr = NewSOCKS5Fixed(m.storage, m.cfg, addr, binding.ProxyAddress).Serve(listener)
		} else {
			serveErr = NewFixed(m.storage, m.cfg, addr, binding.ProxyAddress).Serve(listener)
		}
		if serveErr != nil && !errors.Is(serveErr, net.ErrClosed) && !errors.Is(serveErr, http.ErrServerClosed) {
			log.Printf("[fixed] 端口 %d 服务异常退出: %v", binding.Port, serveErr)
		}
	}()
}

// Stop 关闭所有固定端口监听器
func (m *FixedPortManager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for port, current := range m.listeners {
		if err := current.listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			log.Printf("[fixed] 关闭端口 %d 失败: %v", port, err)
		}
	}
	m.listeners = make(map[int]*fixedListener)
}
