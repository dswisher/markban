---
color: yellow
priority: low
---

# Rendering Improvements

A few things to improve the usability of the tool.

# Subtasks

- [ ] If the directory name contains a hyphen, like `in-progress`, it should be removed, so the column displays as "IN PROGRESS"
- [ ] Rather than rendering to a temp file with a random name, render to a special directory, like `.build` within the board directory with a consistent name; the special directory should be ignored when scanning for columns
- [ ] Add a flag to the `render` verb that does not open a new browser window, something like `--no-open`; that way (with the fixed file name from above), the file can be rebuilt and the user can just refresh their already-open page
- [ ] When rendering the description, use markdown formatting (backticks for preformatted text, bold, italic, etc)
- [ ] If the YAML front matter has a priority, display "high", "medium", or "low" in the upper-right of the card in small, all-caps text
- [ ] If the YAML front matter has a color, use that as the background for the card. There should be a small list of acceptable colors (light pastels: yellow, green, blue, red, orange, purple, magenta and cyan)

