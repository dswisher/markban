---
color: green
priority: medium
---

# Sort Done Column by Completed Date

Add a `done` date to task YAML frontmatter, and use it to sort the "Done" column.

# Subtasks

- [X] Parse done column and use it for sorting
- [ ] In text mode, display done date instead of priority, if present
- [ ] If done is set on a card that isn't in the done column, add a warning
- [ ] If there are warnings, display a message on the last line of the board (otherwise it scrolls off the top)
- [ ] In web view, display done date
- [ ] In text mode, consider hiding done items more than 10 days old, with option to override time period or even disable
