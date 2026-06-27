package validator

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"goproxy/config"
	"goproxy/proxyutil"
	"goproxy/storage"
)

type Validator struct {
	concurrency   int
	timeout       time.Duration
	validateURL   string
	maxResponseMs int
	cfg           *config.Config
}

func concurrencyBuffer(total, concurrency int) int {
	if total < concurrency*10 {
		return total
	}
	return concurrency * 10
}

func New(concurrency, timeoutSec int, validateURL string) *Validator {
	cfg := config.Get()
	maxMs := 0
	if cfg != nil {
		maxMs = cfg.MaxResponseMs
	}
	return &Validator{
		concurrency:   concurrency,
		timeout:       time.Duration(timeoutSec) * time.Second,
		validateURL:   validateURL,
		maxResponseMs: maxMs,
		cfg:           cfg,
	}
}

type Result struct {
	Proxy         storage.Proxy
	Valid         bool
	Latency       time.Duration
	ExitIP        string
	ExitLocation  string
	IPType        string
	RiskScore     float64
	RiskLevel     string
	IsResidential bool
	Reason        string
}

type exitIPInfo struct {
	IP            string
	Location      string
	IPType        string
	RiskScore     float64
	RiskLevel     string
	IsResidential bool
}

// getExitIPInfo 通过代理获取出口 IP 和地理位置
func getExitIPInfo(client *http.Client) exitIPInfo {
	ip, location, isProxy, isHosting, riskKnown := queryIPAPI(client)
	ipinfoIP, ipinfoCountry, ipType, isResidential := queryIPInfo(client)

	if ip == "" {
		ip = ipinfoIP
	}
	if location == "" && ipinfoCountry != "" {
		location = ipinfoCountry
	}

	riskScore, riskLevel := calculateRisk(isProxy, isHosting, riskKnown)
	if isHosting {
		ipType = "Datacenter"
		isResidential = false
	}

	return exitIPInfo{
		IP:            ip,
		Location:      location,
		IPType:        ipType,
		RiskScore:     riskScore,
		RiskLevel:     riskLevel,
		IsResidential: isResidential,
	}
}

// queryIPAPI 使用 ip-api.com 获取出口 IP、位置和风险基础信息
func queryIPAPI(client *http.Client) (string, string, bool, bool, bool) {
	resp, err := client.Get("http://ip-api.com/json/?fields=status,country,countryCode,city,query,proxy,hosting")
	if err != nil {
		return "", "", false, false, false
	}
	defer resp.Body.Close()

	var result struct {
		Status      string `json:"status"`
		Query       string `json:"query"` // IP 地址
		Country     string `json:"country"`
		CountryCode string `json:"countryCode"`
		City        string `json:"city"`
		Proxy       bool   `json:"proxy"`
		Hosting     bool   `json:"hosting"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || result.Status != "success" {
		return "", "", false, false, false
	}

	location := result.CountryCode
	if result.City != "" {
		location = fmt.Sprintf("%s %s", result.CountryCode, result.City)
	}

	return result.Query, location, result.Proxy, result.Hosting, true
}

// queryIPInfo 使用 ipinfo.io 获取更丰富的 ASN/公司类型信息，用于判断 IP 类型和住宅属性
func queryIPInfo(client *http.Client) (string, string, string, bool) {
	resp, err := client.Get("https://ipinfo.io/json")
	if err != nil {
		return "", "", "", false
	}
	defer resp.Body.Close()

	var result struct {
		IP      string `json:"ip"`
		Country string `json:"country"`
		Org     string `json:"org"`
		Company struct {
			Type string `json:"type"`
		} `json:"company"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", "", "", false
	}

	if result.Company.Type != "" {
		ipType := result.Company.Type
		return result.IP, result.Country, ipType, strings.EqualFold(ipType, "isp")
	}

	if looksLikeDatacenterOrg(result.Org) {
		return result.IP, result.Country, "Datacenter", false
	}
	if result.Org != "" {
		return result.IP, result.Country, "ISP", true
	}

	return result.IP, result.Country, "", false
}

func calculateRisk(isProxy, isHosting, known bool) (float64, string) {
	if !known {
		return 0, "Unknown"
	}
	switch {
	case isProxy && isHosting:
		return 0.9, "Very High"
	case isProxy:
		return 0.7, "High"
	case isHosting:
		return 0.5, "Medium"
	default:
		return 0.1, "Low"
	}
}

func looksLikeDatacenterOrg(org string) bool {
	org = strings.ToLower(org)
	keywords := []string{
		"hosting", "cloud", "server", "data center", "datacenter", "vps",
		"amazon", "google", "microsoft", "digitalocean", "linode", "vultr",
		"hetzner", "ovh", "contabo", "alibaba", "tencent", "oracle",
	}
	for _, keyword := range keywords {
		if strings.Contains(org, keyword) {
			return true
		}
	}
	return false
}

// HTTPS 测试目标列表，随机选一个验证代理的 CONNECT 隧道能力
var httpsTestTargets = []string{
	"https://www.google.com",
	"https://www.openai.com",
	"https://www.github.com",
	"https://www.cloudflare.com",
	"https://httpbin.org/ip",
}

// checkHTTPSConnect 通过 HTTP 代理实际访问一个随机 HTTPS 网站，验证 CONNECT 隧道是否可用
// 首次失败会换一个目标重试一次，避免目标网站偶尔抽风导致误杀
func checkHTTPSConnect(proxyAddr string, timeout time.Duration) bool {
	proxyURL, err := proxyutil.HTTPURL(proxyAddr)
	if err != nil {
		return false
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy:               http.ProxyURL(proxyURL),
			TLSHandshakeTimeout: timeout,
		},
		Timeout: timeout,
	}

	// 随机起始索引
	start := int(time.Now().UnixNano() % int64(len(httpsTestTargets)))

	for attempt := 0; attempt < 2; attempt++ {
		idx := (start + attempt) % len(httpsTestTargets)
		resp, err := client.Get(httpsTestTargets[idx])
		if err != nil {
			continue
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		// 2xx 或 3xx 都算成功（部分网站会重定向）
		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			return true
		}
	}

	return false
}

// ValidateAll 并发验证所有代理，返回验证结果
func (v *Validator) ValidateAll(proxies []storage.Proxy) []Result {
	var results []Result
	for r := range v.ValidateStream(proxies) {
		results = append(results, r)
	}
	return results
}

// ValidateStream 并发验证，边验证边通过 channel 返回结果
func (v *Validator) ValidateStream(proxies []storage.Proxy) <-chan Result {
	ch := make(chan Result, concurrencyBuffer(len(proxies), v.concurrency))
	sem := make(chan struct{}, v.concurrency)
	var wg sync.WaitGroup

	go func() {
		for _, p := range proxies {
			wg.Add(1)
			sem <- struct{}{}
			go func(px storage.Proxy) {
				defer wg.Done()
				defer func() { <-sem }()
				result := v.validateOneWithReason(px)
				ch <- result
			}(p)
		}
		wg.Wait()
		close(ch)
	}()

	return ch
}

// ValidateOne 验证单个代理是否可用，返回是否有效、延迟、出口IP和地理位置
func (v *Validator) ValidateOne(p storage.Proxy) (bool, time.Duration, string, string) {
	result := v.ValidateOneWithQuality(p)
	return result.Valid, result.Latency, result.ExitIP, result.ExitLocation
}

// ValidateOneWithQuality 验证单个代理并返回完整 IP 画像
func (v *Validator) ValidateOneWithQuality(p storage.Proxy) Result {
	return v.validateOneWithReason(p)
}

func (v *Validator) validateOneWithReason(p storage.Proxy) Result {
	var client *http.Client
	var err error

	switch p.Protocol {
	case "http":
		client, err = newHTTPClient(p.Address, v.timeout)
	case "socks5":
		client, err = newSOCKS5Client(p.Address, v.timeout)
	default:
		log.Printf("unknown protocol %s for %s", p.Protocol, p.Address)
		return Result{Proxy: p, Valid: false, Reason: "unknown_protocol"}
	}

	if err != nil {
		return Result{Proxy: p, Valid: false, Reason: "client_init_failed"}
	}

	start := time.Now()
	resp, err := client.Get(v.validateURL)
	latency := time.Since(start)
	if err != nil {
		return Result{Proxy: p, Valid: false, Reason: "connect_failed"}
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	// 验证状态码（200 或 204 都接受）
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return Result{Proxy: p, Valid: false, Latency: latency, Reason: fmt.Sprintf("validate_status_%d", resp.StatusCode)}
	}

	// 响应时间过滤
	if v.maxResponseMs > 0 && latency > time.Duration(v.maxResponseMs)*time.Millisecond {
		return Result{Proxy: p, Valid: false, Latency: latency, Reason: "latency_exceeded"}
	}

	// 获取出口 IP、地理位置和 IP 画像（仅在验证通过时）
	exitInfo := getExitIPInfo(client)

	// 必须能获取到出口信息
	if exitInfo.IP == "" || exitInfo.Location == "" {
		return Result{
			Proxy:         p,
			Valid:         false,
			Latency:       latency,
			ExitIP:        exitInfo.IP,
			ExitLocation:  exitInfo.Location,
			IPType:        exitInfo.IPType,
			RiskScore:     exitInfo.RiskScore,
			RiskLevel:     exitInfo.RiskLevel,
			IsResidential: exitInfo.IsResidential,
			Reason:        "exit_info_failed",
		}
	}

	// 地理过滤：白名单优先，否则走黑名单
	if v.cfg != nil && len(exitInfo.Location) >= 2 {
		countryCode := exitInfo.Location[:2]
		if len(v.cfg.AllowedCountries) > 0 {
			// 白名单模式：不在白名单中则拒绝
			allowed := false
			for _, a := range v.cfg.AllowedCountries {
				if countryCode == a {
					allowed = true
					break
				}
			}
			if !allowed {
				return Result{Proxy: p, Valid: false, Latency: latency, ExitIP: exitInfo.IP, ExitLocation: exitInfo.Location, IPType: exitInfo.IPType, RiskScore: exitInfo.RiskScore, RiskLevel: exitInfo.RiskLevel, IsResidential: exitInfo.IsResidential, Reason: "geo_blocked"}
			}
		} else if len(v.cfg.BlockedCountries) > 0 {
			// 黑名单模式
			for _, blocked := range v.cfg.BlockedCountries {
				if countryCode == blocked {
					return Result{Proxy: p, Valid: false, Latency: latency, ExitIP: exitInfo.IP, ExitLocation: exitInfo.Location, IPType: exitInfo.IPType, RiskScore: exitInfo.RiskScore, RiskLevel: exitInfo.RiskLevel, IsResidential: exitInfo.IsResidential, Reason: "geo_blocked"}
				}
			}
		}
	}

	// HTTP 代理额外检测：必须支持 HTTPS CONNECT 隧道
	if p.Protocol == "http" {
		if !checkHTTPSConnect(p.Address, v.timeout) {
			return Result{Proxy: p, Valid: false, Latency: latency, ExitIP: exitInfo.IP, ExitLocation: exitInfo.Location, IPType: exitInfo.IPType, RiskScore: exitInfo.RiskScore, RiskLevel: exitInfo.RiskLevel, IsResidential: exitInfo.IsResidential, Reason: "https_connect_failed"}
		}
	}

	return Result{
		Proxy:         p,
		Valid:         true,
		Latency:       latency,
		ExitIP:        exitInfo.IP,
		ExitLocation:  exitInfo.Location,
		IPType:        exitInfo.IPType,
		RiskScore:     exitInfo.RiskScore,
		RiskLevel:     exitInfo.RiskLevel,
		IsResidential: exitInfo.IsResidential,
	}
}

func newHTTPClient(address string, timeout time.Duration) (*http.Client, error) {
	proxyURL, err := proxyutil.HTTPURL(address)
	if err != nil {
		return nil, err
	}
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
		Timeout: timeout,
	}, nil
}

func newSOCKS5Client(address string, timeout time.Duration) (*http.Client, error) {
	dialer, err := proxyutil.SOCKS5Dialer(address)
	if err != nil {
		return nil, err
	}
	return &http.Client{
		Transport: &http.Transport{
			Dial: dialer.Dial,
		},
		Timeout: timeout,
	}, nil
}
