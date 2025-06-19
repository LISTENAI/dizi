# Dizi - MCP 服务器

> 注意：这个项目几乎是纯AI生成的项目，我只有在编译器出错AI无法解的时候介入了一下，现在属于初级阶段，初步验证都可以使用。

**Dizi** 是一个基于 Model Context Protocol (MCP) 的服务器，支持通过 YAML 配置文件动态定义和管理工具。

## 特性

- 🛠️ **配置驱动** - 通过 `dizi.yml` 配置文件定义服务器和工具
- 🔄 **多传输方式** - 支持 stdio 和 SSE (Server-Sent Events) 两种传输方式
- 📦 **多种工具类型** - 支持 command、script 两种工具类型
- 📁 **文件系统工具** - 内置完整的文件系统操作工具集
- 🎯 **参数验证** - 基于 JSON Schema 的参数验证
- 🔗 **动态工具启用** - 支持通过查询参数或命令行动态启用文件系统工具
- ⚡ **高性能** - Go 语言实现，低内存占用，快速响应

## 安装

### 从 GitHub Release 下载（推荐）

访问 [Releases 页面](https://github.com/LISTENAI/dizi/releases) 下载适合您系统的预编译二进制文件：

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

## 快速开始

### 1. 初始化配置

在您的项目目录中运行：

```bash
dizi init
```

这会创建一个默认的 `dizi.yml` 配置文件。

### 2. 配置工具

编辑 `dizi.yml` 文件来定义您的工具：

```yaml
name: "my-mcp-server"
version: "1.0.0"
description: "我的 MCP 服务器"

server:
  port: 8081

tools:
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

  - name: "current_time"
    description: "获取当前时间"
    type: "script"
    script: "date '+%Y-%m-%d %H:%M:%S'"
```

### 3. 启动服务器

```bash
# 使用 SSE 传输启动（默认）
dizi

# 使用 stdio 传输启动
dizi -transport=stdio

# 指定工作目录（适用于 stdio 模式）
dizi -transport=stdio -workdir=/path/to/project

# 启用文件系统工具
dizi -fs-tools

# 指定端口
dizi -port=9000
```

## 文件系统工具

Dizi 内置了完整的文件系统操作工具集，提供安全的文件访问功能。

### 启用方式

1. **命令行启用**：
   ```bash
   dizi -fs-tools                    # 启用文件系统工具（仅限项目目录）
   dizi -fs-tools -fs-root=/path     # 指定其他根目录
   ```

2. **SSE 查询参数**（适用于 Claude code 等客户端）：
   ```
   http://localhost:8081/sse?include_fs_tools=true              # 启用（仅限项目目录）
   http://localhost:8081/sse?include_fs_tools=true&fs_root=/path # 指定其他根目录
   ```

### 安全特性

- **默认项目限制**：默认情况下，文件系统工具只能访问当前项目目录内的文件
- **路径验证**：严格的路径验证防止目录遍历攻击（../）
- **显式启用**：文件系统工具需要明确启用，不会默认开启

### 可用工具

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

执行 shell 脚本：

```yaml
- name: "build_project"
  description: "构建项目"
  type: "script"
  script: |
    echo "开始构建..."
    make build
    echo "构建完成"

- name: "greet_user"
  description: "问候用户"
  type: "script"
  script: "echo 'Hello, {{name}}!'"
  parameters:
    type: "object"
    properties:
      name:
        type: "string"
        description: "用户名"
    required: ["name"]
```

## 配置参考

### 服务器配置

```yaml
name: "my-server"           # 服务器名称
version: "1.0.0"           # 服务器版本
description: "My Server"   # 服务器描述

server:
  port: 8081               # SSE 传输端口
```

### 工具配置

| 字段 | 类型 | 说明 | 必填 |
|------|------|------|------|
| `name` | string | 工具名称 | ✅ |
| `description` | string | 工具描述 | ✅ |
| `type` | string | 工具类型 (command/script) | ✅ |
| `command` | string | 命令 (command 类型) | - |
| `args` | []string | 命令参数 (command 类型) | - |
| `script` | string | 脚本内容 (script 类型) | - |
| `parameters` | object | JSON Schema 参数定义 | - |

### 参数占位符

在 `args` 和 `script` 中可以使用 `{{parameter_name}}` 占位符：

```yaml
# 命令参数占位符
args: ["-la", "{{path}}"]

# 脚本占位符
script: "echo 'Hello, {{name}}!'"
```

## 命令行选项

| 选项 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `init` | 命令 | 创建默认配置文件 | - |
| `-transport` | string | 传输方式 (stdio/sse) | "sse" |
| `-host` | string | SSE 服务器主机 | "localhost" |
| `-port` | int | SSE 服务器端口 | 配置文件值 |
| `-fs-tools` | bool | 启用文件系统工具 | false |
| `-fs-root` | string | 文件系统工具根目录 | "." |
| `-workdir` | string | 服务器工作目录 | 当前目录 |
| `-help` | bool | 显示帮助信息 | false |

## 使用示例

### 与 Claude Desktop 集成

**指定工作目录**（适用于 Claude Desktop 无法 cd 到特定目录的情况）：

```json
{
  "mcpServers": {
    "my-project": {
      "command": "dizi",
      "args": ["-transport=stdio", "-workdir=/path/to/your/project", "-fs-tools"],
      "env": {}
    }
  }
}
```

注意：command如果放在自己指定的PATH路径，要不再env设置PATH地址，要不写全路径（建议）

### 开发环境工具配置

```yaml
tools:
  - name: "test"
    description: "运行测试"
    type: "script"
    script: "npm test"

  - name: "dev_server"
    description: "启动开发服务器"
    type: "script"
    script: "npm run dev"

  - name: "git_commit"
    description: "提交 Git 更改"
    type: "command"
    command: "git"
    args: ["commit", "-m", "{{message}}"]
    parameters:
      type: "object"
      properties:
        message:
          type: "string"
          description: "提交信息"
      required: ["message"]
```

## 许可证

MIT License

## 相关链接

- [Model Context Protocol](https://modelcontextprotocol.io/)
- [MCP 规范](https://spec.modelcontextprotocol.io/)