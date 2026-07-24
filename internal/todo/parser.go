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

// ParseFTSDetailed parses a markdown task file in detail, returning features and tasks.
func ParseFTSDetailed(filePath string, cfg domain.Config) (completed, total int, features []Feature, err error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return 0, 0, nil, err
	}

	lines := strings.Split(string(data), "\n")
	var currentFeature *Feature

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for Feature headers
		isHeader := false
		for _, prefix := range cfg.FtsFeatures {
			if strings.HasPrefix(line, prefix) {
				isHeader = true
				headerName := strings.TrimSpace(strings.TrimPrefix(line, prefix))
				if currentFeature != nil {
					features = append(features, *currentFeature)
				}
				currentFeature = &Feature{
					Name:  headerName,
					Tasks: []TaskItem{},
				}
				break
			}
		}
		if isHeader {
			continue
		}

		// Check for Task / Subtask checkboxes
		isTask := false
		isChecked := false
		var matchedText string

		tokens := append([]string{}, cfg.FtsTasks...)
		tokens = append(tokens, cfg.FtsSubtasks...)

		for _, token := range tokens {
			tokenTrimmed := strings.TrimSpace(token)
			if tokenTrimmed == "" {
				continue
			}

			uncheckedPrefix := tokenTrimmed
			checkedPrefix1 := strings.Replace(uncheckedPrefix, "[ ]", "[x]", 1)
			checkedPrefix2 := strings.Replace(uncheckedPrefix, "[ ]", "[X]", 1)

			if strings.HasPrefix(trimmed, uncheckedPrefix) {
				isTask = true
				matchedText = strings.TrimSpace(strings.TrimPrefix(trimmed, uncheckedPrefix))
				break
			} else if strings.HasPrefix(trimmed, checkedPrefix1) || strings.HasPrefix(trimmed, checkedPrefix2) {
				isTask = true
				isChecked = true
				matchedText = strings.TrimSpace(strings.TrimPrefix(trimmed, checkedPrefix1))
				if len(matchedText) == len(trimmed) {
					// Fallback to checkedPrefix2 if checkedPrefix1 wasn't the actual match
					matchedText = strings.TrimSpace(strings.TrimPrefix(trimmed, checkedPrefix2))
				}
				break
			}
		}

		if isTask {
			total++
			if isChecked {
				completed++
			}

			task := TaskItem{
				Text:      matchedText,
				Completed: isChecked,
			}

			if currentFeature == nil {
				currentFeature = &Feature{
					Name:  "General Tasks",
					Tasks: []TaskItem{},
				}
			}
			currentFeature.Tasks = append(currentFeature.Tasks, task)
		}
	}

	if currentFeature != nil {
		features = append(features, *currentFeature)
	}

	return completed, total, features, nil
}

// FindTodosInFile scans a file and returns all found TODO comments.
func FindTodosInFile(filePath string, cfg domain.Config) ([]TodoComment, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	head := make([]byte, 512)
	n, _ := file.Read(head)
	if isBinary(head[:n]) {
		return nil, nil
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(file)
	var todos []TodoComment
	lineNumber := 0

	upperTokens := make([]string, len(cfg.TodoTokens))
	for i, t := range cfg.TodoTokens {
		upperTokens[i] = strings.ToUpper(t)
	}

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		lineUpper := strings.ToUpper(line)

		for _, token := range upperTokens {
			if strings.Contains(lineUpper, token) {
				// Strip comment markers and extra whitespace
				text := strings.TrimSpace(line)
				todos = append(todos, TodoComment{
					File: filePath,
					Line: lineNumber,
					Text: text,
				})
				break
			}
		}
	}

	return todos, scanner.Err()
}
