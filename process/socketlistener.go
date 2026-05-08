package process

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"
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
}

var GlobalData = &TerminalData{
	Memory: make([]string, 0, 50),
}

func (w *TerminalData) Addline(line string) {
	// Hidding every sensitive content
	line = ansiRegex.ReplaceAllString(line, "")
	line = promptRegex.ReplaceAllString(line, "")
	line = ipRegex.ReplaceAllString(line, "[IP_REDACTED]")
	line = secretRegex.ReplaceAllString(line, "$1=[REDACTED_SECRET]")
	line = jwtRegex.ReplaceAllString(line, "[JWT_REDACTED]")
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}

	// Mutex the process
	w.Lock()
	defer w.Unlock()

	// Adding the new line and truncates if it is bigger than 50 lines
	w.Memory = append(w.Memory, line)
	if len(w.Memory) > 50 {
		w.Memory = w.Memory[1:]
	}

	// Opening/Creating the log file for the terminal history
	vault, err := os.OpenFile("/tmp/sentinel_vault.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		defer vault.Close()
		_, err2 := vault.WriteString(line + "\n")
		if err2 != nil {
			fmt.Println("WriteString is not working at writing in the vault")
		}
	} else {
		fmt.Println("OpenFile couldnt open the Vault File")
	}
}

func StartSocketListener() {
	// Creating the socket path
	socketPath := "/tmp/sentinel.sock"

	// Cleaning the socket path
	_ = os.Remove(socketPath)

	// Creating the socket with the path we just cleaned
	sock, err := net.Listen("unix", socketPath)
	if err != nil {
		fmt.Println("❌ Listener Error:", err)
		return
	}
	defer sock.Close()

	for {
		// Waits for a new connections attempt from a terminal
		conn, err := sock.Accept()
		if err != nil {
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(c net.Conn) {
	defer c.Close()

	// Reads everything from the connection
	scanner := bufio.NewScanner(c)

	// Copy everything in our struct
	for scanner.Scan() {
		GlobalData.Addline(scanner.Text())
	}
}
