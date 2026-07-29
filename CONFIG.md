# Configuration

Configuration file location: `$XDG_CONFIG_HOME/super-todo/config.toml`

Most likely `~/.config/super-todo/config.toml`

## Configuration Options

The following options can be set in your configuration file:

### General Settings

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `editor` | string | `nvim` | Editor to use for editing projects and config files |

### Colors Settings (`[colors]`)

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `headings` | string | `"15"` | Headings color |
| `primary` | string | `"02"` | Primary accent color |
| `secondary` | string | `"06"` | Secondary color |
| `text` | string | `"07"` | Text color |
| `text_highlight` | string | `"06"` | Highlighted text color |
| `description_highlight` | string | `"05"` | Highlighted description color |
| `tags` | string | `"13"` | Tags color |
| `flags` | string | `"12"` | Flags color |
| `muted` | string | `"08"` | Muted text color |
| `border` | string | `"08"` | Border color |
| `error` | string | `"01"` | Error color |
| `success` | string | `"02"` | Success color |
| `title` | string | `"15"` | Title color |
| `highlight_border` | string | `"02"` | Highlighted border color |
| `name` | string | `"07"` | Name color |
| `highlight_name` | string | `"10"` | Highlighted name color |
| `tasks` | string | `"13"` | Tasks color |
| `progress_bg` | string | `"08"` | Progress bar background color |
| `progress_fg` | string | `"02"` | Progress bar foreground color |
| `highlight_progress_bg` | string | `"08"` | Highlighted progress bar background color |
| `highlight_progress_fg` | string | `"10"` | Highlighted progress bar foreground color |

### Display Settings

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `home_icon` | string | `~` | Icon to represent home directory in list view |
| `default_sort_by` | string | `newest` | Default sort order. Options: `newest`, `oldest`, `a-z`, `z-a` |
| `list_spacing` | string | `space` | List item spacing. Options: `compact`, `tight`, `space` |

### Todo & Project Settings

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `index_folders` | array | `[]` | Glob patterns to scan for projects to index (e.g. `["~/Projects/*"]`) |
| `ignore_folders` | array | `["node_modules", ...]` | Directories to ignore when scanning for projects |
| `show_taskless` | bool | `false` | When false, hide projects that don't have any tasks |
| `fts_files` | array | `["plan/tasks.md", ...]` | Target files for full text tasks |
| `sync_interval` | int | `3600` | Seconds between project scans |
| `daily_reset` | string | `"04:20"` | Time of day to reset daily.md checkboxes (HH:MM format) |
| `weekly_reset` | string | `"monday"` | Day of week for weekly.md reset at daily_reset time |

## Initializing Configuration

To create a new configuration file with default values:

```bash
super-todo config init
```

To edit your configuration file:

```bash
super-todo config
```
