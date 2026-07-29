# Gloal

create project-hub 

a todo manager for all of my local projects. 

# Context


This project is a fork of a another project I made for managing aliases. 

That project has been renamed to project-hub. But the functionality is all still here. 

It saves shell aliases to a file, and also indexes aliases listed in config. In order to quickly search for and show all of the users aliases. 

I want to repurpose this project to index and and show all of my current project tasks. 

As well as any todo items in a given project folder. 

I track most of my personal projects via tasks.


# Definitions

**fts** — feature/task/subtask; used when referring to an action that could affect any of the three levels

**feature**
```
## Feature {feature_number}: Feature Name
```

**task**
```
- [ ] {feature_number}.{task_number} Task description
```

**subtask**
```
- [ ] {feature_number}.{task_number}.{subtask_number} Subtask description
```

# Commands

root: show interactive tui listing projects
soft remove all other commands

# UI

active project has ansi border. 

all projects have 

Name
progress bar
( ## / ## ) tasks complete + # todos

- name: based on project folder by .git ? 
- progress bar: ansi bar showing percent complete of tasks 
- (number of tasks complete / total) + todos based on todo comments in code

# Indexing

add ignore_folders prop, by default ignore node_modules, dist, bin, and other standard build folders
use index folders (same as current indexing) to find projects. 
add fts tokens to config to define which tasks to look for. allow for multiple definitions of each token to account for different project shapes
