# GoProxy 测试脚本

本目录包含用于测试 GoProxy 代理服务的脚本。所有脚本都采用**持续运行模式**（类似 `ping` 命令），按 `Ctrl+C` 停止并显示统计。

## 📝 脚本列表

| 脚本 | 语言 | 依赖 | 运行模式 | 推荐度 |
|------|------|------|----------|--------|
| `test_proxy.sh` | Bash | curl + Python3 | 持续运行 | ⭐⭐⭐ |
| `test_socks5.sh` | Bash | curl + Python3 | 持续运行 | ⭐⭐⭐ |
| `test_proxy.go` | Go | `golang.org/x/net/proxy` | 持续运行 | ⭐⭐ |
| `test_proxy.py` | Python | `requests`, `pysocks` | 持续运行 | ⭐⭐ |

## 🚀 快速使用

### Bash 脚本（推荐）

**HTTP 代理测试**：
```bash
# 测试 7777 端口（随机轮换）
./test/test_proxy.sh 7777

# 测试 7776 端口（最低延迟）
./test/test_proxy.sh 7776

# 按 Ctrl+C 停止并查看统计
```

**SOCKS5 代理测试**：
```bash
# 测试 7777 端口（与 HTTP 共用，随机轮换）
./test/test_socks5.sh 7777

# 测试 7776 端口（与 HTTP 共用，最低延迟）
./test/test_socks5.sh 7776

# 按 Ctrl+C 停止并查看统计
```

### Go 脚本

```bash
# 安装依赖
go get golang.org/x/net/proxy

# 运行测试
go run test/test_proxy.go

# 或编译后运行
cd test
go build -o test_proxy test_proxy.go
./test_proxy
```

### Python 脚本

```bash
# 安装依赖
pip install requests pysocks

# 运行测试
python test/test_proxy.py
```

## 📊 测试内容

所有脚本都会：
1. 通过指定端口代理发送请求（默认 `127.0.0.1:7777`）
2. 访问 `http://ip-api.com/json` 获取出口 IP 和国家信息
3. **持续发送请求**，间隔 1 秒（类似 `ping` 命令）
4. 实时显示国旗 emoji、出口 IP 和延迟
5. 按 `Ctrl+C` 停止并显示统计摘要

## 📖 详细文档

完整的测试指南、故障排查、高级用法，请查看：

👉 [TEST_GUIDE.md](./TEST_GUIDE.md)

## 🔀 测试不同端口策略

### HTTP 代理端口对比

```bash
# 随机轮换模式 - IP 高度分散
./test/test_proxy.sh 7777

# 最低延迟模式 - 固定使用最快代理
./test/test_proxy.sh 7776
```

**观察要点**：
- **7777 端口**：每次请求的出口 IP 应该不同（证明在轮换）
- **7776 端口**：连续多次请求的出口 IP 基本相同（证明固定使用最优代理）

### SOCKS5 代理端口对比

```bash
# 随机轮换模式 - 与 HTTP 随机共用 7777
./test/test_socks5.sh 7777

# 最低延迟模式 - 与 HTTP 稳定共用 7776
./test/test_socks5.sh 7776
```

**观察要点**：
- **7777 端口**：每次连接的出口 IP 应该不同
- **7776 端口**：连续多次连接的出口 IP 基本相同
- 同一个端口也可通过 `test_proxy.sh` 测试 HTTP 入口协议

> 💡 **提示**：SOCKS5 测试脚本使用 `-k` 参数跳过 SSL 证书验证，因为免费上游代理常有证书问题。生产环境建议使用质量更好的付费代理。

## 🔍 预期输出

```
PROXY 127.0.0.1:7777 (http://ip-api.com/json/?fields=countryCode,query): continuous mode

proxy from 🇺🇸 203.0.113.45: seq=1 time=1234ms
proxy from 🇩🇪 198.51.100.78: seq=2 time=987ms
proxy from 🇬🇧 192.0.2.123: seq=3 time=1567ms
proxy #4: request failed (timeout)
proxy from 🇯🇵 198.51.100.12: seq=5 time=890ms
proxy from 🇫🇷 192.0.2.234: seq=6 time=1456ms
...
（持续运行，按 Ctrl+C 停止）

^C
---
50 requests transmitted, 47 received, 3 failed, 6.0% packet loss
```

**输出风格**：
- 简洁清晰，类似 `ping` 命令
- 一行一个结果
- 显示国旗 emoji、出口 IP、序号、延迟
- 统计信息简洁明了

**观察要点**：
- 每次请求的出口 IP 应该不同（证明代理轮换）
- 延迟应该在合理范围（< 2000ms）
- 丢包率应该 < 10%
- 可以长时间运行观察稳定性

## 📝 注意事项

1. 确保 GoProxy 服务已启动：`./goproxy`
2. 首次启动需等待代理池就绪（约 30-60 秒）
3. 可配合 WebUI (http://localhost:7778) 查看实时状态
