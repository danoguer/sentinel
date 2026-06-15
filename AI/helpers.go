package AI

import (
	"fmt"
	"os"
	"runtime"
)

func localGetHistory() string {
	vaultPath := "/tmp/sentinel_vault.log"
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
		bytes, _ := os.ReadFile(vaultPath)
		return string(bytes)
	}

	file.Seek(-4000, 2)
	buffer := make([]byte, 4000)
	file.Read(buffer)

	return string(buffer)
}

func GetLocalSystemSnapshot() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	ramUsageMB := m.Alloc / 1024 / 1024

	return fmt.Sprintf("OS: %s | ARCH: %s | CPUs: %d | Active Goroutines: %d | RAM Allocated: %v MB",
		runtime.GOOS, runtime.GOARCH, runtime.NumCPU(), runtime.NumGoroutine(), ramUsageMB)
}
