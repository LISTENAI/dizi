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
func customSSEHandler(sseServer *server.SSEServer, enableFsTools bool, fsRootDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters for filesystem tools (if needed for future enhancement)
		query := r.URL.Query()

		// Log query parameters if filesystem tools are requested
		if includeFsTools := query.Get("include_fs_tools"); includeFsTools != "" {
			logger.InfoLog("SSE request with include_fs_tools=%s", includeFsTools)
		}

		if rootDir := query.Get("fs_root"); rootDir != "" {
			logger.InfoLog("SSE request with fs_root=%s", rootDir)
		}

		// Handle the SSE connection with the shared server instance
		sseServer.ServeHTTP(w, r)
	}
}

// StartCustomSSEServer starts the SSE server with custom handling
func StartCustomSSEServer(cfg *config.Config, host string, port int, enableFsTools bool, fsRootDir string) error {
	// Create a single MCP server instance to be shared
	mcpServer := server.NewMCPServer(cfg.Name, cfg.Version)

	// Register basic tools
	if err := tools.RegisterTools(mcpServer, cfg.Tools); err != nil {
		return err
	}

	// Register filesystem tools if enabled
	if enableFsTools {
		fsConfig := &tools.FilesystemConfig{}

		// Use custom root if provided, otherwise default to project directory
		if fsRootDir != "" {
			fsConfig.RootDirectory = fsRootDir
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
			return err
		}

		logger.InfoLog("Filesystem tools enabled with root: %s", fsConfig.RootDirectory)
	}

	// Create SSE server with the shared MCP server
	sseServer := server.NewSSEServer(mcpServer)

	mux := http.NewServeMux()

	// Handle SSE endpoint with the shared server
	mux.HandleFunc("/sse", customSSEHandler(sseServer, enableFsTools, fsRootDir))

	// Handle message endpoint
	mux.Handle("/message", sseServer.MessageHandler())

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
