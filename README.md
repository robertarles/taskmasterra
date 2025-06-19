# Taskmasterra

A task management tool for maintaining markdown-based todo lists with journal and archive capabilities.

## Installation

You can install taskmasterra using Go:

```bash
# Latest version (version info will be 'latest')
go install github.com/robertarles/taskmasterra/v2/cmd/taskmasterra@v2.0.22

```

## Usage

```bash
# Show version
taskmasterra version

# Show help
taskmasterra help

# Update reminders with today's tasks
taskmasterra updatereminders -i path/to/todo.md

# Record tasks to journal and archive
taskmasterra recordkeep -i path/to/todo.md

# Generate task statistics report
taskmasterra stats -i path/to/todo.md [-o path/to/report.md]

# Validate task file format
taskmasterra validate -i path/to/todo.md

# Manage configuration
taskmasterra config -init    # Initialize default config
taskmasterra config -show    # Show current config
```

## New Features

### Task Statistics (`stats` command)
Generate comprehensive reports about your tasks:
- Task completion rates
- Priority breakdown
- Effort estimation analysis
- Progress tracking

### Task Validation (`validate` command)
Validate your task file format and get suggestions:
- Check task syntax
- Verify priority and effort formats
- Identify potential issues
- Get formatting suggestions

### Configuration Management (`config` command)
Manage application settings:
- Initialize default configuration
- View current settings
- Customize reminder times, file suffixes, and more

### Enhanced Task Priority System
Support for priority levels and effort estimation:
- Priority levels: A (Critical), B (High), C (Medium), D (Low)
- Fibonacci effort estimation: 1, 2, 3, 5, 8, 13, 21, 34, 55, 89
- Example: `- [ ] A1 !! Critical task with effort 1`

## Development

Build and test locally:

```bash
# Build all platforms
make cross-build

# Run tests
make test
# -OR-
make test-coverage # to generate coverage/coverage.html

# Format code
make fmt

# Create a new release
make release-patch  # For patch version (x.y.Z)
make release-minor  # For minor version (x.Y.0)
make release-major  # For major version (X.0.0)
```

Manage my tasks with `recordkeep` for journaling and archiving activities.

Manage my Reminders with `updatereminders` (previously named `updatecal`) which creates/updates active (/\[.\] !! /) todo items in MacOS Reminders.app

The command was previously named `updatecal` because dated reminders DO appear in Calendar.app, and the goal is to have tasks that can be marked completed, -AND- show in my calendar.

taskmasterra replaces the existing notesutil written in Rust.

## Use

`taskmasterra recordkeep -i <markdownfilepath>`

completed(- [x]) will be removed after being added to the archive.

Touched(capital status, e.g. [BWX]) will be added to the journal.

`taskmasterra updatereminders -i <markdownfilepath>`

All tasks are added to Reminders.app, and Active(!!) + Touched(capital status [BW]) that are not-completed, will be added to TODAY reminders.

Markdown file, task examples:

``` markdown
- [w] !! A1 task is active (!!) and has some work done ([w]), A priority and fibonacci est effort 1, active or touched so added to reminders due today
- [X] B1 task that is touched (capitalized) today, and is completed ([X]) active or touched (capital X) so added to reminders due today
- [b] B2 blocked task that is not active or touched today. will not be added to reminders
- [W] C8 worked on, and touched today (capital W)
```

Taskra (seperate nvim extension) will highlight the task priority with fibonacci entry

## Configuration

Taskmasterra supports configuration via a JSON file located at `~/.taskmasterra/config.json`. You can initialize this file using:

```bash
taskmasterra config -init
```

Configuration options:
- `default_due_hour`: Default hour for reminder due times (default: 16)
- `default_due_minute`: Default minute for reminder due times (default: 0)
- `reminder_list_name`: Name of the Reminders.app list (default: "Taskmasterra")
- `journal_suffix`: Suffix for journal files (default: ".xjournal.md")
- `archive_suffix`: Suffix for archive files (default: ".xarchive.md")
- `active_marker`: Marker for active tasks (default: "!!")

## Building and Versioning

- Change the version in this readme, and in the main.go file
- ./build.sh...
- git add and commit...

Versioning is handled by git tags.

```bash
git tag -a v2.0.4 -m "Version 2.0.5"
git push origin v2.0.5
```
