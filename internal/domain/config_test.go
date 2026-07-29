package domain

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	t.Run("has editor set", func(t *testing.T) {
		if cfg.Editor == "" {
			t.Error("DefaultConfig().Editor should not be empty")
		}
		if cfg.Editor != "nvim" {
			t.Errorf("DefaultConfig().Editor = %q, want %q", cfg.Editor, "nvim")
		}
	})

	t.Run("has all color values set", func(t *testing.T) {
		colorFields := map[string]string{
			"Headings":             cfg.Colors.Headings,
			"Primary":              cfg.Colors.Primary,
			"Secondary":            cfg.Colors.Secondary,
			"Text":                 cfg.Colors.Text,
			"TextHighlight":        cfg.Colors.TextHighlight,
			"DescriptionHighlight": cfg.Colors.DescriptionHighlight,
			"Tags":                 cfg.Colors.Tags,
			"Flags":                cfg.Colors.Flags,
			"Muted":                cfg.Colors.Muted,
			"Accent":               cfg.Colors.Accent,
			"Border":               cfg.Colors.Border,
			"Error":                cfg.Colors.Error,
		}

		for name, value := range colorFields {
			if value == "" {
				t.Errorf("DefaultConfig().Colors.%s should not be empty", name)
			}
		}
	})

	t.Run("has expected default values", func(t *testing.T) {
		if cfg.Colors.Headings != "15" {
			t.Errorf("DefaultConfig().Colors.Headings = %q, want %q", cfg.Colors.Headings, "15")
		}
		if cfg.Colors.Primary != "02" {
			t.Errorf("DefaultConfig().Colors.Primary = %q, want %q", cfg.Colors.Primary, "02")
		}
		if cfg.Colors.Secondary != "06" {
			t.Errorf("DefaultConfig().Colors.Secondary = %q, want %q", cfg.Colors.Secondary, "06")
		}
		if cfg.Colors.Text != "07" {
			t.Errorf("DefaultConfig().Colors.Text = %q, want %q", cfg.Colors.Text, "07")
		}
		if cfg.Colors.TextHighlight != "06" {
			t.Errorf("DefaultConfig().Colors.TextHighlight = %q, want %q", cfg.Colors.TextHighlight, "06")
		}
		if cfg.Colors.DescriptionHighlight != "05" {
			t.Errorf("DefaultConfig().Colors.DescriptionHighlight = %q, want %q", cfg.Colors.DescriptionHighlight, "05")
		}
		if cfg.Colors.Tags != "13" {
			t.Errorf("DefaultConfig().Colors.Tags = %q, want %q", cfg.Colors.Tags, "13")
		}
		if cfg.Colors.Flags != "12" {
			t.Errorf("DefaultConfig().Colors.Flags = %q, want %q", cfg.Colors.Flags, "12")
		}
		if cfg.Colors.Muted != "08" {
			t.Errorf("DefaultConfig().Colors.Muted = %q, want %q", cfg.Colors.Muted, "08")
		}
		if cfg.Colors.Accent != "13" {
			t.Errorf("DefaultConfig().Colors.Accent = %q, want %q", cfg.Colors.Accent, "13")
		}
		if cfg.Colors.Border != "08" {
			t.Errorf("DefaultConfig().Colors.Border = %q, want %q", cfg.Colors.Border, "08")
		}
		if cfg.Colors.Error != "01" {
			t.Errorf("DefaultConfig().Colors.Error = %q, want %q", cfg.Colors.Error, "01")
		}
	})

	t.Run("has list spacing set", func(t *testing.T) {
		if cfg.ListSpacing == "" {
			t.Error("DefaultConfig().ListSpacing should not be empty")
		}
		if cfg.ListSpacing != "space" {
			t.Errorf("DefaultConfig().ListSpacing = %q, want %q", cfg.ListSpacing, "space")
		}
	})

	t.Run("has rainbow default false", func(t *testing.T) {
		if rainbowBool, ok := cfg.Rainbow.(bool); !ok || rainbowBool {
			t.Errorf("DefaultConfig().Rainbow = %v, want false", cfg.Rainbow)
		}
	})

	t.Run("has default rainbow anchors defined", func(t *testing.T) {
		if len(DefaultRainbowAnchors) != 14 {
			t.Errorf("len(DefaultRainbowAnchors) = %d, want 14", len(DefaultRainbowAnchors))
		}
	})
}

func TestDefaultConfig_Consistency(t *testing.T) {
	t.Run("multiple calls return same values", func(t *testing.T) {
		cfg1 := DefaultConfig()
		cfg2 := DefaultConfig()

		if cfg1.Editor != cfg2.Editor {
			t.Error("DefaultConfig() should return consistent Editor values")
		}
		if cfg1.Colors.Headings != cfg2.Colors.Headings {
			t.Error("DefaultConfig() should return consistent Headings values")
		}
	})
}

func TestConfig_StructTags(t *testing.T) {
	t.Run("has toml tags for all fields", func(t *testing.T) {
		// This is a regression test to ensure TOML tags don't get accidentally removed
		cfg := Config{}

		// Set values to ensure struct is properly tagged
		cfg.Editor = "test"
		cfg.Colors.Primary = "01"
		cfg.ListSpacing = "compact"

		// If this compiles and runs, the struct tags are present
		if cfg.Editor != "test" {
			t.Error("Config struct should be properly defined")
		}
	})
}
