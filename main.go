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
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type projectRow struct {
	name    string
	details config.ProjectDetails
}

type keyMap struct {
	Up   key.Binding
	Down key.Binding
	Quit key.Binding
	Edit key.Binding
	Star key.Binding
	Hide key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Edit, k.Star, k.Hide, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Up, k.Down, k.Edit, k.Star, k.Hide, k.Quit}}
}

var keys = keyMap{
	Up:   key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k/up", "move up")),
	Down: key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j/down", "move down")),
	Quit: key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q/ctrl+c", "quit")),
	Edit: key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "edit desc")),
	Star: key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "star")),
	Hide: key.NewBinding(key.WithKeys("H"), key.WithHelp("H", "hide")),
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
	textInput  textinput.Model
	editing    bool
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

	ti := textinput.New()
	ti.Placeholder = "Enter description..."
	ti.Focus()

	h := help.New()
	h.Styles.ShortKey = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("15"))
	h.Styles.FullKey = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("15"))

	return &model{
		config:     cfg,
		configPath: configPath,
		loading:    true,
		spinner:    s,
		help:       h,
		textInput:  ti,
	}, nil
}

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
		m.viewport = viewport.New(msg.Width, msg.Height-3)
		m.viewport.SetContent(m.formatProjects())
	case tea.KeyMsg:
		if m.editing {
			switch msg.String() {
			case "enter":
				p := &m.projects[m.cursor]
				p.details.Desc = m.textInput.Value()
				m.config.Projects[p.name] = p.details
				config.SaveConfig(m.configPath, m.config)
				m.editing = false
				m.textInput.SetValue("")
			case "esc":
				m.editing = false
				m.textInput.SetValue("")
			}
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}

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
		case key.Matches(msg, keys.Edit):
			m.editing = true
			m.textInput.SetValue(m.projects[m.cursor].details.Desc)
		case key.Matches(msg, keys.Star):
			p := &m.projects[m.cursor]
			p.details.Starred = !p.details.Starred
			m.config.Projects[p.name] = p.details
			config.SaveConfig(m.configPath, m.config)
			m.prepareProjectList()
			m.viewport.SetContent(m.formatProjects())
			m.ensureCursorVisible()
		case key.Matches(msg, keys.Hide):
			p := &m.projects[m.cursor]
			p.details.Show = false
			m.config.Projects[p.name] = p.details
			config.SaveConfig(m.configPath, m.config)
			m.prepareProjectList()
			if m.cursor >= len(m.projects) && m.cursor > 0 {
				m.cursor--
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
	if m.cursor < m.viewport.YOffset {
		m.viewport.YOffset = m.cursor
	} else if m.cursor >= m.viewport.YOffset+m.viewport.Height {
		m.viewport.YOffset = m.cursor - m.viewport.Height + 1
	}
}

func (m *model) formatProjects() string {
	var sb strings.Builder
	for i, p := range m.projects {
		star := "☆"
		if p.details.Starred {
			star = "★"
		}

		// Fixed-width name (up to 20 chars)
		name := p.name
		if len(name) > 20 {
			name = name[:17] + "..."
		}

		// Truncated description
		desc := p.details.Desc
		maxDescWidth := m.width - 28 // 20(name) + 2(star) + 3(sep) + 3(padding/buffer)
		if len(desc) > maxDescWidth && maxDescWidth > 3 {
			desc = desc[:maxDescWidth-3] + "..."
		}

		// Highlight style
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
		if i == m.cursor {
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("205"))
		}

		// Combine and render row
		line := fmt.Sprintf("%-2s %-20s - %s", star, name, desc)
		sb.WriteString(style.Width(m.width-2).Render(line) + "\n")
	}
	return sb.String()
}

func (m *model) View() string {
	if m.loading {
		return fmt.Sprintf("\033[H\n  %s Scanning...\n\n", m.spinner.View())
	}

	if m.editing {
		return fmt.Sprintf("\n  Edit description for %s:\n\n%s\n\n(press enter to save, esc to cancel)",
			m.projects[m.cursor].name, m.textInput.View())
	}

	listHeight := (m.height - 3) * 4 / 10
	metaHeight := (m.height - 3) - listHeight

	m.viewport.Height = listHeight
	m.viewport.SetContent(m.formatProjects())
	listView := m.viewport.View()

	metaView := m.formatMetadata(metaHeight)

	return fmt.Sprintf("\033[H%s\n%s\n%s", listView, metaView, m.help.View(keys))
}

func (m *model) formatMetadata(height int) string {
	if len(m.projects) == 0 {
		return ""
	}
	p := m.projects[m.cursor]
	d := p.details

	info := fmt.Sprintf("Last Commit: %s\nBuild/Test/Install: %s\nAgent Docs: %v\n\nLanguages:",
		d.LastCommitDate, d.BuildTestInstall, d.AgentDocs)
	for _, lang := range d.Languages {
		info += fmt.Sprintf("\n- %s", lang)
	}

	readme := d.ReadmePreview
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
	infoWidth := (m.width - 5) * 33 / 100
	readmeWidth := (m.width - 5) - infoWidth

	infoStyle := lipgloss.NewStyle().
		Width(infoWidth).
		BorderRight(true).
		BorderRightForeground(lipgloss.Color("240"))
	readmeStyle := lipgloss.NewStyle().Width(readmeWidth)

	return lipgloss.NewStyle().
		Width(m.width).
		Height(height).
		Border(lipgloss.NormalBorder()).
		Render(lipgloss.JoinHorizontal(lipgloss.Top, infoStyle.Render(info), readmeStyle.Render(readmeView)))
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
