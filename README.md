# TaskMasterRa

Manage my tasks with `recordkeep` for journaling and archiving activities.

Manage my Reminders with `updatecal` which creates/updates active([.:]) todo items in MacOS Reminders.app
    
`updatecal is` "slightly mis-named" because Reminders.app is obviously not Calendar.app, but dated reminders DO appear in Calendar.app, and the goal is to have tasks that can be marked completed, -AND- show in my calendar.


This replaces the existing notesutil written in Rust.

`taskmasterra recordkeep -i <markdownfilepath>` 

    completed(- [x]) will be removed after being added to the archive.
    Touched(:) will be added to the journal.
    
`taskmasterra updatecal -i <markdownfilepath>` 
    
    Active(.) and Touched(:) will be added to reminders.

markdown file, task examples:

```markdown
- [w] . A1 task is active (.) and has some work done ([w]), A priority and fibonacci est effort 1, active or touched so added to reminders due today
- [x] : B1 task that is touched (:) today, and is completed ([x]) active or touched so added to reminders due today
- [b] B2 blocked task that is not active or touched today. will not be added to reminders\
```

Taskra (seperate nvim extension) will highlight the task priority with fibonacci entry
