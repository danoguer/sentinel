package process

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
	"Sentinel/metrics"
)

var (
	ansiRegex   = regexp.MustCompile(`\x1b\[[0-9;]*[mK]`)
	promptRegex = regexp.MustCompile(`^.*?@.*?:.*?[\$#]\s*`)
	ipRegex     = regexp.MustCompile(`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)
	secretRegex = regexp.MustCompile(`(?i)(api_key|apikey|password|pass|secret|token)[=:\s]+([a-zA-Z0-9\-_=]{8,})`)
	jwtRegex    = regexp.MustCompile(`eyJ[a-zA-Z0-9_-]{10,}\.[a-zA-Z0-9_-]{10,}\.[a-zA-Z0-9_-]{10,}`)
)

type TerminalData struct {
	sync.Mutex
	Memory []string
	vault  *os.File
}

func NewTerminalData(logPath string) (*TerminalData, error) {
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &TerminalData{
		Memory: make([]string, 0, 50),
		vault:  f,
	}, nil
}

func (w *TerminalData) Close() {
	w.Lock()
	defer w.Unlock()
	if w.vault != nil {
		w.vault.Close()
	}
}

func (w *TerminalData) AddLine(line string) {
	startTime := time.Now()

	defer func() {
		metrics.LinesProcessed.Inc()
		metrics.SanitizationDuration.Observe(time.Since(startTime).Seconds())
	}()

	line = ansiRegex.ReplaceAllString(line, "")
	line = promptRegex.ReplaceAllString(line, "")

	beforeSecurity := line

	line = ipRegex.ReplaceAllString(line, "[IP_REDACTED]")
	line = secretRegex.ReplaceAllString(line, "$1=[REDACTED_SECRET]")
	line = jwtRegex.ReplaceAllString(line, "[JWT_REDACTED]")
	line = strings.TrimSpace(line)

	if line != beforeSecurity {
		metrics.SecretsRedacted.Inc()
	}

	if line == "" {
		return
	}

	w.Lock()
	defer w.Unlock()

	w.Memory = append(w.Memory, line)
	if len(w.Memory) > 50 {
		w.Memory = w.Memory[1:]
	}

	if _, err := w.vault.WriteString(line + "\n"); err != nil {
		fmt.Fprintf(os.Stderr, "sentinel: failed to write to vault: %v\n", err)
	}
}

func StartSocketListener(vault *TerminalData, socketPath string) {
	_ = os.Remove(socketPath)

	sock, err := net.Listen("unix", socketPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sentinel: failed to start socket listener: %v\n", err)
		return
	}
	defer sock.Close()

	for {
		conn, err := sock.Accept()
		if err != nil {
			continue
		}

		go handleConnection(conn, vault)
	}
}

func handleConnection(c net.Conn, vault *TerminalData) {
	defer c.Close()

	scanner := bufio.NewScanner(c)

	for scanner.Scan() {
		vault.AddLine(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "sentinel: error reading from socket: %v\n", err)
	}
}
