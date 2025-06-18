// Package tools provides tool registration and execution for the MCP server.
// This file handles registration of builtin, command, and script type tools.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"dizi/internal/config"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterTools registers all tools from the configuration
func RegisterTools(mcpServer *server.MCPServer, tools []config.ToolConfig) error {
	for _, tool := range tools {
		// Marshal the parameters to JSON
		var schemaBytes []byte
		var err error
		if tool.Parameters != nil {
			schemaBytes, err = json.Marshal(tool.Parameters)
			if err != nil {
				return fmt.Errorf("failed to marshal parameters for tool %s: %w", tool.Name, err)
			}
		} else {
			// Default empty object schema
			schemaBytes = []byte(`{"type": "object", "properties": {}}`)
		}

		// Create MCP tool with raw schema
		mcpTool := mcp.NewToolWithRawSchema(tool.Name, tool.Description, json.RawMessage(schemaBytes))

		// Create handler based on tool type
		var handler func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

		switch tool.Type {
		case "builtin":
			handler = createBuiltinHandler(tool)
		case "command":
			handler = createCommandHandler(tool)
		case "script":
			handler = createScriptHandler(tool)
		default:
			return fmt.Errorf("unsupported tool type: %s for tool %s", tool.Type, tool.Name)
		}

		// Register the tool
		mcpServer.AddTool(mcpTool, handler)
	}

	return nil
}

// createBuiltinHandler creates a handler for builtin tools
func createBuiltinHandler(tool config.ToolConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		switch tool.Name {
		case "echo":
			return handleEcho(request)
		default:
			return mcp.NewToolResultError(fmt.Sprintf("Unknown builtin tool: %s", tool.Name)), nil
		}
	}
}

// createCommandHandler creates a handler for command tools
func createCommandHandler(tool config.ToolConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract arguments
		arguments, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return mcp.NewToolResultError("Invalid arguments format"), nil
		}

		// Replace placeholders in args
		processedArgs := make([]string, len(tool.Args))
		for i, arg := range tool.Args {
			processedArgs[i] = replacePlaceholders(arg, arguments)
		}

		// Execute command
		cmd := exec.Command(tool.Command, processedArgs...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Command failed: %v\nOutput: %s", err, string(output))), nil
		}

		return mcp.NewToolResultText(string(output)), nil
	}
}

// createScriptHandler creates a handler for script tools
func createScriptHandler(tool config.ToolConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract arguments
		arguments, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return mcp.NewToolResultError("Invalid arguments format"), nil
		}

		// Replace placeholders in script
		processedScript := replacePlaceholders(tool.Script, arguments)

		// Execute script with shell
		cmd := exec.Command("sh", "-c", processedScript)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Script failed: %v\nOutput: %s", err, string(output))), nil
		}

		return mcp.NewToolResultText(string(output)), nil
	}
}

// handleEcho handles the builtin echo tool
func handleEcho(request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract arguments
	arguments, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments format"), nil
	}

	message, ok := arguments["message"].(string)
	if !ok {
		return mcp.NewToolResultError("Missing or invalid message parameter"), nil
	}

	return mcp.NewToolResultText(message), nil
}

// replacePlaceholders replaces {{parameter_name}} placeholders with actual values
func replacePlaceholders(text string, arguments map[string]interface{}) string {
	result := text
	for key, value := range arguments {
		placeholder := fmt.Sprintf("{{%s}}", key)
		replacement := fmt.Sprintf("%v", value)
		result = strings.ReplaceAll(result, placeholder, replacement)
	}
	return result
}

