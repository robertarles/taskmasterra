# Taskmasterra

A modern CLI tool for managing markdown-based todo lists, with journaling, archiving, validation, and seamless integration with macOS Reminders.

---

## 🚀 Quick Start

1. **Install:**
   ```bash
   go install github.com/robertarles/taskmasterra/v2/cmd/taskmasterra@latest
   ```
1. **Initialize config:**
   ```bash
   taskmasterra config -init
   ```
1. **Start managing your tasks:**
   ```bash
   taskmasterra recordkeep -i path/to/todo.md
   taskmasterra updatereminders -i path/to/todo.md
   ```

---

## 📄 File Format Example

```markdown
# My Tasks
- [ ] !! A1 Write project proposal
  - Draft outline
- [w] B2 !! Review code and add tests
- [x] C3 Submit final report
- [b] D5 Blocked by client feedback
```

**Legend:**
- `[ ]` = open, `[x]` = completed, `[w]` = worked, `[b]` = blocked
- `!!` = active today (must be immediately after status)
- `A1`, `B2`, etc. = priority (A=Critical, B=High, C=Medium, D=Low) and effort (Fibonacci)
- Indented lines are details/notes

---

## Usage

```bash
# Show version
$ taskmasterra version

# Show help
$ taskmasterra help

# Update Reminders.app with today's active tasks
$ taskmasterra updatereminders -i todo.md

# Record completed/touched tasks to journal/archive
$ taskmasterra recordkeep -i todo.md

# Generate a statistics report
$ taskmasterra stats -i todo.md -o report.md

# Validate your todo file
$ taskmasterra validate -i todo.md

# Manage configuration
$ taskmasterra config -init    # Initialize default config
$ taskmasterra config -show    # Show current config
```

---

## Features
- **Markdown-based workflow**: Use your favorite editor
- **Journaling & archiving**: Keep a history of what you did and when
- **macOS Reminders integration**: Sync active tasks to Reminders.app
- **Priority & effort**: A/B/C/D + Fibonacci estimation
- **Validation**: Catch formatting issues and get suggestions
- **Statistics**: Visualize your productivity

---

## Configuration

THIS SECTION IS A STUB, COMPLETE AND IS NOT CURRENTLY INTENDED FOR USER CONFIG
THIS WILL HIDDEN OR UPDATED IN A FUTURE RELEASE

Taskmasterra uses a JSON config at `~/.taskmasterra/config.json`. Initialize it with:
```bash
taskmasterra config -init
```

**Config options:**
- `default_due_hour`: Default hour for reminders (0-23)
- `default_due_minute`: Default minute for reminders (0-59)
- `reminder_list_name`: Reminders.app list name
- `journal_suffix`: Suffix for journal files
- `archive_suffix`: Suffix for archive files
- `active_marker`: Marker for active tasks (default: "!!")

---

## Development

```bash
make build         # Build the binary
make test          # Run all tests
make test-coverage # Generate coverage report
make fmt           # Format code
```

---

## Troubleshooting & FAQ

**Q: Why do I get validation warnings?**
- Your file may have formatting issues (e.g., misplaced `!!`, missing priorities). Run `taskmasterra validate -i todo.md` for details.

**Q: How do I archive completed tasks?**
- When you run `taskmasterra recordkeep -i todo.md`. Completed tasks are moved to the archive file with a timestamp.

**Q: Can I use this on Windows/Linux?**
- Most features work cross-platform, but Reminders integration is macOS-only.

**Q: How do I customize priorities or effort values?**
- Priorities are A/B/C/D. Effort is Fibonacci (1,2,3,5,8,13,21,34,55,89). These are not currently customizable.

---

## Versioning

Semantic versioning is used, this is automated in the makefile with 'release' commands.

---

For more, see the code and comments, or run `taskmasterra help`.
