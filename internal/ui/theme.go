package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/super-todo/internal/domain"
)

// Theme holds configurable colors for UI output.
type Theme struct {
	Headings                 lipgloss.Color
	Primary                  lipgloss.Color
	Secondary                lipgloss.Color
	Text                     lipgloss.Color
	TextHighlight            lipgloss.Color
	DescriptionHighlight     lipgloss.Color
	Tags                     lipgloss.Color
	Flags                    lipgloss.Color
	Muted                    lipgloss.Color
	Border                   lipgloss.Color
	Error                    lipgloss.Color
	Success                  lipgloss.Color
	ColorTitle               lipgloss.Color
	ColorHighlightBorder     lipgloss.Color
	ColorName                lipgloss.Color
	ColorHighlightName       lipgloss.Color
	ColorTasks               lipgloss.Color
	ColorProgressBg          lipgloss.Color
	ColorProgressFg          lipgloss.Color
	ColorHighlightProgressBg  lipgloss.Color
	ColorHighlightProgressFg lipgloss.Color
}

// ThemeFromConfig builds a theme with safe fallbacks.
func ThemeFromConfig(cfg domain.Config) Theme {
	headings := resolveColor(cfg.Headings, "15")
	primary := resolveColor(cfg.Primary, "02")
	secondary := resolveColor(cfg.Secondary, "06")
	text := resolveColor(cfg.Text, "07")
	textHighlight := resolveColor(resolveFallback(cfg.TextHighlight, cfg.Secondary), "06")
	muted := resolveColor(cfg.Muted, "08")

	return Theme{
		Headings:                 headings,
		Primary:                  primary,
		Secondary:                secondary,
		Text:                     text,
		TextHighlight:            textHighlight,
		DescriptionHighlight:     resolveColor(resolveFallback(cfg.DescriptionHighlight, cfg.Secondary), "06"),
		Tags:                     resolveColor(resolveFallback(cfg.Tags, cfg.Accent), "13"),
		Flags:                    resolveColor(cfg.Flags, "12"),
		Muted:                    muted,
		Border:                   resolveColor(cfg.Border, "08"),
		Error:                    resolveColor(cfg.Error, "01"),
		Success:                  resolveColor(resolveFallback(cfg.Success, cfg.Primary), "02"),
		ColorTitle:               resolveColor(cfg.ColorTitle, string(primary)),
		ColorHighlightBorder:     resolveColor(cfg.ColorHighlightBorder, string(primary)),
		ColorName:                resolveColor(cfg.ColorName, string(text)),
		ColorHighlightName:       resolveColor(cfg.ColorHighlightName, string(textHighlight)),
		ColorTasks:               resolveColor(cfg.ColorTasks, string(secondary)),
		ColorProgressBg:          resolveColor(cfg.ColorProgressBg, string(muted)),
		ColorProgressFg:          resolveColor(cfg.ColorProgressFg, string(primary)),
		ColorHighlightProgressBg:  resolveColor(cfg.ColorHighlightProgressBg, string(muted)),
		ColorHighlightProgressFg: resolveColor(cfg.ColorHighlightProgressFg, string(textHighlight)),
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
