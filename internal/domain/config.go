package domain

import (
	"os"
	"path/filepath"
	"strings"
)

// Config describes the resolved configuration.
type Config struct {
	Editor                   string `toml:"editor"`
	Primary                  string `toml:"primary"`
	Secondary                string `toml:"secondary"`
	Headings                 string `toml:"headings"`
	Text                     string `toml:"text"`
	TextHighlight            string `toml:"text_highlight"`
	DescriptionHighlight     string `toml:"description_highlight"`
	Tags                     string `toml:"tags"`
	Flags                    string `toml:"flags"`
	Muted                    string `toml:"muted"`
	Accent                   string `toml:"accent"`
	Border                   string `toml:"border"`
	Error                    string `toml:"error"`
	Success                  string `toml:"success"`
	ColorTitle               string `toml:"color_title"`
	ColorHighlightBorder     string `toml:"color_highlight_border"`
	ColorName                string `toml:"color_name"`
	ColorHighlightName       string `toml:"color_highlight_name"`
	ColorTasks               string `toml:"color_tasks"`
	ColorProgressBg          string `toml:"color_progress_bg"`
	ColorProgressFg          string `toml:"color_progress_fg"`
	ColorHighlightProgressBg string `toml:"color_highlight_progress_bg"`
	ColorHighlightProgressFg string `toml:"color_highlight_progress_fg"`
	ListSpacing              string `toml:"list_spacing"`

	// System settings
	IndexFolders  []string          `toml:"index_folders"`
	CacheInterval int               `toml:"cache_interval"`
	HomeIcon      string            `toml:"home_icon"`
	DefaultSortBy string            `toml:"default_sort_by"`

	// Todo settings
	IgnoreFolders []string `toml:"ignore_folders"`
	FtsFiles      []string `toml:"fts_files"`
	FtsFeatures   []string `toml:"fts_features"`
	FtsTasks      []string `toml:"fts_tasks"`
	FtsSubtasks   []string `toml:"fts_subtasks"`
	TodoTokens    []string `toml:"todo_tokens"`
	ShowTaskless  bool     `toml:"show_taskless"`
	SyncInterval  int      `toml:"sync_interval"`
	DailyReset    string   `toml:"daily_reset"`
	WeeklyReset   string   `toml:"weekly_reset"`

	// Rainbow settings
	Rainbow interface{} `toml:"rainbow"`
}

// DefaultRainbowAnchors is the keyframe sequence for the default rainbow.
var DefaultRainbowAnchors = []string{
	"#a6da95", // Green
	"#8bd5ca", // Teal
	"#91d7e3", // Sky
	"#7dc4e4", // Sapphire
	"#8aadf4", // Blue
	"#b7bdf8", // Lavender
	"#c6a0f6", // Mauve
	"#f5bde6", // Pink
	"#f0c6c6", // Flamingo
	"#f4dbd6", // Rosewater
	"#eed49f", // Yellow
	"#f5a97f", // Peach
	"#ee99a0", // Maroon
	"#ed8796", // Red
}

// DefaultConfig returns the default configuration values.
func DefaultConfig() Config {
	return Config{
		Editor:                   "nvim",
		Headings:                 "15",
		Primary:                  "02",
		Secondary:                "06",
		Text:                     "07",
		TextHighlight:            "06",
		DescriptionHighlight:     "05",
		Tags:                     "13",
		Flags:                    "12",
		Muted:                    "08",
		Accent:                   "13",
		Border:                   "08",
		Error:                    "01",
		Success:                  "02",
		ColorTitle:               "15",
		ColorHighlightBorder:     "02",
		ColorName:                "07",
		ColorHighlightName:       "10",
		ColorTasks:               "13",
		ColorProgressBg:          "08",
		ColorProgressFg:          "02",
		ColorHighlightProgressBg: "08",
		ColorHighlightProgressFg: "10",
		ListSpacing:              "space",
		IndexFolders:             []string{},
		CacheInterval:            300,
		HomeIcon:                 "~",
		DefaultSortBy:            "newest",

		// Todo defaults
		IgnoreFolders: []string{"node_modules", "dist", "bin", ".git", ".vscode", "vendor", "build", "target", "pkg", ".venv"},
		FtsFiles:      []string{"plan/tasks.md", "plan/simple.md", "todo.md", "tasks.md", "PLAN.md"},
		FtsFeatures:   []string{"## Feature ", "### Feature "},
		FtsTasks:      []string{"- [ ] ", "* [ ] "},
		FtsSubtasks:   []string{"- [ ] ", "* [ ] "},
		TodoTokens:    []string{"TODO", "FIXME"},
		ShowTaskless:  false,
		SyncInterval:  3600,
		DailyReset:    "04:20",
		WeeklyReset:   "monday",
		Rainbow:       false,
	}
}

func expandPath(value string) string {
	expanded := os.ExpandEnv(value)
	if expanded == "" {
		return expanded
	}
	if expanded == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			return home
		}
		return expanded
	}
	if strings.HasPrefix(expanded, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(expanded, "~/"))
		}
	}
	return expanded
}

func xdgHome(envKey, fallbackSuffix string) string {
	if value := os.Getenv(envKey); value != "" {
		return value
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, fallbackSuffix)
}
