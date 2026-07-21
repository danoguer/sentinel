package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"Sentinel/internal/render"
	"Sentinel/pkg/contextbuilder"
	"Sentinel/pkg/transport"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sentinel",
	Short: "Sentinel is an AI-powered Context Engine",
}

func attachLocalPaths(args []string, ctx *contextbuilder.Context) {
	for _, arg := range args {
		info, err := os.Stat(arg)
		if err != nil {
			continue
		}

		if info.IsDir() {
			_ = filepath.WalkDir(arg, func(path string, d os.DirEntry, walkErr error) error {
				if walkErr != nil {
					return nil
				}
				if d.IsDir() {
					name := d.Name()
					if name == ".git" || name == "vendor" || name == "node_modules" || name == "bin" || name == ".idea" {
						return filepath.SkipDir
					}
					return nil
				}
				if stat, statErr := d.Info(); statErr == nil && stat.Size() < 50*1024 {
					if content, readErr := os.ReadFile(path); readErr == nil {
						ctx.Files = append(ctx.Files, contextbuilder.FileEntry{
							Path:    path,
							Content: string(content),
						})
					}
				}
				return nil
			})
		} else {
			if content, readErr := os.ReadFile(arg); readErr == nil {
				ctx.Files = append(ctx.Files, contextbuilder.FileEntry{
					Path:    arg,
					Content: string(content),
				})
			}
		}
	}
}

func renderAnalysisFormated(rawAnswer string) string {
	var processedLines []string
	rawLines := strings.Split(rawAnswer, "\n")

	for _, line := range rawLines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "```") {
			continue
		}

		if strings.HasPrefix(trimmed, "`") && strings.HasSuffix(trimmed, "`") && strings.Count(trimmed, "`") == 2 {
			processedLines = append(processedLines, render.Code(strings.Trim(trimmed, "`")))
		} else if strings.HasPrefix(strings.ToLower(trimmed), "docker ") || strings.HasPrefix(strings.ToLower(trimmed), "nginx ") || strings.HasPrefix(trimmed, "$ ") {
			processedLines = append(processedLines, render.Code(trimmed))
		} else {
			if strings.Count(trimmed, "`") >= 2 {
				parts := strings.Split(trimmed, "`")
				for i := 1; i < len(parts); i += 2 {
					parts[i] = render.InlineCode(parts[i])
				}
				trimmed = strings.Join(parts, "")
			}
			processedLines = append(processedLines, trimmed)
		}
	}
	return render.SubBlock(strings.Join(processedLines, "\n\n"))
}

var explainCmd = &cobra.Command{
	Use:   "explain [target] [question] or explain [generic question]",
	Short: "Analyze an issue using host context and AI",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		totalStart := time.Now()
		var contextStart time.Time

		var target string
		var question string
		var ctx contextbuilder.Context
		var vectors []string
		collectorName := "Generic"

		firstArg := strings.ToLower(args[0])
		if collector, exists := contextbuilder.Registry[firstArg]; exists {
			target = firstArg
			question = strings.Join(args[1:], " ")
			if question == "" {
				question = fmt.Sprintf("Analiza el estado actual del servicio %s", target)
			}

			contextStart = time.Now()
			if collector.Supports() {
				if collectedCtx, err := collector.Collect(); err == nil {
					ctx = collectedCtx
					switch target {
					case "docker":
						collectorName = "Docker"
						vectors = []string{"docker ps -a", "docker inspect", "docker core metrics", "container tail logs"}
					case "nginx":
						collectorName = "Nginx"
						vectors = []string{"nginx -V", "nginx status endpoint", "error.log tails"}
					default:
						collectorName = strings.ToUpper(target[:1]) + target[1:]
						vectors = []string{"service status", "service logs"}
					}
				}
			}
			ctx.System = contextbuilder.CollectBaseSystem()
			vectors = append(vectors, "host core load", "ram footprint", "current working directory")
			ctx.Sanitize()
		} else {
			target = "generic"
			question = strings.Join(args, " ")

			contextStart = time.Now()
			attachLocalPaths(args, &ctx)
			ctx.System = contextbuilder.CollectBaseSystem()
			vectors = []string{"host core load", "ram footprint", "current working directory"}
			ctx.Sanitize()
		}

		contextDuration := time.Since(contextStart).Milliseconds()

		cwd, _ := os.Getwd()
		envelope := transport.ContextEnvelope{
			Command:  "explain",
			Target:   target,
			Question: question,
			CWD:      cwd,
			Context:  ctx,
		}

		payloadBytes, _ := json.Marshal(envelope)
		payloadSizeKB := float64(len(payloadBytes)) / 1024.0

		socketPath := os.Getenv("SENTINEL_SOCKET_PATH")
		if socketPath == "" {
			socketPath = "/run/sentinel/sentinel.sock"
		}

		centerBlock := func(s string) string {
			return lipgloss.PlaceHorizontal(80, lipgloss.Center, s)
		}

		fmt.Println()
		fmt.Println(centerBlock(render.HeaderBox("🛡 SENTINEL EXPLAIN")))
		fmt.Println(centerBlock(render.TopBar("v2.0.0", "Connected", "Gemini Flash", "AWS: IAM Role")))

		loading := render.NewSpinner("Gathering context and negotiating with Sentinel Agent")
		loading.Start()

		resp, err := transport.SendToAgent(socketPath, envelope)
		loading.Stop()

		if err != nil {
			fmt.Println(render.Banner(render.Error, fmt.Sprintf(" ✗ COMMUNICATION ERROR: %v ", err)))
			os.Exit(1)
		}

		if resp.Error != "" {
			fmt.Println(render.Banner(render.Error, fmt.Sprintf(" ✗ AGENT ERROR: %s ", resp.Error)))
			os.Exit(1)
		}

		totalDuration := time.Since(totalStart).Milliseconds()

		infoBoxStyle := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(0, 1)

		contextData := fmt.Sprintf("%s\n%s\n%s",
			render.KeyValue("Service", target),
			render.KeyValue("Collector", collectorName),
			render.KeyValue("Payload", fmt.Sprintf("%.2f KB", payloadSizeKB)),
		)
		fmt.Println(infoBoxStyle.Render(contextData))

		fmt.Println(render.Section(fmt.Sprintf("Collected Context (%d sources)", len(vectors))))
		var contextLines []string
		for _, v := range vectors {
			contextLines = append(contextLines, render.Item(render.Success, v, ""))
		}
		if len(ctx.Files) > 0 {
			contextLines = append(contextLines, render.Item(render.Success, fmt.Sprintf("%d local files/folders attached", len(ctx.Files)), ""))
		}
		fmt.Println(render.SubBlock(strings.Join(contextLines, "\n")))

		fmt.Println(render.Section("Analysis"))
		fmt.Println(renderAnalysisFormated(resp.Answer))

		footerData := fmt.Sprintf("%s\n%s\n%s",
			render.KeyValue("Context collected", fmt.Sprintf("%d ms", contextDuration)),
			render.KeyValue("LLM inference", fmt.Sprintf("%d ms", resp.DurationMS)),
			render.KeyValue("Total execution", fmt.Sprintf("%d ms", totalDuration)),
		)
		fmt.Println()
		fmt.Println(infoBoxStyle.Render(footerData))
		fmt.Println()
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check the health of the Sentinel Agent",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(render.HeaderBox("🛡️  SENTINEL AGENT STATUS"))
	},
}

func init() {
	rootCmd.AddCommand(explainCmd)
	rootCmd.AddCommand(statusCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
