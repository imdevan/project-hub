package shell

import (
	"os"
	"path/filepath"
	"strings"
)

// Adapter handles shell-specific operations.
type Adapter struct {
	shellType string
}

// New creates a new shell adapter for the given shell type.
func New(shellType string) *Adapter {
	return &Adapter{shellType: shellType}
}

// GetFileExtension returns the appropriate file extension for the shell.
func GetFileExtension(shellType string) string {
	switch shellType {
	case "fish":
		return ".fish"
	case "nu", "nushell":
		return ".nu"
	default: // bash, zsh, sh
		return ".sh"
	}
}

// DetectShell attempts to detect the current shell from environment variables.
// Returns the detected shell type or "bash" as a fallback.
func DetectShell() string {
	// Check SHELL environment variable (most common)
	if shellPath := os.Getenv("SHELL"); shellPath != "" {
		shellName := filepath.Base(shellPath)
		return normalizeShellName(shellName)
	}

	// Check for shell-specific environment variables
	if os.Getenv("BASH_VERSION") != "" {
		return "bash"
	}
	if os.Getenv("ZSH_VERSION") != "" {
		return "zsh"
	}
	if os.Getenv("FISH_VERSION") != "" {
		return "fish"
	}
	if os.Getenv("NU_VERSION") != "" {
		return "nu"
	}

	// Fallback to bash
	return "bash"
}

// normalizeShellName normalizes shell names to standard identifiers.
func normalizeShellName(name string) string {
	name = strings.ToLower(name)

	// Handle common variations
	switch {
	case strings.Contains(name, "bash"):
		return "bash"
	case strings.Contains(name, "zsh"):
		return "zsh"
	case strings.Contains(name, "fish"):
		return "fish"
	case strings.Contains(name, "nu"):
		return "nu"
	case name == "sh":
		return "sh"
	default:
		// If unknown, assume POSIX-compatible
		return "bash"
	}
}
