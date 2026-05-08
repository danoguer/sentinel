package AI

import (
	"context"
	"fmt"
	"log"
	"os"

	"google.golang.org/genai"
)

func Run(prompt string) {
	apiKey := os.Getenv("AI_API_KEY")
	if apiKey == "" {
		log.Fatal("Error: AI_API_KEY is not set in secrets/.env")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	systemInstruction := &genai.Content{
		Parts: []*genai.Part{
			{
				Text: "You are a Senior SRE. Your answers must be concise, terminal-focused, and prioritize Linux/Kubernetes best practices. Always provide copy-pasteable commands when applicable.",
			},
		},
	}

	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-3-flash-preview",
		genai.Text(prompt),
		&genai.GenerateContentConfig{
			SystemInstruction: systemInstruction,
			Temperature:       genai.Ptr(float32(0.2)),
		},
	)

	if err != nil {
		log.Fatalf("API request failed: %v", err)
	}

	if len(result.Candidates) > 0 && len(result.Candidates[0].Content.Parts) > 0 {
		fmt.Println("\n🤖 Sentinel AI:")
		fmt.Println(result.Candidates[0].Content.Parts[0].Text)
	}
}
