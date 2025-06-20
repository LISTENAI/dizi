# Dizi MCP 服务器

Dizi 是一个功能强大的 MCP (Model Context Protocol) 服务器，内置完整的 Lua 5.1 脚本支持，可以通过配置文件定义各种工具和命令，为 AI 模型提供丰富的交互能力。

## 🚀 快速开始

### 安装

下载最新的二进制文件：

```bash
# macOS/Linux
curl -L https://github.com/LISTENAI/dizi/releases/latest/download/dizi-$(uname -s)-$(uname -m) -o dizi
chmod +x dizi
sudo mv dizi /usr/local/bin/

# Windows (PowerShell)
Invoke-WebRequest -Uri "https://github.com/LISTENAI/dizi/releases/latest/download/dizi-Windows-x86_64.exe" -OutFile "dizi.exe"
```

### 初始化项目

**首先必须使用 init 命令初始化项目：**

```bash
# 在当前目录初始化项目
dizi init

# 在指定目录初始化项目  
dizi init -workdir=/path/to/project
```

初始化后会创建：
- `dizi.yml` - 主配置文件
- `dizi_bin/` - Lua 脚本目录
  - `hello.lua` - 问候和数学计算示例
  - `system_info.lua` - 系统信息工具
  - `file_utils.lua` - 文件操作工具

### 启动服务器

#### stdio 传输方式（推荐用于 MCP 客户端）

```bash
# 基本启动
dizi -transport=stdio

# 指定工作目录（必须配置项目目录）
dizi -transport=stdio -workdir=/path/to/your/project

# 启用文件系统工具（仅 stdio 模式支持）
dizi -transport=stdio -workdir=/path/to/project -fs-tools

# 自定义文件系统根目录
dizi -transport=stdio -workdir=/path/to/project -fs-tools -fs-root=/safe/directory
```

#### SSE 传输方式（用于 Web 客户端）

```bash
# 基本启动（默认端口 8082）
dizi

# 自定义端口
dizi -port=9000

# 指定工作目录
dizi -workdir=/path/to/project
```

## ⚙️ MCP 客户端配置

### stdio 配置

在你的 MCP 客户端（如 Claude Desktop）中配置：

```json
{
  "mcpServers": {
    "dizi": {
      "command": "dizi",
      "args": ["-transport=stdio", "-workdir=/path/to/your/project"],
      "env": {}
    }
  }
}
```

如果需要文件系统工具：

```json
{
  "mcpServers": {
    "dizi": {
      "command": "dizi", 
      "args": ["-transport=stdio", "-workdir=/path/to/your/project", "-fs-tools"],
      "env": {}
    }
  }
}
```

### SSE 配置

连接到 SSE 端点：
- 基本连接：`http://localhost:8082/sse`
- 启用文件系统工具：`http://localhost:8082/sse?include_fs_tools=true`
- 自定义根目录：`http://localhost:8082/sse?include_fs_tools=true&fs_root=/path`

## 🎯 Lua 功能

### 命令行脚本执行

```bash
# 运行 Lua 脚本
dizi lua dizi_bin/hello.lua

# 运行其他脚本  
dizi lua dizi_bin/system_info.lua
```

### 交互式 REPL

```bash
dizi repl
```

REPL 特性：
- ✨ **表达式自动计算** - 输入 `2 + 3` 显示 `=> 5`
- 🔄 **命令历史** - 上下箭头浏览历史
- 📋 **自动补全** - Tab 键触发补全
- 🔧 **内置命令**：
  - `:help` - 显示帮助
  - `:quit` / `:exit` - 退出
  - `:clear` - 清屏
  - `:reset` - 重置状态
  - `:vars` - 显示变量
  - `:history` - 显示历史

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
```

## 📝 配置文件说明

### dizi.yml 配置示例

```yaml
name: "dizi-server"
version: "1.0.0"
description: "Dizi MCP Server with Lua Support"

server:
  host: "localhost"
  port: 8082
  transport: "sse"

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
    description: "问候工具"
    type: "lua"
    script: "dizi_bin/hello.lua"
    parameters:
      type: "object"
      properties:
        name:
          type: "string"
          description: "要问候的名称"

  # 命令行工具
  - name: "git_status"
    description: "查看 Git 状态"
    type: "command"
    command: "git status --porcelain"

  # 脚本工具
  - name: "backup"
    description: "运行备份脚本"
    type: "script"
    script: "./scripts/backup.sh"
```

### 工具类型

1. **builtin** - 内置工具（`lua_eval`, `shell_eval`, `echo`）
2. **lua** - 执行 Lua 脚本文件
3. **command** - 执行命令行命令
4. **script** - 执行脚本文件

## 🔧 文件系统工具

**重要：文件系统工具仅在 stdio 传输模式下正确支持**

启用后提供以下工具：
- `read_file` - 读取文件
- `write_file` - 写入文件  
- `list_directory` - 列出目录
- `create_directory` - 创建目录
- `delete_file` - 删除文件
- `copy_file` - 复制文件
- `move_file` - 移动文件
- `get_file_info` - 获取文件信息
- `search_files` - 搜索文件

### 安全特性

- 所有文件操作限制在指定根目录内
- 自动路径验证，防止目录遍历攻击
- 符号链接自动解析

## 📚 Lua 脚本开发

### 基本结构

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
    result = greet()
end
```

### 参数传递

工具参数自动设置为 Lua 全局变量：

```json
{
  "name": "hello_tool",
  "arguments": {
    "name": "张三"
  }
}
```

在脚本中：
```lua
print("name:", name)  -- "张三"
result = "Hello, " .. name
```

## 🛠️ 开发

### 命令行选项

```bash
dizi -help  # 查看所有选项
```

主要选项：
- `-transport` - 传输方式（stdio/sse）
- `-workdir` - 工作目录（建议总是指定）
- `-fs-tools` - 启用文件系统工具（仅 stdio）
- `-fs-root` - 文件系统根目录
- `-port` - SSE 端口
- `-host` - SSE 主机

### 测试

```bash
# 运行测试
go test ./...

# 代码检查
golangci-lint run
```

## 📞 支持

- [项目主页](https://github.com/LISTENAI/dizi)
- [问题反馈](https://github.com/LISTENAI/dizi/issues)
- [功能请求](https://github.com/LISTENAI/dizi/discussions)

---

**注意：使用前必须先运行 `dizi init` 初始化项目！**