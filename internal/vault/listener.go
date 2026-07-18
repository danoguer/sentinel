package vault

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"Sentinel/internal/aws"
	"Sentinel/internal/llm"
	"Sentinel/pkg/transport"
)

func StartSocketListener(vault *TerminalData, socketPath string) {
	_ = os.Remove(socketPath)

	sock, err := net.Listen("unix", socketPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sentinel agent: failed to start socket listener: %v\n", err)
		return
	}
	defer sock.Close()

	// CAMBIO AQUÍ: Forzar permisos 0777 al archivo del socket recién creado
	if err := os.Chmod(socketPath, 0777); err != nil {
		fmt.Fprintf(os.Stderr, "sentinel agent: warning: failed to change socket permissions: %v\n", err)
	}

	for {
		conn, err := sock.Accept()
		if err != nil {
			continue
		}

		go handleIPCConnection(conn, vault)
	}
}

func handleIPCConnection(c net.Conn, vault *TerminalData) {
	defer c.Close()
	startTime := time.Now()

	var req transport.ContextEnvelope
	decoder := json.NewDecoder(c)
	if err := decoder.Decode(&req); err != nil {
		fmt.Fprintf(os.Stderr, "sentinel agent: failed to decode envelope: %v\n", err)
		sendError(c, "invalid JSON payload")
		return
	}

	fmt.Printf("📥 Agent received command: %s (target: %s)\n", req.Command, req.Question)

	var answer string
	if req.Command == "explain" {
		apiKey := aws.GetAPIKey()
		if apiKey == "" {
			sendError(c, "AI_API_KEY could not be loaded from env or AWS Secrets Manager")
			return
		}

		fmt.Printf("🧠 Consulting Gemini (Passive Diagnostic Mode)...\n")

		vault.Lock()
		vaultCopy := make([]string, len(vault.Memory))
		copy(vaultCopy, vault.Memory)
		vault.Unlock()

		geminiResponse, err := llm.AnalyzeContext(apiKey, req, vaultCopy)
		if err != nil {
			sendError(c, fmt.Sprintf("LLM Error: %v", err))
			return
		}
		answer = geminiResponse

	} else if req.Command == "status" {
		answer = "Agent is healthy and listening."
	} else {
		answer = "Unknown command."
	}

	resp := transport.AgentResponse{
		Answer:     answer,
		DurationMS: time.Since(startTime).Milliseconds(),
	}

	encoder := json.NewEncoder(c)
	_ = encoder.Encode(resp)
}

func sendError(c net.Conn, errMsg string) {
	resp := transport.AgentResponse{Error: errMsg}
	_ = json.NewEncoder(c).Encode(resp)
}
