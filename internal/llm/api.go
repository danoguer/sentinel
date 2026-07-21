package llm

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"Sentinel/pkg/transport"
	"google.golang.org/genai"
)

func AnalyzeContext(apiKey string, envelope transport.ContextEnvelope, vaultMemory []string) (string, error) {
	ctx := context.Background()
	currentHostTime := time.Now().Format("Monday, 02-Jan-2006 15:04:05 MST")

	sentinelPrompt := fmt.Sprintf(`You are Sentinel, an elite SRE CLI tool.
- Current host system time: %s
- Always respond in the language used by the user.
- CRITICAL: If the user asks about a file, script, or configuration, you MUST use the 'analyze_local_environment' tool with 'files_to_read' to inspect its contents first. Never assume a file is missing without checking via the tool.

CRITICAL ULTRA-BREVITY RULES:
1. Max 1 short sentence for the core diagnosis. Do not describe healthy components, only the absolute failure.
2. Max 3 actionable, hyper-condensed troubleshooting steps using a numbered list (1., 2., 3.).
3. Cut all narrative filler. Keep descriptions under 8 words per step and go straight to the technical action.
4. Terminal commands must be written clearly on their own lines or wrapped in backticks.
5. Absolute ban on architectural essays, introductory text, or subheadings.`, currentHostTime)

	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
	if err != nil {
		return "", fmt.Errorf("failed to create client: %v", err)
	}

	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{Role: "system", Parts: []*genai.Part{{Text: sentinelPrompt}}},
		Temperature:       genai.Ptr(float32(0.1)),
		Tools: []*genai.Tool{{
			FunctionDeclarations: []*genai.FunctionDeclaration{
				{
					Name:        "analyze_local_environment",
					Description: "Fetches extra context like system stats, terminal history log or reads files from the workspace.",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"fetch_terminal_vault": {
								Type:        genai.TypeBoolean,
								Description: "Set to true to read the sanitized terminal log history.",
							},
							"fetch_system_telemetry": {
								Type:        genai.TypeBoolean,
								Description: "Set to true to read system load, hostname and OS distribution data.",
							},
							"files_to_read": {
								Type:        genai.TypeArray,
								Description: "List of file paths or file names to read from the project or system for analysis.",
								Items: &genai.Schema{
									Type: genai.TypeString,
								},
							},
						},
					},
				},
			},
		}},
	}

	var userPromptBuilder strings.Builder
	userPromptBuilder.WriteString(fmt.Sprintf("User Question: %s\n", envelope.Question))
	userPromptBuilder.WriteString(fmt.Sprintf("Current Working Directory (CWD): %s\n", envelope.CWD))

	if envelope.Target != "generic" && envelope.Target != "" {
		userPromptBuilder.WriteString(fmt.Sprintf("\n--- [DETERMINISTIC CONTEXT FOR TARGET: %s] ---\n", strings.ToUpper(envelope.Target)))

		if len(envelope.Context.Commands) > 0 {
			userPromptBuilder.WriteString("\n[Executed Diagnostics Commands]:\n")
			for _, cmd := range envelope.Context.Commands {
				userPromptBuilder.WriteString(fmt.Sprintf("$ %s\n%s\n", cmd.Name, cmd.Output))
			}
		}

		if len(envelope.Context.Logs) > 0 {
			userPromptBuilder.WriteString("\n[Collected Service Logs]:\n")
			for _, log := range envelope.Context.Logs {
				userPromptBuilder.WriteString(fmt.Sprintf("[%s] %s\n", log.Source, log.Line))
			}
		}
		userPromptBuilder.WriteString("-----------------------------------------------\n")
	}

	if len(envelope.Context.Files) > 0 {
		userPromptBuilder.WriteString("\n[Inspected Configuration/Error Files]:\n")
		for _, file := range envelope.Context.Files {
			userPromptBuilder.WriteString(fmt.Sprintf("--- File: %s ---\n%s\n", file.Path, file.Content))
		}
	}

	var history []*genai.Content
	history = append(history, &genai.Content{
		Role:  "user",
		Parts: []*genai.Part{{Text: userPromptBuilder.String()}},
	})

	executeRequestWithRetry := func(currentHistory []*genai.Content) (*genai.GenerateContentResponse, error) {
		var lastErr error
		baseDelay := 2 * time.Second
		maxAttempts := 5

		for i := 0; i < maxAttempts; i++ {
			res, err := client.Models.GenerateContent(ctx, "gemini-3.1-flash-lite", currentHistory, config)
			if err == nil {
				return res, nil
			}
			lastErr = err
			fmt.Printf("⚠️  API error, retrying (%d/%d): %v\n", i+1, maxAttempts, err)
			time.Sleep((baseDelay * time.Duration(1<<i)) + (time.Duration(rand.Intn(400)) * time.Millisecond))
		}
		return nil, lastErr
	}

	result, err := executeRequestWithRetry(history)
	if err != nil {
		return "", fmt.Errorf("API request failed after retries: %v", err)
	}

	history = append(history, result.Candidates[0].Content)
	part := result.Candidates[0].Content.Parts[0]

	for part.FunctionCall != nil {
		funcName := part.FunctionCall.Name
		var toolResult string

		if funcName == "analyze_local_environment" {
			fmt.Printf("🛡️  Sentinel: Gemini requested auxiliary context (Files/Telemetry/Vault). Fulfilling...\n")
			toolResult = ExecuteAnalyzeLocalEnvironment(part.FunctionCall.Args)
		} else {
			toolResult = fmt.Sprintf("Error: Tool '%s' not implemented.", funcName)
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
			return "", fmt.Errorf("Failed to send tool data: %v", err)
		}

		history = append(history, toolResponseContent)
		history = append(history, result.Candidates[0].Content)
		part = result.Candidates[0].Content.Parts[0]
	}

	if part.Text != "" {
		return part.Text, nil
	}
	return "Process finalized but no readable text was returned.", nil
}
