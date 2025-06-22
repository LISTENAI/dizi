// Package shell provides cross-platform shell environment loading functionality
package shell

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// GetShellConfigFiles returns the list of shell configuration files to source
// based on the current platform and shell
func GetShellConfigFiles() []string {
	var configFiles []string
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return configFiles
	}

	switch runtime.GOOS {
	case "windows":
		// Windows PowerShell profiles
		configFiles = append(configFiles,
			filepath.Join(homeDir, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1"),
			filepath.Join(homeDir, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1"),
		)
	case "darwin", "linux":
		// Unix-like systems - get current shell
		currentShell := getCurrentShell()
		
		// Always include common profile files
		configFiles = append(configFiles,
			"/etc/profile",
			filepath.Join(homeDir, ".profile"),
		)
		
		// Add shell-specific configuration files
		switch filepath.Base(currentShell) {
		case "bash":
			configFiles = append(configFiles,
				"/etc/bash.bashrc",
				"/etc/bashrc",
				filepath.Join(homeDir, ".bashrc"),
				filepath.Join(homeDir, ".bash_profile"),
				filepath.Join(homeDir, ".bash_login"),
			)
		case "zsh":
			configFiles = append(configFiles,
				"/etc/zsh/zshenv",
				"/etc/zshenv",
				filepath.Join(homeDir, ".zshenv"),
				"/etc/zsh/zprofile",
				"/etc/zprofile", 
				filepath.Join(homeDir, ".zprofile"),
				"/etc/zsh/zshrc",
				"/etc/zshrc",
				filepath.Join(homeDir, ".zshrc"),
				"/etc/zsh/zlogin",
				"/etc/zlogin",
				filepath.Join(homeDir, ".zlogin"),
			)
		case "fish":
			configFiles = append(configFiles,
				"/etc/fish/config.fish",
				filepath.Join(homeDir, ".config", "fish", "config.fish"),
			)
		case "tcsh", "csh":
			configFiles = append(configFiles,
				"/etc/csh.cshrc",
				"/etc/csh.login",
				filepath.Join(homeDir, ".cshrc"),
				filepath.Join(homeDir, ".tcshrc"),
				filepath.Join(homeDir, ".login"),
			)
		case "ksh":
			configFiles = append(configFiles,
				"/etc/ksh.kshrc",
				filepath.Join(homeDir, ".kshrc"),
			)
		default:
			// For unknown shells, try common patterns
			shellName := filepath.Base(currentShell)
			if shellName != "" && shellName != "sh" {
				configFiles = append(configFiles,
					filepath.Join(homeDir, "."+shellName+"rc"),
					filepath.Join(homeDir, "."+shellName+"_profile"),
				)
			}
			// Always include bash fallbacks for sh-compatible shells
			configFiles = append(configFiles,
				filepath.Join(homeDir, ".bashrc"),
			)
		}
	}

	// Filter out non-existent files
	var existingFiles []string
	for _, file := range configFiles {
		if _, err := os.Stat(file); err == nil {
			existingFiles = append(existingFiles, file)
		}
	}

	return existingFiles
}

// getCurrentShell returns the current shell, trying multiple methods
func getCurrentShell() string {
	// Method 1: Check SHELL environment variable
	if shell := os.Getenv("SHELL"); shell != "" {
		return shell
	}
	
	// Method 2: Check parent process (works on Unix systems)
	if runtime.GOOS != "windows" {
		if shell := getParentShell(); shell != "" {
			return shell
		}
	}
	
	// Method 3: Check common shell locations
	commonShells := []string{
		"/bin/bash", "/usr/bin/bash", "/usr/local/bin/bash",
		"/bin/zsh", "/usr/bin/zsh", "/usr/local/bin/zsh",
		"/opt/homebrew/bin/zsh", // Homebrew on Apple Silicon
		"/usr/local/bin/fish", "/opt/homebrew/bin/fish",
		"/bin/sh", "/usr/bin/sh",
	}
	
	for _, shell := range commonShells {
		if _, err := os.Stat(shell); err == nil {
			return shell
		}
	}
	
	// Method 4: Try to find shells in PATH
	pathShells := []string{"zsh", "bash", "fish", "sh"}
	for _, shell := range pathShells {
		if path, err := exec.LookPath(shell); err == nil {
			return path
		}
	}
	
	return "/bin/sh" // Ultimate fallback
}

// getParentShell tries to determine the parent shell process
func getParentShell() string {
	// This is a simplified approach - in production you might want to use
	// more sophisticated process tree analysis
	cmd := exec.Command("ps", "-p", fmt.Sprintf("%d", os.Getppid()), "-o", "comm=")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	
	parentComm := strings.TrimSpace(string(output))
	if parentComm == "" {
		return ""
	}
	
	// Try to find the full path
	if path, err := exec.LookPath(parentComm); err == nil {
		return path
	}
	
	return parentComm
}

// CreateShellCommand creates a command that runs in the user's configured shell environment
func CreateShellCommand(command string, args ...string) *exec.Cmd {
	switch runtime.GOOS {
	case "windows":
		return createWindowsCommand(command, args...)
	default:
		return createUnixCommand(command, args...)
	}
}

// CreateShellScriptCommand creates a command that runs a script in the user's shell environment
func CreateShellScriptCommand(script string) *exec.Cmd {
	switch runtime.GOOS {
	case "windows":
		return createWindowsScriptCommand(script)
	default:
		return createUnixScriptCommand(script)
	}
}

// createUnixCommand creates a command for Unix-like systems
func createUnixCommand(command string, args ...string) *exec.Cmd {
	shell := getCurrentShell()
	shellName := filepath.Base(shell)
	
	// Build the command with environment loading
	var fullCommand strings.Builder
	
	// Source configuration files based on shell type
	configFiles := GetShellConfigFiles()
	
	switch shellName {
	case "fish":
		// Fish shell has different syntax
		for _, file := range configFiles {
			fullCommand.WriteString(fmt.Sprintf("test -f '%s'; and source '%s'; ", file, file))
		}
	case "csh", "tcsh":
		// C shell family has different syntax
		for _, file := range configFiles {
			fullCommand.WriteString(fmt.Sprintf("if (-f '%s') source '%s'; ", file, file))
		}
	default:
		// Bourne shell family (bash, zsh, sh, ksh, etc.)
		for _, file := range configFiles {
			fullCommand.WriteString(fmt.Sprintf("[ -f '%s' ] && source '%s' 2>/dev/null; ", file, file))
		}
	}
	
	// Add the actual command
	fullCommand.WriteString(command)
	for _, arg := range args {
		// Escape arguments to prevent shell injection
		escapedArg := strings.ReplaceAll(arg, "'", "'\"'\"'")
		fullCommand.WriteString(fmt.Sprintf(" '%s'", escapedArg))
	}
	
	// Use appropriate shell flags
	var shellArgs []string
	switch shellName {
	case "fish":
		shellArgs = []string{"-c"}
	case "csh", "tcsh":
		shellArgs = []string{"-c"}
	default:
		// Most shells support -c for command execution
		shellArgs = []string{"-c"}
	}
	
	shellArgs = append(shellArgs, fullCommand.String())
	return exec.Command(shell, shellArgs...)
}

// createUnixScriptCommand creates a script command for Unix-like systems
func createUnixScriptCommand(script string) *exec.Cmd {
	shell := getCurrentShell()
	shellName := filepath.Base(shell)
	
	// Build the script with environment loading
	var fullScript strings.Builder
	
	// Source configuration files
	configFiles := GetShellConfigFiles()
	
	switch shellName {
	case "fish":
		for _, file := range configFiles {
			fullScript.WriteString(fmt.Sprintf("test -f '%s'; and source '%s'\n", file, file))
		}
	case "csh", "tcsh":
		for _, file := range configFiles {
			fullScript.WriteString(fmt.Sprintf("if (-f '%s') source '%s'\n", file, file))
		}
	default:
		for _, file := range configFiles {
			fullScript.WriteString(fmt.Sprintf("[ -f '%s' ] && source '%s' 2>/dev/null\n", file, file))
		}
	}
	
	// Add the actual script
	fullScript.WriteString(script)
	
	return exec.Command(shell, "-c", fullScript.String())
}

// createWindowsCommand creates a command for Windows systems
func createWindowsCommand(command string, args ...string) *exec.Cmd {
	var psCommand strings.Builder
	
	// Load PowerShell profiles
	configFiles := GetShellConfigFiles()
	for _, file := range configFiles {
		psCommand.WriteString(fmt.Sprintf("if (Test-Path '%s') { . '%s' }; ", file, file))
	}
	
	// Add the actual command
	psCommand.WriteString(command)
	for _, arg := range args {
		// Escape PowerShell arguments
		escapedArg := strings.ReplaceAll(arg, "'", "''")
		psCommand.WriteString(fmt.Sprintf(" '%s'", escapedArg))
	}
	
	return exec.Command("powershell", "-NoProfile", "-Command", psCommand.String())
}

// createWindowsScriptCommand creates a script command for Windows systems
func createWindowsScriptCommand(script string) *exec.Cmd {
	var psScript strings.Builder
	
	// Load PowerShell profiles
	configFiles := GetShellConfigFiles()
	for _, file := range configFiles {
		psScript.WriteString(fmt.Sprintf("if (Test-Path '%s') { . '%s' }\n", file, file))
	}
	
	// Add the actual script
	psScript.WriteString(script)
	
	return exec.Command("powershell", "-NoProfile", "-Command", psScript.String())
}