package vault

import (
	"fmt"
	"os"
	"sync"
	"time"

	"Sentinel/pkg/sanitize"
	"Sentinel/internal/metrics"
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

func (w *TerminalData) AddLine(rawLine string) {
	startTime := time.Now()

	defer func() {
		metrics.LinesProcessed.Inc()
		metrics.SanitizationDuration.Observe(time.Since(startTime).Seconds())
	}()

	sanitizedLine, redacted := sanitize.SanitizeLine(rawLine)

	if redacted {
		metrics.SecretsRedacted.Inc()
	}

	if sanitizedLine == "" {
		return
	}

	w.Lock()
	defer w.Unlock()

	w.Memory = append(w.Memory, sanitizedLine)
	if len(w.Memory) > 50 {
		w.Memory = w.Memory[1:]
	}

	if _, err := w.vault.WriteString(sanitizedLine + "\n"); err != nil {
		fmt.Fprintf(os.Stderr, "sentinel: failed to write to vault: %v\n", err)
	}
}
