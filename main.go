package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	var (
		transport = flag.String("transport", "stdio", "Transport method: stdio or sse")
		host      = flag.String("host", "localhost", "Host for SSE transport")
		port      = flag.Int("port", 8080, "Port for SSE transport")
		help      = flag.Bool("help", false, "Show help information")
	)
	
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// Create MCP server
	mcpServer := server.NewMCPServer("dizi", "1.0.0")

	// Add example echo tool
	echoTool := mcp.NewTool("echo",
		mcp.WithDescription("Echo back the input message"),
		mcp.WithString("message", mcp.Required(), mcp.Description("Message to echo back")),
	)

	mcpServer.AddTool(echoTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		arguments, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return mcp.NewToolResultError("Invalid arguments format"), nil
		}
		
		message, ok := arguments["message"].(string)
		if !ok {
			return mcp.NewToolResultError("Invalid message parameter"), nil
		}
		
		return mcp.NewToolResultText(fmt.Sprintf("Echo: %s", message)), nil
	})

	// Start server based on transport
	switch *transport {
	case "stdio":
		log.Println("Starting MCP server with stdio transport...")
		if err := server.ServeStdio(mcpServer); err != nil {
			log.Fatalf("Failed to start stdio server: %v", err)
		}
	case "sse":
		log.Printf("Starting MCP server with SSE transport on %s:%d...", *host, *port)
		// Create SSE server
		sseServer := server.NewSSEServer(mcpServer)
		if err := sseServer.Start(fmt.Sprintf("%s:%d", *host, *port)); err != nil {
			log.Fatalf("Failed to start SSE server: %v", err)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unsupported transport: %s\n", *transport)
		showHelp()
		os.Exit(1)
	}
}

func showHelp() {
	fmt.Println("Dizi - MCP Server")
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
	fmt.Println("        Port for SSE transport (default 8080)")
	fmt.Println("  -help")
	fmt.Println("        Show this help information")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  dizi                           # Start with stdio transport")
	fmt.Println("  dizi -transport=sse            # Start with SSE transport on localhost:8080")
	fmt.Println("  dizi -transport=sse -port=9000 # Start with SSE transport on localhost:9000")
	fmt.Println("  dizi -help                     # Show this help")
}