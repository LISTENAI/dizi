package tools

import (
	"context"
	"strings"
	"testing"

	"dizi/internal/config"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func TestRegisterTools(t *testing.T) {
	mcpServer := server.NewMCPServer("test", "1.0.0")
	
	tools := []config.ToolConfig{
		{
			Name:        "echo",
			Description: "Echo tool",
			Type:        "builtin",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"message": map[string]interface{}{
						"type":        "string",
						"description": "Message to echo",
					},
				},
				"required": []string{"message"},
			},
		},
		{
			Name:        "test_command",
			Description: "Test command",
			Type:        "command",
			Command:     "echo",
			Args:        []string{"Hello", "{{name}}"},
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name parameter",
					},
				},
				"required": []string{"name"},
			},
		},
		{
			Name:        "test_script",
			Description: "Test script",
			Type:        "script",
			Script:      "echo 'Hello {{name}}'",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Name parameter",
					},
				},
				"required": []string{"name"},
			},
		},
	}
	
	err := RegisterTools(mcpServer, tools)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Verify tools were registered (we can't easily inspect the server's internal state,
	// but we can test that no error occurred during registration)
}

func TestRegisterToolsUnsupportedType(t *testing.T) {
	mcpServer := server.NewMCPServer("test", "1.0.0")
	
	tools := []config.ToolConfig{
		{
			Name:        "invalid_tool",
			Description: "Invalid tool",
			Type:        "unsupported_type",
		},
	}
	
	err := RegisterTools(mcpServer, tools)
	if err == nil {
		t.Error("Expected error for unsupported tool type, got nil")
	}
	
	if !strings.Contains(err.Error(), "unsupported tool type: unsupported_type") {
		t.Errorf("Expected error to contain 'unsupported tool type: unsupported_type', got '%s'", err.Error())
	}
}

func TestRegisterToolsInvalidParameters(t *testing.T) {
	mcpServer := server.NewMCPServer("test", "1.0.0")
	
	// Create parameters that can't be marshaled to JSON
	invalidParams := map[string]interface{}{
		"invalid": make(chan int), // channels can't be marshaled to JSON
	}
	
	tools := []config.ToolConfig{
		{
			Name:        "invalid_params_tool",
			Description: "Tool with invalid parameters",
			Type:        "builtin",
			Parameters:  invalidParams,
		},
	}
	
	err := RegisterTools(mcpServer, tools)
	if err == nil {
		t.Error("Expected error for invalid parameters, got nil")
	}
}

func TestHandleEcho(t *testing.T) {
	tests := []struct {
		name        string
		arguments   map[string]interface{}
		expectError bool
		expectedMsg string
	}{
		{
			name: "valid message",
			arguments: map[string]interface{}{
				"message": "Hello, World!",
			},
			expectError: false,
			expectedMsg: "Hello, World!",
		},
		{
			name: "empty message",
			arguments: map[string]interface{}{
				"message": "",
			},
			expectError: false,
			expectedMsg: "",
		},
		{
			name: "numeric message",
			arguments: map[string]interface{}{
				"message": 123,
			},
			expectError: true,
		},
		{
			name:        "missing message",
			arguments:   map[string]interface{}{},
			expectError: true,
		},
		{
			name:        "nil arguments",
			arguments:   nil,
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: tt.arguments,
				},
			}
			
			result, err := handleEcho(request)
			
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			if tt.expectError {
				if !result.IsError {
					t.Error("Expected error result, got success")
				}
			} else {
				if result.IsError {
					t.Errorf("Expected success, got error: %v", result.Content)
				}
				
				if len(result.Content) != 1 {
					t.Fatalf("Expected 1 content item, got %d", len(result.Content))
				}
				
				content := result.Content[0]
				if textContent, ok := content.(mcp.TextContent); ok {
					if textContent.Type != "text" {
						t.Errorf("Expected content type 'text', got '%s'", textContent.Type)
					}
					
					if textContent.Text != tt.expectedMsg {
						t.Errorf("Expected message '%s', got '%s'", tt.expectedMsg, textContent.Text)
					}
				} else {
					t.Error("Expected TextContent")
				}
			}
		})
	}
}

func TestCreateBuiltinHandler(t *testing.T) {
	tool := config.ToolConfig{
		Name: "echo",
		Type: "builtin",
	}
	
	handler := createBuiltinHandler(tool)
	if handler == nil {
		t.Error("Expected handler function, got nil")
	}
	
	// Test with unknown builtin tool
	unknownTool := config.ToolConfig{
		Name: "unknown_builtin",
		Type: "builtin",
	}
	
	unknownHandler := createBuiltinHandler(unknownTool)
	
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{},
		},
	}
	
	result, err := unknownHandler(context.Background(), request)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if !result.IsError {
		t.Error("Expected error for unknown builtin tool")
	}
}

func TestCreateCommandHandler(t *testing.T) {
	tool := config.ToolConfig{
		Name:    "test_command",
		Type:    "command",
		Command: "echo",
		Args:    []string{"Hello", "{{name}}"},
	}
	
	handler := createCommandHandler(tool)
	if handler == nil {
		t.Error("Expected handler function, got nil")
	}
	
	// Test valid execution
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"name": "World",
			},
		},
	}
	
	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if result.IsError {
		t.Errorf("Expected success, got error: %v", result.Content)
	}
}

func TestCreateScriptHandler(t *testing.T) {
	tool := config.ToolConfig{
		Name:   "test_script",
		Type:   "script",
		Script: "echo 'Hello {{name}}'",
	}
	
	handler := createScriptHandler(tool)
	if handler == nil {
		t.Error("Expected handler function, got nil")
	}
	
	// Test valid execution
	request := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]interface{}{
				"name": "World",
			},
		},
	}
	
	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if result.IsError {
		t.Errorf("Expected success, got error: %v", result.Content)
	}
}

func TestReplacePlaceholders(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		arguments map[string]interface{}
		expected  string
	}{
		{
			name: "single placeholder",
			text: "Hello {{name}}",
			arguments: map[string]interface{}{
				"name": "World",
			},
			expected: "Hello World",
		},
		{
			name: "multiple placeholders",
			text: "{{greeting}} {{name}}! You are {{age}} years old.",
			arguments: map[string]interface{}{
				"greeting": "Hello",
				"name":     "Alice",
				"age":      30,
			},
			expected: "Hello Alice! You are 30 years old.",
		},
		{
			name:      "no placeholders",
			text:      "Static text with no replacements",
			arguments: map[string]interface{}{},
			expected:  "Static text with no replacements",
		},
		{
			name: "unused arguments",
			text: "Hello {{name}}",
			arguments: map[string]interface{}{
				"name":   "World",
				"unused": "value",
			},
			expected: "Hello World",
		},
		{
			name: "missing arguments",
			text: "Hello {{name}} and {{missing}}",
			arguments: map[string]interface{}{
				"name": "World",
			},
			expected: "Hello World and {{missing}}",
		},
		{
			name:      "empty arguments",
			text:      "{{empty}} placeholder",
			arguments: map[string]interface{}{},
			expected:  "{{empty}} placeholder",
		},
		{
			name: "numeric values",
			text: "Value: {{number}}",
			arguments: map[string]interface{}{
				"number": 42,
			},
			expected: "Value: 42",
		},
		{
			name: "boolean values",
			text: "Enabled: {{flag}}",
			arguments: map[string]interface{}{
				"flag": true,
			},
			expected: "Enabled: true",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := replacePlaceholders(tt.text, tt.arguments)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestRegisterToolsEmptyList(t *testing.T) {
	mcpServer := server.NewMCPServer("test", "1.0.0")
	
	err := RegisterTools(mcpServer, []config.ToolConfig{})
	if err != nil {
		t.Errorf("Expected no error for empty tools list, got %v", err)
	}
}

func TestRegisterToolsNilParameters(t *testing.T) {
	mcpServer := server.NewMCPServer("test", "1.0.0")
	
	tools := []config.ToolConfig{
		{
			Name:        "no_params_tool",
			Description: "Tool with no parameters",
			Type:        "builtin",
			Parameters:  nil, // Explicitly nil
		},
	}
	
	err := RegisterTools(mcpServer, tools)
	if err != nil {
		t.Errorf("Expected no error for nil parameters, got %v", err)
	}
}