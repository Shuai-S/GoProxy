# GoProxy 架构设计

## 系统概览

GoProxy 是一个单二进制的智能代理池系统，由多个协作的 goroutine 组成。核心设计理念：**免费池自动运转 + 订阅池按需导入，两池独立管理但统一对外服务**。

```
                         ┌─────────────────────────┐
                         │       WebUI (:7778)      │
                         │  管理面板 / 订阅管理 / 设置 │
                         └──────────┬──────────────┘
                                    │
  ┌──────────────────┐   ┌──────────┴──────────────┐   ┌──────────────────┐
  │  免费代理源 (20+) │──▶│       SQLite 代理池       │◀──│  订阅源 (Clash/   │
  │  公开列表自动抓取  │   │  proxies / subscriptions  │   │  V2ray/Base64)   │
  └──────────────────┘   └──────────┬──────────────┘   └──────────────────┘
                                    │                           │
                         ┌──────────┴──────────────┐   ┌───────┴────────┐
                         │      对外代理服务          │   │   sing-box     │
                         │ HTTP+SOCKS5 (:7776/:7777) │   │  协议转换进程   │
                         │ 固定双协议端口 (:7779+)    │   │  vmess/trojan  │
                         └───────────────────────────┘   │  → local socks5│
                                                         └────────────────┘
```

## 模块依赖

```
main.go (orchestrator)
  ├── config/       配置管理（环境变量 + config.json），sync.RWMutex 线程安全
  ├── storage/      SQLite 持久化（proxies + subscriptions + source_status）
  ├── fetcher/      免费代理抓取（20+ 源，断路器保护）
  ├── validator/    代理验证（连接 + 出口 IP + 地理位置 + 延迟 + HTTPS 隧道）
  ├── pool/         池子管理（slot 准入、替换逻辑、状态机）
  ├── checker/      健康检查（批次验证，free 删除 / custom 禁用）
  ├── optimizer/    质量优化（替换慢代理，仅作用于免费池）
  ├── custom/       订阅管理
  │   ├── parser.go     格式自动识别（Clash YAML / V2ray 链接 / Base64 / 纯文本）
  │   ├── singbox.go    sing-box 进程管理（配置生成 + 启停 + 端口映射）
  │   └── manager.go    刷新循环 + 探测唤醒 + 过期清理
  ├── proxy/        对外代理服务
  │   ├── multiplexed_server.go HTTP/SOCKS5 同端口协议分流
  │   ├── server.go             HTTP 代理（支持 CONNECT 隧道）
  │   ├── socks5_server.go      SOCKS5 代理（原生协议实现）
  │   └── fixedport.go          动态固定节点双协议端口
  ├── webui/        管理面板（嵌入式 HTML + REST API）
  └── logger/       内存日志收集（供 WebUI 展示）
```

## 数据模型

### proxies 表

| 字段 | 类型 | 说明 |
|------|------|------|
| address | TEXT UNIQUE | 代理地址 (host:port) |
| protocol | TEXT | http / socks5 |
| source | TEXT | free / custom |
| subscription_id | INTEGER | 所属订阅 ID（0=免费） |
| exit_ip | TEXT | 出口 IP |
| exit_location | TEXT | "国家代码 城市" |
| latency | INTEGER | 延迟 (ms) |
| quality_grade | TEXT | S(<=500ms) / A(<=1000) / B(<=2000) / C(>2000) |
| status | TEXT | active / degraded / disabled / candidate_replace |
| fail_count | INTEGER | 连续失败次数 |
| use_count / success_count | INTEGER | 使用统计 |

### subscriptions 表

| 字段 | 类型 | 说明 |
|------|------|------|
| url | TEXT | 订阅 URL |
| file_path | TEXT | 本地文件路径 |
| format | TEXT | auto（全自动识别） |
| status | TEXT | active / paused |
| last_fetch | DATETIME | 最后拉取时间 |
| last_success | DATETIME | 最后有可用节点的时间 |
| contributed | INTEGER | 是否为访客贡献 |

## 免费池状态机

```
                   总量>95%
  ┌─────────────────────────────────────┐
  │                                     ▼
┌──────┐  总量<95%  ┌─────────┐  协议<20%  ┌──────────┐  缺失协议或<10%  ┌───────────┐
│HEALTHY│──────────▶│ WARNING │───────────▶│ CRITICAL │─────────────────▶│ EMERGENCY │
└──────┘           └─────────┘            └──────────┘                  └───────────┘
```

各状态对应的行为：

| 状态 | 延迟阈值 | 抓取模式 | 说明 |
|------|---------|---------|------|
| healthy | 2000ms | 不抓取 | 优化器定期替换慢代理 |
| warning | 4000ms | refill (快源) | 补充到 95% |
| critical | 4000ms | refill (快源) | 优先补充缺失协议 |
| emergency | 4000ms | emergency (全源) | 忽略断路器，全力抓取 |

## 订阅池生命周期

```
添加订阅 → 验证(拉取+解析) → 入库
                                │
        ┌───────────────────────┘
        ▼
  刷新订阅（定时/手动）
        │
        ├── 拉取内容（直连 → 代理 fallback）
        ├── 自动识别格式
        ├── 删除该订阅旧代理
        ├── 分类节点
        │   ├── HTTP/SOCKS5 → 直接入池
        │   └── vmess/trojan/... → sing-box 转换 → 本地 SOCKS5 入池
        ├── 验证入池代理（含地理过滤）
        │   ├── 通过 → status=active, 更新 last_success
        │   └── 失败/被过滤 → status=disabled
        └── 更新订阅 proxy_count
```

### 探测唤醒

```
每 N 分钟（默认 10）
    │
    ├── 获取所有 source=custom AND status=disabled 的代理
    ├── 逐个验证
    │   ├── 通过 + 未被地理过滤 → EnableProxy → 更新 last_success
    │   └── 失败 或 被地理过滤 → 保持 disabled
    └── 日志输出恢复数量
```

### 自动清理

每分钟检查：创建超过 7 天且 `last_success` 距今超过 7 天的订阅 → 删除订阅 + 关联代理 + 重建 sing-box。

## 代理选择策略

5 种模式通过 `CustomProxyMode` + `CustomPriority` + `CustomFreePriority` 三个配置组合：

| UI 选项 | Mode | Priority | FreePriority | 选择逻辑 |
|---------|------|----------|-------------|---------|
| 混合·订阅优先 | mixed | true | false | 先 custom，无可用 fallback 全部 |
| 混合·免费优先 | mixed | false | true | 先 free，无可用 fallback 全部 |
| 混合·平等 | mixed | false | false | sourceFilter=""，不区分来源 |
| 仅订阅 | custom_only | - | - | sourceFilter="custom" |
| 仅免费 | free_only | - | - | sourceFilter="free" |

HTTP 代理服务可使用 HTTP 或 SOCKS5 上游代理。SOCKS5 代理服务仅使用 SOCKS5 上游代理。

## sing-box 集成

```
订阅节点 (vmess/vless/trojan/ss/hysteria2/anytls)
    │
    ▼
生成 sing-box JSON 配置：
  inbounds:  每个节点一个本地 SOCKS5 端口 (20001, 20002, ...)
  outbounds: 每个节点对应一个加密协议出站
  route:     inbound → outbound 一一映射
    │
    ▼
启动 sing-box 子进程 (sing-box run -c config.json)
    │
    ▼
本地 socks5://127.0.0.1:20001 ... 入池为 source=custom 代理
```

Docker 镜像自带 sing-box 二进制，支持 amd64/arm64。本地运行需手动安装。

## 后台 Goroutine

| Goroutine | 间隔 | 职责 |
|-----------|------|------|
| 状态监控 | 30s | 检查免费池状态，触发 smartFetchAndFill |
| 健康检查 | 5min | 批量验证代理，free 删除 / custom 禁用 |
| 优化轮换 | 30min | 替换免费池中的慢代理 |
| 订阅刷新 | 1min tick | 检查到期订阅，执行刷新 |
| 探测唤醒 | 10min | 探测 disabled 的订阅代理 |
| 配置监听 | event | WebUI 配置变更后调整 slot |

## 地理过滤

全局生效，对两个池子行为不同：

| 操作 | 免费代理 | 订阅代理 |
|------|---------|---------|
| 启动清理 | DELETE | status → disabled |
| 验证时不通过 | 不入池 | 入池但 disabled |
| 探测唤醒 | - | 被过滤的不启用 |
| 配置变更 | 下次清理生效 | 下次清理生效 |

白名单（`AllowedCountries`）优先于黑名单（`BlockedCountries`）。

## 端口映射

| 端口 | 服务 | 模式 |
|------|------|------|
| 7776 | HTTP + SOCKS5 代理 | 最低延迟 |
| 7777 | HTTP + SOCKS5 代理 | 随机轮换 |
| 7778 | WebUI | 管理面板 |
| 7779+ | HTTP + SOCKS5 固定端口 | 绑定指定节点 |
| 20001+ | sing-box 本地 | 仅 127.0.0.1 |
