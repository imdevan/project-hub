package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"

	"github.com/project-hub/internal/config"
	"github.com/project-hub/internal/flags"
	"github.com/project-hub/internal/index"
	pkg "github.com/project-hub/internal/package"
	"github.com/project-hub/internal/todo"
	"github.com/project-hub/internal/ui"
)

// Metadata loaded from package.toml at build time
var (
	version = pkg.Version()
	name    = pkg.Name()
	short   = pkg.Short()
)

type rootOptions struct {
	configPath  string
	showVersion bool
}

var rootCmd = newRootCmd()

// Execute is the CLI entrypoint.
func Execute() error {
	return rootCmd.Execute()
}

func newRootCmd() *cobra.Command {
	opts := &rootOptions{}
	cmd := &cobra.Command{
		Use:   name,
		Short: short,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.showVersion {
				ver := resolvedVersion()
				cmd.Printf("%s\n", ver)
				return nil
			}

			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}

			cfg := config.Load(cwd, opts.configPath)

			store, err := index.NewStore()
			if err == nil {
				defer store.Close()
				indexer := index.NewIndexer(store, cfg)
				if indexer.NeedsRefresh() {
					_, _ = ui.RefreshWithSpinner(indexer, ui.ThemeFromConfig(cfg))
				}
			}

			return todo.RunProjectListing(cfg, cwd)
		},
	}

	flags.SetPersistent(cmd, &opts.configPath, "config", "c", "config file path")
	flags.Set(cmd, &opts.showVersion, "version", "v", "print version information")

	cmd.AddCommand(newConfigCmd())
	cmd.AddCommand(newExportCmd())

	return cmd
}

func resolvedVersion() string {
	ver := version
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ver
	}
	if ver == "dev" && strings.TrimSpace(info.Main.Version) != "" && info.Main.Version != "(devel)" {
		ver = info.Main.Version
	}
	return ver
}
