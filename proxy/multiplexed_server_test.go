package proxy

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"goproxy/config"
	"goproxy/storage"
)

func TestMultiplexedServerRoutesHTTPAndSOCKS5(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.ProxyAuthEnabled = true
	cfg.ProxyAuthUsername = "proxy"
	cfg.ProxyAuthPassword = "secret"

	server := NewMultiplexed(nil, cfg, "random", "")
	server.detectTimeout = 200 * time.Millisecond
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	serveDone := make(chan error, 1)
	go func() {
		serveDone <- server.Serve(listener)
	}()
	defer func() {
		_ = listener.Close()
		select {
		case <-serveDone:
		case <-time.After(time.Second):
			t.Fatal("server did not stop")
		}
	}()

	t.Run("http", func(t *testing.T) {
		conn, err := net.Dial("tcp", listener.Addr().String())
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		if _, err := io.WriteString(conn, "GET http://example.com/ HTTP/1.1\r\nHost: example.com\r\n\r\n"); err != nil {
			t.Fatal(err)
		}
		status, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(status, "407") {
			t.Fatalf("expected HTTP 407, got %q", status)
		}
	})

	t.Run("socks5", func(t *testing.T) {
		conn, err := net.Dial("tcp", listener.Addr().String())
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()

		if _, err := conn.Write([]byte{0x05, 0x01, 0x00}); err != nil {
			t.Fatal(err)
		}
		reply := make([]byte, 2)
		if _, err := io.ReadFull(conn, reply); err != nil {
			t.Fatal(err)
		}
		if reply[0] != 0x05 || reply[1] != 0xff {
			t.Fatalf("expected SOCKS5 method rejection, got %v", reply)
		}
	})
}

func TestMultiplexedFixedServerForwardsHTTPAndSOCKS5(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "multiplexed-ok")
	}))
	defer target.Close()

	upstream := newTestHTTPProxy(t)
	defer upstream.Close()

	store, err := storage.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.AddProxyWithSource(upstream.Addr().String(), "http", "custom"); err != nil {
		t.Fatal(err)
	}

	cfg := config.DefaultConfig()
	cfg.ValidateTimeout = 2
	server := NewMultiplexedFixed(store, cfg, "", upstream.Addr().String())
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	serveDone := make(chan error, 1)
	go func() { serveDone <- server.Serve(listener) }()
	defer func() {
		_ = listener.Close()
		<-serveDone
	}()

	t.Run("http", func(t *testing.T) {
		proxyURL, err := url.Parse("http://" + listener.Addr().String())
		if err != nil {
			t.Fatal(err)
		}
		client := &http.Client{
			Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)},
			Timeout:   2 * time.Second,
		}
		resp, err := client.Get(target.URL)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		if string(body) != "multiplexed-ok" {
			t.Fatalf("unexpected body %q", body)
		}
	})

	t.Run("socks5", func(t *testing.T) {
		conn, err := net.Dial("tcp", listener.Addr().String())
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close()
		_ = conn.SetDeadline(time.Now().Add(2 * time.Second))

		if _, err := conn.Write([]byte{0x05, 0x01, 0x00}); err != nil {
			t.Fatal(err)
		}
		methodReply := make([]byte, 2)
		if _, err := io.ReadFull(conn, methodReply); err != nil {
			t.Fatal(err)
		}
		if methodReply[1] != 0x00 {
			t.Fatalf("unexpected method reply %v", methodReply)
		}

		targetURL, err := url.Parse(target.URL)
		if err != nil {
			t.Fatal(err)
		}
		host, portText, err := net.SplitHostPort(targetURL.Host)
		if err != nil {
			t.Fatal(err)
		}
		port, err := strconv.Atoi(portText)
		if err != nil {
			t.Fatal(err)
		}
		ip := net.ParseIP(host).To4()
		request := []byte{0x05, 0x01, 0x00, 0x01, ip[0], ip[1], ip[2], ip[3], byte(port >> 8), byte(port)}
		if _, err := conn.Write(request); err != nil {
			t.Fatal(err)
		}
		reply := make([]byte, 10)
		if _, err := io.ReadFull(conn, reply); err != nil {
			t.Fatal(err)
		}
		if reply[1] != 0x00 {
			t.Fatalf("unexpected connect reply %v", reply)
		}

		if _, err := fmt.Fprintf(conn, "GET / HTTP/1.1\r\nHost: %s\r\nConnection: close\r\n\r\n", targetURL.Host); err != nil {
			t.Fatal(err)
		}
		resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		if string(body) != "multiplexed-ok" {
			t.Fatalf("unexpected body %q", body)
		}
	})
}

func newTestHTTPProxy(t *testing.T) net.Listener {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodConnect {
			upstream, err := net.Dial("tcp", r.Host)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}
			hijacker := w.(http.Hijacker)
			client, _, err := hijacker.Hijack()
			if err != nil {
				upstream.Close()
				return
			}
			_, _ = io.WriteString(client, "HTTP/1.1 200 Connection Established\r\n\r\n")
			go transfer(upstream, client)
			go transfer(client, upstream)
			return
		}

		req := r.Clone(r.Context())
		req.RequestURI = ""
		resp, err := http.DefaultTransport.RoundTrip(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		for key, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
	})}
	go server.Serve(listener)
	return listener
}

func TestMultiplexedServerSlowClientDoesNotBlock(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.ProxyAuthEnabled = true
	server := NewMultiplexed(nil, cfg, "random", "")
	server.detectTimeout = 100 * time.Millisecond

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	serveDone := make(chan error, 1)
	go func() {
		serveDone <- server.Serve(listener)
	}()
	defer func() {
		_ = listener.Close()
		<-serveDone
	}()

	idle, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer idle.Close()

	active, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer active.Close()
	_ = active.SetDeadline(time.Now().Add(time.Second))
	if _, err := io.WriteString(active, "GET http://example.com/ HTTP/1.1\r\nHost: example.com\r\n\r\n"); err != nil {
		t.Fatal(err)
	}
	status, err := bufio.NewReader(active).ReadString('\n')
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(status, "407") {
		t.Fatalf("expected HTTP 407, got %q", status)
	}

	_ = idle.SetReadDeadline(time.Now().Add(time.Second))
	if _, err := idle.Read(make([]byte, 1)); err == nil {
		t.Fatal("expected idle connection to close after detection timeout")
	}
}
