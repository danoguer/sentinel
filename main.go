package main

import (
	"Sentinel/AI"
	"Sentinel/process"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Overload("secrets/.env")
	flag.Parse()

	userArgs := flag.Args()

	uid := os.Getuid()
	vaultPath := fmt.Sprintf("/tmp/sentinel_vault_%d.log", uid)

	vault, err := process.NewTerminalData(vaultPath)
	if err != nil {
		log.Fatalf("Critical: %v", err)
	}
	defer vault.Close()

	if len(userArgs) > 0 {
		question := strings.Join(userArgs, " ")
		fmt.Println("🛡️ Sentinel: Initializing 1-shot sequence...")

		AI.Run(question)
		return
	}

	runDaemonMode(vault)
}

func runDaemonMode(vault *process.TerminalData) {
	uid := os.Getuid()
	socketPath := fmt.Sprintf("/tmp/sentinel_%d.sock", uid)

	fmt.Println("🚀 Sentinel Daemon starting...")

	_ = os.Remove(socketPath)

	go process.StartSocketListener(vault, socketPath)

	fmt.Printf("🛡️  Sentinel Warehouse open at %s\n", socketPath)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	fmt.Println("\n🛡️ Sentinel: Shutting down...")
}
