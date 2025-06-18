package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// registerTools registers all tools from configuration
func registerTools(mcpServer *server.MCPServer, toolConfigs []ToolConfig) error {
	for _, toolConfig := range toolConfigs {
		if err := registerTool(mcpServer, toolConfig); err != nil {
			return fmt.Errorf("failed to register tool %s: %w", toolConfig.Name, err)
		}
	}
	return nil
}

// registerTool registers a single tool based on its configuration
func registerTool(mcpServer *server.MCPServer, toolConfig ToolConfig) error {
	switch toolConfig.Type {
	case "builtin":
		return registerBuiltinTool(mcpServer, toolConfig)
	case "command":
		return registerCommandTool(mcpServer, toolConfig)
	case "script":
		return registerScriptTool(mcpServer, toolConfig)
	default:
		return fmt.Errorf("unsupported tool type: %s", toolConfig.Type)
	}
}

// registerBuiltinTool registers built-in tools
func registerBuiltinTool(mcpServer *server.MCPServer, toolConfig ToolConfig) error {
	switch toolConfig.Name {
	case "echo":
		tool := mcp.NewTool("echo",
			mcp.WithDescription(toolConfig.Description),
			mcp.WithString("message", mcp.Required(), mcp.Description("Message to echo back")),
		)
		
		mcpServer.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
		
		return nil
	default:
		return fmt.Errorf("unknown builtin tool: %s", toolConfig.Name)
	}
}

// registerCommandTool registers command-based tools
func registerCommandTool(mcpServer *server.MCPServer, toolConfig ToolConfig) error {
	if toolConfig.Command == "" {
		return fmt.Errorf("command is required for command-type tool")
	}
	
	// Create tool with dynamic schema if provided, otherwise create basic tool
	var tool mcp.Tool
	if toolConfig.Schema != nil {
		// Build tool options from schema
		toolOptions := []mcp.ToolOption{mcp.WithDescription(toolConfig.Description)}
		
		// Parse schema properties for parameters
		if schemaMap, ok := toolConfig.Schema["properties"].(map[string]interface{}); ok {
			required := make(map[string]bool)
			if reqArray, ok := toolConfig.Schema["required"].([]interface{}); ok {
				for _, req := range reqArray {
					if reqStr, ok := req.(string); ok {
						required[reqStr] = true
					}
				}
			}
			
			for propName, propDef := range schemaMap {
				if propMap, ok := propDef.(map[string]interface{}); ok {
					if propType, ok := propMap["type"].(string); ok {
						desc := ""
						if d, ok := propMap["description"].(string); ok {
							desc = d
						}
						
						var opts []mcp.PropertyOption
						if required[propName] {
							opts = append(opts, mcp.Required())
						}
						if desc != "" {
							opts = append(opts, mcp.Description(desc))
						}
						
						switch propType {
						case "string":
							toolOptions = append(toolOptions, mcp.WithString(propName, opts...))
						case "number", "integer":
							toolOptions = append(toolOptions, mcp.WithNumber(propName, opts...))
						case "boolean":
							toolOptions = append(toolOptions, mcp.WithBoolean(propName, opts...))
						}
					}
				}
			}
		}
		
		tool = mcp.NewTool(toolConfig.Name, toolOptions...)
	} else {
		tool = mcp.NewTool(toolConfig.Name, mcp.WithDescription(toolConfig.Description))
	}
	
	mcpServer.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return executeCommand(toolConfig, request)
	})
	
	return nil
}

// registerScriptTool registers script-based tools
func registerScriptTool(mcpServer *server.MCPServer, toolConfig ToolConfig) error {
	if toolConfig.Script == "" {
		return fmt.Errorf("script is required for script-type tool")
	}
	
	tool := mcp.NewTool(toolConfig.Name, mcp.WithDescription(toolConfig.Description))
	
	mcpServer.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return executeScript(toolConfig, request)
	})
	
	return nil
}

// executeCommand executes a command-based tool
func executeCommand(toolConfig ToolConfig, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Prepare command arguments
	args := make([]string, len(toolConfig.Args))
	copy(args, toolConfig.Args)
	
	// Replace placeholders in arguments with actual values from request
	if arguments, ok := request.Params.Arguments.(map[string]interface{}); ok {
		for i := range args {
			// Simple placeholder replacement: {{param_name}}
			for key, value := range arguments {
				placeholder := fmt.Sprintf("{{%s}}", key)
				if valueStr, ok := value.(string); ok {
					args[i] = strings.ReplaceAll(args[i], placeholder, valueStr)
				}
			}
		}
	}
	
	// Execute command
	cmd := exec.Command(toolConfig.Command, args...)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Command failed: %v\nOutput: %s", err, string(output))), nil
	}
	
	return mcp.NewToolResultText(string(output)), nil
}

// executeScript executes a script-based tool
func executeScript(toolConfig ToolConfig, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// For now, treat script similar to command
	// In the future, this could support inline scripts, etc.
	cmd := exec.Command("sh", "-c", toolConfig.Script)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Script failed: %v\nOutput: %s", err, string(output))), nil
	}
	
	return mcp.NewToolResultText(string(output)), nil
}