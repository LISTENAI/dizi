package main

import (
	"context"
	"encoding/json"
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
		var tool mcp.Tool
		if toolConfig.Parameters != nil {
			// Use raw schema for builtin tools with parameters
			schemaBytes, err := json.Marshal(toolConfig.Parameters)
			if err != nil {
				return fmt.Errorf("failed to marshal tool parameters: %w", err)
			}
			tool = mcp.NewToolWithRawSchema(toolConfig.Name, toolConfig.Description, json.RawMessage(schemaBytes))
		} else {
			// Fallback to basic tool
			tool = mcp.NewTool(toolConfig.Name, mcp.WithDescription(toolConfig.Description))
		}
		
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
	
	// Create tool with inputSchema using raw JSON
	var tool mcp.Tool
	if toolConfig.Parameters != nil {
		schemaBytes, err := json.Marshal(toolConfig.Parameters)
		if err != nil {
			return fmt.Errorf("failed to marshal tool parameters: %w", err)
		}
		tool = mcp.NewToolWithRawSchema(toolConfig.Name, toolConfig.Description, json.RawMessage(schemaBytes))
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
	
	// Create tool with inputSchema if parameters exist
	var tool mcp.Tool
	if toolConfig.Parameters != nil {
		schemaBytes, err := json.Marshal(toolConfig.Parameters)
		if err != nil {
			return fmt.Errorf("failed to marshal tool parameters: %w", err)
		}
		tool = mcp.NewToolWithRawSchema(toolConfig.Name, toolConfig.Description, json.RawMessage(schemaBytes))
	} else {
		tool = mcp.NewTool(toolConfig.Name, mcp.WithDescription(toolConfig.Description))
	}
	
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
	
	// Debug logging (will be discarded in stdio mode)
	infoLog("Executing tool name: %s, command: %s with args: %v", toolConfig.Name, toolConfig.Command, args)
	
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
	// Replace placeholders in script with actual values from request
	script := toolConfig.Script
	if arguments, ok := request.Params.Arguments.(map[string]interface{}); ok {
		for key, value := range arguments {
			placeholder := fmt.Sprintf("{{%s}}", key)
			if valueStr, ok := value.(string); ok {
				script = strings.ReplaceAll(script, placeholder, valueStr)
			}
		}
	}
	
	// Debug logging (will be discarded in stdio mode)
	infoLog("Executing tool name: %s, script: %s", toolConfig.Name, script)
	
	// Execute script
	cmd := exec.Command("sh", "-c", script)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Script failed: %v\nOutput: %s", err, string(output))), nil
	}
	
	return mcp.NewToolResultText(string(output)), nil
}