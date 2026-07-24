package todo

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
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
	theme := ui.ThemeFromConfig(cfg)

	dirs, err := scanner.ResolveProjectDirs(cwd)
	if err != nil {
		return fmt.Errorf("failed to resolve project directories: %w", err)
	}

	if len(dirs) == 0 {
		fmt.Println("No projects discovered in index folders or current directory.")
		return nil
	}

	projects := make([]Project, len(dirs))
	for i, dir := range dirs {
		proj := Project{
			Name: filepath.Base(dir),
			Path: dir,
		}
		projects[i] = proj
	}

	items := buildListItems(projects, cfg.ShowTaskless, cfg)

	delegate := projectDelegate{theme: theme}

	listModel := ui.NewListModel(items, delegate, 80, 20, theme)
	listModel.Title = fmt.Sprintf("%s Projects (%d)", icon.Bookmarks.String(), len(items))
	listModel.SetShowStatusBar(false)
	listModel.SetFilteringEnabled(true)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.Primary)

	model := projectListModel{
		list:          listModel,
		theme:         theme,
		responsive:    ui.NewResponsiveManager(80),
		config:        cfg,
		cwd:           cwd,
		projects:      projects,
		scanner:       scanner,
		spinner:       s,
		indexing:      true,
		indexingIndex: 0,
	}

	model.list.AdditionalShortHelpKeys = model.getShortHelpKeys
	model.list.AdditionalFullHelpKeys = model.allHelpKeys

	opts := tty.GetProgramOptions(tea.WithoutSignalHandler())

	p := tea.NewProgram(model, opts...)
	resModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("failed to run interactive projects list: %w", err)
	}

	if fm, ok := resModel.(projectListModel); ok && fm.selectedCmd != "" {
		fmt.Println(fm.selectedCmd)
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

	// Styles
	titleStyle := lipgloss.NewStyle().Bold(true)
	var barStr string
	statsStyle := lipgloss.NewStyle().Foreground(d.theme.ColorTasks)
	todoStyle := lipgloss.NewStyle().Foreground(d.theme.Error)
	pathStyle := lipgloss.NewStyle().Foreground(d.theme.Muted)

	if isSelected {
		titleStyle = titleStyle.Foreground(d.theme.ColorHighlightName)

		fgStyle := lipgloss.NewStyle().Foreground(d.theme.ColorHighlightProgressFg)
		bgStyle := lipgloss.NewStyle().Foreground(d.theme.ColorHighlightProgressBg)
		filledPart := fgStyle.Render(strings.Repeat("█", filled))
		emptyPart := bgStyle.Render(strings.Repeat("░", barWidth-filled))
		barStr = filledPart + emptyPart
	} else {
		titleStyle = titleStyle.Foreground(d.theme.ColorName)

		fgStyle := lipgloss.NewStyle().Foreground(d.theme.ColorProgressFg)
		bgStyle := lipgloss.NewStyle().Foreground(d.theme.ColorProgressBg)
		filledPart := fgStyle.Render(strings.Repeat("█", filled))
		emptyPart := bgStyle.Render(strings.Repeat("░", barWidth-filled))
		barStr = filledPart + emptyPart
	}

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
			BorderForeground(d.theme.ColorHighlightBorder).
			Padding(0, 1).
			Width(targetTotalWidth).
			Render(content)
	} else {
		rendered = lipgloss.NewStyle().
			Border(lipgloss.Border{}, false, false, false, false).
			Padding(1, 2).
			Width(targetTotalWidth).
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

type scanProjectResultMsg struct {
	index int
	proj  Project
	err   error
}

func scanProjectCmd(scanner *Scanner, index int, dir string) tea.Cmd {
	return func() tea.Msg {
		proj, err := scanner.ScanProject(dir)
		return scanProjectResultMsg{index: index, proj: proj, err: err}
	}
}

type clearIndexingStatusMsg struct{}

func delayClearIndexingCmd() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return clearIndexingStatusMsg{}
	})
}

type projectListModel struct {
	list        list.Model
	theme       ui.Theme
	responsive  *ui.ResponsiveManager
	config      domain.Config
	cwd         string
	message     string
	selectedCmd string
	projects    []Project
	scanner     *Scanner
	spinner     spinner.Model
	indexing    bool
	indexingIndex int
	indexingComplete bool
	screenW     int
	screenH     int
}

func (m projectListModel) Init() tea.Cmd {
	if len(m.projects) > 0 {
		return tea.Batch(
			m.spinner.Tick,
			scanProjectCmd(m.scanner, 0, m.projects[0].Path),
		)
	}
	return nil
}

func (m projectListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.screenW = msg.Width
		m.screenH = msg.Height
		m.responsive.SetWidth(msg.Width)
		width, height := m.responsive.GetListDimensions(msg.Width, msg.Height)
		if m.indexing || m.indexingComplete {
			height -= 2
		}
		m.list.SetSize(width, height)
		return m, nil

	case scanProjectResultMsg:
		if msg.err == nil {
			m.projects[msg.index] = msg.proj

			items := buildListItems(m.projects, m.config.ShowTaskless, m.config)
			m.list.SetItems(items)
			m.list.Title = fmt.Sprintf("%s Projects (%d)", icon.Bookmarks.String(), len(items))
		}

		nextIndex := msg.index + 1
		if nextIndex < len(m.projects) {
			m.indexingIndex = nextIndex
			return m, scanProjectCmd(m.scanner, nextIndex, m.projects[nextIndex].Path)
		} else {
			m.indexing = false
			m.indexingComplete = true
			return m, delayClearIndexingCmd()
		}

	case clearIndexingStatusMsg:
		m.indexingComplete = false
		width, height := m.responsive.GetListDimensions(m.screenW, m.screenH)
		m.list.SetSize(width, height)
		return m, nil

	case spinner.TickMsg:
		if m.indexing {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
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
				relPath := "."
				targetPath := item.Project.Path
				for _, ftsFile := range m.config.FtsFiles {
					testPath := filepath.Join(item.Project.Path, ftsFile)
					if _, err := os.Stat(testPath); err == nil {
						relPath = ftsFile
						targetPath = testPath
						break
					}
				}

				if m.config.ReturnWithCd {
					editorName := m.config.Editor
					if editorName == "" {
						editorName = "nvim"
					}

					escapedProjPath := quoteShellArg(item.Project.Path)
					escapedRelPath := quoteShellArg(relPath)
					m.selectedCmd = fmt.Sprintf("cd %s && %s %s", escapedProjPath, editorName, escapedRelPath)
					return m, tea.Quit
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

	if m.indexing {
		currentProjName := m.projects[m.indexingIndex].Name
		statusStr := fmt.Sprintf("\n%s indexing tasks: %s (%d/%d)...",
			m.spinner.View(),
			currentProjName,
			m.indexingIndex+1,
			len(m.projects))
		body += lipgloss.NewStyle().Foreground(m.theme.Muted).Render(statusStr)
	} else if m.indexingComplete {
		statusStr := fmt.Sprintf("\n%s indexing tasks complete",
			lipgloss.NewStyle().Foreground(m.theme.Success).Render("✔"))
		body += lipgloss.NewStyle().Foreground(m.theme.Muted).Render(statusStr)
	}

	return m.responsive.FullWidthFrameStyle(m.theme).Render(body)
}

func buildListItems(projects []Project, showTaskless bool, config domain.Config) []list.Item {
	var items []list.Item
	for _, p := range projects {
		if !showTaskless && p.Scanned && p.TotalTasks == 0 {
			continue
		}
		items = append(items, projectItem{
			Project: p,
			Config:  config,
		})
	}
	return items
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

func quoteShellArg(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
