package llm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func resolveFilePath(requestedPath string) string {
	if _, err := os.Stat(requestedPath); err == nil {
		return requestedPath
	}

	targetName := filepath.Base(requestedPath)
	var foundPath string

	_ = filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "vendor" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() == targetName {
			foundPath = path
			return filepath.SkipDir
		}
		return nil
	})

	if foundPath != "" {
		return foundPath
	}
	return requestedPath
}

func ExecuteAnalyzeLocalEnvironment(args map[string]any) string {
	var combinedContext strings.Builder

	if reqVault, ok := args["fetch_terminal_vault"].(bool); ok && reqVault {
		combinedContext.WriteString("--- TERMINAL VAULT ---\n")
		combinedContext.WriteString(localGetHistory())
		combinedContext.WriteString("\n\n")
	}

	if reqSys, ok := args["fetch_system_telemetry"].(bool); ok && reqSys {
		combinedContext.WriteString("--- SYSTEM TELEMETRY ---\n")
		combinedContext.WriteString(GetLocalSystemSnapshot())
		combinedContext.WriteString("\n\n")
	}

	if filesArg, ok := args["files_to_read"].([]any); ok {
		for _, fileAny := range filesArg {
			if rawPath, isStr := fileAny.(string); isStr {
				actualPath := resolveFilePath(rawPath)

				stat, err := os.Stat(actualPath)
				if err != nil || stat.IsDir() {
					combinedContext.WriteString(fmt.Sprintf("--- FILE: %s [SKIPPED: Not found or is a directory] ---\n", actualPath))
					continue
				}

				if stat.Size() > 50*1024 {
					combinedContext.WriteString(fmt.Sprintf("--- FILE: %s [SKIPPED: Too large (%d bytes)] ---\n", actualPath, stat.Size()))
					continue
				}

				content, err := os.ReadFile(actualPath)
				if err == nil {
					combinedContext.WriteString(fmt.Sprintf("--- FILE: %s ---\n```\n%s\n```\n\n", actualPath, string(content)))
				}
			}
		}
	}

	result := combinedContext.String()
	if result == "" {
		return "No specific context was found or files were skipped."
	}
	return result
}
