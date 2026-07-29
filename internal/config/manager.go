package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"

	"github.com/super-todo/internal/domain"
	"github.com/super-todo/internal/utils"
)

// ManagerImpl loads and saves configuration files.
type ManagerImpl struct {
	cwd string
}

// NewManager returns a config manager rooted at the provided cwd.
func NewManager(cwd string) *ManagerImpl {
	return &ManagerImpl{cwd: cwd}
}

// LoadWithOverride loads config from a specific path, layered on defaults.
func (m *ManagerImpl) LoadWithOverride(path string) (domain.Config, error) {
	config := domain.DefaultConfig()
	if strings.TrimSpace(path) == "" {
		return m.Load()
	}
	partial, err := readConfig(path)
	if err != nil {
		return domain.Config{}, err
	}
	if partial != nil {
		applyPartial(&config, partial)
	}
	return config, nil
}

// Load reads config with precedence: defaults < global < local.
func (m *ManagerImpl) Load() (domain.Config, error) {
	config := domain.DefaultConfig()

	globalPath := utils.ConfigPathGlobal()
	if partial, err := readConfig(globalPath); err != nil {
		return domain.Config{}, err
	} else if partial != nil {
		applyPartial(&config, partial)
	}

	localPath := utils.ConfigPathLocal(m.cwd)
	if partial, err := readConfig(localPath); err != nil {
		return domain.Config{}, err
	} else if partial != nil {
		applyPartial(&config, partial)
	}

	return config, nil
}

// Save persists config to the global config path.
func (m *ManagerImpl) Save(config domain.Config) error {
	path := utils.ConfigPathGlobal()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := toml.Marshal(config)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// Exists reports whether a local or global config file exists.
func (m *ManagerImpl) Exists() (bool, error) {
	globalPath := utils.ConfigPathGlobal()
	if exists, err := fileExists(globalPath); err != nil {
		return false, err
	} else if exists {
		return true, nil
	}
	localPath := utils.ConfigPathLocal(m.cwd)
	return fileExists(localPath)
}

type partialConfig struct {
	Editor                   *string     `toml:"editor"`
	Primary                  *string     `toml:"primary"`
	Secondary                *string     `toml:"secondary"`
	Headings                 *string     `toml:"headings"`
	Text                     *string     `toml:"text"`
	TextHighlight            *string     `toml:"text_highlight"`
	DescriptionHighlight     *string     `toml:"description_highlight"`
	Tags                     *string     `toml:"tags"`
	Flags                    *string     `toml:"flags"`
	Muted                    *string     `toml:"muted"`
	Accent                   *string     `toml:"accent"`
	Border                   *string     `toml:"border"`
	Error                    *string     `toml:"error"`
	Success                  *string     `toml:"success"`
	ColorTitle               *string     `toml:"color_title"`
	ColorHighlightBorder     *string     `toml:"color_highlight_border"`
	ColorName                *string     `toml:"color_name"`
	ColorHighlightName       *string     `toml:"color_highlight_name"`
	ColorTasks               *string     `toml:"color_tasks"`
	ColorProgressBg          *string     `toml:"color_progress_bg"`
	ColorProgressFg          *string     `toml:"color_progress_fg"`
	ColorHighlightProgressBg *string     `toml:"color_highlight_progress_bg"`
	ColorHighlightProgressFg *string     `toml:"color_highlight_progress_fg"`
	InteractiveDefault       *bool       `toml:"interactive_default"`
	PlainText                *bool       `toml:"plain_text"`
	ConfirmDelete            *bool       `toml:"confirm_delete"`
	ListSpacing              *string     `toml:"list_spacing"`
	ShowTaskless             *bool       `toml:"show_taskless"`
	IndexFolders             interface{} `toml:"index_folders"`
	CacheInterval            *int        `toml:"cache_interval"`
	Shell                    *string     `toml:"shell"`
	HomeIcon                 *string     `toml:"home_icon"`
	DefaultSortBy            *string     `toml:"default_sort_by"`
	IgnoreFolders            interface{} `toml:"ignore_folders"`
	FtsFiles                 interface{} `toml:"fts_files"`
	FtsFeatures              interface{} `toml:"fts_features"`
	FtsTasks                 interface{} `toml:"fts_tasks"`
	FtsSubtasks               interface{} `toml:"fts_subtasks"`
	TodoTokens               interface{} `toml:"todo_tokens"`
	SyncInterval             *int        `toml:"sync_interval"`
	DailyReset               *string     `toml:"daily_reset"`
	WeeklyReset              *string     `toml:"weekly_reset"`
	ExitWithAlias            *bool       `toml:"exit_with_alias"`
	Rainbow                  interface{} `toml:"rainbow"`
}

func readConfig(path string) (*partialConfig, error) {
	if exists, err := fileExists(path); err != nil || !exists {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var partial partialConfig
	if err := toml.Unmarshal(data, &partial); err != nil {
		return nil, err
	}
	return &partial, nil
}

func applyPartial(config *domain.Config, partial *partialConfig) {
	if partial.Editor != nil {
		config.Editor = *partial.Editor
	}
	if partial.Primary != nil {
		config.Primary = *partial.Primary
	}
	if partial.Secondary != nil {
		config.Secondary = *partial.Secondary
	}
	if partial.Headings != nil {
		config.Headings = *partial.Headings
	}
	if partial.Text != nil {
		config.Text = *partial.Text
	}
	if partial.TextHighlight != nil {
		config.TextHighlight = *partial.TextHighlight
	}
	if partial.DescriptionHighlight != nil {
		config.DescriptionHighlight = *partial.DescriptionHighlight
	}
	if partial.Tags != nil {
		config.Tags = *partial.Tags
	}
	if partial.Flags != nil {
		config.Flags = *partial.Flags
	}
	if partial.Muted != nil {
		config.Muted = *partial.Muted
	}
	if partial.Accent != nil {
		config.Accent = *partial.Accent
	}
	if partial.Border != nil {
		config.Border = *partial.Border
	}
	if partial.Error != nil {
		config.Error = *partial.Error
	}
	if partial.Success != nil {
		config.Success = *partial.Success
	}
	if partial.ColorTitle != nil {
		config.ColorTitle = *partial.ColorTitle
	}
	if partial.ColorHighlightBorder != nil {
		config.ColorHighlightBorder = *partial.ColorHighlightBorder
	}
	if partial.ColorName != nil {
		config.ColorName = *partial.ColorName
	}
	if partial.ColorHighlightName != nil {
		config.ColorHighlightName = *partial.ColorHighlightName
	}
	if partial.ColorTasks != nil {
		config.ColorTasks = *partial.ColorTasks
	}
	if partial.ColorProgressBg != nil {
		config.ColorProgressBg = *partial.ColorProgressBg
	}
	if partial.ColorProgressFg != nil {
		config.ColorProgressFg = *partial.ColorProgressFg
	}
	if partial.ColorHighlightProgressBg != nil {
		config.ColorHighlightProgressBg = *partial.ColorHighlightProgressBg
	}
	if partial.ColorHighlightProgressFg != nil {
		config.ColorHighlightProgressFg = *partial.ColorHighlightProgressFg
	}
	if partial.InteractiveDefault != nil {
		config.InteractiveDefault = *partial.InteractiveDefault
	}
	if partial.PlainText != nil {
		config.PlainText = *partial.PlainText
	}
	if partial.ConfirmDelete != nil {
		config.ConfirmDelete = *partial.ConfirmDelete
	}
	if partial.ListSpacing != nil {
		config.ListSpacing = *partial.ListSpacing
	}
	if partial.ShowTaskless != nil {
		config.ShowTaskless = *partial.ShowTaskless
	}
	if partial.IndexFolders != nil {
		switch v := partial.IndexFolders.(type) {
		case string:
			config.IndexFolders = []string{v}
		case []interface{}:
			var folders []string
			for _, elem := range v {
				if str, ok := elem.(string); ok {
					folders = append(folders, str)
				}
			}
			config.IndexFolders = folders
		case []string:
			config.IndexFolders = v
		}
	}
	if partial.CacheInterval != nil {
		config.CacheInterval = *partial.CacheInterval
	}
	if partial.Shell != nil {
		config.Shell = *partial.Shell
	}
	if partial.HomeIcon != nil {
		config.HomeIcon = *partial.HomeIcon
	}
	if partial.DefaultSortBy != nil {
		config.DefaultSortBy = *partial.DefaultSortBy
	}
	if folders := parseStringSlice(partial.IgnoreFolders); folders != nil {
		config.IgnoreFolders = folders
	}
	if files := parseStringSlice(partial.FtsFiles); files != nil {
		config.FtsFiles = files
	}
	if features := parseStringSlice(partial.FtsFeatures); features != nil {
		config.FtsFeatures = features
	}
	if tasks := parseStringSlice(partial.FtsTasks); tasks != nil {
		config.FtsTasks = tasks
	}
	if subtasks := parseStringSlice(partial.FtsSubtasks); subtasks != nil {
		config.FtsSubtasks = subtasks
	}
	if tokens := parseStringSlice(partial.TodoTokens); tokens != nil {
		config.TodoTokens = tokens
	}
	if partial.SyncInterval != nil {
		config.SyncInterval = *partial.SyncInterval
	}
	if partial.DailyReset != nil {
		config.DailyReset = *partial.DailyReset
	}
	if partial.WeeklyReset != nil {
		config.WeeklyReset = *partial.WeeklyReset
	}
	if partial.ExitWithAlias != nil {
		config.ExitWithAlias = *partial.ExitWithAlias
	}
	if partial.Rainbow != nil {
		switch v := partial.Rainbow.(type) {
		case bool:
			config.Rainbow = v
		case []interface{}:
			var colors []string
			for _, elem := range v {
				if str, ok := elem.(string); ok {
					colors = append(colors, str)
				}
			}
			config.Rainbow = colors
		case []string:
			config.Rainbow = v
		default:
			config.Rainbow = partial.Rainbow
		}
	}
}

func parseStringSlice(field interface{}) []string {
	if field == nil {
		return nil
	}
	switch v := field.(type) {
	case string:
		return []string{v}
	case []interface{}:
		var result []string
		for _, elem := range v {
			if str, ok := elem.(string); ok {
				result = append(result, str)
			}
		}
		return result
	case []string:
		return v
	}
	return nil
}

func fileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err == nil {
		return !info.IsDir(), nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}
