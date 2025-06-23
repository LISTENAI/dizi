// Package tools provides filesystem tools for the MCP server.
// Based on the reference implementation from mcp-filesystem-server.
package tools

import (
	"bufio"
	"context"
	"dizi/internal/gitls"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

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
	config         *FilesystemConfig
	readTimestamps map[string]int64       // Track file modification times when read
	gitIgnoreCache map[string][]glob.Glob // Cache parsed .gitignore patterns
	maxFileSize    int64                  // Maximum file size for reading (256KB)
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
		config:         config,
		readTimestamps: make(map[string]int64),
		gitIgnoreCache: make(map[string][]glob.Glob),
		maxFileSize:    262144, // 256KB
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
			"list_project_files",
			"Returns a list of files in the project. By default, when no arguments are passed, it returns all files in the project that are not ignored by .gitignore. Optionally, a glob_pattern can be passed to filter this list.",
			map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"glob_pattern": map[string]interface{}{
						"type":        "string",
						"description": "Optional: a glob pattern to filter the listed files.",
					},
					"include_ignored": map[string]interface{}{
						"type":        "boolean",
						"description": "Optional: whether to include files that are ignored by .gitignore. Defaults to false. WARNING: Use with targeted glob patterns to avoid listing excessive files from dependencies or build directories.",
					},
				},
				"required": []string{},
			},
			fs.handleListProjectFiles,
		},
		{
			"read_project_file",
			"Returns the contents of the given file. Supports an optional line_offset and count. To read the full file, only the path needs to be passed. For security reasons, this tool only works for files that are relative to the project root.",
			map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "The path to the file to read. It is relative to the project root.",
					},
					"line_offset": map[string]interface{}{
						"type":        "integer",
						"description": "Optional: the starting line offset from which to read. Defaults to 0.",
					},
					"count": map[string]interface{}{
						"type":        "integer",
						"description": "Optional: the number of lines to read. Defaults to all.",
					},
				},
				"required": []string{"path"},
			},
			fs.handleReadProjectFile,
		},
		{
			"write_project_file",
			"Writes a file to the file system. If the file already exists, it will be overwritten. Before writing to a file, ensure it was read using the read_project_file tool.",
			map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "The path to the file to write. It is relative to the project root.",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "The content to write to the file",
					},
				},
				"required": []string{"path", "content"},
			},
			fs.handleWriteProjectFile,
		},
		{
			"edit_project_file",
			"A tool for editing parts of a file. It can find and replace text inside a file. For moving or deleting files, use other tools instead. For large edits, use the write_project_file tool instead and overwrite the entire file. Before editing, ensure to read the source file using the read_project_file tool. To use this tool, provide the path to the file, the old_string to search for, and the new_string to replace it with. If the old_string is found multiple times, an error will be returned. To ensure uniqueness, include a couple of lines before and after the edit. All whitespace must be preserved as in the original file. This tool can only do a single edit at a time.",
			map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "The path to the file to edit. It is relative to the project root.",
					},
					"old_string": map[string]interface{}{
						"type":        "string",
						"description": "The string to search for",
					},
					"new_string": map[string]interface{}{
						"type":        "string",
						"description": "The string to replace the old_string with",
					},
				},
				"required": []string{"path", "old_string", "new_string"},
			},
			fs.handleEditProjectFile,
		},
		{
			"grep_project_files",
			"Searches for text patterns in files using regular expressions or plain text search.",
			map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"pattern": map[string]interface{}{
						"type":        "string",
						"description": "The pattern to search for",
					},
					"glob": map[string]interface{}{
						"type":        "string",
						"description": "Optional glob pattern to filter which files to search in, e.g., \"**/*.go\". Note that if a glob pattern is used, the .gitignore file will be ignored.",
					},
					"case_sensitive": map[string]interface{}{
						"type":        "boolean",
						"description": "Whether the search should be case-sensitive. Defaults to false.",
					},
					"max_results": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of results to return. Defaults to 100.",
					},
				},
				"required": []string{"pattern"},
			},
			fs.handleGrepProjectFiles,
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

// HandleListProjectFiles is an exported version for testing
func (fs *FilesystemServer) HandleListProjectFiles(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return fs.handleListProjectFiles(ctx, request)
}

func (fs *FilesystemServer) handleListProjectFiles(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments format"), nil
	}

	var opts []gitls.ListFilesOption
	opts = append(opts, gitls.WithDirectory(fs.config.RootDirectory))
	if pattern, ok := arguments["glob_pattern"].(string); ok && pattern != "" {
		opts = append(opts, gitls.WithGlob(pattern))
	}
	if ignored, ok := arguments["include_ignored"].(bool); ok && ignored {
		opts = append(opts, gitls.WithIncludeIgnored())
	}

	files, err := gitls.ListFiles(opts...)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list files: %v", err)), nil
	}

	if len(files) == 0 {
		return mcp.NewToolResultText("No files found."), nil
	}

	return mcp.NewToolResultText(strings.Join(files, "\n")), nil
}

func (fs *FilesystemServer) handleReadProjectFile(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments format"), nil
	}

	path, ok := arguments["path"].(string)
	if !ok {
		return mcp.NewToolResultError("Missing or invalid path parameter"), nil
	}

	lineOffset := 0
	if offset, exists := arguments["line_offset"].(float64); exists {
		lineOffset = int(offset)
	}

	count := -1 // -1 means read all lines
	if c, exists := arguments["count"].(float64); exists {
		count = int(c)
	}

	content, err := fs.readProjectFile(path, lineOffset, count)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read file: %v", err)), nil
	}

	return mcp.NewToolResultText(content), nil
}

func (fs *FilesystemServer) handleWriteProjectFile(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	err := fs.writeProjectFile(path, content)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to write file: %v", err)), nil
	}

	return mcp.NewToolResultText("Success!"), nil
}

func (fs *FilesystemServer) handleEditProjectFile(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments format"), nil
	}

	path, ok := arguments["path"].(string)
	if !ok {
		return mcp.NewToolResultError("Missing or invalid path parameter"), nil
	}

	oldString, ok := arguments["old_string"].(string)
	if !ok {
		return mcp.NewToolResultError("Missing or invalid old_string parameter"), nil
	}

	newString, ok := arguments["new_string"].(string)
	if !ok {
		return mcp.NewToolResultError("Missing or invalid new_string parameter"), nil
	}

	err := fs.editProjectFile(path, oldString, newString)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to edit file: %v", err)), nil
	}

	return mcp.NewToolResultText("Success!"), nil
}

func (fs *FilesystemServer) handleGrepProjectFiles(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	arguments, ok := request.Params.Arguments.(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("Invalid arguments format"), nil
	}

	pattern, ok := arguments["pattern"].(string)
	if !ok {
		return mcp.NewToolResultError("Missing or invalid pattern parameter"), nil
	}

	globPattern := ""
	if glob, exists := arguments["glob"].(string); exists {
		globPattern = glob
	}

	caseSensitive := false
	if cs, exists := arguments["case_sensitive"].(bool); exists {
		caseSensitive = cs
	}

	maxResults := 100
	if maxVal, exists := arguments["max_results"].(float64); exists {
		maxResults = int(maxVal)
	}

	results, err := fs.grepProjectFiles(pattern, globPattern, caseSensitive, maxResults)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to search files: %v", err)), nil
	}

	jsonResult, err := json.Marshal(results)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to encode results: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonResult)), nil
}

// Core implementation functions

// ListProjectFiles lists all files in the project, optionally filtering by glob pattern
func (fs *FilesystemServer) ListProjectFiles(globPattern string, includeIgnored bool) ([]string, error) {
	rootAbs, err := filepath.Abs(fs.config.RootDirectory)
	if err != nil {
		return nil, fmt.Errorf("invalid root directory: %w", err)
	}

	globMatcher, altGlobMatcher, err := fs.compileGlobMatchers(globPattern)
	if err != nil {
		return nil, err
	}

	ignorePatterns := fs.getIgnorePatterns(includeIgnored)

	var files []string
	err = filepath.Walk(rootAbs, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil // Continue walking even if we can't access some files
		}

		relPath, err := filepath.Rel(rootAbs, path)
		if err != nil {
			return nil
		}

		if fs.shouldIncludeFile(relPath, globPattern, globMatcher, altGlobMatcher, ignorePatterns, includeIgnored) {
			files = append(files, relPath)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return files, nil
}

// compileGlobMatchers compiles glob patterns for file matching
func (fs *FilesystemServer) compileGlobMatchers(globPattern string) (glob.Glob, glob.Glob, error) {
	if globPattern == "" {
		return nil, nil, nil
	}

	globMatcher, err := glob.Compile(globPattern)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid glob pattern: %w", err)
	}

	var altGlobMatcher glob.Glob
	// Handle ** patterns - also create a matcher for root level files
	if strings.HasPrefix(globPattern, "**/") {
		altPattern := globPattern[3:] // Remove "**/" prefix
		if altGlobMatcher, err = glob.Compile(altPattern); err != nil {
			altGlobMatcher = nil // Ignore if compilation fails
		}
	}

	return globMatcher, altGlobMatcher, nil
}

// getIgnorePatterns loads .gitignore patterns if needed
func (fs *FilesystemServer) getIgnorePatterns(includeIgnored bool) []glob.Glob {
	if includeIgnored {
		return nil
	}

	ignorePatterns, err := fs.loadGitignorePatterns()
	if err != nil {
		// Continue even if we can't load .gitignore
		return []glob.Glob{}
	}
	return ignorePatterns
}

// shouldIncludeFile determines whether a file should be included in the results
func (fs *FilesystemServer) shouldIncludeFile(relPath, globPattern string, globMatcher, altGlobMatcher glob.Glob, ignorePatterns []glob.Glob, includeIgnored bool) bool {
	// Apply glob filter if specified
	if globMatcher != nil && !fs.matchesGlobPattern(relPath, globPattern, globMatcher, altGlobMatcher) {
		return false
	}

	// Apply .gitignore patterns if not including ignored files
	if !includeIgnored {
		for _, pattern := range ignorePatterns {
			if pattern.Match(relPath) {
				return false // Skip ignored file
			}
		}
	}

	return true
}

// matchesGlobPattern checks if a file path matches the glob pattern
func (fs *FilesystemServer) matchesGlobPattern(relPath, globPattern string, globMatcher, altGlobMatcher glob.Glob) bool {
	matched := globMatcher.Match(relPath)

	// For ** patterns, also try matching against root-level pattern
	if !matched && altGlobMatcher != nil {
		matched = altGlobMatcher.Match(relPath)
	}

	// Special handling for patterns without ** or / that should only match root level
	if matched && !strings.Contains(globPattern, "**/") && !strings.Contains(globPattern, "/") {
		// If pattern doesn't contain path separators, it should only match root level files
		if strings.Contains(relPath, "/") {
			matched = false
		}
	}

	return matched
}

// loadGitignorePatterns loads and compiles .gitignore patterns
func (fs *FilesystemServer) loadGitignorePatterns() ([]glob.Glob, error) {
	// Check cache first
	if patterns, exists := fs.gitIgnoreCache[fs.config.RootDirectory]; exists {
		return patterns, nil
	}

	gitignorePath := filepath.Join(fs.config.RootDirectory, ".gitignore")
	file, err := os.Open(gitignorePath)
	if err != nil {
		// If .gitignore doesn't exist, return empty patterns
		if os.IsNotExist(err) {
			return []glob.Glob{}, nil
		}
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var patterns []glob.Glob
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Convert gitignore pattern to glob pattern
		globPattern := fs.gitignoreToGlob(line)
		if globPattern != "" { // Skip empty patterns (e.g., negation patterns not implemented)
			if compiled, err := glob.Compile(globPattern); err == nil {
				patterns = append(patterns, compiled)
			}
		}
	}

	// Cache the patterns
	fs.gitIgnoreCache[fs.config.RootDirectory] = patterns
	return patterns, scanner.Err()
}

// gitignoreToGlob converts a .gitignore pattern to a glob pattern
func (fs *FilesystemServer) gitignoreToGlob(pattern string) string {
	// Handle negation patterns (not implemented yet, just skip)
	if strings.HasPrefix(pattern, "!") {
		return ""
	}

	// Handle patterns starting with / (anchor to root)
	if strings.HasPrefix(pattern, "/") {
		pattern = pattern[1:]
		if strings.HasSuffix(pattern, "/") {
			return pattern + "**"
		}
		return pattern
	}

	// Handle directory patterns
	if strings.HasSuffix(pattern, "/") {
		return "{" + pattern + "**," + "**/" + pattern + "**}"
	}

	// Handle patterns with ** already
	if strings.Contains(pattern, "**") {
		return pattern
	}

	// Default case - match anywhere in the tree (including root)
	return "{" + pattern + "," + "**/" + pattern + "}"
}

// readProjectFile reads a file with optional line offset and count
func (fs *FilesystemServer) readProjectFile(path string, lineOffset, count int) (string, error) {
	validPath, err := fs.validatePath(path)
	if err != nil {
		return "", err
	}

	// Check file size
	stat, err := os.Stat(validPath)
	if err != nil {
		return "", fmt.Errorf("file does not exist: %w", err)
	}

	if stat.Size() > fs.maxFileSize {
		return "", fmt.Errorf("file is too large to read (%d bytes). Maximum size is %d bytes", stat.Size(), fs.maxFileSize)
	}

	if !stat.Mode().IsRegular() {
		return "", fmt.Errorf("cannot read non-regular file")
	}

	content, err := os.ReadFile(validPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Check if content is valid UTF-8
	if !utf8.Valid(content) {
		return "", fmt.Errorf("cannot read file, because it contains invalid UTF-8 characters")
	}

	// Track file modification time
	fs.readTimestamps[validPath] = stat.ModTime().Unix()

	// Apply line offset and count if specified
	contentStr := string(content)
	if lineOffset > 0 || count > 0 {
		lines := strings.Split(contentStr, "\n")

		if lineOffset >= len(lines) {
			return "", nil
		}

		endIndex := len(lines)
		if count > 0 && lineOffset+count < len(lines) {
			endIndex = lineOffset + count
		}

		contentStr = strings.Join(lines[lineOffset:endIndex], "\n")
	}

	return contentStr, nil
}

// writeProjectFile writes content to a file with staleness check
func (fs *FilesystemServer) writeProjectFile(path, content string) error {
	validPath, err := fs.validatePath(path)
	if err != nil {
		return err
	}

	// Check if file has been read and is stale
	if err := fs.checkStale(validPath, true); err != nil {
		return err
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(validPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(validPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Update modification timestamp
	if stat, err := os.Stat(validPath); err == nil {
		fs.readTimestamps[validPath] = stat.ModTime().Unix()
	}

	return nil
}

// checkStale checks if a file has been modified since last read
func (fs *FilesystemServer) checkStale(path string, allowNotFound bool) error {
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) && allowNotFound {
			return nil
		}
		return fmt.Errorf("file does not exist: %w", err)
	}

	lastRead, exists := fs.readTimestamps[path]
	if !exists {
		return fmt.Errorf("file has not been read yet. Use read_project_file first before overwriting it")
	}

	if stat.ModTime().Unix() > lastRead {
		return fmt.Errorf("file has been modified since last read. Use read_project_file first to read it again")
	}

	return nil
}

// editProjectFile performs a find-and-replace edit on a file
func (fs *FilesystemServer) editProjectFile(path, oldString, newString string) error {
	validPath, err := fs.validatePath(path)
	if err != nil {
		return err
	}

	// Check if file has been read and is stale
	if err := fs.checkStale(validPath, false); err != nil {
		return err
	}

	content, err := os.ReadFile(validPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	contentStr := string(content)

	// Ensure old_string appears exactly once
	matches := strings.Count(contentStr, oldString)
	if matches == 0 {
		return fmt.Errorf("the original substring was not found in the file. No edits were made")
	}
	if matches > 1 {
		return fmt.Errorf("the substring was found more than once (%d times) in the file. No edits were made. Ensure uniqueness by providing more context", matches)
	}

	// Perform the replacement
	newContent := strings.Replace(contentStr, oldString, newString, 1)

	if err := os.WriteFile(validPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Update modification timestamp
	if stat, err := os.Stat(validPath); err == nil {
		fs.readTimestamps[validPath] = stat.ModTime().Unix()
	}

	return nil
}

// GrepResult represents a single search result
type GrepResult struct {
	Path    string `json:"path"`
	Line    int    `json:"line"`
	Content string `json:"content"`
}

// grepProjectFiles searches for patterns in project files
func (fs *FilesystemServer) grepProjectFiles(pattern, globPattern string, caseSensitive bool, maxResults int) ([]GrepResult, error) {
	rootAbs, err := filepath.Abs(fs.config.RootDirectory)
	if err != nil {
		return nil, fmt.Errorf("invalid root directory: %w", err)
	}

	searchCtx := &grepSearchContext{
		pattern:        pattern,
		caseSensitive:  caseSensitive,
		maxResults:     maxResults,
		globMatcher:    nil,
		regex:          nil,
		ignorePatterns: nil,
	}

	if err := fs.setupGrepSearch(searchCtx, globPattern); err != nil {
		return nil, err
	}

	var results []GrepResult
	err = filepath.Walk(rootAbs, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || len(results) >= maxResults {
			return nil
		}

		relPath, err := filepath.Rel(rootAbs, path)
		if err != nil {
			return nil
		}

		if fs.shouldSearchFile(relPath, searchCtx, globPattern) {
			fileResults := fs.searchInFile(path, relPath, searchCtx)
			results = append(results, fileResults...)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	return results, nil
}

// grepSearchContext holds the search configuration
type grepSearchContext struct {
	pattern        string
	caseSensitive  bool
	maxResults     int
	globMatcher    glob.Glob
	regex          *regexp.Regexp
	ignorePatterns []glob.Glob
}

// setupGrepSearch prepares the search context
func (fs *FilesystemServer) setupGrepSearch(ctx *grepSearchContext, globPattern string) error {
	// Compile glob pattern if provided
	if globPattern != "" {
		var err error
		ctx.globMatcher, err = glob.Compile(globPattern)
		if err != nil {
			return fmt.Errorf("invalid glob pattern: %w", err)
		}
	}

	// Compile regex pattern
	var err error
	if !ctx.caseSensitive {
		ctx.regex, err = regexp.Compile("(?i)" + ctx.pattern)
	} else {
		ctx.regex, err = regexp.Compile(ctx.pattern)
	}
	if err != nil {
		// Fallback to literal string search
		ctx.regex = nil
	}

	// Load .gitignore patterns if no glob pattern is specified
	if globPattern == "" {
		ctx.ignorePatterns, _ = fs.loadGitignorePatterns()
	}

	return nil
}

// shouldSearchFile determines if a file should be searched
func (fs *FilesystemServer) shouldSearchFile(relPath string, ctx *grepSearchContext, globPattern string) bool {
	// Apply glob filter
	if ctx.globMatcher != nil && !ctx.globMatcher.Match(relPath) {
		return false
	}

	// Apply .gitignore patterns if no glob specified
	if globPattern == "" {
		for _, pattern := range ctx.ignorePatterns {
			if pattern.Match(relPath) {
				return false
			}
		}
	}

	return true
}

// searchInFile searches for patterns within a single file
func (fs *FilesystemServer) searchInFile(path, relPath string, ctx *grepSearchContext) []GrepResult {
	content, err := os.ReadFile(path)
	if err != nil || !utf8.Valid(content) {
		return nil
	}

	var results []GrepResult
	lines := strings.Split(string(content), "\n")

	for lineNum, line := range lines {
		if len(results) >= ctx.maxResults {
			break
		}

		if fs.lineMatches(line, ctx) {
			truncatedContent := line
			if len(line) > 200 {
				truncatedContent = line[:200] + "..."
			}

			results = append(results, GrepResult{
				Path:    relPath,
				Line:    lineNum + 1,
				Content: truncatedContent,
			})
		}
	}

	return results
}

// lineMatches checks if a line matches the search pattern
func (fs *FilesystemServer) lineMatches(line string, ctx *grepSearchContext) bool {
	if ctx.regex != nil {
		return ctx.regex.MatchString(line)
	}

	// Fallback to simple string search
	searchLine := line
	searchPattern := ctx.pattern
	if !ctx.caseSensitive {
		searchLine = strings.ToLower(line)
		searchPattern = strings.ToLower(ctx.pattern)
	}
	return strings.Contains(searchLine, searchPattern)
}
