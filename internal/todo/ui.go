package todo

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/super-todo/internal/adapters/editor"
	"github.com/super-todo/internal/adapters/icon"
	"github.com/super-todo/internal/adapters/tty"
	"github.com/super-todo/internal/domain"
	"github.com/super-todo/internal/ui"
)

// RunProjectListing launches the interactive TUI for listing projects.
func RunProjectListing(cfg domain.Config, cwd string) error {
	scanner := NewScanner(cfg)
	projects, err := scanner.ScanProjects(cwd)
	if err != nil {
		return fmt.Errorf("failed to scan projects: %w", err)
	}

	if len(projects) == 0 {
		fmt.Println("No projects discovered in index folders or current directory.")
		return nil
	}

	items := make([]list.Item, len(projects))
	for i, proj := range projects {
		items[i] = projectItem{
			Project: proj,
			Config:  cfg,
		}
	}

	theme := ui.ThemeFromConfig(cfg)
	delegate := projectDelegate{theme: theme}

	listModel := ui.NewListModel(items, delegate, 80, 20, theme)
	listModel.Title = fmt.Sprintf("%s Projects (%d)", icon.Bookmarks.String(), len(items))
	listModel.SetShowStatusBar(false)
	listModel.SetFilteringEnabled(true)

	model := projectListModel{
		list:       listModel,
		theme:      theme,
		responsive: ui.NewResponsiveManager(80),
		config:     cfg,
		cwd:        cwd,
	}

	model.list.AdditionalShortHelpKeys = model.getShortHelpKeys
	model.list.AdditionalFullHelpKeys = model.allHelpKeys

	opts := tty.GetProgramOptions(tea.WithoutSignalHandler())

	p := tea.NewProgram(model, opts...)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run interactive projects list: %w", err)
	}

	return nil
}

type projectItem struct {
	Project Project
	Config  domain.Config
}

func (p projectItem) Title() string {
	return p.Project.Name
}

func (p projectItem) Description() string {
	return p.Project.Path
}

func (p projectItem) FilterValue() string {
	return p.Project.Name + " " + p.Project.Path
}

type projectDelegate struct {
	theme ui.Theme
}

func (d projectDelegate) Height() int {
	return 5
}

func (d projectDelegate) Spacing() int {
	return 0
}

func (d projectDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d projectDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	proj, ok := item.(projectItem)
	if !ok {
		return
	}

	isSelected := index == m.Index()

	// Calculate target total width for the outer box (including borders/padding)
	// list width - 8 (leaves safety columns to prevent wrapping/overflow)
	targetTotalWidth := m.Width() - 8
	if targetTotalWidth < 20 {
		targetTotalWidth = 20
	}

	// For Selected: rounded border adds 2, padding(0, 1) adds 2. Total extra = 4.
	// For Unselected: padding(1, 2) adds 4. Total extra = 4.
	// Therefore, content width inside borders/padding is targetTotalWidth - 4.
	contentWidth := targetTotalWidth - 4
	if contentWidth < 16 {
		contentWidth = 16
	}

	title := truncateString(proj.Project.Name, contentWidth)
	completed := proj.Project.CompletedTasks
	total := proj.Project.TotalTasks
	todoCount := proj.Project.TodoCount

	// Render progress bar (full width of the content area)
	barWidth := contentWidth
	filled := 0
	if total > 0 {
		filled = (completed * barWidth) / total
	}

	bar := ""
	for i := 0; i < barWidth; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}

	// Styles
	titleStyle := lipgloss.NewStyle().Foreground(d.theme.Text).Bold(true)
	barStyle := lipgloss.NewStyle().Foreground(d.theme.Primary)
	statsStyle := lipgloss.NewStyle().Foreground(d.theme.Secondary)
	todoStyle := lipgloss.NewStyle().Foreground(d.theme.Error)
	pathStyle := lipgloss.NewStyle().Foreground(d.theme.Muted)

	if isSelected {
		titleStyle = titleStyle.Foreground(d.theme.TextHighlight)
		barStyle = barStyle.Foreground(d.theme.TextHighlight)
	}

	barStr := barStyle.Render(bar)
	statsStr := statsStyle.Render(fmt.Sprintf("(%d/%d) tasks complete", completed, total))
	todoStr := todoStyle.Render(fmt.Sprintf("%d todos", todoCount))

	// Truncate path based on remaining space
	// stats: ~22 chars, todos: ~10 chars, spacers: ~6 chars
	nonPathLen := 22 + 10 + 6
	pathMaxLen := contentWidth - nonPathLen

	var metadataLine string
	if pathMaxLen >= 10 {
		truncatedPath := truncatePath(proj.Project.Path, pathMaxLen)
		metadataLine = fmt.Sprintf("%s  %s  %s", statsStr, todoStr, pathStyle.Render(truncatedPath))
	} else if contentWidth-34 >= 10 {
		truncatedPath := truncatePath(proj.Project.Path, contentWidth-26)
		metadataLine = fmt.Sprintf("%s  %s", statsStr, pathStyle.Render(truncatedPath))
	} else {
		metadataLine = statsStr
	}

	content := titleStyle.Render(title) + "\n" + barStr + "\n" + metadataLine

	var rendered string
	if isSelected {
		rendered = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(d.theme.Primary).
			Padding(0, 1).
			Width(contentWidth).
			Render(content)
	} else {
		rendered = lipgloss.NewStyle().
			Border(lipgloss.Border{}, false, false, false, false).
			Padding(1, 2).
			Width(contentWidth).
			Render(content)
	}

	fmt.Fprintf(w, "%s", rendered)
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return "..."
	}
	return s[:maxLen-3] + "..."
}

func truncatePath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	if maxLen <= 3 {
		return "..."
	}
	return "..." + path[len(path)-(maxLen-3):]
}

type projectListModel struct {
	list       list.Model
	theme      ui.Theme
	responsive *ui.ResponsiveManager
	config     domain.Config
	cwd        string
	message    string
}

func (m projectListModel) Init() tea.Cmd {
	return nil
}

func (m projectListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.responsive.SetWidth(msg.Width)
		width, height := m.responsive.GetListDimensions(msg.Width, msg.Height)
		m.list.SetSize(width, height)
		return m, nil

	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			switch msg.String() {
			case "q", "esc", "ctrl+c":
				return m, tea.Quit
			}
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit

		case "enter", "o":
			if item, ok := m.list.SelectedItem().(projectItem); ok {
				// Find FTS file or fallback to project directory
				targetPath := item.Project.Path
				for _, ftsFile := range m.config.FtsFiles {
					testPath := filepath.Join(item.Project.Path, ftsFile)
					if _, err := os.Stat(testPath); err == nil {
						targetPath = testPath
						break
					}
				}

				editorAdapter := editor.New(m.config.Editor)
				c, err := editorAdapter.BuildCmd(targetPath, 1)
				if err != nil {
					m.message = fmt.Sprintf("✗ %s", err)
					return m, nil
				}

				return m, tea.ExecProcess(c, func(err error) tea.Msg {
					if err != nil {
						return tea.Quit
					}
					return nil
				})
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m projectListModel) View() string {
	var body string
	if m.message != "" {
		body = lipgloss.NewStyle().Foreground(m.theme.Error).Render(m.message) + "\n\n"
	}
	body += m.list.View()

	return m.responsive.FullWidthFrameStyle(m.theme).Render(body)
}

func (m projectListModel) allHelpKeys() []key.Binding {
	return []key.Binding{
		key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter/o", "open project in editor"),
		),
	}
}

func (m projectListModel) getShortHelpKeys() []key.Binding {
	return m.allHelpKeys()
}
