package AI

import (
	"context"
	"fmt"
	"google.golang.org/genai"
	"log"
	"os"
	"time"
)

// Run sends prompt to Gemini and executes any tool calls it requests
// until a final text response is produced (the agent loop).
func Run(prompt string) {

	// Instructs the model to behave as a minimal CLI tool, avoiding verbose explanations.
	const sentinelPrompt = `You are a CLI utility. 
	- Answer directly, as shorter you could. 
	- Use tools (search, read, history) only if the question cannot be answered otherwise. 
	- Output ONLY the requested information.`

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

	// Temperature 0.2: keeps answer focused which matters for a CLI where consistency is expected.
	chat, err := client.Chats.Create(ctx, "gemini-3-flash-preview", &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{Role: "system", Parts: []*genai.Part{{Text: sentinelPrompt}}},
		Temperature:       genai.Ptr(float32(0.2)),
		Tools: []*genai.Tool{{
			FunctionDeclarations: []*genai.FunctionDeclaration{
				{Name: "read_file", Description: "Reads a file.", Parameters: &genai.Schema{Type: genai.TypeObject, Properties: map[string]*genai.Schema{"filepath": {Type: genai.TypeString}}}},
				{Name: "search_file", Description: "Searches for a file.", Parameters: &genai.Schema{Type: genai.TypeObject, Properties: map[string]*genai.Schema{"filename": {Type: genai.TypeString}, "start_dir": {Type: genai.TypeString}}}},
				{Name: "get_terminal_history", Description: "Fetches recent terminal history.", Parameters: &genai.Schema{Type: genai.TypeObject}},
			},
		}},
	}, nil)
	if err != nil {
		log.Fatalf("Failed to create chat: %v", err)
	}

	// Wraps every API call with 3 retries (2s apart) to handle transient rate limits or network errors.
	sendMessageWithRetry := func(msg genai.Part) (*genai.GenerateContentResponse, error) {
		var lastErr error
		for i := 0; i < 3; i++ {
			res, err := chat.SendMessage(ctx, msg)
			if err == nil {
				return res, nil
			}
			lastErr = err
			fmt.Printf("⚠️  API error, retrying (%d/3): %v\n", i+1, err)
			time.Sleep(2 * time.Second)
		}
		return nil, lastErr
	}

	result, err := sendMessageWithRetry(genai.Part{Text: prompt})
	if err != nil {
		log.Fatalf("API request failed after retries: %v", err)
	}
	part := result.Candidates[0].Content.Parts[0]

	// Resolve function calls until the model provides a final text response.
	for part.FunctionCall != nil {
		funcName := part.FunctionCall.Name
		var toolResult string

		switch funcName {
		case "read_file":
			toolResult = localReadFile(part.FunctionCall.Args["filepath"].(string))
		case "search_file":
			toolResult = localSearchFile(part.FunctionCall.Args["filename"].(string), part.FunctionCall.Args["start_dir"].(string))
		case "get_terminal_history":
			fmt.Printf("📜 AI is reviewing terminal history...\n")
			toolResult = localGetHistory()
		}

		// Return tool output to the model for final synthesis.
		result, err = sendMessageWithRetry(genai.Part{
			FunctionResponse: &genai.FunctionResponse{
				Name:     funcName,
				Response: map[string]any{"content": toolResult},
			},
		})
		if err != nil {
			log.Fatalf("Failed to send tool data after retries: %v", err)
		}
		part = result.Candidates[0].Content.Parts[0]
	}

	if part.Text != "" {
		fmt.Println("\n🤖 Sentinel AI:")
		fmt.Println(part.Text)
	}
}
