package AI

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"google.golang.org/genai"
)

func Run(prompt string) {
	currentHostTime := time.Now().Format("Monday, 02-Jan-2006 15:04:05 MST")

	sentinelPrompt := fmt.Sprintf(`You are an advanced SRE CLI utility. 
	- Current host system time: %s
	- Answer directly, as short as possible. 
	- If you need context to answer the user's prompt (like logs, system stats, or code files), call the 'analyze_local_environment' tool.
	- Output ONLY the requested information.`, currentHostTime)

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

	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{Role: "system", Parts: []*genai.Part{{Text: sentinelPrompt}}},
		Temperature:       genai.Ptr(float32(0.2)),
		Tools: []*genai.Tool{{
			FunctionDeclarations: []*genai.FunctionDeclaration{
				{
					Name:        "analyze_local_environment",
					Description: "Fetches system stats, terminal history, or specific files. Use this to gather context before answering.",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"fetch_terminal_vault": {
								Type:        genai.TypeBoolean,
								Description: "Set to true to read the last 50 lines of the terminal log.",
							},
							"fetch_system_telemetry": {
								Type:        genai.TypeBoolean,
								Description: "Set to true to read RAM, Disk usage, and active network ports.",
							},
							"files_to_read": {
								Type:        genai.TypeArray,
								Description: "A list of specific file paths to read (e.g., ['main.go', 'ai.go']).",
								Items:       &genai.Schema{Type: genai.TypeString},
							},
						},
					},
				},
			},
		}},
	}

	var history []*genai.Content

	history = append(history, &genai.Content{
		Role:  "user",
		Parts: []*genai.Part{{Text: prompt}},
	})

	executeRequestWithRetry := func(currentHistory []*genai.Content) (*genai.GenerateContentResponse, error) {
		var lastErr error
		baseDelay := 5 * time.Second
		maxAttempts := 8

		for i := 0; i < maxAttempts; i++ {
			res, err := client.Models.GenerateContent(ctx, "gemini-3.1-flash-lite", currentHistory, config)
			if err == nil {
				return res, nil
			}
			lastErr = err

			fmt.Printf("⚠️  API error, retrying (%d/%d): %v\n", i+1, maxAttempts, err)

			delay := baseDelay * time.Duration(1<<i)
			jitter := time.Duration(rand.Intn(400)) * time.Millisecond
			time.Sleep(delay + jitter)
		}
		return nil, lastErr
	}

	result, err := executeRequestWithRetry(history)
	if err != nil {
		log.Fatalf("API request failed after retries: %v", err)
	}

	history = append(history, result.Candidates[0].Content)

	var part *genai.Part
	if len(result.Candidates) > 0 && result.Candidates[0].Content != nil && len(result.Candidates[0].Content.Parts) > 0 {
		part = result.Candidates[0].Content.Parts[0]
	} else {
		log.Fatalf("Critical: Initial API handshake returned empty content arrays.")
	}

	for part.FunctionCall != nil {
		funcName := part.FunctionCall.Name
		var toolResult string

		switch funcName {
		case "analyze_local_environment":
			fmt.Printf("🛡️  Sentinel: Gathering requested context...\n")
			toolResult = ExecuteAnalyzeLocalEnvironment(part.FunctionCall.Args)
		default:
			toolResult = fmt.Sprintf("Error: Tool '%s' is not implemented.", funcName)
		}

		toolResponseContent := &genai.Content{
			Role: "user",
			Parts: []*genai.Part{{
				FunctionResponse: &genai.FunctionResponse{
					Name:     funcName,
					Response: map[string]any{"content": toolResult},
				},
			}},
		}

		nextTurnHistory := append(history, toolResponseContent)

		result, err = executeRequestWithRetry(nextTurnHistory)
		if err != nil {
			log.Fatalf("Failed to send tool data after retries: %v", err)
		}

		history = append(history, toolResponseContent)
		history = append(history, result.Candidates[0].Content)

		if len(result.Candidates) > 0 && result.Candidates[0].Content != nil && len(result.Candidates[0].Content.Parts) > 0 {
			part = result.Candidates[0].Content.Parts[0]
		} else {
			log.Fatalf("Critical: Model returned an empty payload loop sequence.")
		}
	}

	if len(result.Candidates) > 0 && result.Candidates[0].Content != nil && len(result.Candidates[0].Content.Parts) > 0 {
		finalPart := result.Candidates[0].Content.Parts[0]
		if finalPart.Text != "" {
			fmt.Println("\n🤖 Sentinel AI:")
			fmt.Println(finalPart.Text)
		}
	} else {
		fmt.Println("\n⚠️  Sentinel AI: Process finalized but no readable text chunks were found.")
	}
}

