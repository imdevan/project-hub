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
- [x] 4.1 add sync_interval to config
  - [x] 4.1.1 add sync_interval field to Config struct (default 3600s = 1 hour)
  - [x] 4.1.2 add sync_interval to partialConfig parsing in config manager
  - [x] 4.1.3 add projects_cache table to SQLite store with SaveProjectsCache/LoadProjectsCache
  - [x] 4.1.4 load cached projects on startup when sync_interval hasn't expired, skip background scan
  - [x] 4.1.5 save scanned projects to cache after indexing completes
  - [x] 4.1.6 add 's' hotkey to force rescan all projects and reset interval
  - [x] 4.1.7 set sync_interval=0 to always rescan (disable caching)
  - [x] 4.1.8 update config template, example config, and help keys

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

## Feature 9: add show_taskless
- [x] 9.1 add show_taskless to config default to false
  - [x] 9.1.1 when false do not show projects without any tasks found


## Feature 10: replace run with cd
- [x] 10.1 replace run_with_cd with exit_with_alias
  - [x] 10.1.1 use go template structure (e.g. cd {{.Project}} && nvim {{.TaskFile}})
  - [x] 10.1.2 keep set/not set handling (if not set open inline in TUI)

## Feature 11: view project
- [x] 11.1 on enter navigate to "Project page"
- [x] 11.2 new view showing list of project todos and tasks

example 

```
 ╭────────────────────────────────────────────────────────────────╮
 │                                                                │
 │   aliases                               ...clis/aliases        │
 │   █████████████████████████████████████████████████████        │
 │   (26/26) tasks complete  3 todos                              │
 │                                                                │
 │ │ Todos:                                                       │
 │ │ // TODO: this is a todo line item                            │
 │ │ .../The Wheel/b                                              │
 │                                                                │
 │   // TODO: this is a todo line item                            │
 │   .../The Wheel/b                                              │
 │                                                                │
 │   plan/tasks.md                                                │
 │   █████████████████████████████████████████████████████        │
 │   (51/51) tasks complete  0 todos  .../The Wheel/l             │
 │                                                                │
 │   plan/kicked.md                                               │
 │   █████████████████████████████████████████████████████        │
 │   (51/51) tasks complete  0 todos  .../The Wheel/l             │

```
- [x] 11.3 the project page should be a completely separate ui from the list view
  - [x] 11.3.1 create separate UI under project_ui.go
  - [x] 11.3.2 project title section showing name (left) and location (justified right), total progress bar, and completion/todo count
  - [x] 11.3.3 scrollable list containing individual highlightable TODO comments (with text and relative path location) and task files (with path, progress bar, completion stats)
  - [x] 11.3.4 enter on TODO comment opens that todo at the correct line number, enter on task file opens that task file
- [x] 11.4 show separate tabs on Project page (TODOs tab and Tasks tab) with separate scrolling, selection state, and pagination

## Feature 12: pin a project
- [x] 12.1 in project list view, hotkey 'p' will pin a project
  - [x] 12.1.1 pinned projects appear at the top of the project lists
  - [x] 12.1.2 use nf-oct-pin (\uf435) to indicate a project is pinned next to its name
  - [x] 12.1.3 follow project to beginning of list and stay focused on item
  - [x] 12.1.4 store pinned state in db


## Feature 13: rainbow
- [x] 13.1 add rainbow to config 
  - [x] 13.1.1 defaults to false
  - [x] 13.1.2 if true or an array present projects in a rainbow cascading color that loops based on length of array
  - [x] 13.1.3 default rainbow value is an array of colors (DefaultRainbowAnchors)

// defaultRainbowAnchors is the keyframe sequence for the default rainbow.
var defaultRainbowAnchors = []lipgloss.Color{
	"#a6da95", // Green
	"#8bd5ca", // Teal
	"#91d7e3", // Sky
	"#7dc4e4", // Sapphire
	"#8aadf4", // Blue
	"#b7bdf8", // Lavender
	"#c6a0f6", // Mauve
	"#f5bde6", // Pink
	"#f0c6c6", // Flamingo
	"#f4dbd6", // Rosewater
	"#eed49f", // Yellow
	"#f5a97f", // Peach
	"#ee99a0", // Maroon
	"#ed8796", // Red
}
