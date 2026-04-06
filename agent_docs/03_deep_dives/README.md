# Deep Dives: Projector

## Scanner Logic
The scanner performs deep introspection of each project folder:
- **Git**: Checks `.git`, retrieves last commit date via `git -C ... log -1 --format=%cd --date=short`.
- **README**: Searches `README.md`, `README`, `README.txt`, captures first 10 lines as a preview.
- **Languages**: Deep walks the directory recursively to find files matching defined extensions (Ruby, HTML, Go, C/C++, Shell, JS/TS, CSS/Sass, Java, Python, Kotlin, Swift).
- **Metadata**: Stores star/category/hidden status.

## TUI Rendering
- **Main View**: Uses `tea.WithAltScreen` for full console capture.
- **Top Pane**: A scrollable `viewport` displaying the table of projects (`star`, `name`, `category`, `description`). Uses `lipgloss` for full-width row highlighting.
- **Bottom Pane**: A 33/67 horizontal split (`lipgloss.JoinHorizontal`) showing project info and README preview.
- **Interaction**: Navigation wrapping via modulo arithmetic (`m.cursor = (m.cursor + 1) % len(m.projects)`), and scrolling logic via `ensureCursorVisible()` which manages the `viewport.YOffset`.
