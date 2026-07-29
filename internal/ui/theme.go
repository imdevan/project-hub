package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/super-todo/internal/domain"
)

// Theme holds configurable colors for UI output.
type Theme struct {
	Headings            lipgloss.Color
	Primary             lipgloss.Color
	Secondary           lipgloss.Color
	Text                lipgloss.Color
	TextHighlight       lipgloss.Color
	DescriptionHighlight lipgloss.Color
	Tags                lipgloss.Color
	Flags               lipgloss.Color
	Muted               lipgloss.Color
	Border              lipgloss.Color
	Error               lipgloss.Color
	Success             lipgloss.Color
	Title               lipgloss.Color
	HighlightBorder     lipgloss.Color
	Name                lipgloss.Color
	HighlightName       lipgloss.Color
	Tasks               lipgloss.Color
	ProgressBg          lipgloss.Color
	ProgressFg          lipgloss.Color
	HighlightProgressBg lipgloss.Color
	HighlightProgressFg lipgloss.Color
}

// ThemeFromConfig builds a theme with safe fallbacks.
func ThemeFromConfig(cfg domain.Config) Theme {
	headings := resolveColor(cfg.Colors.Headings, "15")
	primary := resolveColor(cfg.Colors.Primary, "02")
	secondary := resolveColor(cfg.Colors.Secondary, "06")
	text := resolveColor(cfg.Colors.Text, "07")
	textHighlight := resolveColor(resolveFallback(cfg.Colors.TextHighlight, cfg.Colors.Secondary), "06")
	muted := resolveColor(cfg.Colors.Muted, "08")

	return Theme{
		Headings:            headings,
		Primary:             primary,
		Secondary:           secondary,
		Text:                text,
		TextHighlight:       textHighlight,
		DescriptionHighlight: resolveColor(resolveFallback(cfg.Colors.DescriptionHighlight, cfg.Colors.Secondary), "06"),
		Tags:                resolveColor(resolveFallback(cfg.Colors.Tags, cfg.Colors.Accent), "13"),
		Flags:               resolveColor(cfg.Colors.Flags, "12"),
		Muted:               muted,
		Border:              resolveColor(cfg.Colors.Border, "08"),
		Error:               resolveColor(cfg.Colors.Error, "01"),
		Success:             resolveColor(resolveFallback(cfg.Colors.Success, cfg.Colors.Primary), "02"),
		Title:               resolveColor(cfg.Colors.Title, string(primary)),
		HighlightBorder:     resolveColor(cfg.Colors.HighlightBorder, string(primary)),
		Name:                resolveColor(cfg.Colors.Name, string(text)),
		HighlightName:       resolveColor(cfg.Colors.HighlightName, string(textHighlight)),
		Tasks:               resolveColor(cfg.Colors.Tasks, string(secondary)),
		ProgressBg:          resolveColor(cfg.Colors.ProgressBg, string(muted)),
		ProgressFg:          resolveColor(cfg.Colors.ProgressFg, string(primary)),
		HighlightProgressBg: resolveColor(cfg.Colors.HighlightProgressBg, string(muted)),
		HighlightProgressFg: resolveColor(cfg.Colors.HighlightProgressFg, string(textHighlight)),
	}
}

// GetRainbowColors returns resolved lipgloss colors based on config.Rainbow, or empty slice if disabled.
func GetRainbowColors(rainbow interface{}) []lipgloss.Color {
	if rainbow == nil {
		return nil
	}

	var colorStrings []string
	switch v := rainbow.(type) {
	case bool:
		if !v {
			return nil
		}
		colorStrings = domain.DefaultRainbowAnchors
	case []string:
		if len(v) == 0 {
			return nil
		}
		colorStrings = v
	case []interface{}:
		for _, item := range v {
			if str, ok := item.(string); ok {
				colorStrings = append(colorStrings, str)
			}
		}
		if len(colorStrings) == 0 {
			return nil
		}
	default:
		return nil
	}

	res := make([]lipgloss.Color, len(colorStrings))
	for i, c := range colorStrings {
		res[i] = lipgloss.Color(c)
	}
	return res
}

func resolveColor(value, fallback string) lipgloss.Color {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		trimmed = fallback
	}
	return lipgloss.Color(trimmed)
}

func resolveFallback(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
