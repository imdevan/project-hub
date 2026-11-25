# project-hub

A hub for your projects.

## Features

- TUI for scanning and viewing todo tasks across projects
- Interactive task and project listing

## Commands

```bash
project-hub                 # Opens project and todo TUI list
project-hub config          # View or edit configuration
project-hub config init     # Generate default config file
project-hub export          # Write projects.lock (dirs, remotes, branches, commit) and print tree
project-hub export -o f     # Write lock file to path f
project-hub import          # Clone (or pull if present) projects from <workspace>/projects.lock into the workspace
project-hub import -l f     # Import from lock file f
project-hub -w ~/Projects   # Override workspace_location for any command
project-hub completion      # Generate shell completion scripts
```

## Configuration

Configuration file location: `~/.config/project-hub/config.toml` (`$XDG_CONFIG_HOME/project-hub/config.toml`)

See [CONFIG.md](./CONFIG.md) for configuration details.

## How project-hub works

**Parsing**

- Discovery: `workspace_location` is always walked, and directories matching `index_folders` globs are walked in addition to it. Any directory containing `.git` is a project root. Folders in `ignore_folders` are skipped. Without either, the current directory is walked.
- Import: `import` looks for `<workspace>/projects.lock`, then `<workspace>/projects.lock/projects.lock`. Each project is cloned from its `origin` (or first) remote into the workspace, keeping its path relative to the exported workspace (or the projects' common parent); other remotes are added and the recorded branch is checked out. Projects that already exist are updated with `git pull --ff-only`.
- Task files: files matching `fts_files` (e.g. `plan/tasks.md`, `todo.md`) are parsed as markdown. Lines starting with `fts_tasks` / `fts_subtasks` prefixes (`- [ ] `) count as tasks; `[x]` counts as completed. Lines starting with `fts_features` (`## Feature `) group tasks into features.
- Code TODOs: every other file is searched for `todo_tokens` (`TODO`, `FIXME`). Binary files are skipped.

**Storage**

- Scan results are cached as JSON in a SQLite database at `~/.local/share/project-hub/index.db` (`$XDG_DATA_HOME/project-hub/index.db`).
- The same database stores pinned projects, tmux aliases, and ignored task files.
- Cache is reused until `sync_interval` seconds pass (`0` always rescans). Press `s` in the TUI to force a rescan.

**Display**

- Projects load from the cache into an inline Bubble Tea list (not full screen).
- Each project shows task progress (completed / total, summed over non-ignored task files) and its TODO count.
- Sorted with pinned projects first, then by name. Colors and spacing come from `[colors]` and `list_spacing`.
