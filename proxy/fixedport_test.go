package proxy

import (
	"bufio"
	"io"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"goproxy/config"
)

func TestFixedPortLegacyBindingsServeBothProtocols(t *testing.T) {
	for _, legacyProtocol := range []string{"http", "socks5"} {
		t.Run(legacyProtocol, func(t *testing.T) {
			port := freeTCPPort(t)
			cfg := config.DefaultConfig()
			cfg.ProxyAuthEnabled = true
			manager := NewFixedPortManager(nil, cfg)
			manager.Apply([]config.FixedPortBinding{{
				Port:         port,
				Protocol:     legacyProtocol,
				ProxyAddress: "127.0.0.1:1",
			}})
			defer manager.Stop()

			address := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
			waitForTCPPort(t, address)

			t.Run("http", func(t *testing.T) {
				conn, err := net.Dial("tcp", address)
				if err != nil {
					t.Fatal(err)
				}
				defer conn.Close()
				_ = conn.SetDeadline(time.Now().Add(time.Second))
				_, _ = io.WriteString(conn, "GET http://example.com/ HTTP/1.1\r\nHost: example.com\r\n\r\n")
				status, err := bufio.NewReader(conn).ReadString('\n')
				if err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(status, "407") {
					t.Fatalf("expected HTTP 407, got %q", status)
				}
			})

			t.Run("socks5", func(t *testing.T) {
				conn, err := net.Dial("tcp", address)
				if err != nil {
					t.Fatal(err)
				}
				defer conn.Close()
				_ = conn.SetDeadline(time.Now().Add(time.Second))
				_, _ = conn.Write([]byte{0x05, 0x01, 0x00})
				reply := make([]byte, 2)
				if _, err := io.ReadFull(conn, reply); err != nil {
					t.Fatal(err)
				}
				if reply[0] != 0x05 || reply[1] != 0xff {
					t.Fatalf("expected SOCKS5 method rejection, got %v", reply)
				}
			})
		})
	}
}

func freeTCPPort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

func waitForTCPPort(t *testing.T, address string) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", address, 20*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("port %s did not become ready", address)
}
