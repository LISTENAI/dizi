package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func TestNewFilesystemServer(t *testing.T) {
	tests := []struct {
		name   string
		config *FilesystemConfig
	}{
		{
			name: "with config",
			config: &FilesystemConfig{
				RootDirectory: "/tmp",
			},
		},
		{
			name:   "with nil config",
			config: nil,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := NewFilesystemServer(tt.config)
			
			if fs == nil {
				t.Error("Expected FilesystemServer, got nil")
				return
			}
			
			if fs.config == nil {
				t.Error("Expected config to be set")
			}
			
			if tt.config == nil {
				// Should have set default config
				if fs.config.RootDirectory == "" {
					t.Error("Expected default root directory to be set")
				}
			} else {
				if fs.config.RootDirectory != tt.config.RootDirectory {
					t.Errorf("Expected root directory '%s', got '%s'", tt.config.RootDirectory, fs.config.RootDirectory)
				}
			}
		})
	}
}

func TestValidatePath(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	
	fs := NewFilesystemServer(&FilesystemConfig{
		RootDirectory: tempDir,
	})
	
	tests := []struct {
		name        string
		path        string
		expectError bool
	}{
		{
			name:        "valid absolute path within root",
			path:        filepath.Join(tempDir, "test.txt"),
			expectError: false,
		},
		{
			name:        "valid subdirectory path",
			path:        filepath.Join(tempDir, "subdir", "test.txt"),
			expectError: false,
		},
		{
			name:        "root directory itself",
			path:        tempDir,
			expectError: false,
		},
		{
			name:        "absolute path outside root",
			path:        "/etc/passwd",
			expectError: true,
		},
		{
			name:        "path traversal attempt",
			path:        filepath.Join(tempDir, "..", "outside.txt"),
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validPath, err := fs.validatePath(tt.path)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for path '%s', got nil", tt.path)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for path '%s', got %v", tt.path, err)
				}
				
				if validPath == "" {
					t.Errorf("Expected valid path to be returned for '%s'", tt.path)
				}
				
				// Ensure the validated path is absolute
				if !filepath.IsAbs(validPath) {
					t.Errorf("Expected absolute path, got '%s'", validPath)
				}
			}
		})
	}
}

func TestRegisterFilesystemTools(t *testing.T) {
	tempDir := t.TempDir()
	
	mcpServer := server.NewMCPServer("test", "1.0.0")
	config := &FilesystemConfig{
		RootDirectory: tempDir,
	}
	
	err := RegisterFilesystemTools(mcpServer, config)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	// The function should complete without error
	// We can't easily inspect the registered tools, but we test that registration succeeds
}

func TestHandleReadFile(t *testing.T) {
	// Create a temporary directory and file for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "Hello, World!"
	
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	fs := NewFilesystemServer(&FilesystemConfig{
		RootDirectory: tempDir,
	})
	
	tests := []struct {
		name        string
		arguments   map[string]interface{}
		expectError bool
	}{
		{
			name: "valid file",
			arguments: map[string]interface{}{
				"path": testFile, // Use absolute path
			},
			expectError: false,
		},
		{
			name: "nonexistent file",
			arguments: map[string]interface{}{
				"path": filepath.Join(tempDir, "nonexistent.txt"),
			},
			expectError: true,
		},
		{
			name: "path outside root",
			arguments: map[string]interface{}{
				"path": "/etc/passwd",
			},
			expectError: true,
		},
		{
			name:        "missing path parameter",
			arguments:   map[string]interface{}{},
			expectError: true,
		},
		{
			name: "invalid path type",
			arguments: map[string]interface{}{
				"path": 123,
			},
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
			
			result, err := fs.handleReadProjectFile(context.Background(), request)
			
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
					
					// Content should include the test content
					if !strings.Contains(textContent.Text, testContent) {
						t.Errorf("Expected content to contain '%s', got '%s'", testContent, textContent.Text)
					}
				} else {
					t.Error("Expected TextContent")
				}
			}
		})
	}
}

func TestHandleWriteFile(t *testing.T) {
	tempDir := t.TempDir()
	
	fs := NewFilesystemServer(&FilesystemConfig{
		RootDirectory: tempDir,
	})
	
	tests := []struct {
		name        string
		arguments   map[string]interface{}
		expectError bool
	}{
		{
			name: "valid write",
			arguments: map[string]interface{}{
				"path":    filepath.Join(tempDir, "newfile.txt"),
				"content": "Hello, World!",
			},
			expectError: false,
		},
		{
			name: "write to subdirectory",
			arguments: map[string]interface{}{
				"path":    filepath.Join(tempDir, "subdir", "newfile.txt"),
				"content": "Subdirectory content",
			},
			expectError: false,
		},
		{
			name: "path outside root",
			arguments: map[string]interface{}{
				"path":    "/tmp/outside.txt",
				"content": "Should not work",
			},
			expectError: true,
		},
		{
			name: "missing path parameter",
			arguments: map[string]interface{}{
				"content": "No path",
			},
			expectError: true,
		},
		{
			name: "missing content parameter",
			arguments: map[string]interface{}{
				"path": filepath.Join(tempDir, "nocontent.txt"),
			},
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
			
			result, err := fs.handleWriteProjectFile(context.Background(), request)
			
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
				
				// Verify file was actually written
				if path, ok := tt.arguments["path"].(string); ok {
					if content, ok := tt.arguments["content"].(string); ok {
						writtenContent, err := os.ReadFile(path)
						if err != nil {
							t.Errorf("Failed to read written file: %v", err)
						} else if string(writtenContent) != content {
							t.Errorf("Expected file content '%s', got '%s'", content, string(writtenContent))
						}
					}
				}
			}
		})
	}
}

func TestHandleListProjectFiles(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create some test files and directories
	testFile := filepath.Join(tempDir, "test.txt")
	testDir := filepath.Join(tempDir, "testdir")
	
	_ = os.WriteFile(testFile, []byte("test"), 0644)
	_ = os.Mkdir(testDir, 0755)
	_ = os.WriteFile(filepath.Join(testDir, "test2.txt"), []byte("test2"), 0644)
	
	fs := NewFilesystemServer(&FilesystemConfig{
		RootDirectory: tempDir,
	})
	
	tests := []struct {
		name        string
		arguments   map[string]interface{}
		expectError bool
	}{
		{
			name: "list all project files",
			arguments: map[string]interface{}{},
			expectError: false,
		},
		{
			name: "list with glob pattern",
			arguments: map[string]interface{}{
				"glob_pattern": "*.txt",
			},
			expectError: false,
		},
		{
			name: "list including ignored files",
			arguments: map[string]interface{}{
				"include_ignored": true,
			},
			expectError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Arguments: tt.arguments,
				},
			}
			
			result, err := fs.handleListProjectFiles(context.Background(), request)
			
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
					
					// Check for expected files in project listing
					if tt.name == "list all project files" {
						if !strings.Contains(textContent.Text, "test.txt") {
							t.Errorf("Expected file listing to contain 'test.txt', got '%s'", textContent.Text)
						}
						// test2.txt is in subdirectory, should be listed as testdir/test2.txt
						if !strings.Contains(textContent.Text, "testdir/test2.txt") {
							t.Errorf("Expected file listing to contain 'testdir/test2.txt', got '%s'", textContent.Text)
						}
					}
					if tt.name == "list with glob pattern" && tt.arguments["glob_pattern"] == "*.txt" {
						if !strings.Contains(textContent.Text, "test.txt") {
							t.Errorf("Expected glob pattern result to contain 'test.txt', got '%s'", textContent.Text)
						}
					}
				} else {
					t.Error("Expected TextContent")
				}
			}
		})
	}
}

func TestCopyFileOrDir(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create test file
	testFile := filepath.Join(tempDir, "source.txt")
	testContent := "test content"
	_ = os.WriteFile(testFile, []byte(testContent), 0644)
	
	// Create test directory with content
	testDir := filepath.Join(tempDir, "sourcedir")
	_ = os.Mkdir(testDir, 0755)
	_ = os.WriteFile(filepath.Join(testDir, "file.txt"), []byte("dir content"), 0644)
	
	// Test with simple copy function instead of removed copyFileOrDir
	
	tests := []struct {
		name   string
		source string
		dest   string
	}{
		{
			name:   "copy file",
			source: testFile,
			dest:   filepath.Join(tempDir, "dest.txt"),
		},
		{
			name:   "copy directory",
			source: testDir,
			dest:   filepath.Join(tempDir, "destdir"),
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := copyFileForTest(tt.source, tt.dest)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			
			// Verify destination exists
			if _, err := os.Stat(tt.dest); os.IsNotExist(err) {
				t.Errorf("Expected destination '%s' to exist", tt.dest)
			}
		})
	}
}

func TestGlobPatternFiltering(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create test files
	files := []string{
		"main.go",
		"main_test.go",
		"config.yaml",
		"src/utils.go",
		"src/utils_test.go",
		"docs/guide.md",
		"cmd/server/main.go",
	}
	
	for _, file := range files {
		fullPath := filepath.Join(tempDir, file)
		_ = os.MkdirAll(filepath.Dir(fullPath), 0755)
		_ = os.WriteFile(fullPath, []byte("test"), 0644)
	}
	
	// Create .gitignore
	gitignoreContent := "*_test.go"
	_ = os.WriteFile(filepath.Join(tempDir, ".gitignore"), []byte(gitignoreContent), 0644)
	
	fs := NewFilesystemServer(&FilesystemConfig{
		RootDirectory: tempDir,
	})
	
	tests := []struct {
		name           string
		globPattern    string
		includeIgnored bool
		expectedFiles  []string
	}{
		{
			name:           "all go files including ignored",
			globPattern:    "**/*.go",
			includeIgnored: true,
			expectedFiles:  []string{"cmd/server/main.go", "main.go", "main_test.go", "src/utils.go", "src/utils_test.go"},
		},
		{
			name:           "all go files excluding ignored",
			globPattern:    "**/*.go",
			includeIgnored: false,
			expectedFiles:  []string{"cmd/server/main.go", "main.go", "src/utils.go"},
		},
		{
			name:           "only root level go files",
			globPattern:    "*.go",
			includeIgnored: false,
			expectedFiles:  []string{"main.go"},
		},
		{
			name:           "cmd directory go files",
			globPattern:    "cmd/**/*.go",
			includeIgnored: false,
			expectedFiles:  []string{"cmd/server/main.go"},
		},
		{
			name:           "markdown files",
			globPattern:    "**/*.md",
			includeIgnored: false,
			expectedFiles:  []string{"docs/guide.md"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fs.ListProjectFiles(tt.globPattern, tt.includeIgnored)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			// Sort both slices for comparison
			if len(result) != len(tt.expectedFiles) {
				t.Errorf("Expected %d files, got %d. Expected: %v, Got: %v", 
					len(tt.expectedFiles), len(result), tt.expectedFiles, result)
				return
			}
			
			// Check each expected file is present
			resultMap := make(map[string]bool)
			for _, file := range result {
				resultMap[file] = true
			}
			
			for _, expected := range tt.expectedFiles {
				if !resultMap[expected] {
					t.Errorf("Expected file %s not found in results: %v", expected, result)
				}
			}
		})
	}
}

// copyFileForTest is a simple helper for testing - copies functionality removed from main implementation
func copyFileForTest(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	
	if srcInfo.IsDir() {
		return copyDirForTest(src, dst)
	}
	
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	
	return os.WriteFile(dst, data, 0644)
}

func copyDirForTest(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}
	
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		
		if err := copyFileForTest(srcPath, dstPath); err != nil {
			return err
		}
	}
	
	return nil
}

