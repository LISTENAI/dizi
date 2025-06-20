# Dizi - 动态配置的 MCP 服务器

[![GitHub Release](https://img.shields.io/github/v/release/LISTENAI/dizi)](https://github.com/LISTENAI/dizi/releases)
[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](https://opensource.org/licenses/MIT)

> 💡 **注意**：这个项目几乎是纯 AI 生成的项目，目前处于初级阶段，已通过初步验证可正常使用。

**Dizi** 是一个基于 [Model Context Protocol (MCP)](https://modelcontextprotocol.io/) 的高性能服务器，支持通过 YAML 配置文件动态定义和管理工具，让 AI 助手能够与您的开发环境无缝集成。

## 📋 目录

- [✨ 核心特性](#-核心特性)
- [🚀 快速安装](#-快速安装)
- [⚡ 快速开始](#-快速开始)
- [🔧 客户端配置](#-客户端配置)
- [📁 文件系统工具](#-文件系统工具)
- [🛠️ 工具类型详解](#️-工具类型详解)
- [📖 配置参考](#-配置参考)
- [🎯 Lua 脚本功能](#-lua-脚本功能)
- [💻 命令行选项](#-命令行选项)

## ✨ 核心特性

| 特性 | 描述 |
|------|------|
| 🛠️ **配置驱动** | 通过 `dizi.yml` 配置文件定义服务器和工具 |
| 🔄 **多传输方式** | 支持 stdio 和 SSE (Server-Sent Events) 两种传输方式 |
| 📦 **丰富工具类型** | 支持 command、script、lua、builtin 四种工具类型 |
| 📁 **文件系统集成** | 内置完整的文件系统操作工具集，支持安全的文件访问 |
| 🎯 **参数验证** | 基于 JSON Schema 的严格参数验证 |
| 🔗 **动态启用** | 支持通过查询参数或命令行动态启用文件系统工具 |
| ⚡ **高性能** | Go 语言实现，低内存占用，毫秒级响应 |
| 🌟 **Lua 支持** | 内置 Lua 解释器，支持脚本工具和交互式 REPL |

## 🚀 快速安装

### 方式一：从 GitHub Release 下载（推荐）

访问 [Releases 页面](https://github.com/LISTENAI/dizi/releases) 下载预编译二进制文件：

<details>
<summary>📱 各平台安装命令</summary>

**Linux (x86_64):**
```bash
wget https://github.com/LISTENAI/dizi/releases/latest/download/dizi-linux-amd64
chmod +x dizi-linux-amd64
sudo mv dizi-linux-amd64 /usr/local/bin/dizi
```

**Linux (ARM64):**
```bash
wget https://github.com/LISTENAI/dizi/releases/latest/download/dizi-linux-arm64
chmod +x dizi-linux-arm64
sudo mv dizi-linux-arm64 /usr/local/bin/dizi
```

**macOS (Intel):**
```bash
wget https://github.com/LISTENAI/dizi/releases/latest/download/dizi-darwin-amd64
chmod +x dizi-darwin-amd64
sudo mv dizi-darwin-amd64 /usr/local/bin/dizi
```

**macOS (Apple Silicon):**
```bash
wget https://github.com/LISTENAI/dizi/releases/latest/download/dizi-darwin-arm64
chmod +x dizi-darwin-arm64
sudo mv dizi-darwin-arm64 /usr/local/bin/dizi
```

**Windows:**
下载 `dizi-windows-amd64.exe` 并将其添加到系统 PATH。

</details>

### 方式二：从源码编译

```bash
git clone https://github.com/LISTENAI/dizi.git
cd dizi
go build -o dizi
```

## ⚡ 快速开始

### 1️⃣ 初始化项目

```bash
# 在您的项目目录中运行
dizi init
```

这将创建：
```
📁 您的项目/
├── 📄 dizi.yml          # 主配置文件
└── 📁 dizi_bin/         # Lua 脚本目录
    ├── 📜 hello.lua     # 问候和数学计算示例
    ├── 📜 system_info.lua  # 系统信息工具
    └── 📜 file_utils.lua   # 文件操作工具
```

### 2️⃣ 启动服务器

```bash
# 默认启动（SSE 模式）
dizi

# 启用文件系统工具
dizi -fs-tools

# 使用 stdio 模式
dizi -transport=stdio

# 指定端口
dizi -port=9000
```

### 3️⃣ 测试工具

```bash
# 交互式 Lua REPL
dizi repl

# 运行 Lua 脚本
dizi lua dizi_bin/hello.lua
```

## 🔧 客户端配置

### Claude Desktop 配置

**基础配置：**
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

**启用文件系统工具：**
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

### SSE 客户端配置

| 连接类型 | 端点 |
|----------|------|
| 基本连接 | `http://localhost:8081/sse` |
| 启用文件系统工具 | `http://localhost:8081/sse?include_fs_tools=true` |
| 自定义根目录（不建议使用该参数） | `http://localhost:8081/sse?include_fs_tools=true&fs_root=/path` |

## 📁 文件系统工具

Dizi 提供安全、完整的文件系统操作能力。

### 🔐 安全特性

- **🛡️ 路径限制**：默认仅允许访问项目目录内的文件
- **🚫 防止遍历**：严格验证路径，防止 `../` 攻击
- **✅ 显式启用**：文件系统工具需要明确启用

### 📋 可用工具

| 工具 | 功能描述 | 示例用法 |
|------|----------|----------|
| `read_file` | 读取文件内容 | 查看配置文件、源代码 |
| `write_file` | 创建或覆盖文件 | 生成代码、更新配置 |
| `list_directory` | 列出目录内容 | 浏览项目结构 |
| `create_directory` | 创建目录 | 组织项目文件 |
| `delete_file` | 删除文件或目录 | 清理临时文件 |
| `copy_file` | 复制文件或目录 | 备份重要文件 |
| `move_file` | 移动或重命名 | 重构项目结构 |
| `get_file_info` | 获取文件详情 | 检查文件大小、权限 |
| `search_files` | 搜索匹配文件 | 查找特定文件 |

## 🛠️ 工具类型详解

### Command 工具

执行系统命令：

```yaml
- name: "git_status"
  description: "获取 Git 仓库状态"
  type: "command"
  command: "git"
  args: ["status", "--porcelain"]
```

### Script 工具

执行 shell 脚本，支持参数占位符：

```yaml
- name: "greet_user"
  description: "问候用户"
  type: "script"
  script: "echo 'Hello, {{name}}! 今天是 $(date)'"
  parameters:
    type: "object"
    properties:
      name:
        type: "string"
        description: "用户名"
    required: ["name"]
```

### Lua 工具

执行 Lua 脚本文件：

```yaml
- name: "hello_tool"
  description: "问候和基础数学计算工具"
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
        description: "数学操作 (add, multiply, subtract, divide)"
```

### Builtin 工具

内置的特殊工具：

```yaml
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
```

## 📖 配置参考

### 完整配置示例

```yaml
# Dizi MCP Server Configuration
name: "my-mcp-server"
version: "1.0.0"
description: "My Custom MCP Server"

server:
  port: 8081

tools:
  # Shell 命令执行
  - name: "shell_eval"
    description: "执行标准的 bash 命令"
    type: "script"
    script: "{{command}}"
    parameters:
      type: "object"
      properties:
        command:
          type: "string"
          description: "要执行的 shell 命令"
      required: ["command"]

  # Git 操作
  - name: "git_status"
    description: "显示 Git 仓库状态"
    type: "script"
    script: "cd {{path}} && git status --porcelain"
    parameters:
      type: "object"
      properties:
        path:
          type: "string"
          description: "Git 仓库路径"
      required: ["path"]

  # 时间工具
  - name: "current_time"
    description: "获取当前日期和时间"
    type: "script"
    script: "date '+%Y-%m-%d %H:%M:%S'"

  # Zephyr 开发工具
  - name: "zephyr_build"
    description: "编译 Zephyr 项目"
    type: "script"
    script: "source .venv/bin/activate && west build -p -s {{source_dir}} -b {{board}}"
    parameters:
      type: "object"
      properties:
        board:
          type: "string"
          description: "目标开发板"
        source_dir:
          type: "string"
          description: "源码目录（可选）"
      required: ["board"]
```

### 参数配置说明

| 字段 | 类型 | 说明 | 必填 |
|------|------|------|------|
| `name` | string | 工具唯一名称 | ✅ |
| `description` | string | 工具功能描述 | ✅ |
| `type` | string | 工具类型：`command`/`script`/`lua`/`builtin` | ✅ |
| `command` | string | 系统命令（command 类型） | - |
| `args` | []string | 命令参数（command 类型） | - |
| `script` | string | 脚本内容或文件路径 | - |
| `parameters` | object | JSON Schema 参数定义 | - |

## 🎯 Lua 脚本功能

### 命令行脚本执行

```bash
# 运行 Lua 脚本
dizi lua dizi_bin/hello.lua

# 运行系统信息脚本
dizi lua dizi_bin/system_info.lua
```

### 交互式 REPL

启动交互式 Lua 环境：

```bash
dizi repl
```

**REPL 特性：**
- ✨ **自动计算**：输入 `2 + 3` 显示 `=> 5`
- 🔄 **历史记录**：上下箭头浏览命令历史
- 📋 **自动补全**：Tab 键触发智能补全
- 🎨 **语法高亮**：彩色代码显示

**内置命令：**
- `:help` - 显示帮助信息
- `:quit` / `:exit` - 退出 REPL
- `:clear` - 清屏
- `:reset` - 重置 Lua 状态
- `:vars` - 显示所有变量
- `:history` - 显示命令历史

**使用示例：**
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

### Lua 脚本开发指南

- **兼容性**：支持 Lua 5.1 语法
- **输入获取**：使用全局变量 `args` 获取外部输入
- **结果返回**：使用全局变量 `result` 返回结果
- **调试输出**：使用 `print()` 函数打印调试信息

**参考资源：**
- [Gopher-Lua 文档](https://github.com/yuin/gopher-lua)
- [Lua 5.1 手册](https://www.lua.org/manual/5.1/)

## 💻 命令行选项

### 基本命令

| 命令 | 描述 |
|------|------|
| `dizi init` | 创建默认配置文件和示例脚本 |
| `dizi` | 启动服务器（默认 SSE 模式） |
| `dizi repl` | 启动交互式 Lua REPL |
| `dizi lua <script>` | 执行指定的 Lua 脚本 |

### 服务器选项

| 选项 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `-transport` | string | 传输方式：`stdio`/`sse` | `sse` |
| `-host` | string | SSE 服务器主机地址 | `localhost` |
| `-port` | int | SSE 服务器端口 | 配置文件值或 `8081` |
| `-workdir` | string | 服务器工作目录 | 当前目录 |

### 文件系统选项

| 选项 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `-fs-tools` | bool | 启用文件系统工具 | `false` |
| `-fs-root` | string | 文件系统工具根目录，__除非你知道你在干嘛__，不然不要用 | `.` |

### 其他选项

| 选项 | 类型 | 说明 |
|------|------|------|
| `-help` | bool | 显示详细帮助信息 |
| `-version` | bool | 显示版本信息 |

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建功能分支：`git checkout -b feature/amazing-feature`
3. 提交更改：`git commit -m 'Add amazing feature'`
4. 推送分支：`git push origin feature/amazing-feature`
5. 提交 Pull Request

## 📄 许可证

本项目采用 [MIT License](https://opensource.org/licenses/MIT) 开源协议。

## 🔗 相关链接

- [Model Context Protocol 官网](https://modelcontextprotocol.io/)
- [MCP 规范文档](https://spec.modelcontextprotocol.io/)
- [项目 GitHub 仓库](https://github.com/LISTENAI/dizi)
- [问题反馈](https://github.com/LISTENAI/dizi/issues)

---

<div align="center">

**🌟 如果这个项目对您有帮助，请给它一个 Star！**

Made with ❤️ by AI & Human Collaboration

</div>