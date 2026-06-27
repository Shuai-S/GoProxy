package proxyutil

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

// ParsedProxy 统一后的代理地址
type ParsedProxy struct {
	URL      *url.URL
	Protocol string // http 或 socks5
}

// Parse 解析用户输入的代理地址，支持:
// socks5://user:pass@host:1080
// http://host:8080
// https://user:pass@host:443
// 以及不带 scheme 的 host:port（按 http 处理）
func Parse(raw string) (*ParsedProxy, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return nil, fmt.Errorf("proxy address is empty")
	}

	if !strings.Contains(s, "://") {
		s = "http://" + s
	}

	u, err := url.Parse(s)
	if err != nil {
		return nil, err
	}

	u.Scheme = strings.ToLower(u.Scheme)
	u.Path = ""
	u.RawPath = ""
	u.RawQuery = ""
	u.Fragment = ""
	u.ForceQuery = false
	u.Opaque = ""

	switch u.Scheme {
	case "http", "https":
		host := u.Hostname()
		if host == "" {
			return nil, fmt.Errorf("missing proxy host")
		}
		port := u.Port()
		if port == "" {
			if u.Scheme == "https" {
				port = "443"
			} else {
				port = "80"
			}
		}
		u.Host = net.JoinHostPort(host, port)
		return &ParsedProxy{URL: u, Protocol: "http"}, nil
	case "socks5", "socks5h":
		host := u.Hostname()
		if host == "" {
			return nil, fmt.Errorf("missing proxy host")
		}
		port := u.Port()
		if port == "" {
			port = "1080"
		}
		u.Scheme = "socks5"
		u.Host = net.JoinHostPort(host, port)
		return &ParsedProxy{URL: u, Protocol: "socks5"}, nil
	default:
		return nil, fmt.Errorf("unsupported proxy scheme: %s", u.Scheme)
	}
}

// HTTPURL 返回 HTTP/HTTPS 代理的标准化 URL
func HTTPURL(raw string) (*url.URL, error) {
	parsed, err := Parse(raw)
	if err != nil {
		return nil, err
	}
	if parsed.Protocol != "http" {
		return nil, fmt.Errorf("proxy scheme %s is not http/https", parsed.URL.Scheme)
	}
	return parsed.URL, nil
}

// SOCKS5Dialer 返回 SOCKS5 代理拨号器
func SOCKS5Dialer(raw string) (proxy.Dialer, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return nil, fmt.Errorf("proxy address is empty")
	}
	if !strings.Contains(s, "://") {
		s = "socks5://" + s
	}
	parsed, err := Parse(s)
	if err != nil {
		return nil, err
	}
	if parsed.Protocol != "socks5" {
		return nil, fmt.Errorf("proxy scheme %s is not socks5", parsed.URL.Scheme)
	}
	return proxy.FromURL(parsed.URL, proxy.Direct)
}

// ProxyAuthHeader 生成 Proxy-Authorization 头
func ProxyAuthHeader(u *url.URL) string {
	if u == nil || u.User == nil {
		return ""
	}
	user := u.User.Username()
	pass, _ := u.User.Password()
	token := base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
	return "Basic " + token
}

type bufferedConn struct {
	net.Conn
	reader *bufio.Reader
}

func (c *bufferedConn) Read(p []byte) (int, error) {
	return c.reader.Read(p)
}

// DialHTTPProxyConnect 通过 HTTP/HTTPS 代理建立 CONNECT 隧道
func DialHTTPProxyConnect(rawProxy, target string, timeout time.Duration) (net.Conn, error) {
	parsed, err := HTTPURL(rawProxy)
	if err != nil {
		return nil, err
	}

	dialer := &net.Dialer{Timeout: timeout}
	var conn net.Conn
	if parsed.Scheme == "https" {
		conn, err = tls.DialWithDialer(dialer, "tcp", parsed.Host, &tls.Config{
			ServerName: parsed.Hostname(),
		})
	} else {
		conn, err = dialer.Dial("tcp", parsed.Host)
	}
	if err != nil {
		return nil, err
	}

	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		conn.Close()
		return nil, err
	}

	if _, err := fmt.Fprintf(conn, "CONNECT %s HTTP/1.1\r\nHost: %s\r\n", target, target); err != nil {
		conn.Close()
		return nil, err
	}
	if auth := ProxyAuthHeader(parsed); auth != "" {
		if _, err := fmt.Fprintf(conn, "Proxy-Authorization: %s\r\n", auth); err != nil {
			conn.Close()
			return nil, err
		}
	}
	if _, err := io.WriteString(conn, "\r\n"); err != nil {
		conn.Close()
		return nil, err
	}

	reader := bufio.NewReader(conn)
	resp, err := http.ReadResponse(reader, &http.Request{Method: http.MethodConnect})
	if err != nil {
		conn.Close()
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		io.Copy(io.Discard, resp.Body)
		conn.Close()
		return nil, fmt.Errorf("upstream proxy connect failed: %s", resp.Status)
	}

	if err := conn.SetDeadline(time.Time{}); err != nil {
		conn.Close()
		return nil, err
	}

	return &bufferedConn{Conn: conn, reader: reader}, nil
}
