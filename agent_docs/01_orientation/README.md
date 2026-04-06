# Orientation: Projector

Projector is a TUI-based coding project manager. It scans a configured directory, detecting project metadata (git status, languages, READMEs), and provides a navigable list.

## Core Goals
- Maintain a local, searchable list of coding projects.
- Automatically detect project attributes (Git, Language, READMEs, etc.).
- Allow manual overrides (description, category, starred status, hidden status).
- Provide a fast, TUI-driven experience.

## Tech Stack
- **Language**: Go
- **TUI Framework**: [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- **Styling**: [Lipgloss](https://github.com/charmbracelet/lipgloss)
- **Terminal UI Components**: [Bubbles](https://github.com/charmbracelet/bubbles)
- **Config**: YAML (`~/.projector/config.yml`)
