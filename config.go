// Package main provides configuration management for the MCP server.
// This file handles loading and parsing YAML configuration files.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the dizi.yml configuration structure
type Config struct {
	Name        string      `yaml:"name"`
	Version     string      `yaml:"version"`
	Description string      `yaml:"description"`
	Server      ServerConfig `yaml:"server"`
	Tools       []ToolConfig `yaml:"tools"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Port int `yaml:"port"`
}

// ToolConfig represents a tool configuration
type ToolConfig struct {
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description"`
	Type        string                 `yaml:"type"` // "command", "script", etc.
	Command     string                 `yaml:"command,omitempty"`
	Script      string                 `yaml:"script,omitempty"`
	Args        []string               `yaml:"args,omitempty"`
	Parameters  map[string]interface{} `yaml:"parameters,omitempty"`
}

// LoadConfig loads configuration from dizi.yml in the current directory
func LoadConfig() (*Config, error) {
	configPath := filepath.Join(".", "dizi.yml")
	
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return getDefaultConfig(), nil
	}
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	
	// Set defaults if not specified
	if config.Name == "" {
		config.Name = "dizi"
	}
	if config.Version == "" {
		config.Version = "1.0.0"
	}
	if config.Description == "" {
		config.Description = "MCP Server"
	}
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}
	
	return &config, nil
}

// getDefaultConfig returns a default configuration
func getDefaultConfig() *Config {
	return &Config{
		Name:        "dizi",
		Version:     "1.0.0",
		Description: "MCP Server",
		Server: ServerConfig{
			Port: 8080,
		},
		Tools: []ToolConfig{
			{
				Name:        "echo",
				Description: "Echo back the input message",
				Type:        "builtin",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"message": map[string]interface{}{
							"type":        "string",
							"description": "Message to echo back",
						},
					},
					"required": []string{"message"},
				},
			},
		},
	}
}