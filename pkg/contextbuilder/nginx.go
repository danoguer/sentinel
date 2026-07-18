package contextbuilder

import (
	"os"
	"os/exec"
	"strings"
)

type NginxCollector struct{}

func (c *NginxCollector) Name() string {
	return "nginx"
}

func (c *NginxCollector) Supports() bool {
	_, err := exec.LookPath("nginx")
	return err == nil
}

func (c *NginxCollector) Collect() (Context, error) {

	ctx := Context{}

	if out, err := exec.Command(
		"systemctl",
		"status",
		"nginx",
		"--no-pager",
	).CombinedOutput(); err == nil {

		ctx.Commands = append(ctx.Commands,
			CommandResult{
				Name:   "systemctl status nginx",
				Output: string(out),
			},
		)
	}

	if out, err := exec.Command(
		"journalctl",
		"-u",
		"nginx",
		"-n",
		"50",
		"--no-pager",
	).CombinedOutput(); err == nil {

		for _, line := range strings.Split(string(out), "\n") {
			ctx.Logs = append(ctx.Logs,
				LogEntry{
					Source: "journalctl",
					Line:   line,
				},
			)
		}
	}

	if bytes, err := os.ReadFile("/var/log/nginx/error.log"); err == nil {

		ctx.Files = append(ctx.Files,
			FileEntry{
				Path:    "/var/log/nginx/error.log",
				Content: string(bytes),
			},
		)
	}

	return ctx, nil
}
