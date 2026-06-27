package optimizer

import (
	"log"
	"time"

	"goproxy/config"
	"goproxy/fetcher"
	"goproxy/pool"
	"goproxy/storage"
	"goproxy/validator"
)

// Optimizer 优化轮换器
type Optimizer struct {
	storage   *storage.Storage
	fetcher   *fetcher.Fetcher
	validator *validator.Validator
	poolMgr   *pool.Manager
	cfg       *config.Config
}

func NewOptimizer(s *storage.Storage, f *fetcher.Fetcher, v *validator.Validator, pm *pool.Manager, cfg *config.Config) *Optimizer {
	return &Optimizer{
		storage:   s,
		fetcher:   f,
		validator: v,
		poolMgr:   pm,
		cfg:       cfg,
	}
}

// RunOnce 执行一次优化轮换
func (o *Optimizer) RunOnce() {
	start := time.Now()
	log.Println("[optimize] 🎯 开始优化轮换...")

	// 获取池子状态
	status, err := o.poolMgr.GetStatus()
	if err != nil {
		log.Printf("[optimize] 获取状态失败: %v", err)
		return
	}

	// 只有健康状态才执行优化
	if status.State != "healthy" {
		log.Printf("[optimize] 池子状态 %s，跳过优化", status.State)
		return
	}

	// 抓取新的候选代理（优化模式）
	log.Println("[optimize] 抓取新候选代理...")
	candidates, err := o.fetcher.FetchSmart("optimize", "")
	if err != nil {
		log.Printf("[optimize] 抓取失败: %v", err)
		return
	}

	log.Printf("[optimize] 抓取到 %d 个候选代理", len(candidates))

	// 验证候选代理
	validCandidates := []storage.Proxy{}
	for result := range o.validator.ValidateStream(candidates) {
		if result.Valid {
			latencyMs := int(result.Latency.Milliseconds())
			// 只保留延迟在健康标准内的
			if latencyMs <= o.cfg.MaxLatencyHealthy {
				validCandidates = append(validCandidates, storage.Proxy{
					Address:       result.Proxy.Address,
					Protocol:      result.Proxy.Protocol,
					ExitIP:        result.ExitIP,
					ExitLocation:  result.ExitLocation,
					IPType:        result.IPType,
					RiskScore:     result.RiskScore,
					RiskLevel:     result.RiskLevel,
					IsResidential: result.IsResidential,
					Latency:       latencyMs,
				})
			}
		}
	}

	log.Printf("[optimize] 验证通过 %d 个优质候选（延迟<%dms）", len(validCandidates), o.cfg.MaxLatencyHealthy)

	if len(validCandidates) == 0 {
		log.Println("[optimize] 无优质候选，跳过优化")
		return
	}

	// 尝试用优质候选替换延迟高的代理
	replacedCount := 0
	for _, candidate := range validCandidates {
		added, reason := o.poolMgr.TryAddProxy(candidate)
		if added && reason == "replaced" {
			replacedCount++
		}
	}

	elapsed := time.Since(start)
	log.Printf("[optimize] ✅ 完成: 替换 %d 个代理 耗时%v", replacedCount, elapsed)
}

// StartBackground 后台定时优化
func (o *Optimizer) StartBackground() {
	ticker := time.NewTicker(time.Duration(o.cfg.OptimizeInterval) * time.Minute)
	go func() {
		for range ticker.C {
			o.RunOnce()
		}
	}()
	log.Printf("[optimize] 优化轮换器已启动，间隔 %d 分钟", o.cfg.OptimizeInterval)
}
