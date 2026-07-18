package sanitize

import (
	"regexp"
	"strings"
)

var (
	ansiRegex   = regexp.MustCompile(`\x1b\[[0-9;]*[mK]`)
	promptRegex = regexp.MustCompile(`^.*?@.*?:.*?[\$#]\s*`)
	ipRegex     = regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)
	secretRegex = regexp.MustCompile(`(?i)(api_key|apikey|password|pass|secret|token)[=:\s]+([a-zA-Z0-9\-_=]{8,})`)
	jwtRegex    = regexp.MustCompile(`eyJ[a-zA-Z0-9_-]{10,}\.[a-zA-Z0-9_-]{10,}\.[a-zA-Z0-9_-]{10,}`)
)


func SanitizeLine(line string) (string, bool) {
	line = ansiRegex.ReplaceAllString(line, "")
	line = promptRegex.ReplaceAllString(line, "")

	beforeSecurity := line

	line = ipRegex.ReplaceAllString(line, "[IP_REDACTED]")
	line = secretRegex.ReplaceAllString(line, "$1=[REDACTED_SECRET]")
	line = jwtRegex.ReplaceAllString(line, "[JWT_REDACTED]")

	line = strings.TrimSpace(line)
	beforeSecurity = strings.TrimSpace(beforeSecurity)

	redacted := line != beforeSecurity

	return line, redacted
}
