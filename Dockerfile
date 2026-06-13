# 构建阶段：匹配项目 Go 1.25 版本要求
FROM golang:1.25-alpine AS builder

# 国内 Go 代理加速（解决网络问题）
ENV GOPROXY=https://goproxy.cn,direct
WORKDIR /app

# 依赖缓存层（加速重复构建）
COPY go.mod go.sum ./
RUN go mod download

# 编译项目（关闭CGO，静态编译，体积最小）
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o gocms ./cmd/server

# 运行阶段：极小镜像（仅5MB左右）
FROM alpine:3.20

# 安装基础依赖 + 时区
RUN apk add --no-cache ca-certificates tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

# 创建非root运行用户（安全）
RUN adduser -D -u 1000 gocms

WORKDIR /app

# 复制编译产物 + 所有必要目录
COPY --from=builder /app/gocms .
COPY --from=builder /app/config ./config
COPY --from=builder /app/web ./web

# 创建运行时目录并授权
RUN mkdir -p runtime/cache runtime/logs web/uploads && \
    chown -R gocms:gocms /app

USER gocms

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -q --spider http://localhost:8080/ || exit 1

# 启动命令
CMD ["./gocms"]
