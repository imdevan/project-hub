package todo

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/project-hub/internal/domain"
	"github.com/project-hub/internal/utils"
)

// ReadYAMLName attempts to parse a top front-matter YAML block (between --- markers) and extract the "name" field.
func ReadYAMLName(content string) string {
	trimmed := strings.TrimSpace(content)
	if !strings.HasPrefix(trimmed, "---") {
		return ""
	}

	lines := strings.Split(trimmed, "\n")
	if len(lines) < 3 {
		return ""
	}

	for _, line := range lines[1:] {
		lineTrimmed := strings.TrimSpace(line)
		if lineTrimmed == "---" {
			break
		}
		if strings.HasPrefix(lineTrimmed, "name:") {
			val := strings.TrimPrefix(lineTrimmed, "name:")
			val = strings.TrimSpace(val)
			val = strings.Trim(val, "\"'")
			return val
		}
	}

	return ""
}

const recurringResetMetaFormat = "2006-01-02T15:04:05"

// CheckAndResetRecurringTasks checks if a recurring task file (daily.md or weekly.md) needs reset based on config.
func CheckAndResetRecurringTasks(filePath, recurringType string, cfg domain.Config) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	content := string(data)
	lastResetTime, hasMeta := parseLastResetMeta(content)

	now := time.Now()
	nextReset := calculateNextReset(lastResetTime, recurringType, cfg.DailyReset, cfg.WeeklyReset, now)

	if !hasMeta || now.After(nextReset) || now.Equal(nextReset) {
		// Reset all checkboxes to unchecked [ ]
		newContent := resetCheckboxes(content)
		// Update HTML comment metadata at bottom
		newContent = setLastResetMeta(newContent, now.Format(recurringResetMetaFormat))
		return os.WriteFile(filePath, []byte(newContent), 0o644)
	}

	return nil
}

// ToggleTaskInFile toggles the checkbox on line targetLine (1-indexed) in filePath.
func ToggleTaskInFile(filePath string, targetLine int) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	if targetLine < 1 || targetLine > len(lines) {
		return fmt.Errorf("line %d out of bounds", targetLine)
	}

	lineIdx := targetLine - 1
	line := lines[lineIdx]

	if strings.Contains(line, "[ ]") {
		lines[lineIdx] = strings.Replace(line, "[ ]", "[x]", 1)
	} else if strings.Contains(line, "[x]") {
		lines[lineIdx] = strings.Replace(line, "[x]", "[ ]", 1)
	} else if strings.Contains(line, "[X]") {
		lines[lineIdx] = strings.Replace(line, "[X]", "[ ]", 1)
	}

	return os.WriteFile(filePath, []byte(strings.Join(lines, "\n")), 0o644)
}

func parseLastResetMeta(content string) (time.Time, bool) {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "<!-- last_reset:") && strings.HasSuffix(trimmed, "-->") {
			timeStr := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(trimmed, "<!-- last_reset:"), "-->"))
			t, err := time.Parse(recurringResetMetaFormat, timeStr)
			if err == nil {
				return t, true
			}
		}
	}
	return time.Time{}, false
}

func setLastResetMeta(content, timeStr string) string {
	comment := fmt.Sprintf("<!-- last_reset: %s -->", timeStr)
	lines := strings.Split(content, "\n")
	var newLines []string
	found := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "<!-- last_reset:") && strings.HasSuffix(trimmed, "-->") {
			newLines = append(newLines, comment)
			found = true
		} else {
			newLines = append(newLines, line)
		}
	}
	if !found {
		// Ensure non-empty file ends with newline before appending metadata comment
		if len(newLines) > 0 && newLines[len(newLines)-1] != "" {
			newLines = append(newLines, "")
		}
		newLines = append(newLines, comment)
	}
	return strings.Join(newLines, "\n")
}

func resetCheckboxes(content string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.Contains(line, "[x]") {
			lines[i] = strings.Replace(line, "[x]", "[ ]", 1)
		} else if strings.Contains(line, "[X]") {
			lines[i] = strings.Replace(line, "[X]", "[ ]", 1)
		}
	}
	return strings.Join(lines, "\n")
}

func calculateNextReset(lastReset time.Time, recurringType, dailyResetConfig, weeklyResetConfig string, now time.Time) time.Time {
	hour, min := parseResetTime(dailyResetConfig)

	if lastReset.IsZero() {
		// If no prior reset record, trigger reset immediately by returning past time
		return now.Add(-time.Hour)
	}

	if recurringType == "daily" {
		// Next reset is the daily reset time following lastReset
		resetDay := time.Date(lastReset.Year(), lastReset.Month(), lastReset.Day(), hour, min, 0, 0, lastReset.Location())
		if !lastReset.Before(resetDay) {
			// lastReset was at or after resetDay, so next reset is tomorrow at resetDay
			resetDay = resetDay.AddDate(0, 0, 1)
		}
		return resetDay
	}

	// Weekly reset: monday at daily_reset time
	targetWeekday := time.Monday
	if strings.ToLower(weeklyResetConfig) != "" {
		targetWeekday = parseWeekday(weeklyResetConfig)
	}

	resetDay := time.Date(lastReset.Year(), lastReset.Month(), lastReset.Day(), hour, min, 0, 0, lastReset.Location())
	// Advance resetDay to the next target weekday
	daysUntil := (int(targetWeekday) - int(resetDay.Weekday()) + 7) % 7
	if daysUntil == 0 && !lastReset.Before(resetDay) {
		daysUntil = 7
	}
	resetDay = resetDay.AddDate(0, 0, daysUntil)
	return resetDay
}

func parseResetTime(timeStr string) (int, int) {
	parts := strings.Split(timeStr, ":")
	if len(parts) == 2 {
		h, err1 := strconv.Atoi(parts[0])
		m, err2 := strconv.Atoi(parts[1])
		if err1 == nil && err2 == nil {
			return h, m
		}
	}
	return 4, 20 // Default 04:20
}

func parseWeekday(dayStr string) time.Weekday {
	switch strings.ToLower(dayStr) {
	case "sunday", "sun":
		return time.Sunday
	case "tuesday", "tue":
		return time.Tuesday
	case "wednesday", "wed":
		return time.Wednesday
	case "thursday", "thu":
		return time.Thursday
	case "friday", "fri":
		return time.Friday
	case "saturday", "sat":
		return time.Saturday
	default:
		return time.Monday
	}
}

// GetRecurringConfigDir returns ~/.config/super-todo/ or XDG_CONFIG_HOME/super-todo/
func GetRecurringConfigDir() string {
	configHome := utils.XDGConfigHome()
	return filepath.Join(configHome, "super-todo")
}
