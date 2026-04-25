---
color: blue
priority: high
done: 2026-04-17
---

# Live Reload

The `render` verb should run a small web server that watches for changes in the directory and does a live-reload.


# Description

The rendering should be done to a hidden subdirectory in the project directory, for this board, it would be something like `project-board/.build`.
I'd eventually like to support images and whatnot, and having an entire directory would give us a place to put them.


# Implementation Notes

- MacOS keep prompting to "Allow Connections". This has been fixed for `make render` by building a binary and running that, `make test` fixed by removing usage of "run"
- The way to shut down the server is via `ctrl+c`, which took about 10s. This has been fixed by passing contexts down to where they are needed.

