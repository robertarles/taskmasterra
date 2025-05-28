# TaskMasterRa

Manage my tasks with `recordkeep` for journaling and archiving activities.

Manage my Reminders with `updatecal` which creates/updates active (/[.:]/) todo items in MacOS Reminders.app

`updatecal` is "slightly mis-named" because Reminders.app is obviously not Calendar.app, but dated reminders DO appear in Calendar.app, and the goal is to have tasks that can be marked completed, -AND- show in my calendar.

taskmasterra replaces the existing notesutil written in Rust.

## Use

`taskmasterra recordkeep -i <markdownfilepath>`

completed(- [x]) will be removed after being added to the archive.

Touched(:) will be added to the journal.

`taskmasterra updatecal -i <markdownfilepath>`

Active(.) and Touched(:) will be added to reminders.

Markdown file, task examples:

``` markdown
- [w] . A1 task is active (.) and has some work done ([w]), A priority and fibonacci est effort 1, active or touched so added to reminders due today
- [x] : B1 task that is touched (:) today, and is completed ([x]) active or touched so added to reminders due today
- [b] B2 blocked task that is not active or touched today. will not be added to reminders\
```

Taskra (seperate nvim extension) will highlight the task priority with fibonacci entry

## Installation

### Using `go`

#### For the latest version

```bash
go install github.com/robertarles/taskmasterra/v2@latest
```

Or for a specific version:

```bash
go install github.com/robertarles/taskmasterra/v2@v2.0.2
```

Transfer the binary to the target Mac (using any method like AirDrop, scp, or a USB drive):

### Example install using scp (from your current machine)

Download the appropriate binary and rename it to `taskmasterra`

#### Make it executable

```bash
chmod +x ~/taskmasterra
```

#### Move it to a location in your path (e.g. /usr/local/bin (requires sudo))

```bash
sudo mv ~/taskmasterra /usr/local/bin/
taskmasterra --help
```

## Building and Versioning

- Change the version in this readme, and in the main.go file
- ./build.sh...
- git add and commit...

Versioning is handled by git tags.

```bash
git tag -a v2.0.3 -m "Version 2.0.3"
git push origin v2.0.3
```
