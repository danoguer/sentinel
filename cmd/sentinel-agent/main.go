package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"Sentinel/internal/vault"
	"github.com/joho/godotenv"
)

var (
	Version = "2.0.0"
	Commit  = "unknown"
	Date    = "unknown"
)

type AgentConfig struct {
	UID        int
	VaultPath  string
	SocketPath string
}

func loadConfig() AgentConfig {
	_ = godotenv.Overload("secrets/.env")

	uid := os.Getuid()

	vaultPath := os.Getenv("SENTINEL_VAULT_PATH")
	if vaultPath == "" {
		vaultPath = fmt.Sprintf("/tmp/sentinel_vault_%d.log", uid)
	}

	socketPath := os.Getenv("SENTINEL_SOCKET_PATH")
	if socketPath == "" {
		socketPath = fmt.Sprintf("/tmp/sentinel_%d.sock", uid)
	}

	return AgentConfig{
		UID:        uid,
		VaultPath:  vaultPath,
		SocketPath: socketPath,
	}
}

func startHTTPServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("200 OK"))
	})

	go func() {
		fmt.Println("📊 Telemetry and Health server listening on :2112")
		if err := http.ListenAndServe(":2112", mux); err != nil {
			log.Printf("HTTP Server Error: %v", err)
		}
	}()
}

func waitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n🛡️  Sentinel Agent: Shutting down...")
}

func main() {
	cfg := loadConfig()

	v, err := vault.NewTerminalData(cfg.VaultPath)
	if err != nil {
		log.Fatalf("Critical: %v", err)
	}
	defer v.Close()

	fmt.Println("🚀 Sentinel Agent Daemon starting...")

	go vault.StartSocketListener(v, cfg.SocketPath)
	fmt.Printf("🛡️  Sentinel Warehouse open at %s\n", cfg.SocketPath)

	startHTTPServer()

	waitForShutdown()
}
