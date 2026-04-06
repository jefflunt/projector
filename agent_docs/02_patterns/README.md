# Patterns: Projector

## Projector UI Pattern (Bubble Tea)
Projector follows the Model-View-Update pattern.
1. **Model**: Stores project data, configuration, viewport states (project list, metadata pane), text input state, and navigation cursor.
2. **View**: Uses `lipgloss` to manage a 40/60 vertical split layout (top: project list, bottom: metadata). Uses `viewport` for scrolling project list.
3. **Update**: Handles keyboard shortcuts, input state, and project scanning asynchronously via `tea.Cmd`.

## Config Pattern
Configuration follows a flat structure:
- `code_folder`: Root path to scan.
- `new_tab_cmd`: Command string for terminal tab spawning.
- `projects`: Map where keys are project folder names, values are `ProjectDetails` objects.

## Metadata Pattern
- **Scanning**: Automatic scan on startup.
- **Persistence**: Any manual UI updates to project status (`star`, `category`, `desc`, `show`) persist to `~/.projector/config.yml` immediately.
- **Preservation**: The `scanner` merges new findings with existing metadata (desc, category, starred status) to ensure manual overrides are not lost.
