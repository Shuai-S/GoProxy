# 更新日志

所有重要的项目变更都会记录在此文件中。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
版本号遵循 [语义化版本 2.0.0](https://semver.org/lang/zh-CN/2.0.0/)。

## [Unreleased]

### 新增

- `7776`（最低延迟）和 `7777`（随机轮换）现在同时接受 HTTP 和 SOCKS5 客户端连接
- 固定端口统一支持 HTTP + SOCKS5 协议复用，并提供可搜索节点选择与订阅来源展示

### 变更

- 原 SOCKS5 随机端口 `7779` 迁移到 `7777`，原 SOCKS5 稳定端口 `7780` 迁移到 `7776`
- `7779`、`7780` 释放为默认固定节点端口槽位，只有配置绑定后才监听
- 旧固定端口配置中的 `protocol` 字段继续兼容读取，但不再限制客户端入口协议

## [v0.4.1] - 2026-04-04

### 修复

- 升级 sing-box 从 1.11.8 到 **1.13.5**，修复 anytls 等新协议不支持导致订阅节点启动失败的问题
- sing-box 启动前新增 `sing-box check` 配置预检，配置无效时输出详细错误而非静默崩溃
- 捕获 sing-box stderr 输出到 `[sing-box]` 日志，便于排查运行时错误
- 检测 sing-box 进程启动后立即退出的情况，避免误报"端口未就绪"
- Docker healthcheck 从 `wget` 改为 `curl`（debian-slim 无 wget），Dockerfile 增加 curl 安装
- 修复 `docker-compose.dokploy.yml` 服务未加入 `dokploy-network` 的问题
- 修复中英文切换时订阅池统计模块动态文字未更新的问题

## [v0.4.0] - 2026-04-04

### 新增

- **订阅代理导入**
  - 支持通过 WebUI 添加 Clash/V2ray 订阅 URL 或上传配置文件
  - 格式全自动识别：Clash YAML、V2ray 链接（vmess/vless/trojan/ss/hysteria2/anytls）、Base64 编码、纯文本
  - 内置 sing-box 协议转换：加密协议节点自动转为本地 SOCKS5 代理，Docker 镜像自带 sing-box 二进制
  - 订阅定时刷新：可配置刷新间隔，自动拉取最新节点并替换旧节点
  - 添加订阅时先验证（拉取+解析通过后才入库），失败不产生垃圾数据

- **订阅代理保护机制**
  - 软删除：订阅代理健康检查失败不删除只禁用（`status='disabled'`）
  - 探测唤醒：定时探测禁用的订阅代理，恢复可用后自动启用
  - 地理过滤全局化：免费代理删除、订阅代理禁用，探测唤醒时也检查地理规则
  - 自动清理：连续 7 天无可用节点的订阅自动移除

- **5 种代理使用模式**
  - 混合·订阅优先：优先使用订阅代理，无可用时降级到免费
  - 混合·免费优先：优先使用免费代理，无可用时降级到订阅
  - 混合·平等：不区分来源，按延迟/随机选择
  - 仅订阅代理：只使用订阅导入的代理
  - 仅免费代理：只使用公开抓取的代理

- **访客贡献订阅**
  - 未登录用户可通过「贡献订阅」入口提交订阅 URL 或上传配置文件
  - 提交前自动验证，通过后才入库
  - 管理员可刷新、暂停、删除贡献的订阅
  - 贡献订阅在列表中有橙色「贡献」标记

- **WebUI 增强**
  - 免费池 / 订阅池分离展示，各自独立统计
  - 订阅管理面板：订阅列表（名称 + 可用数 + 禁用数）、添加/刷新/暂停/删除
  - 代理列表中订阅代理带黄色标签显示所属订阅名称 + 左侧黄色竖线
  - 系统设置从侧边栏移至顶部齿轮图标，重组为：代理模式 → 免费池 → 订阅池 → 验证检查 → 地理过滤
  - 新增 ~70 个 i18n 翻译 key，覆盖所有新增 UI 元素

- **代理使用统计**
  - HTTP/SOCKS5 代理服务在请求成功/失败时记录使用次数（`RecordProxyUse`）

### 变更

- `Proxy` 结构体新增 `Source`（free/custom）和 `SubscriptionID` 字段
- `Count()`/`CountByProtocol()` 仅统计免费代理（slot 计算不受订阅代理影响）
- 批量删除方法（`DeleteInvalid`/`DeleteBlockedCountries`/`DeleteNotAllowedCountries`/`DeleteWithoutExitInfo`）仅作用于免费代理
- `GetWorstProxies` 排除订阅代理，优化器不替换订阅代理
- Dockerfile 集成 sing-box 二进制（自动检测 amd64/arm64 架构）

### 修复

- 修复 `AddProxy` 未显式设置 `source='free'` 的问题
- 修复 WebUI「刷新代理」「刷新延迟」对订阅代理执行硬删除的问题（改为禁用）
- 修复 `validateCustomProxies` 将所有代理硬编码为 socks5 协议导致 HTTP 直连代理验证失败
- 修复 `CustomPriority` 和 `CustomFreePriority` 可同时为 true 的互斥问题

## [v0.3.0] - 2026-04-01

### 新增

- **地理过滤增强**
  - 支持国家白名单（`ALLOWED_COUNTRIES`）和黑名单（`BLOCKED_COUNTRIES`）配置
  - 白名单优先级高于黑名单：白名单非空时仅允许指定国家，否则使用黑名单屏蔽
  - 支持通过环境变量、配置文件、WebUI 动态配置地理过滤规则
  - 启动时自动清理违反当前过滤规则的已入池代理
  - 详细文档：`GEO_FILTER.md`

- **项目指南文档**
  - 新增 `CLAUDE.md`，提供项目架构、设计模式、代码规范的完整指导
  - 包含模块依赖流程图、后台协程说明、端口映射表等

- **HTTPS 可用性验证增强**
  - HTTP 协议代理入池前增加 HTTPS CONNECT 隧道验证
  - 随机访问真实 HTTPS 网站（Google/GitHub/OpenAI 等）确认可用性
  - 失败自动切换验证站点重试，确保入池的 HTTP 代理都能访问 HTTPS
  - 新增测试脚本：`test/test_http_https.sh` 用于持续测试 HTTPS 访问能力

### 变更

- 默认 HTTP 协议占比从 50% 调整为 30%（配置 `PoolHTTPRatio: 0.3`）
- 地理过滤配置优先级：`config.json` > 环境变量
- WebUI 地理过滤设置界面支持动态修改白名单/黑名单

### 修复

- 修复地理过滤在验证器和存储层的逻辑一致性问题
- 修复启动时地理过滤清理逻辑，正确处理白名单优先场景
- 修复代理池补充逻辑：当 HTTP 和 SOCKS5 协议都缺失时，同时补充两个协议，而非先后补充
- 修复槽位计算问题：调整默认配置比例为 0.3（3:7），符合 HTTP/SOCKS5 实际使用场景

## [v0.2.0] - 2026-03-30

### 新增

- **SOCKS5 协议支持**
  - 实现完整的 SOCKS5 代理服务器（支持 CONNECT 命令）
  - 提供两个 SOCKS5 端口：`:7779`（随机轮换）+ `:7780`（最低延迟）
  - SOCKS5 服务仅使用 SOCKS5 上游代理，避免 HTTP 代理不支持 CONNECT 的问题
  - 协议并发验证：SOCKS5 和 HTTP 分组并发验证，SOCKS5 无额外检测，优先填充
  - 新增测试脚本：`test/test_socks5.sh` 用于测试 SOCKS5 代理

- **配置增强**
  - 新增 `SOCKS5Port` 和 `StableSOCKS5Port` 配置项
  - 支持通过环境变量配置 SOCKS5 端口
  - 优化代理池槽位分配逻辑，支持 HTTP/SOCKS5 比例配置

### 变更

- 存储层新增协议筛选方法 `CountByProtocol`、`GetRandomByProtocol`、`GetLowestLatencyByProtocol`
- 代理池管理器适配双协议槽位计算
- Docker Compose 配置新增 SOCKS5 端口映射

## [v0.1.0] - 2026-03-29

### 新增

- **代理认证功能**
  - HTTP 和 SOCKS5 代理服务支持可选的用户名密码认证
  - 环境变量配置：`PROXY_AUTH_ENABLED`、`PROXY_AUTH_USERNAME`、`PROXY_AUTH_PASSWORD`
  - 默认关闭，开启后可保护代理服务不被未授权访问

- **环境变量支持**
  - `WEBUI_PASSWORD`：自定义 WebUI 管理密码（默认 `goproxy`）
  - `DATA_DIR`：自定义数据目录路径（默认当前目录）
  - `BLOCKED_COUNTRIES`：屏蔽特定国家的代理（如 `CN,RU,KP`）

- **数据目录集中管理**
  - 支持通过 `DATA_DIR` 环境变量指定数据存储位置
  - 配置文件 `config.json` 和数据库 `proxy.db` 统一存放在数据目录

- **智能抓取机制**
  - 智能状态监控：Healthy / Warning / Critical / Emergency 四级状态
  - 按需抓取：根据池子状态自动选择合适的抓取模式
  - 源断路器：连续失败的代理源自动降级或禁用，冷却后恢复

- **WebUI 增强**
  - 实时日志流显示：支持查看最近 1000 条系统日志
  - 代理质量分布图表：S/A/B/C 各等级代理数量可视化
  - 延迟趋势图：HTTP 和 SOCKS5 平均延迟变化趋势

### 变更

- 验证超时从 8 秒增加到 10 秒，适应较慢的代理网络
- 健康检查批次大小从 10 个增加到 20 个，提高检查效率
- 优化配置参数命名，统一使用 `MaxLatency` 前缀

### 文档

- 完善 README.md，新增快速导航、Docker 部署、测试指南等章节
- 新增 `.env.example` 示例环境变量文件
- 更新 Docker Compose 配置示例
- 新增 GitHub Container Registry 镜像源说明

## [v0.0.1] - 2026-03-27

### 新增

- 项目初始化
- 基础 HTTP 代理池功能
- WebUI 管理界面
- SQLite 数据持久化
- 代理验证和健康检查
- Docker 支持

---

## 版本说明

- **主版本号**：不兼容的 API 变更
- **次版本号**：向下兼容的功能新增
- **修订号**：向下兼容的问题修复

## 相关链接

- [项目仓库](https://github.com/isboyjc/GoProxy)
- [Docker Hub](https://hub.docker.com/r/isboyjc/goproxy)
- [GitHub Container Registry](https://github.com/isboyjc/GoProxy/pkgs/container/goproxy)
- [问题反馈](https://github.com/isboyjc/GoProxy/issues)
