package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFileLock(t *testing.T) {
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "test.lock")

	if err := acquireFileLock(lockPath, 1*time.Minute); err != nil {
		t.Fatalf("first acquireFileLock failed: %v", err)
	}

	// 再次获取应失败
	if err := acquireFileLock(lockPath, 1*time.Minute); err == nil {
		t.Fatalf("expected second acquireFileLock to fail, but succeeded")
	}

	releaseFileLock(lockPath)

	// 释放后重新获取应成功
	if err := acquireFileLock(lockPath, 1*time.Minute); err != nil {
		t.Fatalf("acquireFileLock after release failed: %v", err)
	}
	releaseFileLock(lockPath)
}

func TestRotateLogFile(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "cron_history.log")

	// 构造 3000 行日志
	var sb strings.Builder
	for i := 1; i <= 3000; i++ {
		sb.WriteString("log line entry\n")
	}
	content := []byte(sb.String())
	if err := os.WriteFile(logPath, content, 0644); err != nil {
		t.Fatalf("writeFile failed: %v", err)
	}

	// 触发轮转（设定阈值极小）
	rotateLogFile(logPath, 100)

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("readFile failed: %v", err)
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) > 2001 {
		t.Errorf("rotateLogFile expected <= 2001 lines, got %d", len(lines))
	}
}

func TestCopyFile(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "src.txt")
	dst := filepath.Join(tmpDir, "dst.txt")

	testContent := "hello mihomo safe copy"
	if err := os.WriteFile(src, []byte(testContent), 0644); err != nil {
		t.Fatalf("write src failed: %v", err)
	}

	if err := copyFile(src, dst); err != nil {
		t.Fatalf("copyFile failed: %v", err)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst failed: %v", err)
	}

	if string(data) != testContent {
		t.Errorf("copied content mismatch: got %q, want %q", string(data), testContent)
	}
}
