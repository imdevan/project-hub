package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/project-hub/internal/config"
	"github.com/project-hub/internal/export"
	"github.com/project-hub/internal/flags"
	"github.com/project-hub/internal/todo"
)

func newExportCmd() *cobra.Command {
	var output string
	var configPath string
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export project tree with git remotes to a projects.lock file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			cfg := config.Load(cwd, configPath)
			applyWorkspaceFlag(cmd, &cfg)
			dirs, err := todo.NewScanner(cfg).ResolveProjectDirs(cwd)
			if err != nil {
				return err
			}
			if output == "" {
				output = export.DefaultLockFile
			}
			lock := export.Collect(dirs)
			if cfg.WorkspaceLocation != "" {
				lock.Workspace = export.TildePath(expandHome(cfg.WorkspaceLocation))
			}
			cmd.Print(export.Tree(lock))
			if err := export.Write(lock, output); err != nil {
				return err
			}
			cmd.Printf("\nwrote %s (%d projects)\n", output, len(lock.Projects))
			return nil
		},
	}
	flags.Set(cmd, &output, "output", "o", "lock file path (default projects.lock)")
	flags.Set(cmd, &configPath, "config", "c", "config file path")
	return cmd
}
