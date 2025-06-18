# Dizi MCP Server Makefile

# 变量定义
APP_NAME = dizi
VERSION = $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v1.0.0")
BUILD_TIME = $(shell date +%Y-%m-%d_%H:%M:%S)
GO_VERSION = $(shell go version | cut -d' ' -f3)

# Go 构建标志
LDFLAGS = -ldflags="-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# 默认目标
.PHONY: all
all: clean build

# 构建
.PHONY: build
build:
	@echo "Building $(APP_NAME)..."
	go build $(LDFLAGS) -o $(APP_NAME)
	@echo "Build complete: $(APP_NAME)"

# 开发构建（包含调试信息）
.PHONY: build-dev
build-dev:
	@echo "Building $(APP_NAME) (development)..."
	go build -o $(APP_NAME)
	@echo "Development build complete: $(APP_NAME)"

# 跨平台构建
.PHONY: build-all
build-all: clean
	@echo "Building for multiple platforms..."
	@mkdir -p dist
	
	# Linux amd64
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(APP_NAME)-linux-amd64
	
	# Linux arm64
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(APP_NAME)-linux-arm64
	
	# macOS amd64
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(APP_NAME)-darwin-amd64
	
	# macOS arm64
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(APP_NAME)-darwin-arm64
	
	# Windows amd64
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(APP_NAME)-windows-amd64.exe
	
	@echo "Cross-platform build complete. Files in dist/"

# 测试
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# 代码格式化
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# 代码检查
.PHONY: vet
vet:
	@echo "Running go vet..."
	go vet ./...

# 依赖管理
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# 清理
.PHONY: clean
clean:
	@echo "Cleaning..."
	rm -f $(APP_NAME) $(APP_NAME).exe
	rm -rf dist/

# 安装到 GOPATH/bin
.PHONY: install
install:
	@echo "Installing $(APP_NAME)..."
	go install $(LDFLAGS)

# 运行
.PHONY: run
run: build
	./$(APP_NAME)

# 运行开发模式
.PHONY: run-dev
run-dev:
	go run . -port=8082

# 检查配置文件
.PHONY: check-config
check-config:
	@if [ ! -f dizi.yml ]; then \
		echo "Warning: dizi.yml not found. Creating from example..."; \
		cp dizi.example.yml dizi.yml; \
		echo "Created dizi.yml from example. Please customize it."; \
	else \
		echo "Configuration file dizi.yml found."; \
	fi

# 完整的开发工作流
.PHONY: dev
dev: deps fmt vet test build

# 发布准备
.PHONY: release
release: clean deps fmt vet test build-all
	@echo "Release build complete!"

# 帮助
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build      - Build the application"
	@echo "  build-dev  - Build with debug info"
	@echo "  build-all  - Cross-platform build"
	@echo "  test       - Run tests"
	@echo "  fmt        - Format code"
	@echo "  vet        - Run go vet"
	@echo "  deps       - Download dependencies"
	@echo "  clean      - Clean build artifacts"
	@echo "  install    - Install to GOPATH/bin"
	@echo "  run        - Build and run"
	@echo "  run-dev    - Run in development mode"
	@echo "  check-config - Check/create config file"
	@echo "  dev        - Full development workflow"
	@echo "  release    - Prepare release build"
	@echo "  help       - Show this help"