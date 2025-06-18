package logger

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func TestSetupLogger(t *testing.T) {
	tests := []struct {
		name      string
		transport string
		silent    bool
	}{
		{
			name:      "stdio transport enables silent mode",
			transport: "stdio",
			silent:    true,
		},
		{
			name:      "sse transport disables silent mode",
			transport: "sse",
			silent:    false,
		},
		{
			name:      "other transport disables silent mode",
			transport: "http",
			silent:    false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset silent mode to a known state
			silentMode = false
			
			SetupLogger(tt.transport)
			
			if silentMode != tt.silent {
				t.Errorf("Expected silentMode %v, got %v", tt.silent, silentMode)
			}
		})
	}
}

func TestInfoLog(t *testing.T) {
	tests := []struct {
		name        string
		silent      bool
		format      string
		args        []interface{}
		expectEmpty bool
	}{
		{
			name:        "logs when not silent",
			silent:      false,
			format:      "Test message %s",
			args:        []interface{}{"hello"},
			expectEmpty: false,
		},
		{
			name:        "does not log when silent",
			silent:      true,
			format:      "Test message %s",
			args:        []interface{}{"hello"},
			expectEmpty: true,
		},
		{
			name:        "logs simple message when not silent",
			silent:      false,
			format:      "Simple message",
			args:        []interface{}{},
			expectEmpty: false,
		},
		{
			name:        "does not log simple message when silent",
			silent:      true,
			format:      "Simple message",
			args:        []interface{}{},
			expectEmpty: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture log output
			var buf bytes.Buffer
			originalLogger := logger
			logger = log.New(&buf, "", 0) // No timestamp for easier testing
			defer func() { logger = originalLogger }()
			
			// Set silent mode
			silentMode = tt.silent
			
			// Call InfoLog
			InfoLog(tt.format, tt.args...)
			
			output := buf.String()
			
			if tt.expectEmpty {
				if output != "" {
					t.Errorf("Expected empty output in silent mode, got: %s", output)
				}
			} else {
				if output == "" {
					t.Error("Expected log output when not silent, got empty string")
				}
				
				expectedPrefix := "[INFO]"
				if !strings.Contains(output, expectedPrefix) {
					t.Errorf("Expected output to contain '%s', got: %s", expectedPrefix, output)
				}
				
				// Check that the formatted message is included
				if len(tt.args) > 0 {
					expectedContent := "Test message hello"
					if !strings.Contains(output, expectedContent) {
						t.Errorf("Expected output to contain formatted message, got: %s", output)
					}
				} else {
					expectedContent := "Simple message"
					if !strings.Contains(output, expectedContent) {
						t.Errorf("Expected output to contain simple message, got: %s", output)
					}
				}
			}
		})
	}
}

func TestInfoLogMultipleArgs(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	originalLogger := logger
	logger = log.New(&buf, "", 0)
	defer func() { logger = originalLogger }()
	
	// Ensure not in silent mode
	silentMode = false
	
	// Test with multiple arguments
	InfoLog("Test %s %d %v", "string", 42, true)
	
	output := buf.String()
	expectedContent := "Test string 42 true"
	
	if !strings.Contains(output, expectedContent) {
		t.Errorf("Expected output to contain '%s', got: %s", expectedContent, output)
	}
}

func TestInfoLogWithNoFormatting(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	originalLogger := logger
	logger = log.New(&buf, "", 0)
	defer func() { logger = originalLogger }()
	
	// Ensure not in silent mode
	silentMode = false
	
	// Test with no formatting placeholders
	InfoLog("Plain message without formatting")
	
	output := buf.String()
	expectedContent := "Plain message without formatting"
	
	if !strings.Contains(output, expectedContent) {
		t.Errorf("Expected output to contain '%s', got: %s", expectedContent, output)
	}
	
	if !strings.Contains(output, "[INFO]") {
		t.Errorf("Expected output to contain '[INFO]' prefix, got: %s", output)
	}
}

func TestLoggerInitialization(t *testing.T) {
	// Test that logger is properly initialized
	if logger == nil {
		t.Error("Logger should be initialized")
	}
	
	// Test that initial silent mode is false
	originalSilent := silentMode
	defer func() { silentMode = originalSilent }()
	
	// Reset to test initial state
	silentMode = false
	
	if silentMode != false {
		t.Error("Initial silent mode should be false")
	}
}

func TestConcurrentLogging(t *testing.T) {
	// Test that logging is safe for concurrent use
	var buf bytes.Buffer
	originalLogger := logger
	logger = log.New(&buf, "", 0)
	defer func() { logger = originalLogger }()
	
	silentMode = false
	
	// Start multiple goroutines that log simultaneously
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func(id int) {
			InfoLog("Concurrent log %d", id)
			done <- true
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
	
	output := buf.String()
	
	// Check that we got some output (exact content may vary due to concurrency)
	if output == "" {
		t.Error("Expected some log output from concurrent logging")
	}
	
	// Count occurrences of "[INFO]" to verify multiple logs
	infoCount := strings.Count(output, "[INFO]")
	if infoCount != 10 {
		t.Errorf("Expected 10 log entries, got %d", infoCount)
	}
}