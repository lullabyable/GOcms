.PHONY: build run clean test tidy build-linux docker lint

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
	go test ./... -v -count=1

# 测试覆盖率
test-cover:
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -func=coverage.out | tail -1
	go tool cover -html=coverage.out -o coverage.html

# 整理依赖
tidy:
	go mod tidy

# 代码检查
lint:
	go vet ./...

# 交叉编译
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/gocms-linux ./cmd/server

# Docker 构建
docker:
	docker build -t gocms .

# Docker 运行
docker-run:
	docker run -p 8080:8080 -v $(PWD)/config:/app/config -v $(PWD)/runtime:/app/runtime gocms

# 开发模式（热重载需要 air）
dev:
	air -c .air.toml 2>/dev/null || go run ./cmd/server
