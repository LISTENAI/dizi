# Dizi - MCP Server

**Dizi** 是一个基于 Go 语言实现的 Model Context Protocol (MCP) 服务器，支持通过 YAML 配置文件动态定义和管理工具。

## 特性

- 🛠️ **配置驱动** - 通过 `dizi.yml` 配置文件定义服务器和工具
- 🔄 **多传输方式** - 支持 stdio 和 SSE (Server-Sent Events) 两种传输方式
- 📦 **多种工具类型** - 支持 builtin、command、script 三种工具类型
- 📁 **文件系统工具** - 内置完整的文件系统操作工具集（兼容 mcp-filesystem-server）
- 🎯 **参数验证** - 基于 JSON Schema 的参数验证
- 🔗 **动态工具启用** - 支持通过查询参数或命令行动态启用文件系统工具
- 📝 **日志管理** - 智能日志输出，stdio 模式下不干扰协议通信
- ⚡ **高性能** - Go 语言实现，低内存占用，快速响应

## 快速开始

### 安装

```bash
# 克隆仓库
git clone <repository-url>
cd dizi

# 编译
go build -o dizi

# 或者使用 make
make build
```

### 配置

创建 `dizi.yml` 配置文件：

```yaml
name: "my-mcp-server"
version: "1.0.0"
description: "我的 MCP 服务器"

server:
  port: 8081

tools:
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

  - name: "list_files"
    description: "列出目录中的文件"
    type: "command"
    command: "ls"
    args: ["-la", "{{path}}"]
    parameters:
      type: "object"
      properties:
        path:
          type: "string"
          description: "要列出的目录路径"
      required: ["path"]
```

### 使用

```bash
# 使用 SSE 传输启动（默认）
./dizi

# 使用 stdio 传输启动
./dizi -transport=stdio

# 启用文件系统工具
./dizi -fs-tools

# 启用文件系统工具并指定根目录
./dizi -fs-tools -fs-root=/home/user

# 指定端口
./dizi -port=9000

# 显示帮助
./dizi -help
```

### 文件系统工具

Dizi 内置了完整的文件系统操作工具集，兼容 [mcp-filesystem-server](https://github.com/mark3labs/mcp-filesystem-server)：

#### 启用方式

1. **命令行启用**（适用于 stdio 和 SSE 模式）：
   ```bash
   ./dizi -fs-tools                    # 启用文件系统工具（仅限项目目录）
   ./dizi -fs-tools -fs-root=/path     # 指定其他根目录
   ```

2. **SSE 查询参数**（适用于 Claude Desktop 等客户端）：
   ```
   http://localhost:8081/sse?include_fs_tools=true              # 启用（仅限项目目录）
   http://localhost:8081/sse?include_fs_tools=true&fs_root=/path # 指定其他根目录
   ```

#### 安全特性

- **默认项目限制**：默认情况下，文件系统工具只能访问当前项目目录内的文件
- **路径验证**：严格的路径验证防止目录遍历攻击（../）
- **显式启用**：文件系统工具需要明确启用，不会默认开启

#### 可用工具

| 工具名 | 描述 |
|--------|------|
| `read_file` | 读取文件内容 |
| `write_file` | 创建或覆盖文件 |
| `list_directory` | 列出目录内容 |
| `create_directory` | 创建目录 |
| `delete_file` | 删除文件或目录 |
| `copy_file` | 复制文件或目录 |
| `move_file` | 移动或重命名文件 |
| `get_file_info` | 获取文件详细信息 |
| `search_files` | 搜索匹配的文件 |

## 工具类型

### 1. Builtin 工具

内置工具，由程序直接实现：

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

### 2. Command 工具

执行系统命令：

```yaml
- name: "list_files"
  description: "列出目录中的文件"
  type: "command"
  command: "ls"
  args: ["-la", "{{path}}"]
  parameters:
    type: "object"
    properties:
      path:
        type: "string"
        description: "目录路径"
    required: ["path"]
```

### 3. Script 工具

执行 shell 脚本：

```yaml
- name: "current_time"
  description: "获取当前时间"
  type: "script"
  script: "date '+%Y-%m-%d %H:%M:%S'"

- name: "zephyr_build"
  description: "编译 Zephyr 项目"
  type: "script"
  script: "source .venv/bin/activate && west build -p -s {{source_dir}} -b {{board}}"
  parameters:
    type: "object"
    properties:
      board:
        type: "string"
        description: "目标板型号"
      source_dir:
        type: "string"
        description: "源码目录"
    required: ["board"]
```

## 配置说明

### 服务器配置

| 字段 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `name` | string | 服务器名称 | "dizi" |
| `version` | string | 服务器版本 | "1.0.0" |
| `description` | string | 服务器描述 | "MCP Server" |
| `server.port` | int | SSE 传输端口 | 8080 |

### 工具配置

| 字段 | 类型 | 说明 | 必填 |
|------|------|------|------|
| `name` | string | 工具名称 | ✅ |
| `description` | string | 工具描述 | ✅ |
| `type` | string | 工具类型 (builtin/command/script) | ✅ |
| `command` | string | 命令 (command 类型) | - |
| `args` | []string | 命令参数 (command 类型) | - |
| `script` | string | 脚本内容 (script 类型) | - |
| `parameters` | object | JSON Schema 参数定义 | - |

### 参数占位符

在 `args` 和 `script` 中可以使用 `{{parameter_name}}` 占位符，会被实际参数值替换：

```yaml
# 命令参数占位符
args: ["-la", "{{path}}"]

# 脚本占位符
script: "west build -b {{board}} -s {{source_dir}}"
```

## 命令行选项

| 选项 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `-transport` | string | 传输方式 (stdio/sse) | "sse" |
| `-host` | string | SSE 服务器主机 | "localhost" |
| `-port` | int | SSE 服务器端口 (覆盖配置) | 配置文件值 |
| `-fs-tools` | bool | 启用文件系统工具 | false |
| `-fs-root` | string | 文件系统工具根目录 | "." |
| `-help` | bool | 显示帮助信息 | false |

## 开发

### 项目结构

```
dizi/
├── README.md           # 项目文档
├── go.mod             # Go 模块定义
├── go.sum             # Go 依赖锁定
├── Makefile           # 构建脚本
├── dizi.yml           # 配置文件
├── dizi.example.yml   # 示例配置文件
├── install.sh         # 安装脚本
├── cmd/               # 命令行程序
│   └── dizi/
│       └── main.go    # 主程序入口
├── internal/          # 内部包（不对外暴露）
│   ├── config/        # 配置管理
│   │   └── config.go
│   ├── logger/        # 日志管理
│   │   └── logger.go
│   ├── server/        # 服务器实现
│   │   └── sse.go
│   └── tools/         # 工具实现
│       ├── tools.go
│       └── filesystem.go
└── dizi               # 编译后的可执行文件
```

### 构建

```bash
# 开发构建
make build-dev
# 或者
go build -o dizi ./cmd/dizi

# 生产构建
make build
# 或者
go build -ldflags="-s -w" -o dizi ./cmd/dizi

# 跨平台构建
make build-all
```

### 测试

```bash
# 运行测试
go test ./...

# 测试 SSE 模式
./dizi -port=8082

# 测试 stdio 模式
echo '{"jsonrpc":"2.0","method":"tools/list","id":1}' | ./dizi -transport=stdio
```

## 依赖

- [mcp-go](https://github.com/mark3labs/mcp-go) - MCP 协议实现
- [yaml.v3](https://gopkg.in/yaml.v3) - YAML 配置解析

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

## 相关链接

- [Model Context Protocol](https://modelcontextprotocol.io/)
- [MCP Go SDK](https://mcp-go.dev/)