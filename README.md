# Markban

This is a simple Markdown-based Kanban board, inspired by [KanBanLess](https://github.com/markdav-is/KanBanLess).


## Usage

To view the board in the terminal, use the `view` command, as shown below.
If no directory is specified, the app searches for the repository root (looking for a `.git` directory), and then looks for a directory with a `board.toml` file, or a directory with "board" in the name.

```shell
markban view project-board
```

To view the board in a web browser, use the `serve` verb, as shown below.

```shell
markban serve project-board
```


## Task File Format

Tasks are plain markdown files, with optional YAML frontmatter:

```markdown
---
priority: medium
tags: []
---

# Task Title

A 1-liner blurb for the task.

## Description

A longer, free-form description of the task.

## Subtasks

- [ ] Subtask 1
- [ ] Subtask 2
- [ ] Subtask 3
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

It is useful to use a `.gitkeep` file in each directory, such that if the column becomes empty, git will preserve the directory.


## Contributing

Some random notes on coding conventions:
* Idiomatic Go, passing lint and formatting checks
* All code shall have unix line-endings (LF, not CRLF), as enforced by `.editorconfig` and `.gitattributes`


### License

MIT License.

