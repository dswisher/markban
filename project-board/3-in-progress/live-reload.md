---
color: blue
priority: high
---

# Live Reload

The `render` verb should run a small web server that watches for changes in the directory and does a live-reload.


# Description

The rendering should be done to a hidden subdirectory in the project directory, for this board, it would be something like `project-board/.build`.
I'd eventually like to support images and whatnot, and having an entire directory would give us a place to put them.
