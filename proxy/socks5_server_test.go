package proxy

import (
	"bytes"
	"io"
	"net"
	"testing"
	"time"

	"goproxy/config"
)

func TestSOCKS5RequestReadsFragmentedDomain(t *testing.T) {
	server := NewSOCKS5(nil, config.DefaultConfig(), "random", "")
	client, service := net.Pipe()
	defer client.Close()
	defer service.Close()

	result := make(chan struct {
		target string
		err    error
	}, 1)
	go func() {
		target, err := server.readSOCKS5Request(service)
		result <- struct {
			target string
			err    error
		}{target: target, err: err}
	}()

	parts := [][]byte{
		{0x05, 0x01, 0x00, 0x03},
		{0x0b},
		[]byte("example.com"),
		{0x01, 0xbb},
	}
	for _, part := range parts {
		if _, err := client.Write(part); err != nil {
			t.Fatal(err)
		}
	}

	select {
	case got := <-result:
		if got.err != nil {
			t.Fatal(got.err)
		}
		if got.target != "example.com:443" {
			t.Fatalf("expected example.com:443, got %s", got.target)
		}
	case <-time.After(time.Second):
		t.Fatal("request parser timed out")
	}
}

func TestSOCKS5HandshakePreservesFollowingRequest(t *testing.T) {
	cfg := config.DefaultConfig()
	server := NewSOCKS5(nil, cfg, "random", "")
	input := append([]byte{0x05, 0x01, 0x00}, []byte{
		0x05, 0x01, 0x00, 0x01,
		127, 0, 0, 1,
		0x1f, 0x90,
	}...)
	conn := newMemoryConn(input)

	if err := server.socks5Handshake(conn); err != nil {
		t.Fatal(err)
	}
	target, err := server.readSOCKS5Request(conn)
	if err != nil {
		t.Fatal(err)
	}
	if target != "127.0.0.1:8080" {
		t.Fatalf("expected 127.0.0.1:8080, got %s", target)
	}
	if got := conn.writes.Bytes(); !bytes.Equal(got, []byte{0x05, 0x00}) {
		t.Fatalf("unexpected handshake reply: %v", got)
	}
}

func TestSOCKS5UsernamePasswordAuth(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.ProxyAuthEnabled = true
	cfg.ProxyAuthUsername = "proxy"
	cfg.ProxyAuthPassword = "secret"
	server := NewSOCKS5(nil, cfg, "random", "")

	tests := []struct {
		name       string
		password   string
		wantErr    bool
		wantStatus byte
	}{
		{name: "valid", password: "secret", wantStatus: 0x00},
		{name: "invalid", password: "wrong", wantErr: true, wantStatus: 0x01},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := []byte{0x05, 0x01, 0x02, 0x01, 0x05}
			input = append(input, []byte("proxy")...)
			input = append(input, byte(len(tt.password)))
			input = append(input, []byte(tt.password)...)
			conn := newMemoryConn(input)

			err := server.socks5Handshake(conn)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}
			want := []byte{0x05, 0x02, 0x01, tt.wantStatus}
			if got := conn.writes.Bytes(); !bytes.Equal(got, want) {
				t.Fatalf("expected reply %v, got %v", want, got)
			}
		})
	}
}

type memoryConn struct {
	reader *bytes.Reader
	writes bytes.Buffer
}

func newMemoryConn(input []byte) *memoryConn {
	return &memoryConn{reader: bytes.NewReader(input)}
}

func (c *memoryConn) Read(p []byte) (int, error)       { return c.reader.Read(p) }
func (c *memoryConn) Write(p []byte) (int, error)      { return c.writes.Write(p) }
func (c *memoryConn) Close() error                     { return nil }
func (c *memoryConn) LocalAddr() net.Addr              { return memoryAddr("local") }
func (c *memoryConn) RemoteAddr() net.Addr             { return memoryAddr("remote") }
func (c *memoryConn) SetDeadline(time.Time) error      { return nil }
func (c *memoryConn) SetReadDeadline(time.Time) error  { return nil }
func (c *memoryConn) SetWriteDeadline(time.Time) error { return nil }

var _ net.Conn = (*memoryConn)(nil)
var _ io.Reader = (*memoryConn)(nil)

type memoryAddr string

func (a memoryAddr) Network() string { return "memory" }
func (a memoryAddr) String() string  { return string(a) }
