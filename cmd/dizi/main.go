// Package main implements a configurable MCP (Model Context Protocol) server.
// The server supports both stdio and SSE transports and can be configured
// via YAML files to provide various tools including builtin, command, and script types.
package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"

	"dizi/internal/config"
	"dizi/internal/logger"
	"dizi/internal/server"
	"dizi/internal/tools"

	mcpserver "github.com/mark3labs/mcp-go/server"
)

//go:embed dizi.example.yml
var exampleConfig string

// main is the entry point of the MCP server application.
// It parses command line arguments, loads configuration, registers tools,
// and starts the server with the specified transport method.
func main() {
	// Check for subcommand
	if len(os.Args) > 1 && os.Args[1] == "init" {
		initCommand()
		return
	}

	// Parse command line flags
	var (
		transport     = flag.String("transport", "sse", "Transport method: stdio or sse")
		host          = flag.String("host", "localhost", "Host for SSE transport")
		portFlag      = flag.Int("port", 0, "Port for SSE transport (overrides config)")
		enableFsTools = flag.Bool("fs-tools", false, "Enable filesystem tools")
		fsRootDir     = flag.String("fs-root", "", "Root directory for filesystem tools")
		workDir       = flag.String("workdir", "", "Working directory for the server")
		help          = flag.Bool("help", false, "Show help information")
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
		if *fsRootDir != "" {
			fsConfig.RootDirectory = *fsRootDir
		} else {
			// Default to current working directory (project directory)
			pwd, err := os.Getwd()
			if err != nil {
				fsConfig.RootDirectory = "."
			} else {
				fsConfig.RootDirectory = pwd
			}
		}

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
		// Use custom SSE server with query parameter support
		if err := server.StartCustomSSEServer(cfg, *host, port, *enableFsTools, *fsRootDir); err != nil {
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

// initCommand creates a default dizi.yml configuration file in the current directory
func initCommand() {
	const configFile = "dizi.yml"
	
	// Check if dizi.yml already exists
	if _, err := os.Stat(configFile); err == nil {
		fmt.Printf("Configuration file '%s' already exists.\n", configFile)
		fmt.Println("To reinitialize, please remove the existing file first.")
		return
	}
	
	// Write the embedded example config to dizi.yml
	err := os.WriteFile(configFile, []byte(exampleConfig), 0644)
	if err != nil {
		log.Fatalf("Failed to create configuration file: %v", err)
	}
	
	fmt.Printf("Successfully created '%s' configuration file.\n", configFile)
	fmt.Println("You can now edit the configuration file to customize your tools and settings.")
	fmt.Println("Run 'dizi' to start the server with your configuration.")
}
