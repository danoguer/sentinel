package contextbuilder

import (
	"os/exec"
	"strings"
)

type DockerCollector struct{}

func (c *DockerCollector) Name() string {
	return "docker"
}


func (c *DockerCollector) Supports() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

func (c *DockerCollector) Collect() (Context, error) {
	ctx := Context{}

	if out, err := exec.Command("systemctl", "status", "docker", "--no-pager").CombinedOutput(); err == nil {
		ctx.Commands = append(ctx.Commands, CommandResult{
			Name:   "systemctl status docker",
			Output: string(out),
		})
	}

	if out, err := exec.Command("docker", "ps", "-a", "--format", "table {{.ID}}\t{{.Names}}\t{{.Status}}\t{{.Ports}}").CombinedOutput(); err == nil {
		ctx.Commands = append(ctx.Commands, CommandResult{
			Name:   "docker ps -a",
			Output: string(out),
		})
	}


	if out, err := exec.Command("docker", "stats", "--no-stream", "--format", "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}").CombinedOutput(); err == nil {
		ctx.Commands = append(ctx.Commands, CommandResult{
			Name:   "docker stats",
			Output: string(out),
		})
	}

	if out, err := exec.Command("journalctl", "-u", "docker", "-n", "30", "--no-pager").CombinedOutput(); err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			if strings.TrimSpace(line) == "" {
				continue
			}
			ctx.Logs = append(ctx.Logs, LogEntry{
				Source: "journalctl -u docker",
				Line:   line,
			})
		}
	}

	return ctx, nil
}
