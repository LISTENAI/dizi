// Package tools provides filesystem tools for the MCP server.
// Based on the reference implementation from mcp-filesystem-server.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/djherbis/times"
	"github.com/gabriel-vasile/mimetype"
	"github.com/gobwas/glob"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// FilesystemConfig holds configuration for filesystem tools
type FilesystemConfig struct {
	RootDirectory string
}

// FilesystemServer wraps the filesystem functionality
type FilesystemServer struct {
	config *FilesystemConfig
}

// NewFilesystemServer creates a new filesystem server with the given configuration
func NewFilesystemServer(config *FilesystemConfig) *FilesystemServer {
	if config == nil {
		// Default to current working directory (project directory)
		pwd, err := os.Getwd()
		if err != nil {
			pwd = "."
		}
		config = &FilesystemConfig{
			RootDirectory: pwd,
		}
	}

	return &FilesystemServer{
		config: config,
	}
}

// RegisterFilesystemTools registers all filesystem-related tools
func RegisterFilesystemTools(mcpServer *server.MCPServer, config *FilesystemConfig) error {
	fs := NewFilesystemServer(config)

	tools := []struct {
		name    string
		desc    string
		schema  map[string]interface{}
		handler func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
	}{
		{
			"read_file",
			"Read the complete contents of a file from the file system. Handles both text and binary files safely. Can read files of any size. Provides a direct, uncached view of the file content at the time the tool is called.",
			map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Absolute or relative path to the file to read. The path will be resolved relative to the server's allowed directory.",
					},
				},
				"required": []string{"path"},
			},
			fs.handleReadFile,
		},
		{
			"write_file",
			"Create a new file or completely overwrite an existing file with new content. If the file exists, it will be completely replaced. If the file doesn't exist, it will be created along with any necessary parent directories.",
			map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Absolute or relative path where the file should be created or overwritten.",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "The complete content to write to the file. This will replace any existing content.",
					},
				},
				"required": []string{"path", "content"},
			},
			fs.handleWriteFile,
		},
		{
			"list_directory",
			"Get a detailed listing of all files and directories in a specified path. Results include names, types, sizes, and modification times. Can list the contents of any directory the server has access to.",
			map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Absolute or relative path to the directory to list. Use '.' for the current directory.",
					},
				},
				"required": []string{"path"},
			},
			fs.handleListDirectory,
		},
		{
			"create_directory",
			"Create a new directory or ensure a directory exists. Can create multiple nested directories in one operation. If the directory already exists, this operation will succeed silently.",
			map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Absolute or relative path of the directory to create. Parent directories will be created as needed.",
					},
				},
				"required": []string{"path"},
			},
			fs.handleCreateDirectory,
		},
		{
			"delete_file",
			"Delete a file or directory from the file system. For directories, this will recursively delete all contents. Use with caution as this operation cannot be undone.",
			map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Absolute or relative path to the file or directory to delete.",
					},
				},
				"required": []string{"path"},
			},
			fs.handleDeleteFile,
		},
		{
			"copy_file",
			"Copy a file or directory from one location to another. For directories, this will recursively copy all contents. Creates parent directories as needed at the destination.",
			map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"source": map[string]interface{}{
						"type":        "string",
						"description": "Absolute or relative path to the source file or directory to copy.",
					},
					"destination": map[string]interface{}{
						"type":        "string",
						"description": "Absolute or relative path where the file or directory should be copied to.",
					},
				},
				"required": []string{"source", "destination"},
			},
			fs.handleCopyFile,
		},
		{
			"move_file",
			"Move or rename a file or directory. Can move files between directories and rename them in a single operation. Creates parent directories as needed at the destination.",
			map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"source": map[string]interface{}{
						"type":        "string",
						"description": "Absolute or relative path to the source file or directory to move.",
					},
					"destination": map[string]interface{}{
						"type":        "string",
						"description": "Absolute or relative path where the file or directory should be moved to.",
					},
				},
				"required": []string{"source", "destination"},
			},
			fs.handleMoveFile,
		},
		{
			"get_file_info",
			"Get detailed information about a file or directory, including size, creation time, modification time, access time, type, and permissions. This tool provides comprehensive metadata about filesystem objects.",
			map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Absolute or relative path to the file or directory to get information about.",
					},
				},
				"required": []string{"path"},
			},
			fs.handleGetFileInfo,
		},
		{
			"search_files",
			"Search for files and directories matching a pattern. Supports glob patterns with wildcards. Can search recursively through directory trees. Returns detailed information about matching items.",
			map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Absolute or relative path to the directory to search in. The search will include all subdirectories.",
					},
					"pattern": map[string]interface{}{
						"type":        "string",
						"description": "Glob pattern to match against file and directory names (e.g., '*.txt', '**/*.py', 'test_*').",
					},
				},
				"required": []string{"path", "pattern"},
			},
			fs.handleSearchFiles,
		},
	}

	for _, tool := range tools {
		schemaBytes, err := json.Marshal(tool.schema)
		if err != nil {
			return fmt.Errorf("failed to marshal schema for tool %s: %w", tool.name, err)
		}

		mcpTool := mcp.NewToolWithRawSchema(tool.name, tool.desc, json.RawMessage(schemaBytes))
		mcpServer.AddTool(mcpTool, tool.handler)
	}

	return nil
}

// validatePath checks if the path is allowed and safe - only allows access within the root directory
func (fs *FilesystemServer) validatePath(path string) (string, error) {
	// Clean the path to prevent path traversal attacks
	cleanPath := filepath.Clean(path)

	// Make both the input path and root directory absolute
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	// Get absolute root directory
	rootAbs, err := filepath.Abs(fs.config.RootDirectory)
	if err != nil {
		return "", fmt.Errorf("invalid root directory: %w", err)
	}

	// Ensure path is within root directory (strict containment check)
	// This prevents access to files outside the project directory
	if !strings.HasPrefix(absPath+string(filepath.Separator), rootAbs+string(filepath.Separator)) && absPath != rootAbs {
		return "", fmt.Errorf("access denied: path %s is outside project directory %s", absPath, rootAbs)
	}

	return absPath, nil
}

// Tool handlers
func (fs *FilesystemServer) handleReadFile(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments format"), nil
	}

	path, ok := arguments["path"].(string)
	if !ok {
		return mcp.NewToolResultError("Missing or invalid path parameter"), nil
	}

	validPath, err := fs.validatePath(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Access denied: %v", err)), nil
	}

	content, err := os.ReadFile(validPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read file: %v", err)), nil
	}

	// Detect MIME type for better content handling
	mimeType := mimetype.Detect(content)

	// Add metadata about the file
	metadata := ""
	if stat, err := os.Stat(validPath); err == nil {
		metadata = fmt.Sprintf("File: %s\nSize: %d bytes\nMIME Type: %s\nModified: %s\n\n",
			validPath, stat.Size(), mimeType.String(), stat.ModTime().Format(time.RFC3339))
	}

	return mcp.NewToolResultText(metadata + string(content)), nil
}

func (fs *FilesystemServer) handleWriteFile(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments format"), nil
	}

	path, ok := arguments["path"].(string)
	if !ok {
		return mcp.NewToolResultError("Missing or invalid path parameter"), nil
	}

	content, ok := arguments["content"].(string)
	if !ok {
		return mcp.NewToolResultError("Missing or invalid content parameter"), nil
	}

	validPath, err := fs.validatePath(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Access denied: %v", err)), nil
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(validPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create directory: %v", err)), nil
	}

	if err := os.WriteFile(validPath, []byte(content), 0644); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to write file: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), validPath)), nil
}

func (fs *FilesystemServer) handleListDirectory(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments format"), nil
	}

	path, ok := arguments["path"].(string)
	if !ok {
		return mcp.NewToolResultError("Missing or invalid path parameter"), nil
	}

	validPath, err := fs.validatePath(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Access denied: %v", err)), nil
	}

	entries, err := os.ReadDir(validPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read directory: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Directory listing for %s:\n\n", validPath))
	result.WriteString(fmt.Sprintf("%-40s %-10s %-12s %-20s %-10s\n", "Name", "Type", "Size", "Modified", "Permissions"))
	result.WriteString(strings.Repeat("-", 100) + "\n")

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		fileType := "file"
		if entry.IsDir() {
			fileType = "directory"
		}

		// Get extended time information
		fullPath := filepath.Join(validPath, entry.Name())
		stat, err := os.Stat(fullPath)
		if err != nil {
			continue
		}

		// Get creation time if available
		timeSpec := times.Get(stat)
		modTime := timeSpec.ModTime()
		if timeSpec.ChangeTime() == (time.Time{}) {
			modTime = stat.ModTime()
		}

		result.WriteString(fmt.Sprintf("%-40s %-10s %12d %-20s %-10s\n",
			entry.Name(),
			fileType,
			info.Size(),
			modTime.Format("2006-01-02 15:04:05"),
			info.Mode().String(),
		))
	}

	return mcp.NewToolResultText(result.String()), nil
}

func (fs *FilesystemServer) handleCreateDirectory(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments format"), nil
	}

	path, ok := arguments["path"].(string)
	if !ok {
		return mcp.NewToolResultError("Missing or invalid path parameter"), nil
	}

	validPath, err := fs.validatePath(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Access denied: %v", err)), nil
	}

	if err := os.MkdirAll(validPath, 0755); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create directory: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully created directory: %s", validPath)), nil
}

func (fs *FilesystemServer) handleDeleteFile(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments format"), nil
	}

	path, ok := arguments["path"].(string)
	if !ok {
		return mcp.NewToolResultError("Missing or invalid path parameter"), nil
	}

	validPath, err := fs.validatePath(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Access denied: %v", err)), nil
	}

	// Get info about what we're deleting
	stat, err := os.Stat(validPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("File or directory not found: %v", err)), nil
	}

	itemType := "file"
	if stat.IsDir() {
		itemType = "directory"
	}

	if err := os.RemoveAll(validPath); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete %s: %v", itemType, err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully deleted %s: %s", itemType, validPath)), nil
}

func (fs *FilesystemServer) handleCopyFile(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments format"), nil
	}

	source, ok := arguments["source"].(string)
	if !ok {
		return mcp.NewToolResultError("Missing or invalid source parameter"), nil
	}

	destination, ok := arguments["destination"].(string)
	if !ok {
		return mcp.NewToolResultError("Missing or invalid destination parameter"), nil
	}

	validSource, err := fs.validatePath(source)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Source access denied: %v", err)), nil
	}

	validDestination, err := fs.validatePath(destination)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Destination access denied: %v", err)), nil
	}

	if err := fs.copyFileOrDir(validSource, validDestination); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to copy: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully copied %s to %s", validSource, validDestination)), nil
}

func (fs *FilesystemServer) handleMoveFile(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments format"), nil
	}

	source, ok := arguments["source"].(string)
	if !ok {
		return mcp.NewToolResultError("Missing or invalid source parameter"), nil
	}

	destination, ok := arguments["destination"].(string)
	if !ok {
		return mcp.NewToolResultError("Missing or invalid destination parameter"), nil
	}

	validSource, err := fs.validatePath(source)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Source access denied: %v", err)), nil
	}

	validDestination, err := fs.validatePath(destination)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Destination access denied: %v", err)), nil
	}

	// Create destination directory if it doesn't exist
	destDir := filepath.Dir(validDestination)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create destination directory: %v", err)), nil
	}

	if err := os.Rename(validSource, validDestination); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to move: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully moved %s to %s", validSource, validDestination)), nil
}

func (fs *FilesystemServer) handleGetFileInfo(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments format"), nil
	}

	path, ok := arguments["path"].(string)
	if !ok {
		return mcp.NewToolResultError("Missing or invalid path parameter"), nil
	}

	validPath, err := fs.validatePath(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Access denied: %v", err)), nil
	}

	info, err := os.Stat(validPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get file info: %v", err)), nil
	}

	fileType := "file"
	if info.IsDir() {
		fileType = "directory"
	}

	// Get extended time information
	timeSpec := times.Get(info)

	var result strings.Builder
	result.WriteString(fmt.Sprintf("File Information for: %s\n", validPath))
	result.WriteString(fmt.Sprintf("Type: %s\n", fileType))
	result.WriteString(fmt.Sprintf("Size: %d bytes\n", info.Size()))
	result.WriteString(fmt.Sprintf("Permissions: %s\n", info.Mode().String()))
	result.WriteString(fmt.Sprintf("Modified: %s\n", info.ModTime().Format(time.RFC3339)))

	if timeSpec.AccessTime() != (time.Time{}) {
		result.WriteString(fmt.Sprintf("Accessed: %s\n", timeSpec.AccessTime().Format(time.RFC3339)))
	}

	if timeSpec.ChangeTime() != (time.Time{}) {
		result.WriteString(fmt.Sprintf("Changed: %s\n", timeSpec.ChangeTime().Format(time.RFC3339)))
	}

	if timeSpec.BirthTime() != (time.Time{}) {
		result.WriteString(fmt.Sprintf("Created: %s\n", timeSpec.BirthTime().Format(time.RFC3339)))
	}

	// Add MIME type for files
	if !info.IsDir() {
		if data, err := os.ReadFile(validPath); err == nil {
			mimeType := mimetype.Detect(data)
			result.WriteString(fmt.Sprintf("MIME Type: %s\n", mimeType.String()))
		}
	}

	return mcp.NewToolResultText(result.String()), nil
}

func (fs *FilesystemServer) handleSearchFiles(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments format"), nil
	}

	path, ok := arguments["path"].(string)
	if !ok {
		return mcp.NewToolResultError("Missing or invalid path parameter"), nil
	}

	pattern, ok := arguments["pattern"].(string)
	if !ok {
		return mcp.NewToolResultError("Missing or invalid pattern parameter"), nil
	}

	validPath, err := fs.validatePath(path)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Access denied: %v", err)), nil
	}

	// Compile glob pattern
	g, err := glob.Compile(pattern)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid pattern: %v", err)), nil
	}

	var matches []string
	err = filepath.Walk(validPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking even if we can't access some files
		}

		// Get relative path from search root for pattern matching
		relPath, err := filepath.Rel(validPath, filePath)
		if err != nil {
			relPath = filepath.Base(filePath)
		}

		// Match against both filename and relative path
		fileName := filepath.Base(filePath)
		if g.Match(fileName) || g.Match(relPath) {
			matches = append(matches, filePath)
		}

		return nil
	})

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Search failed: %v", err)), nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Search results for pattern '%s' in %s:\n\n", pattern, validPath))

	if len(matches) == 0 {
		result.WriteString("No files found matching the pattern.")
	} else {
		result.WriteString(fmt.Sprintf("Found %d matching items:\n\n", len(matches)))
		result.WriteString(fmt.Sprintf("%-60s %-10s %-12s %-20s\n", "Path", "Type", "Size", "Modified"))
		result.WriteString(strings.Repeat("-", 110) + "\n")

		for _, match := range matches {
			if info, err := os.Stat(match); err == nil {
				fileType := "file"
				if info.IsDir() {
					fileType = "directory"
				}

				result.WriteString(fmt.Sprintf("%-60s %-10s %12d %-20s\n",
					match,
					fileType,
					info.Size(),
					info.ModTime().Format("2006-01-02 15:04:05"),
				))
			}
		}
	}

	return mcp.NewToolResultText(result.String()), nil
}

// copyFileOrDir recursively copies a file or directory
func (fs *FilesystemServer) copyFileOrDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if srcInfo.IsDir() {
		return fs.copyDir(src, dst)
	}
	return fs.copyFile(src, dst)
}

// copyFile copies a single file
func (fs *FilesystemServer) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Create destination directory if it doesn't exist
	destDir := filepath.Dir(dst)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Copy file permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, sourceInfo.Mode())
}

// copyDir recursively copies a directory
func (fs *FilesystemServer) copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Create destination directory
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

		if entry.IsDir() {
			if err := fs.copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := fs.copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}
