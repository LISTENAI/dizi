package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestWorkdirChange 测试工作目录切换功能
func TestWorkdirChange(t *testing.T) {
	// 保存原始工作目录
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	defer func() {
		// 恢复原始工作目录
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("Failed to restore working directory: %v", err)
		}
	}()

	// 创建临时测试目录
	tempDir, err := os.MkdirTemp("", "dizi_workdir_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("Failed to remove temp directory: %v", err)
		}
	}()

	// 测试切换到临时目录
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// 验证工作目录已切换
	currentWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory after change: %v", err)
	}

	expectedPath, err := filepath.Abs(tempDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path of temp directory: %v", err)
	}

	// 解析符号链接以获得真实路径
	expectedRealPath, err := filepath.EvalSymlinks(expectedPath)
	if err != nil {
		t.Fatalf("Failed to evaluate symlinks for expected path: %v", err)
	}

	currentRealPath, err := filepath.EvalSymlinks(currentWd)
	if err != nil {
		t.Fatalf("Failed to evaluate symlinks for current path: %v", err)
	}

	if currentRealPath != expectedRealPath {
		t.Errorf("Working directory not changed correctly. Expected: %s, Got: %s", expectedRealPath, currentRealPath)
	}
}

// TestWorkdirChangeError 测试切换到不存在目录的错误处理
func TestWorkdirChangeError(t *testing.T) {
	// 尝试切换到不存在的目录
	nonExistentDir := "/this/directory/does/not/exist"
	err := os.Chdir(nonExistentDir)
	
	// 应该返回错误
	if err == nil {
		t.Error("Expected error when changing to non-existent directory, but got nil")
	}
}

// TestWorkdirRelativeToAbsolute 测试相对路径转换为绝对路径
func TestWorkdirRelativeToAbsolute(t *testing.T) {
	// 保存原始工作目录
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	defer func() {
		// 恢复原始工作目录
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("Failed to restore working directory: %v", err)
		}
	}()

	// 创建临时测试目录结构
	tempDir, err := os.MkdirTemp("", "dizi_workdir_relative_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("Failed to remove temp directory: %v", err)
		}
	}()

	subDir := filepath.Join(tempDir, "subdir")
	err = os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// 先切换到临时目录
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// 使用相对路径切换到子目录
	err = os.Chdir("subdir")
	if err != nil {
		t.Fatalf("Failed to change to subdirectory using relative path: %v", err)
	}

	// 验证当前目录是子目录
	currentWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}

	expectedPath, err := filepath.Abs(subDir)
	if err != nil {
		t.Fatalf("Failed to get absolute path of subdirectory: %v", err)
	}

	// 解析符号链接以获得真实路径
	expectedRealPath, err := filepath.EvalSymlinks(expectedPath)
	if err != nil {
		t.Fatalf("Failed to evaluate symlinks for expected path: %v", err)
	}

	currentRealPath, err := filepath.EvalSymlinks(currentWd)
	if err != nil {
		t.Fatalf("Failed to evaluate symlinks for current path: %v", err)
	}

	if currentRealPath != expectedRealPath {
		t.Errorf("Working directory not changed to subdirectory correctly. Expected: %s, Got: %s", expectedRealPath, currentRealPath)
	}
}

// TestWorkdirPermissions 测试权限不够的目录切换
func TestWorkdirPermissions(t *testing.T) {
	// 在某些系统上，尝试切换到无权访问的目录
	// 这个测试在不同系统上可能表现不同，所以只做基本检查
	restrictedDirs := []string{
		"/root",     // Linux/macOS 系统上通常无权访问
		"/private",  // macOS 上的受限目录
	}

	for _, dir := range restrictedDirs {
		// 检查目录是否存在
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			continue // 目录不存在，跳过
		}

		// 尝试切换，可能会因权限不足失败
		err := os.Chdir(dir)
		if err != nil {
			// 权限错误是预期的，这是正常情况
			t.Logf("Expected permission error when accessing %s: %v", dir, err)
		} else {
			// 如果成功了，需要切换回来
			originalWd, _ := os.Getwd()
			defer func() {
				if err := os.Chdir(originalWd); err != nil {
					t.Logf("Failed to restore working directory: %v", err)
				}
			}()
		}
	}
}