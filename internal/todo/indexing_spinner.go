package todo

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/super-todo/internal/adapters/tty"
	"github.com/super-todo/internal/ui"
)

type scanResultMsg struct {
	projects []Project
	err      error
}

type projectScanModel struct {
	spinner  spinner.Model
	scanner  *Scanner
	cwd      string
	theme    ui.Theme
	done     bool
	projects []Project
	err      error
	start    time.Time
}

func (m projectScanModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			projs, err := m.scanner.ScanProjects(m.cwd)
			return scanResultMsg{projects: projs, err: err}
		},
	)
}

func (m projectScanModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case scanResultMsg:
		m.done = true
		m.projects = msg.projects
		m.err = msg.err
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.err = fmt.Errorf("interrupted")
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m projectScanModel) View() string {
	if m.done {
		return ""
	}
	elapsed := time.Since(m.start).Round(100 * time.Millisecond)
	textStyle := lipgloss.NewStyle().Foreground(m.theme.Text)
	elapsedStyle := lipgloss.NewStyle().Foreground(m.theme.Muted)
	return fmt.Sprintf("%s %s %s",
		m.spinner.View(),
		textStyle.Render("indexing tasks"),
		elapsedStyle.Render(fmt.Sprintf("(%s)", elapsed)))
}

func scanProjectsWithSpinner(scanner *Scanner, cwd string, theme ui.Theme) ([]Project, error) {
	if !tty.IsTerminal(os.Stdout.Fd()) {
		return scanner.ScanProjects(cwd)
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.Primary)

	m := projectScanModel{
		spinner: s,
		scanner: scanner,
		cwd:     cwd,
		theme:   theme,
		start:   time.Now(),
	}

	opts := tty.GetProgramOptions()
	p := tea.NewProgram(m, opts...)
	resModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	finalModel, ok := resModel.(projectScanModel)
	if !ok {
		return nil, fmt.Errorf("unexpected model type")
	}

	if finalModel.err != nil {
		return nil, finalModel.err
	}

	return finalModel.projects, nil
}
