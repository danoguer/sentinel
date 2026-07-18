package llm

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
)

func GetLocalSystemSnapshot() string {
	if runtime.GOOS != "linux" {
		return fmt.Sprintf("OS: %s | ARCH: %s (Host telemetry only supported natively on Linux kernels)", runtime.GOOS, runtime.GOARCH)
	}

	loadStr := "Unknown Load"
	if loadBytes, err := os.ReadFile("/proc/loadavg"); err == nil {
		fields := strings.Fields(string(loadBytes))
		if len(fields) >= 3 {
			loadStr = strings.Join(fields[:3], " ")
		}
	} else {
		loadStr = fmt.Sprintf("Error reading load: %v", err)
	}

	memStr := "Unknown RAM"
	if memBytes, err := os.ReadFile("/proc/meminfo"); err == nil {
		var total, available string
		lines := strings.Split(string(memBytes), "\n")

		for _, line := range lines {
			if strings.HasPrefix(line, "MemTotal:") {
				total = parseMemLine(line)
			} else if strings.HasPrefix(line, "MemAvailable:") {
				available = parseMemLine(line)
			}
			if total != "" && available != "" {
				break
			}
		}

		if total != "" && available != "" {
			memStr = fmt.Sprintf("%s available / %s total", available, total)
		}
	} else {
		memStr = fmt.Sprintf("Error reading RAM: %v", err)
	}

	return fmt.Sprintf("OS: %s | Host Load (1/5/15m): [%s] | RAM: %s", runtime.GOOS, loadStr, memStr)
}


func parseMemLine(line string) string {
	fields := strings.Fields(line)
	if len(fields) >= 2 {
		val := fields[1]
		if len(fields) > 2 {
			return fmt.Sprintf("%s %s", val, fields[2])
		}
		return val
	}
	return ""
}


func localGetHistory() string {
	uid := os.Getuid()
	vaultPath := fmt.Sprintf("/tmp/sentinel_vault_%d.log", uid)

	file, err := os.Open(vaultPath)
	if err != nil {
		return "Error: Vault is empty or not found."
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return "Error: Cannot read vault stats."
	}

	if stat.Size() < 4000 {
		bytes, err := os.ReadFile(vaultPath)
		if err != nil {
			return fmt.Sprintf("Error reading vault file: %v", err)
		}
		return string(bytes)
	}

	_, err = file.Seek(-4000, io.SeekEnd)
	if err != nil {
		return fmt.Sprintf("Error seeking inside vault: %v", err)
	}

	buffer := make([]byte, 4000)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return fmt.Sprintf("Error reading vault buffer: %v", err)
	}

	return string(buffer[:n])
}
