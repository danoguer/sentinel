package main

import (
	"Sentinel/AI"
	"Sentinel/process"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func main() {
	_ = godotenv.Overload("secrets/.env")
	flag.Parse()

	userArgs := flag.Args()

	if len(userArgs) > 0 {
		if userArgs[0] == "history" || userArgs[0] == "vault" {
			printVaultHistory()
			return
		}
	}

	vault, err := process.NewTerminalData("/tmp/sentinel_vault.log")
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
	fmt.Println("🚀 Sentinel Daemon starting...")

	_ = os.Remove("/tmp/sentinel.sock")

	go process.StartSocketListener(vault)

	fmt.Println("🛡️  Sentinel Warehouse open at /tmp/sentinel.sock")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	fmt.Println("\n🛡️ Sentinel: Shutting down...")
}

func printVaultHistory() {
	vaultPath := "/tmp/sentinel_vault.log"
	data := getRecentVaultHistory(vaultPath)

	if data == "" {
		fmt.Println("🛡️ Sentinel: The vault is currently empty.")
		return
	}

	fmt.Println("🛡️ --- SENTINEL VAULT HISTORY --- 🛡️")
	fmt.Println(data)
	fmt.Println("------------------------------------")
}

func getRecentVaultHistory(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return ""
	}

	if stat.Size() < 4000 {
		bytes, _ := os.ReadFile(filePath)
		return string(bytes)
	}

	file.Seek(-4000, 2)
	buffer := make([]byte, 4000)
	file.Read(buffer)

	return string(buffer)
}
