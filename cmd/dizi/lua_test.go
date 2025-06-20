package main

import (
	"os"
	"strings"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

// TestExecuteLuaREPL tests the Lua REPL execution function
func TestExecuteLuaREPL(t *testing.T) {

	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name:     "Simple expression",
			code:     "2 + 3",
			expected: "=> 5",
		},
		{
			name:     "String concatenation",
			code:     "'Hello ' .. 'World'",
			expected: "=> Hello World",
		},
		{
			name:     "Variable assignment",
			code:     "x = 42",
			expected: "",
		},
		{
			name:     "Result variable",
			code:     "result = 'test result'",
			expected: "result = test result",
		},
		{
			name:     "Function definition",
			code:     "function double(x) return x * 2 end",
			expected: "",
		},
		{
			name:     "Math operation",
			code:     "math.sqrt(16)",
			expected: "=> 4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh Lua state for each test
			testL := lua.NewState()
			defer testL.Close()
			result := executeLuaREPL(testL, tt.code)
			if result != tt.expected {
				t.Errorf("executeLuaREPL(%q) = %q, want %q", tt.code, result, tt.expected)
			}
		})
	}
}

// TestExecuteLuaREPLErrors tests error handling in Lua REPL
func TestExecuteLuaREPLErrors(t *testing.T) {

	tests := []struct {
		name string
		code string
	}{
		{
			name: "Syntax error",
			code: "function invalid syntax",
		},
		{
			name: "Runtime error",
			code: "error('test error')",
		},
		{
			name: "Undefined variable",
			code: "undefined_variable + 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh Lua state for each test
			testL := lua.NewState()
			defer testL.Close()
			result := executeLuaREPL(testL, tt.code)
			if !strings.HasPrefix(result, "Error:") {
				t.Errorf("executeLuaREPL(%q) should return error, got %q", tt.code, result)
			}
		})
	}
}

// TestSetupLuaHelpers tests the Lua helpers setup
func TestSetupLuaHelpers(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	// Setup helpers
	setupLuaHelpers(L)

	// Check if help function is available
	helpFunc := L.GetGlobal("help")
	if helpFunc == lua.LNil {
		t.Error("help function should be available after setupLuaHelpers")
	}

	// Check if vars function is available
	varsFunc := L.GetGlobal("vars")
	if varsFunc == lua.LNil {
		t.Error("vars function should be available after setupLuaHelpers")
	}

	// Test help function execution
	err := L.DoString("help()")
	if err != nil {
		t.Errorf("help() function should execute without error, got: %v", err)
	}

	// Test vars function execution
	err = L.DoString("vars()")
	if err != nil {
		t.Errorf("vars() function should execute without error, got: %v", err)
	}
}

// TestLuaScriptExecution tests embedded Lua script execution
func TestLuaScriptExecution(t *testing.T) {
	// Test if embedded scripts are not empty
	if helloLuaScript == "" {
		t.Error("helloLuaScript should not be empty")
	}
	if systemInfoLuaScript == "" {
		t.Error("systemInfoLuaScript should not be empty")
	}
	if fileUtilsLuaScript == "" {
		t.Error("fileUtilsLuaScript should not be empty")
	}
	if helpersLuaScript == "" {
		t.Error("helpersLuaScript should not be empty")
	}

	// Test execution of embedded scripts
	L := lua.NewState()
	defer L.Close()

	// Test hello script
	err := L.DoString(helloLuaScript)
	if err != nil {
		t.Errorf("helloLuaScript should execute without error, got: %v", err)
	}

	// Test system info script
	L.SetTop(0) // Clear stack
	err = L.DoString(systemInfoLuaScript)
	if err != nil {
		t.Errorf("systemInfoLuaScript should execute without error, got: %v", err)
	}

	// Check if result is set
	result := L.GetGlobal("result")
	if result == lua.LNil {
		t.Error("systemInfoLuaScript should set result variable")
	}
}

// TestLuaCommand tests the lua command functionality
func TestLuaCommand(t *testing.T) {
	// Create a temporary Lua script
	tempScript := `
result = "Test script executed successfully"
print("This is a test script")
`

	tmpFile, err := os.CreateTemp("", "test_script_*.lua")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	_, err = tmpFile.WriteString(tempScript)
	if err != nil {
		t.Fatalf("Failed to write to temporary file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temporary file: %v", err)
	}

	// Test file existence check (this would be part of luaCommand)
	if _, err := os.Stat(tmpFile.Name()); os.IsNotExist(err) {
		t.Errorf("Temporary script file should exist")
	}

	// Test Lua execution
	L := lua.NewState()
	defer L.Close()

	err = L.DoFile(tmpFile.Name())
	if err != nil {
		t.Errorf("Lua script should execute without error, got: %v", err)
	}

	// Check result
	result := L.GetGlobal("result")
	if result == lua.LNil {
		t.Error("Script should set result variable")
	}
}