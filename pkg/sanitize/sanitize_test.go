package sanitize

import (
	"testing"
)

func TestSanitizeLine(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expected         string
		expectedRedacted bool
	}{
		{
			name:             "Pass unchanged if string is clean",
			input:            "ls -la /var/log/nginx",
			expected:         "ls -la /var/log/nginx",
			expectedRedacted: false,
		},
		{
			name:             "Strip ANSI escape color codes",
			input:            "\x1b[32mhello\x1b[0m world",
			expected:         "hello world",
			expectedRedacted: false,
		},
		{
			name:             "Strip shell prompt prefixes (user@host)",
			input:            "admin@production-server:~$ systemctl restart docker",
			expected:         "systemctl restart docker",
			expectedRedacted: false,
		},
		{
			name:             "Redact IPv4 addresses securely",
			input:            "curl http://192.168.1.105:8080/api/v1/status",
			expected:         "curl http://[IP_REDACTED]:8080/api/v1/status",
			expectedRedacted: true,
		},
		{
			name:             "Redact passwords inline assignments",
			input:            "mysql -u root -p password=MySuperSecretPassword123",
			expected:         "mysql -u root -p password=[REDACTED_SECRET]",
			expectedRedacted: true,
		},
		{
			name:             "Redact environment variable API keys",
			input:            "export APP_API_KEY=AIzaSyA1B2C3D4E5F6G7",
			expected:         "export APP_API_KEY=[REDACTED_SECRET]",
			expectedRedacted: true,
		},
		{
			name:             "Redact raw JSON Web Tokens (JWT)",
			input:            "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			expected:         "Authorization: Bearer [JWT_REDACTED]",
			expectedRedacted: true,
		},
		{
			name:             "Ignore inputs that become empty strings after trimming",
			input:            "    \n\t   ",
			expected:         "",
			expectedRedacted: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			output, redacted := SanitizeLine(tc.input)

			if output != tc.expected {
				t.Errorf("\nInput:    %s\nExpected: %s\nActual:   %s", tc.input, tc.expected, output)
			}

			if redacted != tc.expectedRedacted {
				t.Errorf("Expected redacted flag to be %v, got %v", tc.expectedRedacted, redacted)
			}
		})
	}
}

func BenchmarkSanitizationOnly(b *testing.B) {
	heavyInput := "admin@prod-node-01:~$ curl -u webapp_admin:SecretPassword123 -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.SflKxwRJSMe' https://192.168.1.105/api/v1/deploy"

	b.ResetTimer()
	for b.Loop() {
		SanitizeLine(heavyInput)
	}
}
