# TaskMasterRa

Manage my tasks with `recordkeep` for journaling and archiving activities.
Manage my Reminders with (mis-named) `updatecal` which creates/updates active([.:]) todo items in Reminders.app

This replaces the existing notesutil written in Rust.

`recordkeep` 

    completed(- [x]) will be removed after being added to the archive.
    Touched(:) will be added to the journal.
    
`updatecal`
    
    Active(.) and Touched(:) will be added to reminders.

markdown file, task examples:

- [w] . A1 task is active (.) and has some work done ([w]), A priority and fibonacci est effort 1, active or touched so added to reminders due today
- [x] : B1 task that is touched (:) today, and is completed ([x]) active or touched so added to reminders due today
- [b] B2 blocked task that is not active or touched today. will not be added to reminders\

Taskra (seperate nvim extension) will highlight the task priority with fibonacci entry
