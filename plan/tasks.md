# Context

This project is a fork of another project made for managing aliases. The project has been renamed to super-todo, but the functionality remains. It is being repurposed to index and show all current project tasks and todo items in given project folders.

# Definitions

- **fts**: feature/task/subtask; used when referring to an action that could affect any of the three levels
- **feature**: Defined by a markdown header matching `## Feature {feature_number}: Feature Name`
- **task**: Defined by `- [ ] {feature_number}.{task_number} Task description`
- **subtask**: Defined by `- [ ] {feature_number}.{task_number}.{subtask_number} Subtask description`

# v0.1.0

## Feature 1: Commands & CLI Cleanup
  - [x] 1.1 Update CLI entrypoints and structure
    - notes: Align commands with new super-todo scope.
    - [x] 1.1.1 Configure root command to launch the project-listing interactive TUI
    - [x] 1.1.2 Soft remove or disable unrelated alias/bookmark commands
    - [x] 1.1.3 Rename references from the old project/alias schema to super-todo

## Feature 2: Project & Todo Indexing
  - [x] 2.1 Implement project scanner and discovery
    - notes: Locate project root folders based on configuration and directory markers.
    - [x] 2.1.1 Use existing indexing logic to scan target directories for projects
    - [x] 2.1.2 Identify project name from folder name or .git repository
    - [x] 2.1.3 Support standard ignore_folders (e.g., node_modules, dist, bin) in scanner
  - [x] 2.2 ignore_folders should be included in config init

  - [x] 2.3 Scan and parse codebase TODO comments
    - notes: Scan files within projects to extract todo count.
    - [x] 2.3.1 Traverse project directories while respecting ignore lists
    - [x] 2.3.2 Count todo comments (e.g., // TODO, # TODO)
  - [x] 2.4 refine project traversal
    - [x] 2.4.1 index_folders can and likely do contain multiple projects
    - [x] 2.4.2 projects should be parsed by detecting a .git directory in the project root. 
    - tasks are tokens identied above in md files
    - todo items are todo comments in a given project, they are not tasks, and are countered separately.
    - ignore config.ignore_folders when indexing
  
## Feature 3: index in the background
- [x] 3.1 lazy index projects. 
  - [x] 3.1.1 load tui first.
  - [x] 3.1.2 show current project statuses 
  - [x] 3.1.3 show loading index progress on bottom of tui while indexing in the background

## Feature 4: add sync_interval to config
- [ ] 4.1 add sync_interval to config
- only index after interval. time in seconds or time of day. 
- default to 1 minute?
- is there a way to do it on change? or on hover?

## Feature 5: break down each ui element give it's own color in config
- [x] 5.1 Add title, highlight_border, name, highlight_name, tasks, progress_bg, progress_fg, highlight_progress_bg, highlight_progress_fg to config overrides
- [x] 5.2 Default to current color mapping when not specified



## Feature 6: Task & Feature (FTS) Parsing
  - [x] 6.1 Implement Markdown task parser for FTS files
    - notes: Parse markdown planning files matching the FTS definitions.
    - [x] 6.1.1 Scan project folders for task files based on configure patterns
    - [x] 6.1.2 Parse Features, Tasks, and Subtasks based on markdown headers and checkbox patterns
    - [x] 6.1.3 Calculate completion percentage based on task and subtask status (checked vs unchecked)

## Feature 7: Interactive TUI
  - [x] 7.1 Design and build the projects list TUI
    - notes: Use Bubble Tea/Bubbles to build the TUI inline.
    - [x] 7.1.1 Create list view of discovered projects
    - [x] 7.1.2 Display progress bar and completion ratio (complete / total) for each project
    - [x] 7.1.3 Display total count of source code TODO items
    - [x] 7.1.4 Style the active project list item with an ANSI border

## Feature 8: Configuration Updates
  - [x] 8.1 Update configuration model and default TOML
    - notes: Add configuration fields supporting the new indexer and parser.
    - [x] 8.1.1 Add ignore_folders property to the config structure
    - [x] 8.1.2 Add fts tokens and definitions to the config to identify task files
    - [x] 8.1.3 Update example-config.toml and documentation
