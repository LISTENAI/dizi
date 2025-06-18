package config

import (
	"os"
	"testing"
)

func TestLoadDefaultConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	
	// Change to temp directory where no dizi.yml exists
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)
	
	config, err := Load()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Check default values
	if config.Name != "dizi" {
		t.Errorf("Expected name 'dizi', got '%s'", config.Name)
	}
	if config.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", config.Version)
	}
	if config.Description != "MCP Server" {
		t.Errorf("Expected description 'MCP Server', got '%s'", config.Description)
	}
	if config.Server.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", config.Server.Port)
	}
	
	// Check default tools
	if len(config.Tools) != 1 {
		t.Errorf("Expected 1 default tool, got %d", len(config.Tools))
	}
	if config.Tools[0].Name != "echo" {
		t.Errorf("Expected default tool name 'echo', got '%s'", config.Tools[0].Name)
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	
	// Change to temp directory
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)
	
	// Create a test config file
	configContent := `name: "test-server"
version: "2.0.0"
description: "Test MCP Server"
server:
  port: 9999
tools:
  - name: "test_tool"
    description: "A test tool"
    type: "builtin"
    parameters:
      type: "object"
      properties:
        param1:
          type: "string"
          description: "Test parameter"
      required: ["param1"]
`
	
	err := os.WriteFile("dizi.yml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	
	config, err := Load()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Check loaded values
	if config.Name != "test-server" {
		t.Errorf("Expected name 'test-server', got '%s'", config.Name)
	}
	if config.Version != "2.0.0" {
		t.Errorf("Expected version '2.0.0', got '%s'", config.Version)
	}
	if config.Description != "Test MCP Server" {
		t.Errorf("Expected description 'Test MCP Server', got '%s'", config.Description)
	}
	if config.Server.Port != 9999 {
		t.Errorf("Expected port 9999, got %d", config.Server.Port)
	}
	
	// Check tools
	if len(config.Tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(config.Tools))
	}
	if config.Tools[0].Name != "test_tool" {
		t.Errorf("Expected tool name 'test_tool', got '%s'", config.Tools[0].Name)
	}
	if config.Tools[0].Type != "builtin" {
		t.Errorf("Expected tool type 'builtin', got '%s'", config.Tools[0].Type)
	}
}

func TestLoadConfigWithPartialValues(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	
	// Change to temp directory
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)
	
	// Create a config file with only some values
	configContent := `name: "partial-server"
tools:
  - name: "custom_tool"
    description: "Custom tool"
    type: "command"
    command: "echo"
    args: ["hello"]
`
	
	err := os.WriteFile("dizi.yml", []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	
	config, err := Load()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// Check that defaults are applied for missing values
	if config.Name != "partial-server" {
		t.Errorf("Expected name 'partial-server', got '%s'", config.Name)
	}
	if config.Version != "1.0.0" { // Should be default
		t.Errorf("Expected default version '1.0.0', got '%s'", config.Version)
	}
	if config.Description != "MCP Server" { // Should be default
		t.Errorf("Expected default description 'MCP Server', got '%s'", config.Description)
	}
	if config.Server.Port != 8080 { // Should be default
		t.Errorf("Expected default port 8080, got %d", config.Server.Port)
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	
	// Change to temp directory
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)
	
	// Create an invalid YAML file
	invalidYAML := `name: "test
invalid yaml content
`
	
	err := os.WriteFile("dizi.yml", []byte(invalidYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	
	_, err = Load()
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

func TestLoadConfigFileReadError(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	
	// Change to temp directory
	os.Chdir(tempDir)
	defer os.Chdir(originalWd)
	
	// Create a directory with the config filename (should cause read error)
	err := os.Mkdir("dizi.yml", 0755)
	if err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	
	_, err = Load()
	if err == nil {
		t.Error("Expected error for directory instead of file, got nil")
	}
}

func TestGetDefaultConfig(t *testing.T) {
	config := getDefaultConfig()
	
	// Check all default values
	if config.Name != "dizi" {
		t.Errorf("Expected default name 'dizi', got '%s'", config.Name)
	}
	if config.Version != "1.0.0" {
		t.Errorf("Expected default version '1.0.0', got '%s'", config.Version)
	}
	if config.Description != "MCP Server" {
		t.Errorf("Expected default description 'MCP Server', got '%s'", config.Description)
	}
	if config.Server.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", config.Server.Port)
	}
	
	// Check default tools
	if len(config.Tools) != 1 {
		t.Errorf("Expected 1 default tool, got %d", len(config.Tools))
	}
	
	tool := config.Tools[0]
	if tool.Name != "echo" {
		t.Errorf("Expected default tool name 'echo', got '%s'", tool.Name)
	}
	if tool.Type != "builtin" {
		t.Errorf("Expected default tool type 'builtin', got '%s'", tool.Type)
	}
	if tool.Description != "Echo back the input message" {
		t.Errorf("Expected specific description, got '%s'", tool.Description)
	}
	
	// Check parameters structure
	if tool.Parameters == nil {
		t.Error("Expected parameters to be set")
	}
	
	paramType, ok := tool.Parameters["type"].(string)
	if !ok || paramType != "object" {
		t.Errorf("Expected parameters type 'object', got %v", paramType)
	}
	
	required, ok := tool.Parameters["required"].([]string)
	if !ok || len(required) != 1 || required[0] != "message" {
		t.Errorf("Expected required parameters ['message'], got %v", required)
	}
}