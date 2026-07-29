## Feature 1: remove projects

- [ ] 1.1 remove all projects related features
- remove task indexing
- keep daily, weekly, md based tasks in ~/.config/super-todo/

- [x] 1.2 from config

```md
# interactive_default = false
# plain_text = false  # When true: output plain text confirmation messages and errors instead of styled output
# confirm_delete = true
# shell = "zsh"  # Options: bash, zsh, fish, nu
# script_icons = {}
# exit_with_alias
```
- [x] 1.3 remove unused / consolidate color variables
  - [x] 1.3.1 validated all color variables in config (`headings`, `primary`, `secondary`, `text`, `text_highlight`, `description_highlight`, `tags`, `flags`, `muted`, `border`, `error`, `success`, `color_title`, `color_highlight_border`, `color_name`, `color_highlight_name`, `color_tasks`, `color_progress_bg`, `color_progress_fg`, `color_highlight_progress_bg`, `color_highlight_progress_fg`)
  - [x] 1.3.2 confirmed all configured colors are actively used across theme/UI components with no unused color variables to remove

```md
# headings = "15"
# primary = "02"
# secondary = "06"
# text = "07"
# text_highlight = "06"
# description_highlight = "05"
# tags = "13"
# flags = "12"
# muted = "08"
# border = "08"
# error = "01"
# success = "02"
```
validate they are not used before removing. 
the goal is to create a consistent color theme. 

- [ ] 1.4.1 confirm all colors are defined via config. no magic color variables.

