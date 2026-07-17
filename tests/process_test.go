package tests

import (
	"os"
	"path/filepath"
	"testing"
	"Sentinel/process"
)

func TestAddLineSanitization(t *testing.T) {

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Pass unchanged if string is clean",
			input:    "ls -la /var/log/nginx",
			expected: "ls -la /var/log/nginx",
		},
		{
			name:     "Strip ANSI escape color codes",
			input:    "\x1b[32mhello\x1b[0m world",
			expected: "hello world",
		},
		{
			name:     "Strip shell prompt prefixes (user@host)",
			input:    "admin@production-server:~$ systemctl restart docker",
			expected: "systemctl restart docker",
		},
		{
			name:     "Redact IPv4 addresses securely",
			input:    "curl http://192.168.1.105:8080/api/v1/status",
			expected: "curl http://[IP_REDACTED]:8080/api/v1/status",
		},
		{
			name:     "Redact passwords inline assignments",
			input:    "mysql -u root -p password=MySuperSecretPassword123",
			expected: "mysql -u root -p password=[REDACTED_SECRET]",
		},
		{
			name:     "Redact environment variable API keys",
			input:    "export APP_API_KEY=AIzaSyA1B2C3D4E5F6G7",
			expected: "export APP_API_KEY=[REDACTED_SECRET]",
		},
		{
			name:     "Redact raw JSON Web Tokens (JWT)",
			input:    "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			expected: "Authorization: Bearer [JWT_REDACTED]",
		},
		{
			name:     "Ignore inputs that become empty strings after trimming",
			input:    "    \n\t   ",
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			tmpDir := t.TempDir()
			logPath := filepath.Join(tmpDir, "test_sentinel_vault.log")

			vault, err := process.NewTerminalData(logPath)
			if err != nil {
				t.Fatalf("failed to initialize test vault structures: %v", err)
			}
			defer vault.Close()

			vault.AddLine(tc.input)

			if tc.expected == "" {
				if len(vault.Memory) != 0 {
					t.Errorf("expected in-memory buffer to be empty, but contains: %v", vault.Memory)
				}
				return
			}

			if len(vault.Memory) == 0 {
				t.Fatalf("expected in-memory log buffer to contain 1 record, got 0")
			}
			actualMemory := vault.Memory[0]
			if actualMemory != tc.expected {
				t.Errorf("\n[Memory Buffer Mismatch]\nInput:    %s\nExpected: %s\nActual:   %s", tc.input, tc.expected, actualMemory)
			}

			fileBytes, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatalf("failed to read the generated vault log file from disk: %v", err)
			}

			expectedFileContent := tc.expected + "\n"
			if string(fileBytes) != expectedFileContent {
				t.Errorf("\n[Disk File Content Mismatch]\nExpected payload: %q\nActual payload:   %q", expectedFileContent, string(fileBytes))
			}
		})
	}
}
