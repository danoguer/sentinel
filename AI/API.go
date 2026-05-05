package ai

import (
	"context"
	"fmt"
	"log"
	"os"

	"google.golang.org/genai"
)

func Run() {
	apiKey := os.Getenv("AI_API_KEY")
	if apiKey == "" {
		log.Fatal("Error: AI_API_KEY environment variable is not set")
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage: Sentinel 'your question'")
		return
	}
	userPrompt := os.Args[1]

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
		genai.Text(userPrompt),
		&genai.GenerateContentConfig{
			SystemInstruction: systemInstruction,
			Temperature:       genai.Ptr(float32(0.2)),
		},
	)
	if err != nil {
		log.Fatalf("API request failed: %v", err)
	}

	if len(result.Candidates) > 0 && len(result.Candidates[0].Content.Parts) > 0 {
		for _, part := range result.Candidates[0].Content.Parts {
			if part.Text != "" {
				fmt.Println(part.Text)
			}
		}
	} else {
		fmt.Println("AI returned an empty response.")
	}
}
