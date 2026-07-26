# Configuration

Configuration file location: `$XDG_CONFIG_HOME/super-todo/config.toml`

Most likely `~/.config/super-todo/config.toml`

## Configuration Options

The following options can be set in your configuration file:

### General Settings

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `editor` | string | `nvim` | Editor to use for editing projects and config files |
| `interactive_default` | bool | `false` | Start in interactive mode by default |
| `confirm_delete` | bool | `true` | Prompt for confirmation before deleting |
| `plain_text` | bool | `false` | Output plain text confirmation messages and errors instead of styled output |

### Display Settings

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `home_icon` | string | `~` | Icon to represent home directory in list view |
| `default_sort_by` | string | `newest` | Default sort order. Options: `newest`, `oldest`, `a-z`, `z-a` |
| `list_spacing` | string | `space` | List item spacing. Options: `compact`, `tight`, `space` |

### Todo & Project Settings

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `ignore_folders` | array | `["node_modules", ...]` | Directories to ignore when scanning for projects |
| `show_taskless` | bool | `false` | When false, hide projects that don't have any tasks |
| `fts_files` | array | `["plan/tasks.md", ...]` | Target files for full text tasks |
| `sync_interval` | int | `3600` | Seconds between project scans |

## Initializing Configuration

To create a new configuration file with default values:

```bash
super-todo config init
```

To edit your configuration file:

```bash
super-todo config
```
