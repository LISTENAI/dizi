// Package tools provides tool registration and execution for the MCP server.
// This file handles registration of builtin, command, and script type tools.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"dizi/internal/config"
	"dizi/internal/shell"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	libs "github.com/vadv/gopher-lua-libs"
	lua "github.com/yuin/gopher-lua"
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
		case "lua":
			handler = createLuaHandler(tool)
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
	return func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		switch tool.Name {
		case "echo":
			return handleEcho(request)
		case "lua_eval":
			return handleLuaEval(request)
		default:
			return mcp.NewToolResultError(fmt.Sprintf("Unknown builtin tool: %s", tool.Name)), nil
		}
	}
}

// createCommandHandler creates a handler for command tools
func createCommandHandler(tool config.ToolConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

		// Execute command with shell environment
		cmd := shell.CreateShellCommand(tool.Command, processedArgs...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Command failed: %v\nOutput: %s", err, string(output))), nil
		}

		return mcp.NewToolResultText(string(output)), nil
	}
}

// createScriptHandler creates a handler for script tools
func createScriptHandler(tool config.ToolConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract arguments
		arguments, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return mcp.NewToolResultError("Invalid arguments format"), nil
		}

		// Replace placeholders in script
		processedScript := replacePlaceholders(tool.Script, arguments)

		// Execute script with shell environment
		cmd := shell.CreateShellScriptCommand(processedScript)
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

// createLuaHandler creates a handler for lua tools
func createLuaHandler(tool config.ToolConfig) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract arguments
		arguments, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return mcp.NewToolResultError("Invalid arguments format"), nil
		}

		// Create Lua state
		L := lua.NewState()
		defer L.Close()
		
		// Load gopher-lua-libs
		libs.Preload(L)

		// Set arguments as global variables in Lua
		for key, value := range arguments {
			switch v := value.(type) {
			case string:
				L.SetGlobal(key, lua.LString(v))
			case float64:
				L.SetGlobal(key, lua.LNumber(v))
			case bool:
				L.SetGlobal(key, lua.LBool(v))
			default:
				L.SetGlobal(key, lua.LString(fmt.Sprintf("%v", v)))
			}
		}

		// Execute the Lua script from file
		if err := L.DoFile(tool.Script); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Lua script failed: %v", err)), nil
		}

		// Get the result from a global variable called 'result' if it exists
		result := L.GetGlobal("result")
		if result != lua.LNil {
			return mcp.NewToolResultText(result.String()), nil
		}

		return mcp.NewToolResultText("Lua script executed successfully"), nil
	}
}

// handleLuaEval handles the builtin lua_eval tool
func handleLuaEval(request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract arguments
	arguments, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments format"), nil
	}

	code, ok := arguments["code"].(string)
	if !ok {
		return mcp.NewToolResultError("Missing or invalid code parameter"), nil
	}

	// Create Lua state
	L := lua.NewState()
	defer L.Close()
	
	// Load gopher-lua-libs
	libs.Preload(L)

	// Capture print output
	var output strings.Builder
	L.SetGlobal("print", L.NewFunction(func(L *lua.LState) int {
		top := L.GetTop()
		for i := 1; i <= top; i++ {
			if i > 1 {
				output.WriteString("\t")
			}
			output.WriteString(L.Get(i).String())
		}
		output.WriteString("\n")
		return 0
	}))

	// First, check if it's a simple expression by trying "return " + code
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
				result := returnValue.String()
				if output.Len() > 0 {
					result = output.String() + "Return value: " + result
				}
				return mcp.NewToolResultText(result), nil
			}
		}
		L.SetTop(0) // Clear stack
	}

	// Reset output buffer for statement execution
	output.Reset()
	
	// Execute as statement
	if err := L.DoString(code); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Lua code failed: %v", err)), nil
	}

	// Check for result variable
	result := L.GetGlobal("result")
	var resultText string

	if result != lua.LNil {
		resultText = result.String()
	}

	// For complex statements, check if there's a return statement or try to get stack top
	var lastExpressionResult string
	if strings.Contains(code, "return ") {
		// If code contains return, the result should be on stack
		if L.GetTop() > 0 {
			returnValue := L.Get(-1)
			if returnValue != lua.LNil {
				lastExpressionResult = returnValue.String()
			}
		}
	} else if strings.Contains(code, ";") {
		// For other complex statements, try to evaluate the last expression
		parts := strings.Split(code, ";")
		lastPart := strings.TrimSpace(parts[len(parts)-1])
		if lastPart != "" && !strings.HasPrefix(lastPart, "print") && !strings.HasPrefix(lastPart, "local") {
			// Create a new Lua state to avoid interfering with the main one
			tempL := lua.NewState()
			defer tempL.Close()
			
			// Load gopher-lua-libs for the temporary state
			libs.Preload(tempL)
			
			// Copy necessary globals from main state
			if err := tempL.DoString(code); err != nil {
				// Ignore error - this is part of expression evaluation attempt
				_ = err
			}
			
			// Try to evaluate the last part as an expression
			err := tempL.DoString("return " + lastPart)
			if err == nil && tempL.GetTop() > 0 {
				returnValue := tempL.Get(-1)
				if returnValue != lua.LNil {
					lastExpressionResult = returnValue.String()
				}
			}
		}
	}

	// Combine all results
	var finalResult strings.Builder
	
	if output.Len() > 0 {
		finalResult.WriteString(strings.TrimSpace(output.String()))
	}
	
	if resultText != "" {
		if finalResult.Len() > 0 {
			finalResult.WriteString("\nResult: ")
		}
		finalResult.WriteString(resultText)
	}
	
	if lastExpressionResult != "" && resultText == "" {
		if finalResult.Len() > 0 {
			finalResult.WriteString("\nReturn value: ")
		}
		finalResult.WriteString(lastExpressionResult)
	}

	if finalResult.Len() > 0 {
		return mcp.NewToolResultText(finalResult.String()), nil
	}

	return mcp.NewToolResultText("Lua code executed successfully (no output)"), nil
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

