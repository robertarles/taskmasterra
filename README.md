# Taskmasterra

A task management tool for maintaining markdown-based todo lists with journal and archive capabilities.

## Installation

You can install taskmasterra using Go:

```bash
# Latest version (version info will be 'latest')
go install github.com/robertarles/taskmasterra/v2/cmd/taskmasterra@v2.0.21

```

## Usage

```bash
# Show version
taskmasterra version

# Show help
taskmasterra help

# Update calendar with today's tasks
taskmasterra updatecal -i path/to/todo.md

# Record tasks to journal and archive
taskmasterra recordkeep -i path/to/todo.md
```

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

Manage my Reminders with `updatecal` which creates/updates active (/\[.\] !! /) todo items in MacOS Reminders.app

`updatecal` is "slightly mis-named" because Reminders.app is obviously not Calendar.app, but dated reminders DO appear in Calendar.app, and the goal is to have tasks that can be marked completed, -AND- show in my calendar.

taskmasterra replaces the existing notesutil written in Rust.

## Use

`taskmasterra recordkeep -i <markdownfilepath>`

completed(- [x]) will be removed after being added to the archive.

Touched(capital status, e.g. [BWX]) will be added to the journal.

`taskmasterra updatecal -i <markdownfilepath>`

All tasks are added to Reminders.app, and Active(!!) + Touched(capital status [BW]) that are not-completed, will be added to TODAY reminders.

Markdown file, task examples:

``` markdown
- [w] !! A1 task is active (!!) and has some work done ([w]), A priority and fibonacci est effort 1, active or touched so added to reminders due today
- [X] B1 task that is touched (capitalized) today, and is completed ([X]) active or touched (capital X) so added to reminders due today
- [b] B2 blocked task that is not active or touched today. will not be added to reminders
- [W] C8 worked on, and touched today (capital W)
```

Taskra (seperate nvim extension) will highlight the task priority with fibonacci entry

## Building and Versioning

- Change the version in this readme, and in the main.go file
- ./build.sh...
- git add and commit...

Versioning is handled by git tags.

```bash
git tag -a v2.0.4 -m "Version 2.0.5"
git push origin v2.0.5
```
