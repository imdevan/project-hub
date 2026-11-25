## Feature 1: remove .config/project-hub/todos.md

- [x] 1.1 remove .config/project-hub/todos.md related features
  - [x] 1.1.1 remove special case recurring md files (`daily.md`, `weekly.md`, `todo.md`) scanner logic and reset mechanisms (`recurring.go`)
  - [x] 1.1.2 remove recurring tasks UI handling, spacebar toggle keybindings, and custom metadata rendering
  - [x] 1.1.3 remove `daily_reset` and `weekly_reset` configuration options and clean up domain/config models

## Feature 2: open with tmux config
- [x] 2.1 open_with_tmux if true will open project +nvim in a new tmux window with the name of the project
  - [x] 2.1.1 add `open_with_tmux` boolean configuration option to domain config, manager, and templates
  - [x] 2.1.2 add `BuildCommandString` helper to editor adapter for shell string formatting
  - [x] 2.1.3 update project opening logic in `ui.go` and `project_ui.go` to launch a detached `tmux new-window` with project name and directory when `open_with_tmux` is enabled
- [x] 2.2 add add tmux alias
  - [x] 2.2.1 hotkey `a` to add/rename tmux alias, persisting across syncs stored in db
  - [x] 2.2.2 keep project name in list and detail view but use alias to name tmux window if present


## Feature 3: ignore task file
- [x] 3.1 a user can ignore a given projects task file
  - [x] 3.1.1 hotkey `i` on highlighted task file in tasks tab to ignore file
  - [x] 3.1.2 prompt user to confirm choice using `internal/ui` confirmation element
  - [x] 3.1.3 ignored task file is hidden from project task list and omitted from total/completed task counts
- [x] 3.2 a user can hotkey shift+I to see ignored task files

## Feature 4: export projects list
- [x] 4.1 add `export` command outputs tree list of dirs, and projects with remotes
  - [x] 4.1.1 add `internal/export` package collecting path, remotes, branches, current branch, and commit per project
  - [x] 4.1.2 render dir tree with remotes under each project
  - [x] 4.1.3 write `projects.lock` (TOML, `~`-relative paths), `-o` flag to override path
  - [x] 4.1.4 wire `export` cobra command, add test, update `README.md`
- we will be importing these later


```
projects.lock?
```

open question: do i store commit hashes? - certainly branches... 
## Feature 5: import projects
- [x] 5.1 add workspace_location to config options
  - [x] 5.1.1 add optional flag for workspace location (persistent `--workspace`/`-w` overrides `workspace_location`)
  - [x] 5.1.2 add `workspace_location` to domain config, config manager, config template, `example-config.toml`, `CONFIG.md`
  - [x] 5.1.3 `export` records `workspace` in `projects.lock` so import can keep paths relative to it
- [x] 5.2 add import command
  - [x] 5.2.1 import command should attempt to clone projects from projects.lock into workspace (`origin` or first remote, other remotes added, recorded branch checked out)
  - [x] 5.2.2 if a project already exists in the location pull (plain `git pull`, honoring the user's pull config)
  - [x] 5.2.3 project-hub should check workspace/projects.lock then workspace/projects.lock/projects.lock for a projects.lock file
  - [x] 5.2.4 optional flag to use different lock file (`--lock`/`-l`)
  - [x] 5.2.5 add `internal/importer` package (find/read lock, map destinations, clone/pull) with tests
  - [x] 5.2.6 print per-project status (cloned/pulled/skipped/failed) and summary; non-zero exit on failures
- [x] 5.3 update index_folders function to always index the workspace_location and index_folders are folders that are indexed in addition to the workspace
  - [ ] 5.4 update index_folders config comments
  - [x] 5.4.2 add scanner test for workspace + index_folders discovery


## Feature 6: rebase projects
- [x] 6.1 add pull command
  - [x] 6.2 attempt to pull all projects in workspace (projects discovered under `workspace_location` / `--workspace`; `index_folders` are not pulled)
  - [x] 6.1.1 add `importer.Pull` (plain `git pull` honoring the user's `pull.rebase`/`pull.ff` config, skips repos with no upstream branch) and reuse it in `import`, with test
  - [x] 6.1.2 share per-project status output and summary between `import` and `pull`; non-zero exit on failures
  - [x] 6.1.3 update `README.md`
- [ ] 6.3 check repo status before pulling, and only pull repos that need it
  - [ ] 6.3.1 skip repos with uncommitted changes (`git status --porcelain` non-empty) and report them as `dirty`
  - [ ] 6.3.2 `git fetch` each repo, then compare with upstream using `git rev-list --left-right --count HEAD...@{u}`
  - [ ] 6.3.3 `0 0` → report `up to date`, skip pull
  - [ ] 6.3.4 ahead only → report `ahead N`, skip pull
  - [ ] 6.3.5 behind only → `git pull` (fast-forwards under any config), report `behind N` / pulled
  - [ ] 6.3.6 split from remote (ahead and behind) → `git pull` using the user's pull config, report `diverged`
  - [ ] 6.3.7 update summary line to show the new statuses (and hide counts that don't apply, e.g. `cloned` under `pull`)

## Feature 7: online projects

## Feature 8: projects are auto deleted from system after certain period of time but opening the project re clones it? 

## Feature 9: run pull / import in parallel
- [ ] 9.1 run the per-repo work (status check, fetch, pull, clone) for several repos at once instead of one by one
  - [ ] 9.1.1 bounded worker pool (default e.g. 8), configurable with a `parallel` config option and `--jobs`/`-j` flag
  - [ ] 9.1.2 apply to both `pull` and `import`
  - [ ] 9.1.3 keep output readable: print results in a stable order (or live per-repo progress with a Bubble Tea inline spinner/list)
  - [ ] 9.1.4 update `README.md`, `CONFIG.md`, `example-config.toml`

