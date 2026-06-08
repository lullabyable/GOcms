.PHONY: build run clean test tidy

# 构建
build:
	go build -o bin/gocms ./cmd/server

# 运行
run:
	go run ./cmd/server

# 清理
clean:
	rm -rf bin/ runtime/

# 测试
test:
	go test ./... -v

# 整理依赖
tidy:
	go mod tidy

# 交叉编译
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/gocms-linux ./cmd/server

# Docker 构建
docker:
	docker build -t gocms .
