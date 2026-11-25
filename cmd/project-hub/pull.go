package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/project-hub/internal/config"
	"github.com/project-hub/internal/export"
	"github.com/project-hub/internal/flags"
	"github.com/project-hub/internal/importer"
	"github.com/project-hub/internal/todo"
)

func newPullCmd() *cobra.Command {
	var configPath string
	cmd := &cobra.Command{
		Use:   "pull",
		Short: "Pull all projects in the workspace (uses your git pull config)",
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
			if info, err := os.Stat(workspace); err != nil || !info.IsDir() {
				return fmt.Errorf("workspace %s is not a directory", export.TildePath(workspace))
			}

			// Only the workspace: index_folders are not pulled.
			scanCfg := cfg
			scanCfg.WorkspaceLocation = workspace
			scanCfg.IndexFolders = nil
			dirs, err := todo.NewScanner(scanCfg).ResolveProjectDirs(workspace)
			if err != nil {
				return err
			}

			cmd.Printf("pulling %d projects in %s\n\n", len(dirs), export.TildePath(workspace))
			return reportResults(cmd, cfg, "pull", len(dirs), func(i int) importer.Result {
				return importer.Pull(dirs[i])
			})
		},
	}
	flags.Set(cmd, &configPath, "config", "c", "config file path")
	return cmd
}
