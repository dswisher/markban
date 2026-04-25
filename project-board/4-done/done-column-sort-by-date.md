---
color: green
priority: medium
done: 2026-04-25
---

# Sort Done Column by Completed Date

Add a `done` date to task YAML frontmatter, and use it to sort the "Done" column.

# Subtasks

- [X] Parse done column and use it for sorting
- [X] In text mode, display done date instead of priority, if present
- [X] If there are warnings, display a message on the last line of the board (otherwise it scrolls off the top)
- [X] In web view, display done date
- [X] In text mode, consider hiding done items more than 10 days old, with option to override time period or even disable
    - Did not implement - just need to be more diligent about moving items to archive
