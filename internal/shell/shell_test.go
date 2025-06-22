package shell

import (
	"os"
	"runtime"
	"strings"
	"testing"
)

func TestGetCurrentShell(t *testing.T) {
	shell := getCurrentShell()
	if shell == "" {
		t.Error("getCurrentShell returned empty string")
	}
	
	// Should return a valid path
	if !strings.Contains(shell, "/") && runtime.GOOS != "windows" {
		t.Errorf("getCurrentShell returned invalid path: %s", shell)
	}
}

func TestGetShellConfigFiles(t *testing.T) {
	configFiles := GetShellConfigFiles()
	
	// Should return at least some config files
	if len(configFiles) == 0 && runtime.GOOS != "windows" {
		t.Error("GetShellConfigFiles returned no config files")
	}
	
	// All returned files should exist
	for _, file := range configFiles {
		if _, err := os.Stat(file); err != nil {
			t.Errorf("Config file does not exist: %s", file)
		}
	}
}

func TestCreateShellCommand(t *testing.T) {
	cmd := CreateShellCommand("echo", "test")
	if cmd == nil {
		t.Error("CreateShellCommand returned nil")
	}
	
	// Test that command can be executed
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("Command execution failed: %v, output: %s", err, string(output))
	}
	
	if !strings.Contains(string(output), "test") {
		t.Errorf("Command output doesn't contain expected text: %s", string(output))
	}
}

func TestCreateShellScriptCommand(t *testing.T) {
	script := "echo 'script test'"
	cmd := CreateShellScriptCommand(script)
	if cmd == nil {
		t.Error("CreateShellScriptCommand returned nil")
	}
	
	// Test that script can be executed
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("Script execution failed: %v, output: %s", err, string(output))
	}
	
	if !strings.Contains(string(output), "script test") {
		t.Errorf("Script output doesn't contain expected text: %s", string(output))
	}
}