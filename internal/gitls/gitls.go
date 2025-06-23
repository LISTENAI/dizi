package gitls

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// LineEnding represents the type of line ending.
type LineEnding string

const (
	// LF represents the a line feed ending.
	LF LineEnding = "lf"
	// CRLF represents a carriage return line feed ending.
	CRLF LineEnding = "crlf"
)

// ListFilesOptions holds the optional parameters for the ListFiles function.
type ListFilesOptions struct {
	Glob           string
	IncludeIgnored bool
	Directory      string // Working directory to run the command in.
}

// ListFilesOption defines a function that modifies ListFilesOptions.
type ListFilesOption func(*ListFilesOptions)

// WithGlob sets the glob pattern for filtering files.
func WithGlob(pattern string) ListFilesOption {
	return func(opts *ListFilesOptions) {
		opts.Glob = pattern
	}
}

// WithIncludeIgnored makes the listing include files ignored by git.
func WithIncludeIgnored() ListFilesOption {
	return func(opts *ListFilesOptions) {
		opts.IncludeIgnored = true
	}
}

// WithDirectory sets the working directory for the git command.
// If not set, it defaults to the current working directory.
func WithDirectory(dir string) ListFilesOption {
	return func(opts *ListFilesOptions) {
		opts.Directory = dir
	}
}

// ListFiles lists files in a git repository, similar to `git ls-files`.
// It gracefully handles cases where the directory is not a git repository
// by creating a temporary one for the command's execution.
func ListFiles(options ...ListFilesOption) ([]string, error) {
	opts := &ListFilesOptions{} // Default options
	for _, option := range options {
		option(opts)
	}

	workDir, cleanup, err := getGitArgsAndCleanup(opts.Directory)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	args := []string{"ls-files", "--cached", "--others"}

	if opts.Glob != "" {
		args = append(args, opts.Glob)
	}
	if !opts.IncludeIgnored {
		args = append(args, "--exclude-standard")
	}

	cmd := exec.Command("git", args...)
	if workDir != "" {
		cmd.Dir = workDir
	} else if opts.Directory != "" {
		cmd.Dir = opts.Directory
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git command failed: %w\nOutput: %s", err, string(output))
	}

	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(files) == 1 && files[0] == "" {
		return []string{}, nil // Return empty slice for no files found
	}
	return files, nil
}

// DetectLineEndings detects the dominant line ending style (LF or CRLF) in the repository.
func DetectLineEndings(directory ...string) (LineEnding, error) {
	dir := ""
	if len(directory) > 0 {
		dir = directory[0]
	}

	workDir, cleanup, err := getGitArgsAndCleanup(dir)
	if err != nil {
		return "", err
	}
	defer cleanup()

	args := []string{"ls-files", "--cached", "--others", "--exclude-standard", "--eol"}

	cmd := exec.Command("git", args...)
	if workDir != "" {
		cmd.Dir = workDir
	} else if dir != "" {
		cmd.Dir = dir
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git command failed: %w\nOutput: %s", err, string(output))
	}

	return parseLineEndings(string(output)), nil
}

// getGitArgsAndCleanup checks for a .git directory. If it doesn't exist, it creates
// a temporary git repository and returns the necessary working directory and a cleanup function.
func getGitArgsAndCleanup(dir string) (workDir string, cleanup func(), err error) {
	// Default cleanup is a no-op
	cleanup = func() {}

	// Check if git is available
	if _, err := exec.LookPath("git"); err != nil {
		return "", cleanup, fmt.Errorf("this tool requires git to be installed and available in the PATH")
	}

	// Check for .git directory in the target path
	gitDirPath := ".git"
	if dir != "" {
		gitDirPath = filepath.Join(dir, ".git")
	}

	_, err = os.Stat(gitDirPath)
	if err == nil {
		// .git directory exists, use the provided directory or current directory
		if dir != "" {
			return dir, cleanup, nil
		}
		return "", cleanup, nil
	}

	if !os.IsNotExist(err) {
		// Some other error occurred (e.g., permissions)
		return "", cleanup, fmt.Errorf("failed to stat .git directory: %w", err)
	}

	// .git does not exist, create a temporary repo
	tmpDir, err := os.MkdirTemp("", "tidewave-gitls-")
	if err != nil {
		return "", cleanup, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Schedule the temporary directory for removal
	cleanup = func() { _ = os.RemoveAll(tmpDir) }

	// Initialize an empty git repo in the temp directory
	cmd := exec.Command("git", "init", tmpDir)
	if err := cmd.Run(); err != nil {
		cleanup() // Clean up immediately on failure
		return "", func() {}, fmt.Errorf("failed to run 'git init' in temp directory: %w", err)
	}

	// Copy files from the original directory to the temp directory
	sourceDir := dir
	if sourceDir == "" {
		sourceDir, _ = os.Getwd()
	}
	
	if err := copyDirectory(sourceDir, tmpDir); err != nil {
		cleanup() // Clean up immediately on failure
		return "", func() {}, fmt.Errorf("failed to copy files to temp directory: %w", err)
	}

	return tmpDir, cleanup, nil
}

// parseLineEndings analyzes the output of `git ls-files --eol` to determine the dominant line ending.
func parseLineEndings(result string) LineEnding {
	lfCount, crlfCount := 0, 0

	lines := strings.Split(strings.TrimSpace(result), "\n")
	for _, line := range lines {
		// git eol format: i/<eol> w/<eol> attr/<attrs> <path>
		// Example: i/lf    w/lf    attr/text=auto        README.md
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		// The second field contains the working-tree eol info
		workingTreeEol := fields[1]
		switch workingTreeEol {
		case "w/lf":
			lfCount++
		case "w/crlf":
			crlfCount++
			// "w/mixed" and others are ignored, just like in the Elixir version.
		}
	}

	if lfCount >= crlfCount {
		return LF
	}
	return CRLF
}

// copyDirectory recursively copies a directory and its contents
func copyDirectory(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip .git directories to avoid conflicts
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		// Calculate destination path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		// Skip if it's the source directory itself
		if path == src {
			return nil
		}

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return copyFile(path, dstPath, info.Mode())
	})
}

// copyFile copies a single file
func copyFile(src, dst string, mode os.FileMode) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = srcFile.Close() }()

	// Create directory if it doesn't exist
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return err
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = dstFile.Close() }()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return os.Chmod(dst, mode)
}
