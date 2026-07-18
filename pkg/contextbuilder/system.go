package contextbuilder

import (
	"os"
	"strings"
)

func GetSystemLoad() string {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return "unknown system load"
	}

	parts := strings.Fields(string(data))
	if len(parts) >= 3 {
		return strings.Join(parts[:3], " ")
	}

	return strings.TrimSpace(string(data))
}

func GetOSRelease() string {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "Linux"
	}

	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			return strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), `"`)
		}
	}
	return "Linux"
}

func GetHostname() string {
	data, err := os.ReadFile("/proc/sys/kernel/hostname")
	if err != nil {
		return "unknown-host"
	}
	return strings.TrimSpace(string(data))
}

func CollectBaseSystem() SystemContext {
	return SystemContext{
		OS:          GetOSRelease(),
		LoadAverage: GetSystemLoad(),
		Hostname:    GetHostname(),

		Memory:      "OK",
		Uptime:      "Active",
	}
}
