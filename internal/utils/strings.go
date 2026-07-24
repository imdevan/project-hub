package utils

import (
	"path/filepath"
	"strings"
)

// GenerateAlias generates an alias name from a directory path.
func GenerateAlias(path string, separator string, lowercase bool, partLength int) string {
	cleaned := filepath.Clean(path)
	base := filepath.Base(cleaned)
	if base == "/" || base == "." {
		return "h"
	}

	var parts []string
	if separator != "" {
		parts = strings.Split(base, separator)
	} else {
		if strings.Contains(base, "-") {
			parts = strings.Split(base, "-")
		} else if strings.Contains(base, "_") {
			parts = strings.Split(base, "_")
		} else {
			parts = []string{base}
		}
	}

	var aliasParts []string
	for _, part := range parts {
		if len(part) == 0 {
			continue
		}
		length := partLength
		if length > len(part) {
			length = len(part)
		}
		aliasParts = append(aliasParts, part[:length])
	}

	alias := strings.Join(aliasParts, separator)
	if lowercase {
		alias = strings.ToLower(alias)
	}
	return alias
}

// EscapeAliasValue escapes double quotes inside the alias value by replacing them with single quotes.
func EscapeAliasValue(val string) string {
	return strings.ReplaceAll(val, "\"", "'")
}

// WrapText wraps a string to a given width, breaking words if they exceed the width limit.
func WrapText(s string, width int) []string {
	if width <= 0 {
		return []string{s}
	}
	var lines []string
	for _, paragraph := range strings.Split(s, "\n") {
		if paragraph == "" {
			lines = append(lines, "")
			continue
		}
		
		var currentLine strings.Builder
		words := strings.Split(paragraph, " ")
		for _, word := range words {
			if word == "" {
				continue
			}
			
			// Does it fit on the current line?
			if currentLine.Len() == 0 {
				if len(word) <= width {
					currentLine.WriteString(word)
				} else {
					// Word is longer than width, must break it
					for len(word) > 0 {
						if currentLine.Len() == width {
							lines = append(lines, currentLine.String())
							currentLine.Reset()
						}
						spaceLeft := width - currentLine.Len()
						chunkLen := len(word)
						if chunkLen > spaceLeft {
							chunkLen = spaceLeft
						}
						currentLine.WriteString(word[:chunkLen])
						word = word[chunkLen:]
					}
				}
			} else {
				if currentLine.Len()+1+len(word) <= width {
					currentLine.WriteByte(' ')
					currentLine.WriteString(word)
				} else {
					// Push current line, then check word on a new empty line
					lines = append(lines, currentLine.String())
					currentLine.Reset()
					
					if len(word) <= width {
						currentLine.WriteString(word)
					} else {
						// Word is longer than width, must break it
						for len(word) > 0 {
							if currentLine.Len() == width {
								lines = append(lines, currentLine.String())
								currentLine.Reset()
							}
							spaceLeft := width - currentLine.Len()
							chunkLen := len(word)
							if chunkLen > spaceLeft {
								chunkLen = spaceLeft
							}
							currentLine.WriteString(word[:chunkLen])
							word = word[chunkLen:]
						}
					}
				}
			}
		}
		if currentLine.Len() > 0 {
			lines = append(lines, currentLine.String())
		}
	}
	return lines
}
