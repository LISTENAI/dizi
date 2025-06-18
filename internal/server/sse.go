// Package server provides custom SSE handling with query parameter support.
package server

import (
	"net/http"
	"os"
	"strconv"

	"dizi/internal/config"
	"dizi/internal/logger"
	"dizi/internal/tools"

	"github.com/mark3labs/mcp-go/server"
)


// customSSEHandler wraps the SSE server to handle query parameters
func customSSEHandler(cfg *config.Config, enableFsTools bool, fsRootDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters
		query := r.URL.Query()

		// Check if filesystem tools should be enabled for this request
		shouldEnableFsTools := enableFsTools
		customFsRoot := fsRootDir

		if includeFsTools := query.Get("include_fs_tools"); includeFsTools != "" {
			if enabled, err := strconv.ParseBool(includeFsTools); err == nil {
				shouldEnableFsTools = enabled
			}
		}

		if rootDir := query.Get("fs_root"); rootDir != "" {
			customFsRoot = rootDir
		}

		// Create a new MCP server for this specific request
		mcpServer := server.NewMCPServer(cfg.Name, cfg.Version)

		// Register basic tools
		if err := tools.RegisterTools(mcpServer, cfg.Tools); err != nil {
			http.Error(w, "Failed to register tools", http.StatusInternalServerError)
			return
		}

		// Register filesystem tools if requested
		if shouldEnableFsTools {
			fsConfig := &tools.FilesystemConfig{}

			// Use custom root if provided, otherwise default to project directory
			if customFsRoot != "" {
				fsConfig.RootDirectory = customFsRoot
			} else {
				// Default to current working directory (project directory)
				pwd, err := os.Getwd()
				if err != nil {
					fsConfig.RootDirectory = "."
				} else {
					fsConfig.RootDirectory = pwd
				}
			}

			if err := tools.RegisterFilesystemTools(mcpServer, fsConfig); err != nil {
				http.Error(w, "Failed to register filesystem tools", http.StatusInternalServerError)
				return
			}

			logger.InfoLog("SSE request with filesystem tools enabled, root: %s", fsConfig.RootDirectory)
		}

		// Create SSE server for this request
		sseServer := server.NewSSEServer(mcpServer)

		// Handle the SSE connection
		sseServer.ServeHTTP(w, r)
	}
}

// StartCustomSSEServer starts the SSE server with custom handling
func StartCustomSSEServer(cfg *config.Config, host string, port int, enableFsTools bool, fsRootDir string) error {
	mux := http.NewServeMux()

	// Handle SSE endpoint with query parameter support
	mux.HandleFunc("/sse", customSSEHandler(cfg, enableFsTools, fsRootDir))

	// Handle message endpoint - we need to create a basic server for this
	basicMcpServer := server.NewMCPServer(cfg.Name, cfg.Version)
	if err := tools.RegisterTools(basicMcpServer, cfg.Tools); err != nil {
		return err
	}

	basicSSEServer := server.NewSSEServer(basicMcpServer)
	mux.Handle("/message", basicSSEServer.MessageHandler())

	// Add a simple status endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"name": "` + cfg.Name + `",
			"version": "` + cfg.Version + `",
			"description": "` + cfg.Description + `",
			"endpoints": {
				"/sse": "SSE endpoint (supports ?include_fs_tools=true&fs_root=/path)",
				"/message": "Message endpoint",
				"/": "Status endpoint"
			}
		}`))
	})

	addr := host + ":" + strconv.Itoa(port)
	logger.InfoLog("Starting custom SSE server on http://%s", addr)
	logger.InfoLog("SSE endpoint: http://%s/sse", addr)
	logger.InfoLog("With filesystem tools: http://%s/sse?include_fs_tools=true", addr)

	return http.ListenAndServe(addr, mux)
}
