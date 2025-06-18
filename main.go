package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/server"
)

func main() {
	var (
		transport = flag.String("transport", "stdio", "Transport method: stdio or sse")
		host      = flag.String("host", "localhost", "Host for SSE transport")
		portFlag  = flag.Int("port", 0, "Port for SSE transport (overrides config)")
		help      = flag.Bool("help", false, "Show help information")
	)
	
	flag.Parse()

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if *help {
		showHelp(config)
		return
	}

	// Use port from flag if specified, otherwise use config
	port := config.Server.Port
	if *portFlag != 0 {
		port = *portFlag
	}

	// Create MCP server with config values
	mcpServer := server.NewMCPServer(config.Name, config.Version)

	// Register tools from config
	if err := registerTools(mcpServer, config.Tools); err != nil {
		log.Fatalf("Failed to register tools: %v", err)
	}

	// Start server based on transport
	switch *transport {
	case "stdio":
		log.Printf("Starting %s v%s - %s with stdio transport...", config.Name, config.Version, config.Description)
		if err := server.ServeStdio(mcpServer); err != nil {
			log.Fatalf("Failed to start stdio server: %v", err)
		}
	case "sse":
		log.Printf("Starting %s v%s - %s with SSE transport on %s:%d...", config.Name, config.Version, config.Description, *host, port)
		// Create SSE server
		sseServer := server.NewSSEServer(mcpServer)
		if err := sseServer.Start(fmt.Sprintf("%s:%d", *host, port)); err != nil {
			log.Fatalf("Failed to start SSE server: %v", err)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unsupported transport: %s\n", *transport)
		showHelp(config)
		os.Exit(1)
	}
}

func showHelp(config *Config) {
	fmt.Printf("%s v%s - %s\n", config.Name, config.Version, config.Description)
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  dizi [flags]")
	fmt.Println("")
	fmt.Println("Flags:")
	fmt.Println("  -transport string")
	fmt.Println("        Transport method: stdio or sse (default \"stdio\")")
	fmt.Println("  -host string")
	fmt.Println("        Host for SSE transport (default \"localhost\")")
	fmt.Println("  -port int")
	fmt.Printf("        Port for SSE transport (default %d from config)\n", config.Server.Port)
	fmt.Println("  -help")
	fmt.Println("        Show this help information")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  dizi                           # Start with stdio transport")
	fmt.Println("  dizi -transport=sse            # Start with SSE transport on localhost:8080")
	fmt.Println("  dizi -transport=sse -port=9000 # Start with SSE transport on localhost:9000")
	fmt.Println("  dizi -help                     # Show this help")
}