---
color: yellow
priority: medium
---

# Rendering Improvements

A few things to improve the usability of the tool and make it prettier.

# Subtasks

- [X] Add a board.toml file that has a project name that is displayed across the top of the page
- [X] If the project directory is not specified, look for the git root and try to find a subdirectory with "board" in the name that has a `board.toml` file in it
- [X] If the directory name contains a hyphen, like `in-progress`, it should be removed, so the column displays as "IN PROGRESS"
- [X] When there are multiple cards in a column, sort by priority (high on top, low on bottom) and then alphabetical by task title
- [X] If the YAML front matter has a color, use that as the background for the card. There should be a small list of acceptable colors (light pastels: yellow, green, blue, red, orange, purple, magenta and cyan)
- [ ] When rendering the description, use markdown formatting (backticks for preformatted text, bold, italic, etc)
- [ ] If the YAML front matter has a priority, display "high", "medium", or "low" in the upper-right of the card in small, all-caps text
- [ ] If live reload connectivity is lost by the browser, display a message in the upper right of the page so the user is aware

