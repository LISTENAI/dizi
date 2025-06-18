// Package logger provides logging functionality for the MCP server.
// It provides smart logging that disables output in stdio mode to avoid
// interfering with the MCP protocol communication.
package logger

import (
	"log"
	"os"
)

var (
	// silentMode disables logging output (used in stdio mode)
	silentMode = false
	// logger is the standard logger instance
	logger = log.New(os.Stderr, "", log.LstdFlags)
)

// SetupLogger configures logging based on the transport mode
func SetupLogger(transport string) {
	if transport == "stdio" {
		// Disable logging for stdio mode to avoid interfering with protocol
		silentMode = true
	} else {
		silentMode = false
	}
}

// InfoLog logs an info message if not in silent mode
func InfoLog(format string, args ...interface{}) {
	if !silentMode {
		logger.Printf("[INFO] "+format, args...)
	}
}
