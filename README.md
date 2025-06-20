# Dizi MCP 服务器

Dizi 是一个功能强大的 MCP (Model Context Protocol) 服务器，内置完整的 Lua 5.1 脚本支持，可以通过配置文件定义各种工具和命令，为 AI 模型提供丰富的交互能力。

## ✨ 主要特性

- 🚀 **完整的 Lua 支持** - 集成 gopher-lua 提供 Lua 5.1 解释器
- 💻 **命令行脚本执行** - 直接运行 Lua 脚本进行快速测试
- 🎯 **交互式 REPL** - 支持 readline 的专业 Lua 交互环境
- ⚙️ **YAML 配置驱动** - 通过配置文件定义各种工具类型
- 🔧 **内置工具集** - 包含文件操作、命令执行等常用工具
- 📦 **嵌入式脚本** - 使用 go:embed 内置示例脚本
- 🌐 **多种传输方式** - 支持 SSE 和 stdio 传输协议
- 🔒 **安全沙箱** - 文件系统访问限制和路径验证

## 🚀 快速开始

### 安装

```bash
go install github.com/LISTENAI/dizi/cmd/dizi@latest
```

或者克隆源码编译：

```bash
git clone https://github.com/LISTENAI/dizi.git
cd dizi
go build ./cmd/dizi
```

### 初始化项目

```bash
# 在当前目录初始化 Dizi 项目
dizi init

# 在指定目录初始化项目
dizi init -workdir=/path/to/project
```

初始化后会创建：
- `dizi.yml` - 主配置文件
- `dizi_bin/` - Lua 脚本目录
  - `hello.lua` - 问候和数学计算函数
  - `system_info.lua` - 系统信息展示工具
  - `file_utils.lua` - 文件操作工具集

### 启动服务器

```bash
# 使用 SSE 传输启动（默认）
dizi

# 使用 stdio 传输启动
dizi -transport=stdio

# 指定端口和配置文件
dizi -port=8080 -config=custom.yml

# 启用文件系统工具（限制在指定目录）
dizi -enable-fs-tools -fs-root=/safe/directory
```

## 🎯 Lua 功能详解

### 命令行脚本执行

```bash
# 运行指定的 Lua 脚本
dizi lua dizi_bin/hello.lua

# 运行脚本并传递参数
dizi lua dizi_bin/hello.lua name=张三

# 运行系统信息脚本
dizi lua dizi_bin/system_info.lua
```

### 交互式 REPL

```bash
dizi repl
```

REPL 特性：
- ✨ **表达式自动计算** - 输入 `2 + 3` 自动显示 `=> 5`
- 📝 **多行输入支持** - 支持函数定义和复杂代码块
- 🔄 **命令历史** - 上下箭头浏览历史命令
- 📋 **自动补全** - Tab 键触发自动补全
- 🎨 **语法高亮** - 清晰的代码显示
- 🔧 **内置命令**：
  - `:help` - 显示帮助信息
  - `:quit` / `:exit` - 退出 REPL
  - `:clear` - 清屏
  - `:reset` - 重置 Lua 状态
  - `:vars` - 显示所有变量
  - `:history` - 显示命令历史

### REPL 使用示例

```lua
lua:1> 2 + 3
=> 5

lua:2> name = "Dizi"
lua:3> "Hello, " .. name
=> Hello, Dizi

lua:4> function factorial(n)
lua:4>   if n <= 1 then return 1 end
lua:4>   return n * factorial(n-1)
lua:4> end

lua:5> factorial(5)
=> 120

lua:6> :vars
User-defined variables:
  name = Dizi
  factorial = function

lua:7> :help
# Dizi Lua REPL 帮助信息...
```

## ⚙️ 配置说明

### 主配置文件 (dizi.yml)

```yaml
# 服务器基本信息
name: "dizi-server"
version: "1.0.0"
description: "Dizi MCP Server with Lua Support"

# 服务器设置
server:
  host: "localhost"
  port: 8082
  transport: "sse"  # sse 或 stdio

# 工具定义
tools:
  # 内置工具
  - name: "lua_eval"
    description: "执行 Lua 代码片段"
    type: "builtin"
    parameters:
      type: "object"
      properties:
        code:
          type: "string"
          description: "要执行的 Lua 代码"
      required: ["code"]

  - name: "shell_eval"
    description: "执行 Shell 命令"
    type: "builtin"
    parameters:
      type: "object"
      properties:
        command:
          type: "string"
          description: "要执行的命令"
      required: ["command"]

  # Lua 脚本工具
  - name: "hello_tool"
    description: "问候工具，支持自定义名称和数学计算"
    type: "lua"
    script: "dizi_bin/hello.lua"
    parameters:
      type: "object"
      properties:
        name:
          type: "string"
          description: "要问候的名称"
        operation:
          type: "string"
          description: "数学操作 (add, multiply, etc.)"
        a:
          type: "string"
          description: "第一个数字"
        b:
          type: "string"
          description: "第二个数字"

  # 命令行工具
  - name: "git_status"
    description: "查看 Git 状态"
    type: "command"
    command: "git status --porcelain"

  # 脚本工具
  - name: "backup_script"
    description: "运行备份脚本"
    type: "script"
    script: "./scripts/backup.sh"
    parameters:
      type: "object"
      properties:
        target:
          type: "string"
          description: "备份目标路径"
```

### 工具类型说明

#### 1. builtin 类型
内置在服务器中的工具：
- `lua_eval` - 执行 Lua 代码
- `shell_eval` - 执行 Shell 命令
- `echo` - 回显消息

#### 2. lua 类型
执行 Lua 脚本文件的工具：
```yaml
- name: "my_lua_tool"
  type: "lua"
  script: "dizi_bin/my_script.lua"  # 相对于工作目录的路径
```

#### 3. command 类型
直接执行命令行命令：
```yaml
- name: "list_files"
  type: "command"
  command: "ls -la {{path}}"  # 支持参数占位符
```

#### 4. script 类型
执行脚本文件：
```yaml
- name: "deploy"
  type: "script"
  script: "./deploy.sh"
  command: "bash"  # 可选：指定解释器
```

## 📝 Lua 脚本开发

### 脚本结构

Lua 脚本通过全局变量接收参数，通过 `result` 变量返回结果：

```lua
-- hello.lua
function greet(name)
    if name == nil or name == "" then
        name = "World"
    end
    return "Hello, " .. name .. "!"
end

-- 主执行逻辑
if name then
    result = greet(name)
else
    result = greet() .. " 请使用 name 参数指定名称"
end
```

### 参数传递

通过工具配置的参数会自动设置为 Lua 全局变量：

```yaml
# 工具调用
{
  "name": "hello_tool",
  "arguments": {
    "name": "张三",
    "operation": "add",
    "a": "10",
    "b": "20"
  }
}
```

在 Lua 脚本中可以直接使用：
```lua
print("name:", name)        -- "张三"
print("operation:", operation)  -- "add"
print("a:", a)              -- "10"
print("b:", b)              -- "20"
```

### 内置函数和库

Lua 脚本可以使用：
- **标准 Lua 5.1 库**：`string`, `table`, `math`, `io`, `os`
- **自定义 print 函数**：输出会被捕获并返回
- **JSON 处理**：通过 `json` 库编码解码 JSON

```lua
-- 使用标准库
local time = os.time()
local formatted = os.date("%Y-%m-%d %H:%M:%S", time)

-- 数学计算
local sqrt_result = math.sqrt(16)

-- 字符串处理
local upper_name = string.upper(name or "default")

-- 设置返回结果
result = {
    time = formatted,
    sqrt = sqrt_result,
    name = upper_name
}
```

## 🌐 MCP 客户端集成

### SSE 传输

连接到 `http://localhost:8082/sse` 使用 Server-Sent Events：

```javascript
const eventSource = new EventSource('http://localhost:8082/sse');
eventSource.onmessage = function(event) {
    const data = JSON.parse(event.data);
    console.log('Received:', data);
};
```

### stdio 传输

通过标准输入输出通信：

```bash
echo '{"jsonrpc":"2.0","method":"tools/list","id":1}' | dizi -transport=stdio
```

### 工具调用示例

```json
{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "lua_eval",
    "arguments": {
      "code": "return math.sqrt(16) + 10"
    }
  },
  "id": 1
}
```

## 🔒 安全特性

### 文件系统限制

启用文件系统工具时，所有文件操作都限制在指定的根目录：

```bash
dizi -enable-fs-tools -fs-root=/safe/project/directory
```

### 路径验证

- 自动解析符号链接
- 防止路径遍历攻击（`../` 等）
- 限制访问范围在根目录内

### Lua 沙箱

- 限制危险的系统调用
- 控制文件访问权限
- 执行超时保护

## 🛠️ 开发和测试

### 运行测试

```bash
# 运行所有测试
go test ./...

# 运行特定模块测试
go test ./cmd/dizi -v
go test ./internal/tools -v

# 代码覆盖率
go test ./... -cover
```

### 代码质量检查

```bash
# 运行 linter
golangci-lint run

# 格式化代码
go fmt ./...

# 检查依赖
go mod tidy
go mod verify
```

### 构建

```bash
# 构建当前平台
go build ./cmd/dizi

# 交叉编译
GOOS=linux GOARCH=amd64 go build ./cmd/dizi
GOOS=windows GOARCH=amd64 go build ./cmd/dizi
```

## 📚 使用场景

### 1. AI 模型工具扩展
为 Claude、GPT 等模型提供文件操作、代码执行等能力。

### 2. 自动化脚本执行
通过 MCP 协议远程执行各种自动化任务。

### 3. 开发环境集成
在 IDE 或编辑器中集成，提供智能代码分析和操作。

### 4. CI/CD 流水线
作为构建和部署流程中的工具执行器。

## 🤝 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交变更 (`git commit -m 'Add amazing feature'`)
4. 推送分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🔗 相关链接

- [Model Context Protocol 规范](https://github.com/anthropics/mcp)
- [gopher-lua 文档](https://github.com/yuin/gopher-lua)
- [项目问题反馈](https://github.com/LISTENAI/dizi/issues)

## 📞 支持

如果遇到问题或需要帮助：

1. 查看 [FAQ](docs/FAQ.md)
2. 搜索或创建 [Issue](https://github.com/LISTENAI/dizi/issues)
3. 参与 [Discussions](https://github.com/LISTENAI/dizi/discussions)

---

**Dizi** - 让 AI 模型拥有更强大的工具能力！ 🚀