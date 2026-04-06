package main

import (
	"fmt"
	"os"
	"projector/config"
	"projector/scanner"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *model) prepareProjectList() {
	var starred, unstarred []projectRow
	for name, p := range m.config.Projects {
		if !p.Show {
			continue
		}
		row := projectRow{name, p}
		if p.Starred {
			starred = append(starred, row)
		} else {
			unstarred = append(unstarred, row)
		}
	}

	sortFunc := func(i, j int, list []projectRow) bool {
		return strings.ToLower(list[i].name) < strings.ToLower(list[j].name)
	}

	sort.Slice(starred, func(i, j int) bool { return sortFunc(i, j, starred) })
	sort.Slice(unstarred, func(i, j int) bool { return sortFunc(i, j, unstarred) })

	m.projects = append(starred, unstarred...)
}

type projectRow struct {
	name    string
	details config.ProjectDetails
}

type keyMap struct {
	Up   key.Binding
	Down key.Binding
	Quit key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Up, k.Down, k.Quit}}
}

var keys = keyMap{
	Up:   key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k/up", "move up")),
	Down: key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j/down", "move down")),
	Quit: key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q/ctrl+c", "quit")),
}

type model struct {
	config     *config.Config
	configPath string
	projects   []projectRow
	cursor     int
	loading    bool
	spinner    spinner.Model
	viewport   viewport.Model
	help       help.Model
	width      int
	height     int
}

func initialModel() (*model, error) {
	configPath, err := config.EnsureConfigExists()
	if err != nil {
		return nil, err
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	if cfg.CodeFolder == "" {
		fmt.Println("Please set 'code_folder' in ~/.projector/config.yml")
		os.Exit(1)
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return &model{
		config:     cfg,
		configPath: configPath,
		loading:    true,
		spinner:    s,
		help:       help.New(),
	}, nil
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			projects, err := scanner.ScanProjects(m.config.CodeFolder)
			if err != nil {
				return err
			}
			return projects
		},
	)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		m.viewport = viewport.New(msg.Width, msg.Height-3) // Leave space for help bar
		m.viewport.SetContent(m.formatProjects())
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Down):
			if len(m.projects) > 0 {
				m.cursor = (m.cursor + 1) % len(m.projects)
			}
			m.viewport.SetContent(m.formatProjects())
			m.ensureCursorVisible()
		case key.Matches(msg, keys.Up):
			if len(m.projects) > 0 {
				m.cursor = (m.cursor - 1 + len(m.projects)) % len(m.projects)
			}
			m.viewport.SetContent(m.formatProjects())
			m.ensureCursorVisible()
		}
	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case map[string]config.ProjectDetails:
		m.loading = false
		for name, p := range msg {
			if _, ok := m.config.Projects[name]; !ok {
				p.Show = true
				m.config.Projects[name] = p
			} else {
				existing := m.config.Projects[name]
				p.Desc = existing.Desc
				m.config.Projects[name] = p
			}
		}
		config.SaveConfig(m.configPath, m.config)
		m.prepareProjectList()
		m.viewport.SetContent(m.formatProjects())
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, nil
}

func (m *model) ensureCursorVisible() {
	// If cursor is before start of view, jump to it.
	if m.cursor < m.viewport.YOffset {
		m.viewport.YOffset = m.cursor
	} else if m.cursor >= m.viewport.YOffset+m.viewport.Height {
		// If cursor is after end of view, jump to it.
		// Height is lines visible.
		m.viewport.YOffset = m.cursor - m.viewport.Height + 1
	}
}

func (m *model) formatProjects() string {
	var sb strings.Builder
	for i, p := range m.projects {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}

		line := fmt.Sprintf("%s%s - %s", cursor, p.name, p.details.Desc)
		if i == m.cursor {
			line = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render(line)
		}
		sb.WriteString(line + "\n")
	}
	return sb.String()
}

func (m *model) View() string {
	if m.loading {
		return fmt.Sprintf("\033[H\n  %s Scanning...\n\n", m.spinner.View())
	}

	// Calculate 40/60 split
	listHeight := (m.height - 3) * 4 / 10
	metaHeight := (m.height - 3) - listHeight

	// Top pane: list
	m.viewport.Height = listHeight
	m.viewport.SetContent(m.formatProjects())
	listView := m.viewport.View()

	// Bottom pane: metadata
	metaView := m.formatMetadata(metaHeight)

	// Combine
	return fmt.Sprintf("\033[H%s\n%s\n%s", listView, metaView, m.help.View(keys))
}

func (m *model) formatMetadata(height int) string {
	if len(m.projects) == 0 {
		return ""
	}
	p := m.projects[m.cursor]
	d := p.details

	// Info section (left column)
	info := fmt.Sprintf("Last Commit: %s\nLanguages: %s\nBuild/Test/Install: %s\nAgent Docs: %v",
		d.LastCommitDate, strings.Join(d.Languages, ", "), d.BuildTestInstall, d.AgentDocs)

	// README preview (right column)
	readme := d.ReadmePreview
	// Truncate based on height and width
	lines := strings.Split(readme, "\n")
	if len(lines) > height {
		lines = lines[:height]
	}
	for i, line := range lines {
		if len(line) > m.width/2-2 {
			lines[i] = line[:m.width/2-5] + "..."
		}
	}
	readmeView := strings.Join(lines, "\n")

	// Apply styles for horizontal layout
	colWidth := m.width/2 - 2
	infoStyle := lipgloss.NewStyle().Width(colWidth)
	readmeStyle := lipgloss.NewStyle().Width(colWidth)
	// Just use a pipe character as a separator
	spacer := lipgloss.NewStyle().Width(1).Align(lipgloss.Center).Render("|")

	return lipgloss.NewStyle().
		Width(m.width).
		Height(height).
		Border(lipgloss.NormalBorder()).
		Render(lipgloss.JoinHorizontal(lipgloss.Top, infoStyle.Render(info), spacer, readmeStyle.Render(readmeView)))
}

func main() {
	m, err := initialModel()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
