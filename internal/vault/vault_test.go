package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAddLineIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test_sentinel_vault.log")

	vault, err := NewTerminalData(logPath)
	if err != nil {
		t.Fatalf("failed to initialize test vault structures: %v", err)
	}
	defer vault.Close()

	input := "root@prod:~# curl http://10.0.0.1?token=SuperSecret123"
	expected := "curl http://[IP_REDACTED]?token=[REDACTED_SECRET]"

	vault.AddLine(input)

	if len(vault.Memory) != 1 || vault.Memory[0] != expected {
		t.Errorf("Memory buffer state invalid. Got: %v", vault.Memory)
	}

	fileBytes, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read vault file: %v", err)
	}

	expectedFileContent := expected + "\n"
	if string(fileBytes) != expectedFileContent {
		t.Errorf("Disk file content invalid. Expected %q, got %q", expectedFileContent, string(fileBytes))
	}
}
