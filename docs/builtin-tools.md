# 内置工具 (Builtin Tools)

Dizi 提供了一套内置工具，这些工具由 Go 代码直接实现，提供高性能和安全的功能。

## 概述

内置工具是 Dizi 服务器内部实现的工具，不依赖外部命令或脚本。这些工具具有以下特点：

- **高性能**：直接由 Go 代码实现，无需启动外部进程
- **安全性**：严格的参数验证和错误处理
- **一致性**：标准化的接口和行为
- **可靠性**：经过充分测试，适合生产环境使用

## 可用的内置工具

### Echo 工具

**名称：** `echo`

**描述：** 回显输入的消息，主要用于测试和调试。

**配置示例：**
```yaml
- name: "echo"
  description: "回显输入的消息"
  type: "builtin"
  parameters:
    type: "object"
    properties:
      message:
        type: "string"
        description: "要回显的消息"
    required: ["message"]
```

**使用方法：**
```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "echo",
    "arguments": {
      "message": "Hello, World!"
    }
  },
  "id": 1
}
```

**返回结果：**
```json
{
  "jsonrpc": "2.0",
  "result": {
    "content": [
      {
        "type": "text",
        "text": "Hello, World!"
      }
    ],
    "isError": false
  },
  "id": 1
}
```

## 实现原理

### 工具注册

内置工具在服务器启动时自动注册：

```go
func createBuiltinHandler(tool config.ToolConfig) mcp.ToolHandler {
    return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        switch tool.Name {
        case "echo":
            return handleEcho(request)
        default:
            return mcp.NewToolResultError("Unknown builtin tool: " + tool.Name), nil
        }
    }
}
```

### 参数处理

内置工具使用统一的参数处理机制：

```go
func handleEcho(request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    args := request.Params.Arguments
    if args == nil {
        return mcp.NewToolResultError("Missing arguments"), nil
    }
    
    message, ok := args["message"].(string)
    if !ok {
        return mcp.NewToolResultError("Invalid message parameter"), nil
    }
    
    return mcp.NewToolResultText(message), nil
}
```

### 错误处理

内置工具采用统一的错误处理策略：

- **参数验证错误**：返回 `isError: true` 的结果，包含错误描述
- **执行错误**：同样返回错误结果，而不是抛出异常
- **系统错误**：极少数情况下返回 Go error，由框架处理

## 开发新的内置工具

### 1. 定义工具处理函数

```go
func handleMyTool(request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    // 参数验证
    args := request.Params.Arguments
    if args == nil {
        return mcp.NewToolResultError("Missing arguments"), nil
    }
    
    // 获取参数
    param1, ok := args["param1"].(string)
    if !ok {
        return mcp.NewToolResultError("Invalid param1"), nil
    }
    
    // 执行业务逻辑
    result := doSomething(param1)
    
    // 返回结果
    return mcp.NewToolResultText(result), nil
}
```

### 2. 注册工具

在 `createBuiltinHandler` 函数中添加新工具：

```go
case "my_tool":
    return handleMyTool(request)
```

### 3. 配置文件

用户可以在配置文件中使用新工具：

```yaml
- name: "my_tool"
  description: "我的自定义工具"
  type: "builtin"
  parameters:
    type: "object"
    properties:
      param1:
        type: "string"
        description: "参数1"
    required: ["param1"]
```

## 最佳实践

### 参数验证

```go
// 推荐：详细的参数验证
func validateParams(args map[string]interface{}) error {
    if args == nil {
        return fmt.Errorf("missing arguments")
    }
    
    if _, ok := args["required_param"]; !ok {
        return fmt.Errorf("missing required parameter: required_param")
    }
    
    if value, ok := args["string_param"].(string); !ok || value == "" {
        return fmt.Errorf("invalid string_param")
    }
    
    return nil
}
```

### 错误消息

```go
// 推荐：清晰的错误消息
return mcp.NewToolResultError("Invalid file path: path must be absolute"), nil

// 不推荐：模糊的错误消息
return mcp.NewToolResultError("Error"), nil
```

### 资源管理

```go
// 推荐：适当的资源清理
func handleFileOperation(request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    file, err := os.Open(filename)
    if err != nil {
        return mcp.NewToolResultErrorFromErr("Failed to open file", err), nil
    }
    defer file.Close()
    
    // 处理文件
    return result, nil
}
```

## 安全考虑

### 输入验证

- 所有用户输入必须进行严格验证
- 使用白名单而不是黑名单进行验证
- 防止注入攻击

### 权限控制

- 内置工具应该具有最小权限
- 避免执行任意代码或命令
- 限制文件系统访问范围

### 错误信息

- 错误消息不应暴露敏感信息
- 使用统一的错误格式
- 记录详细日志用于调试

## 测试

### 单元测试

```go
func TestHandleEcho(t *testing.T) {
    request := mcp.CallToolRequest{
        Params: mcp.CallToolParams{
            Arguments: map[string]interface{}{
                "message": "test message",
            },
        },
    }
    
    result, err := handleEcho(request)
    if err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }
    
    if result.IsError {
        t.Errorf("Expected success, got error: %v", result.Content)
    }
    
    // 验证结果内容
    // ...
}
```

### 集成测试

内置工具的集成测试应该验证：

- 工具注册是否正确
- 参数验证是否有效
- 错误处理是否正确
- 性能是否符合预期

## 性能优化

### 缓存

对于计算密集型工具，考虑使用缓存：

```go
var cache = make(map[string]string)
var cacheMutex sync.RWMutex

func getCachedResult(key string) (string, bool) {
    cacheMutex.RLock()
    defer cacheMutex.RUnlock()
    result, exists := cache[key]
    return result, exists
}
```

### 并发处理

内置工具应该是并发安全的：

```go
// 使用原子操作或互斥锁保护共享状态
var counter int64

func incrementCounter() int64 {
    return atomic.AddInt64(&counter, 1)
}
```

## 调试

### 日志记录

```go
import "dizi/internal/logger"

func handleMyTool(request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    logger.InfoLog("Processing tool request: %s", request.Params.Name)
    // 处理逻辑
    logger.InfoLog("Tool request completed successfully")
    return result, nil
}
```

### 调试模式

可以添加调试模式来提供更详细的信息：

```go
func handleDebugInfo(request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    info := map[string]interface{}{
        "version": "1.0.0",
        "uptime": time.Since(startTime).String(),
        "memory": getMemoryStats(),
    }
    
    jsonData, _ := json.MarshalIndent(info, "", "  ")
    return mcp.NewToolResultText(string(jsonData)), nil
}
```