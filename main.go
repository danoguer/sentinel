package main

import (
	"Sentinel/AI"              // 1. Matches your module name + folder
	"github.com/joho/godotenv" // 4. Needed for godotenv.Load
	"log"                      // 2. Needed for log.Printf
	"os"                       // 3. Needed for os.Getwd
)

func main() {
	path := "secrets/.env"

	// Load the .env file
	err := godotenv.Load(path)
	if err != nil {
		// This helps you debug exactly where Go is looking
		cwd, _ := os.Getwd()
		log.Printf("Note: No .env file found at %s/%s. Falling back to system env.", cwd, path)
	}

	// Call your AI logic
	ai.Run()
}
