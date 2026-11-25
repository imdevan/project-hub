package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/project-hub/internal/config"
	"github.com/project-hub/internal/domain"
	"github.com/project-hub/internal/export"
	"github.com/project-hub/internal/flags"
	"github.com/project-hub/internal/importer"
	"github.com/project-hub/internal/ui"
)

func newImportCmd() *cobra.Command {
	var lockPath string
	var configPath string
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Clone (or pull) projects from a projects.lock file into the workspace",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			cfg := config.Load(cwd, configPath)
			workspace, err := resolveWorkspace(cmd, &cfg)
			if err != nil {
				return err
			}

			if lockPath == "" {
				if lockPath, err = importer.FindLock(workspace); err != nil {
					return err
				}
			}
			lock, err := importer.Read(expandHome(lockPath))
			if err != nil {
				return err
			}

			cmd.Printf("importing %d projects from %s into %s\n\n", len(lock.Projects), export.TildePath(lockPath), export.TildePath(workspace))
			dests := importer.Destinations(lock, workspace)
			return reportResults(cmd, cfg, "import", len(dests), func(i int) importer.Result {
				return importer.Import(lock.Projects[i], dests[i])
			})
		},
	}
	flags.Set(cmd, &lockPath, "lock", "l", "lock file path (default <workspace>/projects.lock)")
	flags.Set(cmd, &configPath, "config", "c", "config file path")
	return cmd
}

// reportResults runs n git operations, printing a styled status line for each and a summary.
// It returns an error if any operation failed.
func reportResults(cmd *cobra.Command, cfg domain.Config, verb string, n int, run func(i int) importer.Result) error {
	theme := ui.ThemeFromConfig(cfg)
	styles := map[importer.Action]lipgloss.Style{
		importer.Cloned:   lipgloss.NewStyle().Foreground(theme.Success),
		importer.Pulled:   lipgloss.NewStyle().Foreground(theme.Primary),
		importer.UpToDate: lipgloss.NewStyle().Foreground(theme.Muted),
		importer.Ahead:    lipgloss.NewStyle().Foreground(theme.Secondary),
		importer.Dirty:    lipgloss.NewStyle().Foreground(theme.Flags),
		importer.Skipped:  lipgloss.NewStyle().Foreground(theme.Muted),
		importer.Failed:   lipgloss.NewStyle().Foreground(theme.Error),
	}
	muted := lipgloss.NewStyle().Foreground(theme.Muted)
	width := 0
	for _, a := range importer.Actions {
		width = max(width, len(a)+1)
	}
	indent := "\n" + strings.Repeat(" ", width)

	counts := map[importer.Action]int{}
	for i := 0; i < n; i++ {
		res := run(i)
		counts[res.Action]++
		line := styles[res.Action].Render(fmt.Sprintf("%-*s", width, res.Action)) + export.TildePath(res.Dest)
		if res.Detail != "" {
			line += indent + muted.Render(strings.ReplaceAll(res.Detail, "\n", indent))
		}
		cmd.Println(line)
	}

	// Summary only lists statuses that occurred.
	var parts []string
	for _, a := range importer.Actions {
		if counts[a] > 0 {
			parts = append(parts, styles[a].Render(fmt.Sprintf("%d %s", counts[a], a)))
		}
	}
	if len(parts) == 0 {
		parts = append(parts, "nothing to do")
	}
	cmd.Printf("\n%s\n", strings.Join(parts, ", "))
	if counts[importer.Failed] > 0 {
		return fmt.Errorf("%d projects failed to %s", counts[importer.Failed], verb)
	}
	return nil
}

// resolveWorkspace returns the absolute workspace path from config (after the --workspace override).
func resolveWorkspace(cmd *cobra.Command, cfg *domain.Config) (string, error) {
	applyWorkspaceFlag(cmd, cfg)
	if cfg.WorkspaceLocation == "" {
		return "", fmt.Errorf("no workspace set: add workspace_location to config or pass --workspace")
	}
	return filepath.Abs(expandHome(cfg.WorkspaceLocation))
}

// expandHome expands a leading "~" to the user's home directory.
func expandHome(path string) string {
	if path == "~" || strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(path, "~"))
		}
	}
	return path
}
