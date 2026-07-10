# 构建阶段（使用完整 Debian 镜像，内置 gcc，避免 alpine apk 问题）
FROM golang:1.25 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o proxy-pool .

# 下载 sing-box 二进制
ARG SINGBOX_VERSION=1.13.5
RUN ARCH=$(case "$(dpkg --print-architecture)" in amd64) echo "amd64";; arm64) echo "arm64";; *) echo "amd64";; esac) && \
    curl -fsSL "https://github.com/SagerNet/sing-box/releases/download/v${SINGBOX_VERSION}/sing-box-${SINGBOX_VERSION}-linux-${ARCH}.tar.gz" \
    -o /tmp/sing-box.tar.gz && \
    tar -xzf /tmp/sing-box.tar.gz -C /tmp && \
    cp /tmp/sing-box-${SINGBOX_VERSION}-linux-${ARCH}/sing-box /app/sing-box && \
    chmod +x /app/sing-box && \
    rm -rf /tmp/sing-box*

# 运行阶段（使用轻量 debian-slim）
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates tzdata curl && \
    rm -rf /var/lib/apt/lists/*

ENV TZ=Asia/Shanghai

WORKDIR /app
COPY --from=builder /app/proxy-pool .
COPY --from=builder /app/sing-box /usr/local/bin/sing-box

# 7776/7777 为双协议策略端口，7779/7780 为默认固定端口槽位
EXPOSE 7776 7777 7778 7779 7780

CMD ["./proxy-pool"]
