package pool

import (
	"log"

	"goproxy/config"
	"goproxy/storage"
)

// Manager 池子管理器
type Manager struct {
	storage *storage.Storage
	cfg     *config.Config
}

func NewManager(s *storage.Storage, cfg *config.Config) *Manager {
	return &Manager{
		storage: s,
		cfg:     cfg,
	}
}

// PoolStatus 池子状态
type PoolStatus struct {
	Total            int
	HTTP             int
	SOCKS5           int
	HTTPSlots        int
	SOCKS5Slots      int
	State            string // healthy/warning/critical/emergency
	AvgLatencyHTTP   int
	AvgLatencySocks5 int
	CustomCount      int // 订阅代理数量
}

// GetStatus 获取当前池子状态
func (m *Manager) GetStatus() (*PoolStatus, error) {
	total, _ := m.storage.Count()
	httpCount, _ := m.storage.CountByProtocol("http")
	socks5Count, _ := m.storage.CountByProtocol("socks5")

	httpSlots, socks5Slots := m.cfg.CalculateSlots()

	// 计算平均延迟
	avgHTTP, _ := m.storage.GetAverageLatency("http")
	avgSOCKS5, _ := m.storage.GetAverageLatency("socks5")

	// 判断状态
	state := m.determineState(total, httpCount, socks5Count)

	customCount, _ := m.storage.CountBySource("custom")

	return &PoolStatus{
		Total:            total,
		HTTP:             httpCount,
		SOCKS5:           socks5Count,
		HTTPSlots:        httpSlots,
		SOCKS5Slots:      socks5Slots,
		State:            state,
		AvgLatencyHTTP:   avgHTTP,
		AvgLatencySocks5: avgSOCKS5,
		CustomCount:      customCount,
	}, nil
}

// determineState 判断池子状态
func (m *Manager) determineState(total, httpCount, socks5Count int) string {
	httpSlots, socks5Slots := m.cfg.CalculateSlots()

	// 单协议缺失
	if httpCount == 0 || socks5Count == 0 {
		return "emergency"
	}

	// 紧急：总数<10%
	emergencyThreshold := int(float64(m.cfg.PoolMaxSize) * 0.1)
	if total < emergencyThreshold {
		return "emergency"
	}

	// 危急：任一协议<20%槽位
	if httpCount < int(float64(httpSlots)*0.2) || socks5Count < int(float64(socks5Slots)*0.2) {
		return "critical"
	}

	// 警告：总数<95%
	healthyThreshold := int(float64(m.cfg.PoolMaxSize) * 0.95)
	if total < healthyThreshold {
		return "warning"
	}

	// 健康
	return "healthy"
}

// NeedsFetch 判断是否需要抓取以及抓取模式
func (m *Manager) NeedsFetch(status *PoolStatus) (bool, string, string) {
	// 单协议缺失：紧急模式，指定协议
	if status.HTTP == 0 {
		return true, "emergency", "http"
	}
	if status.SOCKS5 == 0 {
		return true, "emergency", "socks5"
	}

	// 紧急状态：紧急模式
	if status.State == "emergency" {
		return true, "emergency", ""
	}

	// 危急或警告：补充模式
	if status.State == "critical" || status.State == "warning" {
		// 判断哪个协议更缺
		httpPct := float64(status.HTTP) / float64(status.HTTPSlots)
		socks5Pct := float64(status.SOCKS5) / float64(status.SOCKS5Slots)

		// 如果两个协议都缺（都<50%），同时补充两个协议
		if httpPct < 0.5 && socks5Pct < 0.5 {
			return true, "refill", ""
		}
		// 只有一个协议缺时，优先补充更缺的
		if httpPct < 0.5 {
			return true, "refill", "http"
		}
		if socks5Pct < 0.5 {
			return true, "refill", "socks5"
		}
		return true, "refill", ""
	}

	// 健康状态：不需要补充抓取
	return false, "", ""
}

// NeedsFetchQuick 快速判断是否还需要抓取（用于提前终止）
func (m *Manager) NeedsFetchQuick(status *PoolStatus) bool {
	need, _, _ := m.NeedsFetch(status)
	return need
}

// TryAddProxy 尝试将代理加入池子
func (m *Manager) TryAddProxy(p storage.Proxy) (bool, string) {
	// 订阅代理直接入池，不受 slot 限制
	if p.Source == "custom" {
		if err := m.storage.AddProxyWithSource(p.Address, p.Protocol, "custom"); err != nil {
			return false, "db_error"
		}
		m.storage.UpdateExitInfoWithQuality(p.Address, p.ExitIP, p.ExitLocation, p.IPType, p.RiskScore, p.RiskLevel, p.IsResidential, p.Latency)
		log.Printf("[pool] ✅ 订阅代理入池: %s (%s) %dms %s", p.Address, p.Protocol, p.Latency, p.ExitLocation)
		return true, "added_custom"
	}

	httpSlots, socks5Slots := m.cfg.CalculateSlots()
	httpCount, _ := m.storage.CountByProtocol("http")
	socks5Count, _ := m.storage.CountByProtocol("socks5")
	total, _ := m.storage.Count()

	var maxSlots int
	var currentCount int
	if p.Protocol == "http" {
		maxSlots = httpSlots
		currentCount = httpCount
	} else {
		maxSlots = socks5Slots
		currentCount = socks5Count
	}

	// 情况1：该协议槽位未满，直接加入
	if currentCount < maxSlots {
		if err := m.storage.AddProxy(p.Address, p.Protocol); err != nil {
			return false, "db_error"
		}
		// 更新完整信息
		m.storage.UpdateExitInfoWithQuality(p.Address, p.ExitIP, p.ExitLocation, p.IPType, p.RiskScore, p.RiskLevel, p.IsResidential, p.Latency)
		log.Printf("[pool] ✅ 直接入池: %s (%s %d/%d) %dms %s %s",
			p.Address, p.Protocol, currentCount+1, maxSlots, p.Latency, p.ExitIP, p.ExitLocation)
		return true, "added"
	}

	// 情况2：槽位满，但允许10%浮动
	allowedFloat := int(float64(maxSlots) * 0.1)
	if total < m.cfg.PoolMaxSize && currentCount < maxSlots+allowedFloat {
		if err := m.storage.AddProxy(p.Address, p.Protocol); err != nil {
			return false, "db_error"
		}
		m.storage.UpdateExitInfoWithQuality(p.Address, p.ExitIP, p.ExitLocation, p.IPType, p.RiskScore, p.RiskLevel, p.IsResidential, p.Latency)
		log.Printf("[pool] ✅ 浮动入池: %s (%s %d/%d+%d) %dms",
			p.Address, p.Protocol, currentCount+1, maxSlots, allowedFloat, p.Latency)
		return true, "added_float"
	}

	// 情况3：池子满了，尝试替换
	if currentCount >= maxSlots || total >= m.cfg.PoolMaxSize {
		return m.tryReplace(p)
	}

	return false, "slots_full"
}

// tryReplace 尝试替换现有代理
func (m *Manager) tryReplace(newProxy storage.Proxy) (bool, string) {
	// 获取同协议中可替换的代理（延迟最高的前10个）
	candidates, err := m.storage.GetWorstProxies(newProxy.Protocol, 10)
	if err != nil || len(candidates) == 0 {
		return false, "no_candidates"
	}

	worst := candidates[0]

	// 判断是否值得替换：新代理需要显著更快
	threshold := m.cfg.ReplaceThreshold
	if float64(newProxy.Latency) < float64(worst.Latency)*threshold {
		if err := m.storage.ReplaceProxy(worst.Address, newProxy); err != nil {
			return false, "replace_error"
		}
		log.Printf("[pool] 🔄 替换: %s(%dms) → %s(%dms) 提升%.0f%%",
			worst.Address, worst.Latency, newProxy.Address, newProxy.Latency,
			(1-float64(newProxy.Latency)/float64(worst.Latency))*100)
		return true, "replaced"
	}

	return false, "not_better"
}

// AdjustForConfigChange 配置变更后调整池子
func (m *Manager) AdjustForConfigChange(oldSize int, oldRatio float64) {
	newHTTP, newSOCKS5 := m.cfg.CalculateSlots()
	oldHTTP := int(float64(oldSize) * oldRatio)
	oldSOCKS5 := oldSize - oldHTTP

	log.Printf("[pool] 配置变更: 容量 %d→%d, HTTP槽位 %d→%d, SOCKS5槽位 %d→%d",
		oldSize, m.cfg.PoolMaxSize, oldHTTP, newHTTP, oldSOCKS5, newSOCKS5)

	// 如果槽位减少且当前超标，标记超标的为替换候选
	httpCount, _ := m.storage.CountByProtocol("http")
	socks5Count, _ := m.storage.CountByProtocol("socks5")

	if httpCount > newHTTP {
		excess := httpCount - newHTTP
		log.Printf("[pool] HTTP 超标 %d 个，标记为替换候选", excess)
		// 这里可以标记延迟高的为候选
	}

	if socks5Count > newSOCKS5 {
		excess := socks5Count - newSOCKS5
		log.Printf("[pool] SOCKS5 超标 %d 个，标记为替换候选", excess)
	}
}
