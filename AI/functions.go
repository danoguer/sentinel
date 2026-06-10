package AI

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func localReadFile(filepath string) string {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Sprintf("Error reading file: %v", err)
	}
	return string(data)
}

func localSearchFile(filename string, startDir string) string {
	var foundPaths []string

	err := filepath.WalkDir(startDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && strings.Contains(strings.ToLower(d.Name()), strings.ToLower(filename)) {
			foundPaths = append(foundPaths, path)
		}
		return nil
	})

	if err != nil {
		return fmt.Sprintf("Error searching: %v", err)
	}
	if len(foundPaths) == 0 {
		return fmt.Sprintf("No files containing '%s' found in %s", filename, startDir)
	}

	return strings.Join(foundPaths, "\n")
}

func localGetHistory() string {
	vaultPath := "/tmp/sentinel_vault.log"
	file, err := os.Open(vaultPath)
	if err != nil {
		return "History vault currently inaccessible."
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil || stat.Size() == 0 {
		return "Vault is currently empty."
	}

	size := stat.Size()
	var toRead int64 = 4000
	if size < toRead {
		toRead = size
	}

	buffer := make([]byte, toRead)
	_, err = file.ReadAt(buffer, size-toRead)
	if err != nil {
		return "Error reading vault history."
	}

	return string(buffer)
}
