package todo

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strings"

	"github.com/super-todo/internal/domain"
)

// ParseFTSContent parses markdown content and returns the number of completed and total tasks.
func ParseFTSContent(content string, cfg domain.Config) (completed, total int) {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		isTask := false
		isChecked := false

		tokens := append([]string{}, cfg.FtsTasks...)
		tokens = append(tokens, cfg.FtsSubtasks...)

		for _, token := range tokens {
			tokenTrimmed := strings.TrimSpace(token)
			if tokenTrimmed == "" {
				continue
			}

			// Support checked version (replace "[ ]" with "[x]" or "[X]")
			uncheckedPrefix := tokenTrimmed
			checkedPrefix1 := strings.Replace(uncheckedPrefix, "[ ]", "[x]", 1)
			checkedPrefix2 := strings.Replace(uncheckedPrefix, "[ ]", "[X]", 1)

			if strings.HasPrefix(trimmed, uncheckedPrefix) {
				isTask = true
				break
			} else if strings.HasPrefix(trimmed, checkedPrefix1) || strings.HasPrefix(trimmed, checkedPrefix2) {
				isTask = true
				isChecked = true
				break
			}
		}

		if isTask {
			total++
			if isChecked {
				completed++
			}
		}
	}
	return completed, total
}

// ParseFTSFile reads a file and parses its FTS tasks.
func ParseFTSFile(filePath string, cfg domain.Config) (completed, total int, err error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return 0, 0, err
	}
	completed, total = ParseFTSContent(string(data), cfg)
	return completed, total, nil
}

// CountTodosInFile counts occurrences of todo tokens in a file.
func CountTodosInFile(filePath string, cfg domain.Config) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	// Check if binary file by reading the first 512 bytes
	head := make([]byte, 512)
	n, _ := file.Read(head)
	if isBinary(head[:n]) {
		return 0, nil
	}

	// Reset file pointer to beginning
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return 0, err
	}

	scanner := bufio.NewScanner(file)
	count := 0

	// Pre-upper case the tokens for faster case-insensitive matching
	upperTokens := make([]string, len(cfg.TodoTokens))
	for i, t := range cfg.TodoTokens {
		upperTokens[i] = strings.ToUpper(t)
	}

	for scanner.Scan() {
		line := scanner.Text()
		lineUpper := strings.ToUpper(line)

		// Check if line contains any of the todo tokens
		for _, token := range upperTokens {
			if strings.Contains(lineUpper, token) {
				count++
				break
			}
		}
	}

	return count, scanner.Err()
}

func isBinary(data []byte) bool {
	return bytes.Contains(data, []byte{0})
}
