package main

import (
	"io"
	"log"
	"os"
)

// setupLogger configures logging based on transport mode
func setupLogger(transport string) {
	if transport == "stdio" {
		// For stdio mode, disable logging to avoid interfering with MCP protocol
		// All log output will be discarded
		log.SetOutput(io.Discard)
		log.SetFlags(0) // Remove timestamp and other flags
	} else {
		// For SSE and other modes, log to stderr with timestamp
		log.SetOutput(os.Stderr)
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}
}

// debugLog logs debug messages only when not in stdio mode
func debugLog(format string, v ...interface{}) {
	log.Printf("[DEBUG] "+format, v...)
}

// infoLog logs info messages
func infoLog(format string, v ...interface{}) {
	log.Printf("[INFO] "+format, v...)
}

// errorLog logs error messages
func errorLog(format string, v ...interface{}) {
	log.Printf("[ERROR] "+format, v...)
}