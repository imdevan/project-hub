package ui

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/super-todo/internal/adapters/tty"
	"github.com/super-todo/internal/index"
)

type indexResultMsg struct {
	res index.IndexResult
}

type indexProgressModel struct {
	spinner spinner.Model
	indexer *index.Indexer
	theme   Theme
	done    bool
	result  index.IndexResult
	err     error
	start   time.Time
	elapsed time.Duration
}

func (m indexProgressModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			res := m.indexer.Refresh()
			return indexResultMsg{res: res}
		},
	)
}

func (m indexProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case indexResultMsg:
		m.done = true
		m.result = msg.res
		m.elapsed = time.Since(m.start)
		return m, tea.Quit

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.err = fmt.Errorf("indexing interrupted")
			return m, tea.Quit
		}

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m indexProgressModel) View() string {
	if m.done {
		if m.err != nil {
			errStr := fmt.Sprintf("Indexing failed: %v", m.err)
			return lipgloss.NewStyle().Foreground(m.theme.Error).Render("✗ " + errStr)
		}

		successMsg := fmt.Sprintf("Indexed %d files (skipped %d unchanged) in %v",
			m.result.FilesScanned,
			m.result.FilesSkipped,
			m.result.Duration.Round(time.Millisecond))

		return fmt.Sprintf("%s %s",
			lipgloss.NewStyle().Foreground(m.theme.Success).Render("✔"),
			lipgloss.NewStyle().Foreground(m.theme.Text).Render(successMsg))
	}

	// Calculate current elapsed time
	elapsed := time.Since(m.start).Round(100 * time.Millisecond)

	textStyle := lipgloss.NewStyle().Foreground(m.theme.Text)
	elapsedStyle := lipgloss.NewStyle().Foreground(m.theme.Muted)

	return fmt.Sprintf("%s %s %s",
		m.spinner.View(),
		textStyle.Render("indexing tasks"),
		elapsedStyle.Render(fmt.Sprintf("(%s)", elapsed)))
}

// RefreshWithSpinner runs the indexing operation with a visual progress spinner.
func RefreshWithSpinner(indexer *index.Indexer, theme Theme) (index.IndexResult, error) {
	if !tty.IsTerminal(os.Stdout.Fd()) {
		return indexer.Refresh(), nil
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.Primary)

	m := indexProgressModel{
		spinner: s,
		indexer: indexer,
		theme:   theme,
		start:   time.Now(),
	}

	opts := tty.GetProgramOptions()
	p := tea.NewProgram(m, opts...)
	resModel, err := p.Run()
	if err != nil {
		return index.IndexResult{}, err
	}

	finalModel, ok := resModel.(indexProgressModel)
	if !ok {
		return index.IndexResult{}, fmt.Errorf("unexpected model type returned from program")
	}

	if finalModel.err != nil {
		return index.IndexResult{}, finalModel.err
	}

	return finalModel.result, nil
}
