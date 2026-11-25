package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

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
	var jobs int
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
			// Sort by destination so parents come before projects nested inside them; a nested
			// project waits for its parent so parallel clones don't race for the same directory.
			dests := importer.Destinations(lock, workspace)
			order := make([]int, len(dests))
			for i := range order {
				order[i] = i
			}
			sort.SliceStable(order, func(a, b int) bool { return dests[order[a]] < dests[order[b]] })
			sorted := make([]string, len(order))
			labels := make([]string, len(order))
			for i, o := range order {
				sorted[i] = dests[o]
				labels[i] = export.TildePath(dests[o])
			}
			parents := importer.ParentIndex(sorted)
			finished := make([]chan struct{}, len(order))
			for i := range finished {
				finished[i] = make(chan struct{})
			}
			return reportResults(cmd, cfg, "importing", labels, resolveJobs(jobs, cfg), func(i int) importer.Result {
				defer close(finished[i])
				if p := parents[i]; p >= 0 {
					<-finished[p]
				}
				return importer.Import(lock.Projects[order[i]], sorted[i])
			})
		},
	}
	addJobsFlag(cmd, &jobs)
	flags.Set(cmd, &lockPath, "lock", "l", "lock file path (default <workspace>/projects.lock)")
	flags.Set(cmd, &configPath, "config", "c", "config file path")
	return cmd
}

// reportResults runs one git job per label on up to jobs workers, printing a styled status
// line per job (in label order) and a summary. It returns an error if any job failed.
func reportResults(cmd *cobra.Command, cfg domain.Config, verb string, labels []string, jobs int, run func(i int) importer.Result) error {
	theme := ui.ThemeFromConfig(cfg)
	results, runErr := ui.RunGitJobs(cmd.OutOrStderr(), theme, verb, labels, jobs, run)

	counts := map[importer.Action]int{}
	for _, r := range results {
		if r.Action != "" {
			counts[r.Action]++
		}
	}
	// Summary only lists statuses that occurred.
	var parts []string
	for _, a := range importer.Actions {
		if counts[a] > 0 {
			parts = append(parts, ui.GitJobStyle(theme, a).Render(fmt.Sprintf("%d %s", counts[a], a)))
		}
	}
	if len(parts) == 0 {
		parts = append(parts, "nothing to do")
	}
	cmd.Printf("\n%s\n", strings.Join(parts, ", "))
	if runErr != nil {
		return runErr
	}
	if counts[importer.Failed] > 0 {
		return fmt.Errorf("%d projects failed", counts[importer.Failed])
	}
	return nil
}

// addJobsFlag registers --jobs/-j on cmd.
func addJobsFlag(cmd *cobra.Command, jobs *int) {
	flags.Set(cmd, jobs, "jobs", "j", "repositories to process at once (overrides parallel, default 8)")
}

// resolveJobs returns the --jobs flag if set, else the parallel config value, at least 1.
func resolveJobs(flagJobs int, cfg domain.Config) int {
	jobs := cfg.Parallel
	if flagJobs > 0 {
		jobs = flagJobs
	}
	return max(jobs, 1)
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
