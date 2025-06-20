// Package main implements a configurable MCP (Model Context Protocol) server.
// The server supports both stdio and SSE transports and can be configured
// via YAML files to provide various tools including builtin, command, and script types.
package main

import (
	_ "embed"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"dizi/internal/config"
	"dizi/internal/logger"
	"dizi/internal/tools"

	"github.com/chzyer/readline"
	mcpserver "github.com/mark3labs/mcp-go/server"
	lua "github.com/yuin/gopher-lua"
)

//go:embed dizi.example.yml
var exampleConfig string

//go:embed embedded_lua/hello.lua
var helloLuaScript string

//go:embed embedded_lua/system_info.lua
var systemInfoLuaScript string

//go:embed embedded_lua/file_utils.lua
var fileUtilsLuaScript string

//go:embed embedded_lua/helpers.lua
var helpersLuaScript string

// main is the entry point of the MCP server application.
// It parses command line arguments, loads configuration, registers tools,
// and starts the server with the specified transport method.
func main() {
	// Check for subcommands before parsing flags
	if len(os.Args) > 1 {
		subcommand := os.Args[1]
		// Don't treat flags as subcommands
		if !strings.HasPrefix(subcommand, "-") {
			switch subcommand {
			case "init":
				initCommand()
				return
			case "lua":
				luaCommand()
				return
			case "repl":
				replCommand()
				return
			}
		}
	}

	// Parse command line flags for server mode
	var (
		transport     = flag.String("transport", "sse", "Transport method: stdio or sse")
		host          = flag.String("host", "localhost", "Host for SSE transport")
		portFlag      = flag.Int("port", 0, "Port for SSE transport (overrides config)")
		enableFsTools = flag.Bool("fs-tools", false, "Enable filesystem tools")
		// fsRootDir     = flag.String("fs-root", "", "Root directory for filesystem tools")
		workDir = flag.String("workdir", "", "Working directory for the server")
		help    = flag.Bool("help", false, "Show help information")
	)

	flag.Parse()

	// Change working directory if specified
	if *workDir != "" {
		if err := os.Chdir(*workDir); err != nil {
			log.Fatalf("Failed to change working directory to %s: %v", *workDir, err)
		}
		logger.InfoLog("Changed working directory to: %s", *workDir)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if *help {
		showHelp(cfg)
		return
	}

	// Use port from flag if specified, otherwise use config
	port := cfg.Server.Port
	if *portFlag != 0 {
		port = *portFlag
	}

	// Create MCP server with config values
	mcpServer := mcpserver.NewMCPServer(cfg.Name, cfg.Version)

	// Register tools from config
	if err := tools.RegisterTools(mcpServer, cfg.Tools); err != nil {
		log.Fatalf("Failed to register tools: %v", err)
	}

	// Register filesystem tools if enabled
	if *enableFsTools {
		fsConfig := &tools.FilesystemConfig{}

		// Use command line fs-root if provided, otherwise default to project directory
		// if *fsRootDir != "" {
		// 	fsConfig.RootDirectory = *fsRootDir
		// } else {
		// Default to current working directory (project directory)
		pwd, err := os.Getwd()
		if err != nil {
			fsConfig.RootDirectory = "."
		} else {
			fsConfig.RootDirectory = pwd
		}
		// }

		if err := tools.RegisterFilesystemTools(mcpServer, fsConfig); err != nil {
			log.Fatalf("Failed to register filesystem tools: %v", err)
		}

		logger.InfoLog("Filesystem tools enabled with root: %s", fsConfig.RootDirectory)
	}

	// Setup logging based on transport mode
	logger.SetupLogger(*transport)

	// Start server based on transport
	switch *transport {
	case "stdio":
		// Silent start for stdio mode
		if err := mcpserver.ServeStdio(mcpServer); err != nil {
			log.Fatalf("Failed to start stdio server: %v", err)
		}
	case "sse":
		logger.InfoLog("Starting %s v%s - %s with SSE transport", cfg.Name, cfg.Version, cfg.Description)

		// Register filesystem tools if enabled
		if *enableFsTools {
			fsConfig := &tools.FilesystemConfig{}

			// Use command line fs-root if provided, otherwise default to project directory
			// if *fsRootDir != "" {
			// 	fsConfig.RootDirectory = *fsRootDir
			// } else {
			// Default to current working directory (project directory)
			pwd, err := os.Getwd()
			if err != nil {
				fsConfig.RootDirectory = "."
			} else {
				fsConfig.RootDirectory = pwd
			}
			// }

			if err := tools.RegisterFilesystemTools(mcpServer, fsConfig); err != nil {
				log.Fatalf("Failed to register filesystem tools: %v", err)
			}

			logger.InfoLog("Filesystem tools enabled with root: %s", fsConfig.RootDirectory)
		}

		// Create and start SSE server
		sseServer := mcpserver.NewSSEServer(mcpServer)
		addr := *host + ":" + strconv.Itoa(port)
		logger.InfoLog("SSE server starting on http://%s", addr)
		if err := sseServer.Start(addr); err != nil {
			log.Fatalf("Failed to start SSE server: %v", err)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unsupported transport: %s\n", *transport)
		showHelp(cfg)
		os.Exit(1)
	}
}

func showHelp(cfg *config.Config) {
	fmt.Printf("%s v%s - %s\n", cfg.Name, cfg.Version, cfg.Description)
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  dizi [flags]")
	fmt.Println("  dizi init")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  init")
	fmt.Println("        Create a default dizi.yml configuration file")
	fmt.Println("  lua <script_file>")
	fmt.Println("        Run a Lua script file")
	fmt.Println("  repl")
	fmt.Println("        Start interactive Lua REPL")
	fmt.Println("")
	fmt.Println("Flags:")
	fmt.Println("  -transport string")
	fmt.Println("        Transport method: stdio or sse (default \"sse\")")
	fmt.Println("  -host string")
	fmt.Println("        Host for SSE transport (default \"localhost\")")
	fmt.Println("  -port int")
	fmt.Printf("        Port for SSE transport (default %d from config)\n", cfg.Server.Port)
	fmt.Println("  -fs-tools")
	fmt.Println("        Enable filesystem tools (restricted to project directory)")
	fmt.Println("  -fs-root string")
	fmt.Println("        Root directory for filesystem tools (default: project directory)")
	fmt.Println("  -workdir string")
	fmt.Println("        Working directory for the server")
	fmt.Println("  -help")
	fmt.Println("        Show this help information")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  dizi                           # Start with SSE transport (default)")
	fmt.Println("  dizi -port=9000                # Start with SSE transport on port 9000")
	fmt.Println("  dizi -transport=stdio          # Start with stdio transport")
	fmt.Println("  dizi -transport=stdio -workdir=/path/to/project  # Start stdio in specific directory")
	fmt.Println("  dizi -fs-tools                 # Enable filesystem tools (project only)")
	fmt.Println("  dizi -fs-tools -fs-root=/home  # Enable filesystem tools with custom root")
	fmt.Println("  dizi lua script.lua            # Run a Lua script")
	fmt.Println("  dizi lua dizi_bin/example.lua  # Run the example Lua script")
	fmt.Println("  dizi repl                      # Start interactive Lua REPL")
	fmt.Println("")
	fmt.Println("SSE Query Parameters:")
	fmt.Println("  ?include_fs_tools=true         # Enable filesystem tools (project only)")
	fmt.Println("  ?fs_root=/path                 # Set custom filesystem root")
	fmt.Println("  Example: http://localhost:8081/sse?include_fs_tools=true&fs_root=/home")
	fmt.Println("")
	fmt.Println("Filesystem Tools (when enabled):")
	fmt.Println("  read_file, write_file, list_directory, create_directory,")
	fmt.Println("  delete_file, copy_file, move_file, get_file_info, search_files")
}

// initCommand creates a complete dizi project with configuration and Lua scripts
func initCommand() {
	fmt.Println("🚀 Initializing Dizi MCP Server project...")
	fmt.Println()

	// Check if dizi.yml already exists
	const configFile = "dizi.yml"
	if _, err := os.Stat(configFile); err == nil {
		fmt.Printf("❌ Configuration file '%s' already exists.\n", configFile)
		fmt.Println("To reinitialize, please remove the existing file first.")
		return
	}

	// Create dizi_bin directory
	const luaDir = "dizi_bin"
	if err := os.MkdirAll(luaDir, 0755); err != nil {
		log.Fatalf("Failed to create %s directory: %v", luaDir, err)
	}
	fmt.Printf("✅ Created directory: %s/\n", luaDir)

	// Write the embedded example config to dizi.yml
	if err := os.WriteFile(configFile, []byte(exampleConfig), 0644); err != nil {
		log.Fatalf("Failed to create configuration file: %v", err)
	}
	fmt.Printf("✅ Created configuration file: %s\n", configFile)

	// Create example Lua scripts using embedded content
	luaScripts := map[string]string{
		"dizi_bin/hello.lua":       helloLuaScript,
		"dizi_bin/system_info.lua": systemInfoLuaScript,
		"dizi_bin/file_utils.lua":  fileUtilsLuaScript,
	}

	// Write Lua scripts
	for filename, content := range luaScripts {
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			log.Fatalf("Failed to create %s: %v", filename, err)
		}
		fmt.Printf("✅ Created Lua script: %s\n", filename)
	}

	// Create README.md
	readmeContent := `# Dizi MCP Server Project

This project contains a configured Dizi MCP (Model Context Protocol) server with Lua scripting support.

## Quick Start

1. **Start the server:**
   ` + "```bash" + `
   dizi                          # Start with SSE transport (default)
   dizi -transport=stdio         # Start with stdio transport
   ` + "```" + `

2. **Test Lua functionality:**
   ` + "```bash" + `
   dizi lua dizi_bin/hello.lua
   dizi repl                     # Interactive Lua REPL
   ` + "```" + `

## Configuration

Edit ` + "`dizi.yml`" + ` to customize:
- Server settings (port, etc.)
- Available tools and commands
- Lua script tools

## Lua Scripts

The ` + "`dizi_bin/`" + ` directory contains example Lua scripts:

- ` + "`hello.lua`" + `      - Basic greeting and math functions
- ` + "`system_info.lua`" + ` - System information display
- ` + "`file_utils.lua`" + `  - File operation utilities

## Available Tools

Check ` + "`dizi.yml`" + ` for configured tools including:
- ` + "`lua_eval`" + `        - Execute Lua code directly
- ` + "`shell_eval`" + `      - Execute shell commands
- Lua script tools using your custom scripts

## Usage Examples

### Using MCP Tools

Connect your MCP client to:
- SSE: ` + "`http://localhost:8082/sse`" + `
- stdio: Run with ` + "`-transport=stdio`" + `

### Direct Lua Execution

` + "```bash" + `
# Run a script
dizi lua dizi_bin/hello.lua

# Interactive REPL
dizi repl
lua:1> 2 + 3
=> 5
lua:2> print('Hello, Dizi!')
Hello, Dizi!
` + "```" + `

## Adding Custom Tools

1. Create new Lua scripts in ` + "`dizi_bin/`" + `
2. Add tool definitions to ` + "`dizi.yml`" + `
3. Restart the server

Example tool definition:
` + "```yaml" + `
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
` + "```" + `

For more information, run ` + "`dizi -help`" + `.
`

	if err := os.WriteFile("README.md", []byte(readmeContent), 0644); err != nil {
		log.Fatalf("Failed to create README.md: %v", err)
	}
	fmt.Printf("✅ Created documentation: README.md\n")

	fmt.Println()
	fmt.Println("🎉 Dizi project initialized successfully!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Review and edit dizi.yml configuration")
	fmt.Println("  2. Test Lua scripts: dizi lua dizi_bin/hello.lua")
	fmt.Println("  3. Try the REPL: dizi repl")
	fmt.Println("  4. Start the server: dizi")
	fmt.Println()
	fmt.Println("📖 See README.md for detailed usage instructions.")
}

// luaCommand runs a Lua script file
func luaCommand() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: dizi lua <script_file>\n")
		fmt.Fprintf(os.Stderr, "Example: dizi lua dizi_bin/example.lua\n")
		os.Exit(1)
	}

	scriptFile := os.Args[2]

	// Check if file exists
	if _, err := os.Stat(scriptFile); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Lua script file '%s' not found\n", scriptFile)
		os.Exit(1)
	}

	// Create Lua state
	L := lua.NewState()
	defer L.Close()

	// Capture print output by redirecting Lua's print function
	L.SetGlobal("print", L.NewFunction(func(L *lua.LState) int {
		top := L.GetTop()
		for i := 1; i <= top; i++ {
			if i > 1 {
				fmt.Print("\t")
			}
			fmt.Print(L.Get(i).String())
		}
		fmt.Println()
		return 0
	}))

	// Execute the Lua script
	if err := L.DoFile(scriptFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing Lua script: %v\n", err)
		os.Exit(1)
	}

	// Get the result from a global variable called 'result' if it exists
	result := L.GetGlobal("result")
	if result != lua.LNil {
		fmt.Printf("Result: %s\n", result.String())
	}
}

// replCommand starts an interactive Lua REPL with readline support
func replCommand() {
	fmt.Println("Dizi Lua REPL v1.0.0")
	fmt.Println("Enter Lua code. Type :help for commands, :quit to exit.")
	fmt.Println("Features: history (↑↓), auto-completion (Tab), line editing")
	fmt.Println()

	// Create Lua state
	L := lua.NewState()
	defer L.Close()

	// Setup readline with custom completer
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "lua:1> ",
		HistoryFile:     "/tmp/.dizi_lua_history",
		AutoComplete:    luaCompleter(L),
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})
	if err != nil {
		fmt.Printf("Error initializing readline: %v\n", err)
		return
	}
	defer func() {
		if err := rl.Close(); err != nil {
			fmt.Printf("Warning: Failed to close readline: %v\n", err)
		}
	}()

	// Capture print output
	var lastOutput strings.Builder
	L.SetGlobal("print", L.NewFunction(func(L *lua.LState) int {
		top := L.GetTop()
		for i := 1; i <= top; i++ {
			if i > 1 {
				lastOutput.WriteString("\t")
			}
			lastOutput.WriteString(L.Get(i).String())
		}
		lastOutput.WriteString("\n")
		fmt.Print(lastOutput.String())
		lastOutput.Reset()
		return 0
	}))

	// Add some helpful globals and utilities
	setupLuaHelpers(L)

	lineNum := 1

	for {
		// Update prompt
		rl.SetPrompt(fmt.Sprintf("lua:%d> ", lineNum))

		// Read input
		input, err := rl.Readline()
		if err == readline.ErrInterrupt {
			fmt.Println("Use :quit to exit")
			continue
		} else if err == io.EOF {
			break
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Handle REPL commands
		switch input {
		case ":help":
			if err := L.DoString("help()"); err != nil {
				fmt.Printf("Error executing help(): %v\n", err)
			}
			continue
		case ":quit", ":exit":
			fmt.Println("Bye!")
			return
		case ":clear":
			// Clear screen (works on Unix-like systems)
			fmt.Print("\033[2J\033[H")
			continue
		case ":reset":
			L.Close()
			L = lua.NewState()
			setupLuaHelpers(L)
			// Re-setup print function
			L.SetGlobal("print", L.NewFunction(func(L *lua.LState) int {
				top := L.GetTop()
				for i := 1; i <= top; i++ {
					if i > 1 {
						lastOutput.WriteString("\t")
					}
					lastOutput.WriteString(L.Get(i).String())
				}
				lastOutput.WriteString("\n")
				fmt.Print(lastOutput.String())
				lastOutput.Reset()
				return 0
			}))
			fmt.Println("Lua state reset.")
			lineNum = 1
			continue
		case ":version":
			fmt.Println(L.GetGlobal("_VERSION").String())
			continue
		case ":vars":
			if err := L.DoString("vars()"); err != nil {
				fmt.Printf("Error executing vars(): %v\n", err)
			}
			continue
		case ":history":
			showHistory(rl)
			continue
		}

		// Execute Lua code
		result := executeLuaREPL(L, input)
		if result != "" {
			fmt.Println(result)
		}

		lineNum++
	}
}

// executeLuaREPL executes Lua code in REPL mode and returns the result
func executeLuaREPL(L *lua.LState, code string) string {
	// Reset stack
	L.SetTop(0)

	// Try as expression first (for immediate results)
	if !strings.Contains(code, ";") && !strings.Contains(code, "\n") &&
		!strings.HasPrefix(strings.TrimSpace(code), "function") &&
		!strings.HasPrefix(strings.TrimSpace(code), "local") &&
		!strings.HasPrefix(strings.TrimSpace(code), "if") &&
		!strings.HasPrefix(strings.TrimSpace(code), "for") &&
		!strings.HasPrefix(strings.TrimSpace(code), "while") &&
		!strings.HasPrefix(strings.TrimSpace(code), "print") {

		expressionCode := "return " + code
		err := L.DoString(expressionCode)
		if err == nil && L.GetTop() > 0 {
			returnValue := L.Get(-1)
			if returnValue != lua.LNil {
				return "=> " + returnValue.String()
			}
		}
	}

	// Execute as statement
	if err := L.DoString(code); err != nil {
		return "Error: " + err.Error()
	}

	// Check for result variable
	result := L.GetGlobal("result")
	if result != lua.LNil {
		return "result = " + result.String()
	}

	// Check if there's a return value on stack
	if L.GetTop() > 0 {
		returnValue := L.Get(-1)
		if returnValue != lua.LNil {
			return "=> " + returnValue.String()
		}
	}

	return "" // No output needed
}

// setupLuaHelpers sets up helpful Lua functions using embedded script
func setupLuaHelpers(L *lua.LState) {
	if err := L.DoString(helpersLuaScript); err != nil {
		fmt.Printf("Error loading Lua helpers: %v\n", err)
	}
}

// luaCompleter creates an auto-completer for Lua keywords and variables
func luaCompleter(_ *lua.LState) readline.AutoCompleter {
	return readline.NewPrefixCompleter(
		// REPL commands
		readline.PcItem(":help"),
		readline.PcItem(":quit"),
		readline.PcItem(":exit"),
		readline.PcItem(":clear"),
		readline.PcItem(":reset"),
		readline.PcItem(":version"),
		readline.PcItem(":vars"),
		readline.PcItem(":history"),

		// Lua keywords
		readline.PcItem("and"),
		readline.PcItem("break"),
		readline.PcItem("do"),
		readline.PcItem("else"),
		readline.PcItem("elseif"),
		readline.PcItem("end"),
		readline.PcItem("false"),
		readline.PcItem("for"),
		readline.PcItem("function"),
		readline.PcItem("if"),
		readline.PcItem("in"),
		readline.PcItem("local"),
		readline.PcItem("nil"),
		readline.PcItem("not"),
		readline.PcItem("or"),
		readline.PcItem("repeat"),
		readline.PcItem("return"),
		readline.PcItem("then"),
		readline.PcItem("true"),
		readline.PcItem("until"),
		readline.PcItem("while"),

		// Common functions
		readline.PcItem("print"),
		readline.PcItem("type"),
		readline.PcItem("tostring"),
		readline.PcItem("tonumber"),
		readline.PcItem("pairs"),
		readline.PcItem("ipairs"),
		readline.PcItem("next"),
		readline.PcItem("table"),
		readline.PcItem("string"),
		readline.PcItem("math"),
		readline.PcItem("io"),
		readline.PcItem("os"),

		// Helper functions
		readline.PcItem("help()"),
		readline.PcItem("vars()"),
	)
}

// filterInput filters input characters
func filterInput(r rune) (rune, bool) {
	switch r {
	// Block some control characters that might cause issues
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

// showHistory displays recent command history
func showHistory(rl *readline.Instance) {
	fmt.Println("Recent command history:")
	fmt.Println("(Use ↑↓ arrows to navigate)")
	// Note: readline doesn't provide direct access to history
	// but the user can use arrows to navigate
}
