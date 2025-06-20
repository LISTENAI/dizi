# Dizi MCP Server Project

This project contains a configured Dizi MCP (Model Context Protocol) server with Lua scripting support.

## Quick Start

1. **Start the server:**
   ```bash
   dizi                          # Start with SSE transport (default)
   dizi -transport=stdio         # Start with stdio transport
   ```

2. **Test Lua functionality:**
   ```bash
   dizi lua dizi_bin/hello.lua
   dizi repl                     # Interactive Lua REPL
   ```

## Configuration

Edit `dizi.yml` to customize:
- Server settings (port, etc.)
- Available tools and commands
- Lua script tools

## Lua Scripts

The `dizi_bin/` directory contains example Lua scripts:

- `hello.lua`      - Basic greeting and math functions
- `system_info.lua` - System information display
- `file_utils.lua`  - File operation utilities

## Available Tools

Check `dizi.yml` for configured tools including:
- `lua_eval`        - Execute Lua code directly
- `shell_eval`      - Execute shell commands
- Lua script tools using your custom scripts

## Usage Examples

### Using MCP Tools

Connect your MCP client to:
- SSE: `http://localhost:8082/sse`
- stdio: Run with `-transport=stdio`

### Direct Lua Execution

```bash
# Run a script
dizi lua dizi_bin/hello.lua

# Interactive REPL
dizi repl
lua:1> 2 + 3
=> 5
lua:2> print('Hello, Dizi!')
Hello, Dizi!
```

## Adding Custom Tools

1. Create new Lua scripts in `dizi_bin/`
2. Add tool definitions to `dizi.yml`
3. Restart the server

Example tool definition:
```yaml
- name: "my_tool"
  description: "My custom tool"
  type: "lua"
  script: "dizi_bin/my_script.lua"
  parameters:
    type: "object"
    properties:
      input:
        type: "string"
        description: "Input parameter"
```

For more information, run `dizi -help`.
