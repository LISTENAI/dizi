# 开发指南

本文档为 Dizi 项目的开发者提供详细的开发环境搭建、代码结构说明和贡献指南。

## 开发环境搭建

### 系统要求

- **Go 1.21+** - 编程语言运行时
- **Git** - 版本控制
- **Make** (可选) - 构建工具

### 克隆项目

```bash
git clone https://github.com/your-org/dizi.git
cd dizi
```

### 安装依赖

```bash
go mod download
go mod verify
```

### 验证安装

```bash
# 运行测试
go test ./...

# 构建项目
go build ./cmd/dizi

# 测试二进制文件
./dizi init
./dizi -help
```

## 项目结构

```
dizi/
├── README.md              # 用户文档
├── go.mod                 # Go 模块定义
├── go.sum                 # 依赖版本锁定
├── Makefile               # 构建脚本
├── dizi.example.yml       # 示例配置文件
├── .github/               # GitHub Actions
│   └── workflows/
│       ├── ci.yml         # 持续集成
│       └── release.yml    # 发布流程
├── cmd/                   # 命令行程序
│   └── dizi/
│       ├── main.go        # 主程序入口
│       └── dizi.example.yml # 嵌入的配置模板
├── internal/              # 内部包（不对外暴露）
│   ├── config/            # 配置管理
│   │   ├── config.go
│   │   └── config_test.go
│   ├── logger/            # 日志管理
│   │   ├── logger.go
│   │   └── logger_test.go
│   ├── server/            # 服务器实现
│   │   ├── sse.go
│   │   └── sse_test.go
│   └── tools/             # 工具实现
│       ├── tools.go
│       ├── tools_test.go
│       ├── filesystem.go
│       └── filesystem_test.go
└── docs/                  # 技术文档
    ├── README.md          # 文档索引
    ├── architecture.md    # 架构设计
    ├── builtin-tools.md   # 内置工具
    ├── development.md     # 开发指南
    └── ...
```

### 目录说明

#### `cmd/dizi/`
命令行程序的主入口，包含：
- `main.go`: 主程序逻辑，参数解析，服务器启动
- `dizi.example.yml`: 嵌入的配置模板

#### `internal/`
内部实现包，不对外暴露：

**`config/`** - 配置管理
- 配置文件加载和解析
- 默认配置提供
- 配置验证

**`logger/`** - 日志管理
- 智能日志输出
- 传输方式适配
- 线程安全日志

**`server/`** - 服务器实现
- SSE 自定义处理
- 查询参数解析
- 动态工具启用

**`tools/`** - 工具系统
- 工具注册和管理
- 多种工具类型支持
- 文件系统工具实现

## 开发工作流

### 1. 创建功能分支

```bash
git checkout -b feature/new-feature
```

### 2. 开发和测试

```bash
# 运行测试
go test ./...

# 运行特定包的测试
go test ./internal/tools/

# 运行测试并查看覆盖率
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 3. 代码质量检查

```bash
# 格式化代码
go fmt ./...

# 检查代码
go vet ./...

# 运行 lint（需要安装 golangci-lint）
golangci-lint run
```

### 4. 构建和测试

```bash
# 开发构建
go build -o dizi ./cmd/dizi

# 生产构建
go build -ldflags="-s -w" -o dizi ./cmd/dizi

# 跨平台构建
GOOS=linux GOARCH=amd64 go build -o dizi-linux-amd64 ./cmd/dizi
GOOS=darwin GOARCH=arm64 go build -o dizi-darwin-arm64 ./cmd/dizi
```

### 5. 提交和推送

```bash
git add .
git commit -m "feat: add new feature"
git push origin feature/new-feature
```

## 添加新功能

### 1. 添加新的内置工具

**步骤 1：** 实现工具处理函数

在 `internal/tools/tools.go` 中添加：

```go
func handleMyTool(request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    // 参数验证
    args := request.Params.Arguments
    if args == nil {
        return mcp.NewToolResultError("Missing arguments"), nil
    }
    
    // 获取参数
    param, ok := args["param"].(string)
    if !ok {
        return mcp.NewToolResultError("Invalid param"), nil
    }
    
    // 业务逻辑
    result := processParam(param)
    
    // 返回结果
    return mcp.NewToolResultText(result), nil
}
```

**步骤 2：** 注册工具

在 `createBuiltinHandler` 函数中添加：

```go
case "my_tool":
    return handleMyTool(request)
```

**步骤 3：** 添加测试

在 `internal/tools/tools_test.go` 中添加：

```go
func TestHandleMyTool(t *testing.T) {
    tests := []struct {
        name        string
        arguments   map[string]interface{}
        expectError bool
        expectedMsg string
    }{
        // 测试用例
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            request := mcp.CallToolRequest{
                Params: mcp.CallToolParams{
                    Arguments: tt.arguments,
                },
            }
            
            result, err := handleMyTool(request)
            // 验证逻辑
        })
    }
}
```

### 2. 添加新的配置选项

**步骤 1：** 更新配置结构

在 `internal/config/config.go` 中：

```go
type Config struct {
    // 现有字段...
    NewOption string `yaml:"new_option"`
}

func getDefaultConfig() *Config {
    return &Config{
        // 现有默认值...
        NewOption: "default_value",
    }
}
```

**步骤 2：** 更新配置测试

在 `internal/config/config_test.go` 中添加相应测试。

**步骤 3：** 更新示例配置

在 `dizi.example.yml` 中添加新选项的说明。

### 3. 添加新的传输方式

**步骤 1：** 在 `internal/server/` 中创建新文件

**步骤 2：** 实现传输接口

**步骤 3：** 在 `main.go` 中添加支持

## 测试指南

### 单元测试

每个包都应该有对应的测试文件：

```go
func TestFunctionName(t *testing.T) {
    // 表格驱动测试
    tests := []struct {
        name     string
        input    InputType
        expected OutputType
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    validInput,
            expected: expectedOutput,
            wantErr:  false,
        },
        {
            name:    "invalid input",
            input:   invalidInput,
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := FunctionName(tt.input)
            
            if tt.wantErr {
                if err == nil {
                    t.Error("Expected error, got nil")
                }
                return
            }
            
            if err != nil {
                t.Errorf("Unexpected error: %v", err)
            }
            
            if !reflect.DeepEqual(result, tt.expected) {
                t.Errorf("Expected %v, got %v", tt.expected, result)
            }
        })
    }
}
```

### 集成测试

创建端到端测试：

```go
func TestServerIntegration(t *testing.T) {
    // 创建临时配置
    tempDir := t.TempDir()
    configFile := filepath.Join(tempDir, "dizi.yml")
    
    config := &config.Config{
        Name:    "test-server",
        Version: "1.0.0",
        Tools: []config.ToolConfig{
            {
                Name: "echo",
                Type: "builtin",
            },
        },
    }
    
    // 启动服务器
    // 发送请求
    // 验证响应
}
```

### 测试覆盖率

目标覆盖率：
- **整体覆盖率**: > 80%
- **关键组件**: > 90%
- **新增代码**: 100%

检查覆盖率：

```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

## 代码规范

### Go 代码风格

遵循官方 Go 代码风格：

```go
// 好的例子
func ProcessUserRequest(ctx context.Context, userID string) (*User, error) {
    if userID == "" {
        return nil, fmt.Errorf("userID cannot be empty")
    }
    
    user, err := getUserFromDB(ctx, userID)
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    
    return user, nil
}

// 避免的例子
func processUserRequest(userID string) *User {
    user, _ := getUserFromDB(userID) // 忽略错误
    return user
}
```

### 命名约定

- **包名**: 小写，简短，有意义
- **函数名**: 驼峰命名，首字母大写表示公开
- **变量名**: 驼峰命名，短作用域用短名字
- **常量名**: 全大写，下划线分隔

### 错误处理

```go
// 推荐：包装错误，提供上下文
if err != nil {
    return fmt.Errorf("failed to process request: %w", err)
}

// 推荐：自定义错误类型
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error in field %s: %s", e.Field, e.Message)
}
```

### 文档注释

```go
// ProcessRequest 处理用户请求并返回结果。
// 参数 ctx 用于控制请求的取消和超时。
// 参数 req 包含请求的详细信息。
// 返回处理结果和可能的错误。
func ProcessRequest(ctx context.Context, req *Request) (*Response, error) {
    // 实现
}
```

## 性能优化

### 1. 内存优化

```go
// 使用对象池
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 0, 1024)
    },
}

func processData(data []byte) {
    buf := bufferPool.Get().([]byte)
    defer bufferPool.Put(buf[:0])
    
    // 使用 buf 处理数据
}
```

### 2. 并发优化

```go
// 使用带缓冲的 channel
resultChan := make(chan Result, workerCount)

// 使用 context 控制超时
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
```

### 3. I/O 优化

```go
// 使用缓冲读写
reader := bufio.NewReader(file)
writer := bufio.NewWriter(file)
defer writer.Flush()
```

## 调试技巧

### 1. 使用调试器

```bash
# 安装 Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# 调试程序
dlv debug ./cmd/dizi
```

### 2. 添加调试日志

```go
import "dizi/internal/logger"

func debugFunction() {
    logger.InfoLog("Debug: entering function with param: %v", param)
    defer logger.InfoLog("Debug: exiting function")
    
    // 函数逻辑
}
```

### 3. 性能分析

```bash
# CPU 分析
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# 内存分析
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof
```

## 发布流程

### 1. 版本标记

```bash
# 创建新版本标签
git tag -a v1.2.0 -m "Release version 1.2.0"
git push origin v1.2.0
```

### 2. 自动构建

GitHub Actions 会自动：
- 运行所有测试
- 构建多平台二进制文件
- 创建 GitHub Release
- 上传构建产物

### 3. 手动发布

如需手动构建：

```bash
# 使用 Makefile
make build-all

# 或手动构建各平台
./scripts/build-release.sh
```

## 贡献指南

### 提交信息格式

使用约定式提交格式：

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

类型：
- `feat`: 新功能
- `fix`: 错误修复
- `docs`: 文档更新
- `style`: 代码格式
- `refactor`: 重构
- `test`: 测试相关
- `chore`: 构建工具等

示例：
```
feat(tools): add new builtin math tool

Add a new builtin tool for basic mathematical operations.
Supports addition, subtraction, multiplication, and division.

Closes #123
```

### Pull Request 流程

1. Fork 项目
2. 创建功能分支
3. 编写代码和测试
4. 确保所有测试通过
5. 提交 Pull Request
6. 等待代码审查
7. 根据反馈进行修改
8. 合并到主分支

### 代码审查清单

- [ ] 代码遵循项目规范
- [ ] 包含充分的测试
- [ ] 测试覆盖率符合要求
- [ ] 文档已更新
- [ ] 没有引入安全问题
- [ ] 性能影响可接受
- [ ] 兼容性考虑
- [ ] 错误处理适当

## 常见问题

### Q: 如何添加新的工具类型？

A: 在 `internal/tools/tools.go` 的 `RegisterTools` 函数中添加新的 case 分支，并实现相应的处理逻辑。

### Q: 如何调试 MCP 协议通信？

A: 使用 SSE 模式启动服务器，通过浏览器开发者工具或 curl 观察网络请求。

### Q: 如何优化工具执行性能？

A: 考虑使用缓存、对象池、并发处理等技术。具体方法参考性能优化章节。

### Q: 如何处理跨平台兼容性？

A: 使用 Go 标准库的跨平台 API，避免平台特定的系统调用。在 CI 中测试多个平台。

## 相关资源

- [Go 官方文档](https://golang.org/doc/)
- [MCP 协议规范](https://spec.modelcontextprotocol.io/)
- [项目 Issue 跟踪](https://github.com/your-org/dizi/issues)
- [讨论区](https://github.com/your-org/dizi/discussions)