# Markban

This is a simple Markdown-based Kanban board, inspired by [KanBanLess](https://github.com/markdav-is/KanBanLess).


## Usage

Point the app at a directory with the proper structure, and it will open a web browser with a rendering of the board:

```shell
markban project-board
```


## Task File Format

Tasks are plain markdown files, with optional YAML frontmatter:

```markdown
---
priority: medium
tags: []
---

# Task Title

Brief description of the task.

## Checklist

- [ ] Step one
- [ ] Step two
- [ ] Step three
```

The only required bit is the task title.
Everything else is optional.

## Directory Structure

Markdown files are grouped into folders, with each folder representing a column in the Kanban board.
The names of the folders are up to the user, and may optionally be prefixed with one or more digits and a hyphen (`1-`) to order the columns.
If numbers are omitted, the app makes its best guess based on convention ("backlog" first, "done" last, etc), resorting to alphabetic as a last resort.
When the column names are rendered, any numeric prefixes and hyphens are removed.

```
my-board/
  1-backlog/
    task-slug.md
  2-todo/
    another-task.md
  3-in-progress/
    in-progress-task.md
  4-done/
    completed-task.md
```

