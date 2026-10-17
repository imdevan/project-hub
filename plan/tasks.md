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

## Feature 6: rebase projects
- [ ] 6.1 attempt to rebase all projects

## Feature 7: online projects

## Feature 8: projects are auto deleted from system after certain period of time but opening the project re clones it? 

