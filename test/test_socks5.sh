#!/bin/bash
###
 # @LastEditTime: 2026-03-29 23:26:29
 # @Description: ...
 # @Date: 2026-03-29 23:14:32
 # @Author: isboyjc
 # @LastEditors: isboyjc
### 

# GoProxy SOCKS5 代理测试脚本
# 用法: ./test_socks5.sh [端口号，默认7777]

PROXY_HOST="${PROXY_HOST:-127.0.0.1}"
PROXY_PORT="${1:-7777}"
TEST_URL="https://httpbin.org/ip"
DELAY=1

# 统计变量
total=0
success=0
fail=0

# 获取毫秒时间戳（兼容 macOS 和 Linux）
get_ms_time() {
    python3 -c 'import time; print(int(time.time() * 1000))'
}

# 国家代码转 emoji 旗帜
country_to_emoji() {
    local country_code="$1"
    if [ -z "$country_code" ] || [ "$country_code" = "null" ]; then
        echo "🌐"
        return
    fi
    
    local first="${country_code:0:1}"
    local second="${country_code:1:1}"
    python3 -c "print(chr(127462 + ord('$first') - ord('A')) + chr(127462 + ord('$second') - ord('A')))"
}

# 捕获 Ctrl+C 信号
trap ctrl_c INT
function ctrl_c() {
    echo ""
    echo "---"
    loss_rate=$(awk "BEGIN {printf \"%.1f\", ($total - $success)/$total*100}")
    echo "$total requests transmitted, $success received, $((total - success)) failed, ${loss_rate}% packet loss"
    exit 0
}

echo "SOCKS5 PROXY ${PROXY_HOST}:${PROXY_PORT}: continuous mode"
echo ""

while true; do
    total=$((total + 1))
    
    # 使用 curl 的 SOCKS5 支持（-k 跳过 SSL 验证，因为免费代理证书常有问题）
    start=$(get_ms_time)
    response=$(curl -s -k --socks5-hostname ${PROXY_HOST}:${PROXY_PORT} ${TEST_URL} --max-time 10 2>&1)
    end=$(get_ms_time)
    latency=$((end - start))
    
    if echo "$response" | grep -q '"origin"'; then
        success=$((success + 1))
        origin=$(echo "$response" | grep -o '"origin":"[^"]*"' | cut -d'"' -f4 | cut -d',' -f1)
        country=$(curl -s "http://ip-api.com/json/${origin}?fields=countryCode" 2>/dev/null | grep -o '"countryCode":"[^"]*"' | cut -d'"' -f4)
        emoji=$(country_to_emoji "$country")
        echo "socks5 #${total}: ${origin} ${emoji} ${country} (${latency}ms)"
    else
        fail=$((fail + 1))
        echo "socks5 #${total}: request failed"
    fi
    
    sleep $DELAY
done
