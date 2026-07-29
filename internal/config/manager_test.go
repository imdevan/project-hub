package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/project-hub/internal/domain"
	"github.com/project-hub/internal/utils"
)

func TestManagerLoadsDefaults(t *testing.T) {
	root := t.TempDir()
	cwd := filepath.Join(root, "project")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatalf("mkdir cwd: %v", err)
	}
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(root, "data"))

	manager := NewManager(cwd)
	cfg, err := manager.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if cfg.Editor == "" {
		t.Fatal("expected editor to have default value")
	}
}

func TestManagerLoadsFromFile(t *testing.T) {
	root := t.TempDir()
	cwd := filepath.Join(root, "project")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatalf("mkdir cwd: %v", err)
	}
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))

	configPath := utils.ConfigPathGlobal()
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}

	data := []byte("editor = \"vim\"\n")
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	manager := NewManager(cwd)
	cfg, err := manager.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if cfg.Editor != "vim" {
		t.Fatalf("expected editor from config, got %q", cfg.Editor)
	}
}

func TestManagerSavesConfig(t *testing.T) {
	root := t.TempDir()
	cwd := filepath.Join(root, "project")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatalf("mkdir cwd: %v", err)
	}
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))

	manager := NewManager(cwd)
	cfg := domain.DefaultConfig()
	cfg.Editor = "emacs"

	if err := manager.Save(cfg); err != nil {
		t.Fatalf("save: %v", err)
	}

	configPath := utils.ConfigPathGlobal()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("expected config file to exist")
	}
}

func TestManagerExists(t *testing.T) {
	t.Run("returns false when no config exists", func(t *testing.T) {
		root := t.TempDir()
		cwd := filepath.Join(root, "project")
		if err := os.MkdirAll(cwd, 0o755); err != nil {
			t.Fatalf("mkdir cwd: %v", err)
		}
		t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))

		manager := NewManager(cwd)
		exists, err := manager.Exists()
		if err != nil {
			t.Fatalf("exists: %v", err)
		}
		if exists {
			t.Error("expected config to not exist")
		}
	})

	t.Run("returns true when global config exists", func(t *testing.T) {
		root := t.TempDir()
		cwd := filepath.Join(root, "project")
		if err := os.MkdirAll(cwd, 0o755); err != nil {
			t.Fatalf("mkdir cwd: %v", err)
		}
		t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))

		configPath := utils.ConfigPathGlobal()
		if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
			t.Fatalf("mkdir config dir: %v", err)
		}
		if err := os.WriteFile(configPath, []byte("editor = \"vim\"\n"), 0o644); err != nil {
			t.Fatalf("write config: %v", err)
		}

		manager := NewManager(cwd)
		exists, err := manager.Exists()
		if err != nil {
			t.Fatalf("exists: %v", err)
		}
		if !exists {
			t.Error("expected config to exist")
		}
	})

	t.Run("returns true when local config exists", func(t *testing.T) {
		root := t.TempDir()
		cwd := filepath.Join(root, "project")
		if err := os.MkdirAll(cwd, 0o755); err != nil {
			t.Fatalf("mkdir cwd: %v", err)
		}
		t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))

		localConfigPath := utils.ConfigPathLocal(cwd)
		if err := os.MkdirAll(filepath.Dir(localConfigPath), 0o755); err != nil {
			t.Fatalf("mkdir local config dir: %v", err)
		}
		if err := os.WriteFile(localConfigPath, []byte("editor = \"code\"\n"), 0o644); err != nil {
			t.Fatalf("write local config: %v", err)
		}

		manager := NewManager(cwd)
		exists, err := manager.Exists()
		if err != nil {
			t.Fatalf("exists: %v", err)
		}
		if !exists {
			t.Error("expected config to exist")
		}
	})
}

func TestManagerLoadWithOverride(t *testing.T) {
	t.Run("loads from override path", func(t *testing.T) {
		root := t.TempDir()
		cwd := filepath.Join(root, "project")
		if err := os.MkdirAll(cwd, 0o755); err != nil {
			t.Fatalf("mkdir cwd: %v", err)
		}

		overridePath := filepath.Join(root, "custom-config.toml")
		data := []byte("editor = \"emacs\"\nprimary = \"03\"\n")
		if err := os.WriteFile(overridePath, data, 0o644); err != nil {
			t.Fatalf("write override config: %v", err)
		}

		manager := NewManager(cwd)
		cfg, err := manager.LoadWithOverride(overridePath)
		if err != nil {
			t.Fatalf("load with override: %v", err)
		}

		if cfg.Editor != "emacs" {
			t.Errorf("expected editor from override, got %q", cfg.Editor)
		}
		if cfg.Colors.Primary != "03" {
			t.Errorf("expected primary from override, got %q", cfg.Colors.Primary)
		}
	})

	t.Run("falls back to Load when path is empty", func(t *testing.T) {
		root := t.TempDir()
		cwd := filepath.Join(root, "project")
		if err := os.MkdirAll(cwd, 0o755); err != nil {
			t.Fatalf("mkdir cwd: %v", err)
		}
		t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))

		manager := NewManager(cwd)
		cfg, err := manager.LoadWithOverride("")
		if err != nil {
			t.Fatalf("load with empty override: %v", err)
		}

		if cfg.Editor == "" {
			t.Error("expected default config to be loaded")
		}
	})
}

func TestManagerLocalOverridesGlobal(t *testing.T) {
	root := t.TempDir()
	cwd := filepath.Join(root, "project")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatalf("mkdir cwd: %v", err)
	}
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))

	// Create global config
	globalPath := utils.ConfigPathGlobal()
	if err := os.MkdirAll(filepath.Dir(globalPath), 0o755); err != nil {
		t.Fatalf("mkdir global config dir: %v", err)
	}
	globalData := []byte("editor = \"vim\"\n[colors]\nprimary = \"01\"\n")
	if err := os.WriteFile(globalPath, globalData, 0o644); err != nil {
		t.Fatalf("write global config: %v", err)
	}

	// Create local config
	localPath := utils.ConfigPathLocal(cwd)
	if err := os.MkdirAll(filepath.Dir(localPath), 0o755); err != nil {
		t.Fatalf("mkdir local config dir: %v", err)
	}
	localData := []byte("editor = \"emacs\"\n")
	if err := os.WriteFile(localPath, localData, 0o644); err != nil {
		t.Fatalf("write local config: %v", err)
	}

	manager := NewManager(cwd)
	cfg, err := manager.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if cfg.Editor != "emacs" {
		t.Errorf("expected editor from local config, got %q", cfg.Editor)
	}
	if cfg.Colors.Primary != "01" {
		t.Errorf("expected primary from global config, got %q", cfg.Colors.Primary)
	}
}

func TestManagerPartialConfig(t *testing.T) {
	t.Run("only overrides specified fields", func(t *testing.T) {
		root := t.TempDir()
		cwd := filepath.Join(root, "project")
		if err := os.MkdirAll(cwd, 0o755); err != nil {
			t.Fatalf("mkdir cwd: %v", err)
		}
		t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))

		configPath := utils.ConfigPathGlobal()
		if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
			t.Fatalf("mkdir config dir: %v", err)
		}

		// Only set editor, leave other fields as defaults
		data := []byte("editor = \"code\"\n")
		if err := os.WriteFile(configPath, data, 0o644); err != nil {
			t.Fatalf("write config: %v", err)
		}

		manager := NewManager(cwd)
		cfg, err := manager.Load()
		if err != nil {
			t.Fatalf("load: %v", err)
		}

		if cfg.Editor != "code" {
			t.Errorf("expected editor from config, got %q", cfg.Editor)
		}
		// Check that defaults are still present
		if cfg.Colors.Primary != "02" {
			t.Errorf("expected default primary, got %q", cfg.Colors.Primary)
		}
		if cfg.Colors.Headings != "15" {
			t.Errorf("expected default headings, got %q", cfg.Colors.Headings)
		}
	})
}

func TestManagerColorOverrides(t *testing.T) {
	root := t.TempDir()
	cwd := filepath.Join(root, "project")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatalf("mkdir cwd: %v", err)
	}
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))

	configPath := utils.ConfigPathGlobal()
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}

	data := []byte(`
[colors]
headings = "10"
primary = "04"
secondary = "05"
text = "09"
text_highlight = "11"
description_highlight = "12"
tags = "14"
flags = "15"
muted = "07"
accent = "16"
border = "06"
`)
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	manager := NewManager(cwd)
	cfg, err := manager.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"Headings", cfg.Colors.Headings, "10"},
		{"Primary", cfg.Colors.Primary, "04"},
		{"Secondary", cfg.Colors.Secondary, "05"},
		{"Text", cfg.Colors.Text, "09"},
		{"TextHighlight", cfg.Colors.TextHighlight, "11"},
		{"DescriptionHighlight", cfg.Colors.DescriptionHighlight, "12"},
		{"Tags", cfg.Colors.Tags, "14"},
		{"Flags", cfg.Colors.Flags, "15"},
		{"Muted", cfg.Colors.Muted, "07"},
		{"Accent", cfg.Colors.Accent, "16"},
		{"Border", cfg.Colors.Border, "06"},
	}

	for _, tt := range tests {
		if tt.got != tt.expected {
			t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.expected)
		}
	}
}

func TestManagerBooleanOverrides(t *testing.T) {
	root := t.TempDir()
	cwd := filepath.Join(root, "project")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatalf("mkdir cwd: %v", err)
	}
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))

	configPath := utils.ConfigPathGlobal()
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}

	data := []byte("show_taskless = true\n")
	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	manager := NewManager(cwd)
	cfg, err := manager.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if !cfg.ShowTaskless {
		t.Error("expected show_taskless to be true from config")
	}
}

func TestManagerIndexFoldersTOML(t *testing.T) {
	t.Run("single string index_folders", func(t *testing.T) {
		root := t.TempDir()
		cwd := filepath.Join(root, "project")
		_ = os.MkdirAll(cwd, 0o755)
		t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))

		configPath := utils.ConfigPathGlobal()
		_ = os.MkdirAll(filepath.Dir(configPath), 0o755)

		data := []byte("index_folders = \"~/dotfiles/*.zsh\"\n")
		_ = os.WriteFile(configPath, data, 0o644)

		manager := NewManager(cwd)
		cfg, err := manager.Load()
		if err != nil {
			t.Fatalf("Load single string index_folders: %v", err)
		}
		if len(cfg.IndexFolders) != 1 || cfg.IndexFolders[0] != "~/dotfiles/*.zsh" {
			t.Errorf("unexpected index_folders: %+v", cfg.IndexFolders)
		}
	})

	t.Run("array string index_folders", func(t *testing.T) {
		root := t.TempDir()
		cwd := filepath.Join(root, "project")
		_ = os.MkdirAll(cwd, 0o755)
		t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))

		configPath := utils.ConfigPathGlobal()
		_ = os.MkdirAll(filepath.Dir(configPath), 0o755)

		data := []byte("index_folders = [\"~/dotfiles/*.zsh\", \"!~/dotfiles/ignored.zsh\"]\n")
		_ = os.WriteFile(configPath, data, 0o644)

		manager := NewManager(cwd)
		cfg, err := manager.Load()
		if err != nil {
			t.Fatalf("Load array string index_folders: %v", err)
		}
		if len(cfg.IndexFolders) != 2 || cfg.IndexFolders[0] != "~/dotfiles/*.zsh" || cfg.IndexFolders[1] != "!~/dotfiles/ignored.zsh" {
			t.Errorf("unexpected index_folders: %+v", cfg.IndexFolders)
		}
	})
}

func TestManagerLoadsColorOverrides(t *testing.T) {
	root := t.TempDir()
	cwd := filepath.Join(root, "project")
	_ = os.MkdirAll(cwd, 0o755)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))

	configPath := utils.ConfigPathGlobal()
	_ = os.MkdirAll(filepath.Dir(configPath), 0o755)

	tomlContent := `
show_taskless = true

[colors]
title = "12"
highlight_border = "13"
name = "14"
highlight_name = "15"
tasks = "16"
progress_bg = "17"
progress_fg = "18"
highlight_progress_bg = "19"
highlight_progress_fg = "20"
`
	_ = os.WriteFile(configPath, []byte(tomlContent), 0o644)

	manager := NewManager(cwd)
	cfg, err := manager.Load()
	if err != nil {
		t.Fatalf("Load color overrides failed: %v", err)
	}

	if cfg.Colors.Title != "12" {
		t.Errorf("expected Colors.Title 12, got %q", cfg.Colors.Title)
	}
	if cfg.Colors.HighlightBorder != "13" {
		t.Errorf("expected Colors.HighlightBorder 13, got %q", cfg.Colors.HighlightBorder)
	}
	if cfg.Colors.Name != "14" {
		t.Errorf("expected Colors.Name 14, got %q", cfg.Colors.Name)
	}
	if cfg.Colors.HighlightName != "15" {
		t.Errorf("expected Colors.HighlightName 15, got %q", cfg.Colors.HighlightName)
	}
	if cfg.Colors.Tasks != "16" {
		t.Errorf("expected Colors.Tasks 16, got %q", cfg.Colors.Tasks)
	}
	if cfg.Colors.ProgressBg != "17" {
		t.Errorf("expected Colors.ProgressBg 17, got %q", cfg.Colors.ProgressBg)
	}
	if cfg.Colors.ProgressFg != "18" {
		t.Errorf("expected Colors.ProgressFg 18, got %q", cfg.Colors.ProgressFg)
	}
	if cfg.Colors.HighlightProgressBg != "19" {
		t.Errorf("expected Colors.HighlightProgressBg 19, got %q", cfg.Colors.HighlightProgressBg)
	}
	if cfg.Colors.HighlightProgressFg != "20" {
		t.Errorf("expected Colors.HighlightProgressFg 20, got %q", cfg.Colors.HighlightProgressFg)
	}
	if !cfg.ShowTaskless {
		t.Errorf("expected ShowTaskless true, got false")
	}
}

func TestManagerRainbowTOML(t *testing.T) {
	t.Run("boolean rainbow", func(t *testing.T) {
		root := t.TempDir()
		cwd := filepath.Join(root, "project")
		_ = os.MkdirAll(cwd, 0o755)
		t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))

		configPath := utils.ConfigPathGlobal()
		_ = os.MkdirAll(filepath.Dir(configPath), 0o755)

		_ = os.WriteFile(configPath, []byte("rainbow = true\n"), 0o644)

		manager := NewManager(cwd)
		cfg, err := manager.Load()
		if err != nil {
			t.Fatalf("Load rainbow bool failed: %v", err)
		}
		if rb, ok := cfg.Rainbow.(bool); !ok || !rb {
			t.Errorf("expected Rainbow true, got %v", cfg.Rainbow)
		}
	})

	t.Run("array rainbow", func(t *testing.T) {
		root := t.TempDir()
		cwd := filepath.Join(root, "project")
		_ = os.MkdirAll(cwd, 0o755)
		t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))

		configPath := utils.ConfigPathGlobal()
		_ = os.MkdirAll(filepath.Dir(configPath), 0o755)

		_ = os.WriteFile(configPath, []byte("rainbow = [\"#ff0000\", \"#00ff00\", \"#0000ff\"]\n"), 0o644)

		manager := NewManager(cwd)
		cfg, err := manager.Load()
		if err != nil {
			t.Fatalf("Load rainbow array failed: %v", err)
		}
		colors, ok := cfg.Rainbow.([]string)
		if !ok || len(colors) != 3 || colors[0] != "#ff0000" {
			t.Errorf("expected Rainbow array, got %v", cfg.Rainbow)
		}
	})
}

