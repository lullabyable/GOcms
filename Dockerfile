FROM golang:1.22-alpine AS builder

WORKDIR /app

# 依赖缓存层
COPY go.mod go.sum ./
RUN go mod download

# 构建
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o gocms ./cmd/server

# --- 运行阶段 ---
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

# 创建非root用户
RUN adduser -D -u 1000 gocms

WORKDIR /app

# 复制二进制和配置
COPY --from=builder /app/gocms .
COPY --from=builder /app/config ./config

# 创建运行时目录
RUN mkdir -p runtime/cache runtime/logs web/uploads && \
    chown -R gocms:gocms /app

USER gocms

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -q --spider http://localhost:8080/ || exit 1

CMD ["./gocms"]
