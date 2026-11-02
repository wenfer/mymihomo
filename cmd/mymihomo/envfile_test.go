package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadWriteEnvFile(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, "test.env")

	data := map[string]string{
		"CONF_URL":        "https://example.com/sub",
		"CRON_EXPRESSION": "0 3 * * *",
		"MIXED_PORT":      "7892",
	}

	if err := writeEnvFile(envPath, data); err != nil {
		t.Fatalf("writeEnvFile failed: %v", err)
	}

	read, err := readEnvFile(envPath)
	if err != nil {
		t.Fatalf("readEnvFile failed: %v", err)
	}

	for k, want := range data {
		got, ok := read[k]
		if !ok || got != want {
			t.Errorf("readEnvFile key %s = %q; want %q", k, got, want)
		}
	}
}

func TestLoadPersistedEnvOverrides(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, "test_override.env")

	// 模拟已存在环境变量
	os.Setenv("WEB_ENV_FILE", envPath)
	os.Setenv("TEST_KEY_EXISTING", "original_value")

	data := map[string]string{
		"TEST_KEY_EXISTING": "overridden_value",
	}
	if err := writeEnvFile(envPath, data); err != nil {
		t.Fatalf("writeEnvFile failed: %v", err)
	}

	if err := loadPersistedEnv(); err != nil {
		t.Fatalf("loadPersistedEnv failed: %v", err)
	}

	if got := os.Getenv("TEST_KEY_EXISTING"); got != "overridden_value" {
		t.Errorf("loadPersistedEnv failed to override existing variable; got %q, want %q", got, "overridden_value")
	}
}
