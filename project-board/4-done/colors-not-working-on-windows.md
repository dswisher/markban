---
color: blue
priority: high
done: 2026-04-17
---

# Text Colors Not Working on Windows

When I use the `markban view` command on windows, the escape sequences to change colors are not working.

# Notes

In both MacOS and Windows, I run `wezterm`.
On MacOS, the text mode colors look great.
On Windows, I see the escape sequences (`←[94`, etc).
They work fine in "git bash", just not in wezterm.

If I run `markban view --no-colors`, I see the bold/unbold escape sequences.
If `--no-colors` is specified, we should also disable bolding.

We need to figure out if there is a way to get the escape sequences in Wezterm working.
Maybe they need something slightly different, or my config is wrong.
My config can be found [here](https://github.com/dswisher/dotfiles/blob/main/dot_config/wezterm/wezterm.lua), or in `~/.local/share/chezmoi/dot_config/wezterm/wezterm.lua`.

