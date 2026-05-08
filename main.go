package main

import (
	"Sentinel/AI"
	"Sentinel/process"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"strings"
)

func main() {
	_ = godotenv.Load("secrets/.env")

	flag.Parse()

	userArgs := flag.Args()

	if len(userArgs) > 0 {
		if userArgs[0] == "history" || userArgs[0] == "vault" {
			printVaultHistory()
			return
		}
		question := strings.Join(userArgs, " ")
		runClientMode(question)
		return
	}

	runDaemonMode()
}

func runDaemonMode() {
	fmt.Println("🚀 Sentinel Daemon starting...")

	_ = os.Remove("/tmp/sentinel.sock")

	go process.StartSocketListener()

	fmt.Println("🛡️  Sentinel Warehouse open at /tmp/sentinel.sock")
	select {}
}

func runClientMode(question string) {
	if len(flag.Args()) < 1 {
		fmt.Println("❌ Error: Please provide a question.")
		return
	}
	userQuestion := flag.Args()[0]

	history, err := os.ReadFile("/tmp/sentinel_vault.log")
	if err != nil {
		fmt.Println("⚠️  Note: No history found yet. Asking Gemini without context.")
	}

	fullPrompt := fmt.Sprintf("CONTEXT (Terminal History):\n%s\n\nUSER QUESTION: %s", string(history), userQuestion)

	AI.Run(fullPrompt)
}

func printVaultHistory() {
	vaultPath := "/tmp/sentinel_vault.log"

	// We can reuse the safe reader we made earlier!
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
