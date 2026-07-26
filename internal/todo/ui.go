package todo

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"sort"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/super-todo/internal/adapters/editor"
	"github.com/super-todo/internal/adapters/icon"
	"github.com/super-todo/internal/adapters/tty"
	"github.com/super-todo/internal/domain"
	"github.com/super-todo/internal/index"
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

	store, err := index.NewStore()
	var pinnedMap map[string]int
	if err == nil {
		defer store.Close()
		pinnedMap, _ = store.GetPinnedProjects()
	}

	// Check cache: if sync_interval > 0 and cache is fresh, load from cache
	var projects []Project
	usedCache := false
	if store != nil && cfg.SyncInterval > 0 {
		cached, scannedAt, cacheErr := store.LoadProjectsCache()
		if cacheErr == nil && len(cached) > 0 && time.Since(scannedAt) < time.Duration(cfg.SyncInterval)*time.Second {
			// Restore pinned state from DB (cache may be stale on pins)
			for i := range cached {
				order, pinned := pinnedMap[cached[i].Path]
				cached[i].Pinned = pinned
				cached[i].PinOrder = order
			}
			projects = cached
			usedCache = true
		}
	}

	recurringProjects := scanner.ScanRecurringProjects()

	if !usedCache {
		projects = make([]Project, len(dirs))
		for i, dir := range dirs {
			order, pinned := pinnedMap[dir]
			proj := Project{
				Name:     filepath.Base(dir),
				Path:     dir,
				Pinned:   pinned,
				PinOrder: order,
			}
			projects[i] = proj
		}
	}

	// Filter out any cached recurring projects to ensure fresh state, then prepend recurring projects
	var regularProjects []Project
	for _, p := range projects {
		if !p.IsRecurring {
			regularProjects = append(regularProjects, p)
		}
	}
	projects = append(recurringProjects, regularProjects...)

	sortProjects(projects)

	items := buildListItems(projects, cfg.ShowTaskless, cfg)

	barAnim := 0.0
	delegate := projectDelegate{theme: theme, barAnimProgress: &barAnim}

	listModel := ui.NewListModel(items, delegate, 80, 20, theme)
	listModel.Title = fmt.Sprintf("%s Projects (%d)", icon.Bookmarks.String(), len(items))
	listModel.SetShowStatusBar(false)
	listModel.SetFilteringEnabled(true)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.Primary)

	model := projectListModel{
		list:            listModel,
		theme:           theme,
		responsive:      ui.NewResponsiveManager(80),
		config:          cfg,
		cwd:             cwd,
		projects:        projects,
		scanner:         scanner,
		spinner:         s,
		indexing:        !usedCache,
		indexingIndex:   0,
		activeView:      "list",
		barAnimProgress: &barAnim,
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
	theme           ui.Theme
	barAnimProgress *float64
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

	titleText := proj.Project.Name
	if proj.Project.Pinned {
		titleText = titleText + " " + icon.Pin.String()
	}
	title := truncateString(titleText, contentWidth)
	completed := proj.Project.CompletedTasks
	total := proj.Project.TotalTasks
	todoCount := proj.Project.TodoCount

	// Render progress bar (full width of the content area)
	barWidth := contentWidth
	filled := 0
	if total > 0 {
		filled = (completed * barWidth) / total
	}
	// Apply animation scaling
	if d.barAnimProgress != nil && *d.barAnimProgress < 1.0 {
		filled = int(float64(filled) * *d.barAnimProgress)
	}

	// Styles
	titleStyle := lipgloss.NewStyle().Bold(true)
	var barStr string
	statsStyle := lipgloss.NewStyle().Foreground(d.theme.ColorTasks)
	todoStyle := lipgloss.NewStyle().Foreground(d.theme.Error)
	pathStyle := lipgloss.NewStyle().Foreground(d.theme.Muted)

	rainbowColors := ui.GetRainbowColors(proj.Config.Rainbow)
	useRainbow := len(rainbowColors) > 0

	if isSelected {
		if useRainbow {
			c := rainbowColors[index%len(rainbowColors)]
			titleStyle = titleStyle.Foreground(c)
			fgStyle := lipgloss.NewStyle().Foreground(c)
			bgStyle := lipgloss.NewStyle().Foreground(d.theme.ColorHighlightProgressBg)
			filledPart := fgStyle.Render(strings.Repeat("█", filled))
			emptyPart := bgStyle.Render(strings.Repeat("░", barWidth-filled))
			barStr = filledPart + emptyPart
		} else {
			titleStyle = titleStyle.Foreground(d.theme.ColorHighlightName)

			fgStyle := lipgloss.NewStyle().Foreground(d.theme.ColorHighlightProgressFg)
			bgStyle := lipgloss.NewStyle().Foreground(d.theme.ColorHighlightProgressBg)
			filledPart := fgStyle.Render(strings.Repeat("█", filled))
			emptyPart := bgStyle.Render(strings.Repeat("░", barWidth-filled))
			barStr = filledPart + emptyPart
		}
	} else {
		if useRainbow {
			c := rainbowColors[index%len(rainbowColors)]
			titleStyle = titleStyle.Foreground(c)
			fgStyle := lipgloss.NewStyle().Foreground(c)
			bgStyle := lipgloss.NewStyle().Foreground(d.theme.ColorProgressBg)
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
	}

	statsText := fmt.Sprintf("(%d/%d) tasks complete", completed, total)
	todoText := fmt.Sprintf("%d todos", todoCount)
	leftPart := fmt.Sprintf("%s  %s", statsStyle.Render(statsText), todoStyle.Render(todoText))
	leftWidth := lipgloss.Width(leftPart)

	var metadataLine string
	// Room available for path to the right
	pathMaxLen := contentWidth - leftWidth - 2
	if pathMaxLen >= 10 {
		truncatedPath := truncatePath(proj.Project.Path, pathMaxLen)
		renderedPath := pathStyle.Render(truncatedPath)
		pathWidth := lipgloss.Width(renderedPath)
		spaces := contentWidth - leftWidth - pathWidth
		if spaces < 2 {
			spaces = 2
		}
		metadataLine = leftPart + strings.Repeat(" ", spaces) + renderedPath
	} else if contentWidth-lipgloss.Width(statsStyle.Render(statsText))-2 >= 10 {
		leftPartShort := statsStyle.Render(statsText)
		leftWShort := lipgloss.Width(leftPartShort)
		truncatedPath := truncatePath(proj.Project.Path, contentWidth-leftWShort-2)
		renderedPath := pathStyle.Render(truncatedPath)
		pathWidth := lipgloss.Width(renderedPath)
		spaces := contentWidth - leftWShort - pathWidth
		if spaces < 2 {
			spaces = 2
		}
		metadataLine = leftPartShort + strings.Repeat(" ", spaces) + renderedPath
	} else {
		metadataLine = leftPart
	}

	content := titleStyle.Render(title) + "\n" + barStr + "\n" + metadataLine

	var rendered string
	if isSelected {
		borderCol := d.theme.ColorHighlightBorder
		if useRainbow {
			borderCol = rainbowColors[index%len(rainbowColors)]
		}
		rendered = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderCol).
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

type barAnimTickMsg struct{}

func barAnimTickCmd() tea.Cmd {
	return tea.Tick(25*time.Millisecond, func(t time.Time) tea.Msg {
		return barAnimTickMsg{}
	})
}

type projectListModel struct {
	list             list.Model
	theme            ui.Theme
	responsive       *ui.ResponsiveManager
	config           domain.Config
	cwd              string
	message          string
	selectedCmd      string
	projects         []Project
	scanner          *Scanner
	spinner          spinner.Model
	indexing         bool
	indexingIndex    int
	indexingComplete bool
	screenW          int
	screenH          int
	barAnimProgress  *float64
	barAnimDone      bool

	activeView       string // "list" or "project"
	projectPageModel *projectPageModel
}

func (m projectListModel) Init() tea.Cmd {
	cmds := []tea.Cmd{barAnimTickCmd()}
	if m.indexing && len(m.projects) > 0 {
		cmds = append(cmds, m.spinner.Tick, scanProjectCmd(m.scanner, 0, m.projects[0].Path))
	}
	return tea.Batch(cmds...)
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
		if m.projectPageModel != nil {
			m.projectPageModel.width = msg.Width
			m.projectPageModel.height = msg.Height
		}
		return m, nil

	case scanProjectResultMsg:
		if msg.err == nil {
			pinned := m.projects[msg.index].Pinned
			pinOrder := m.projects[msg.index].PinOrder
			m.projects[msg.index] = msg.proj
			m.projects[msg.index].Pinned = pinned
			m.projects[msg.index].PinOrder = pinOrder

			sortProjects(m.projects)

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
			// Save scanned projects to cache
			if s, err := index.NewStore(); err == nil {
				_ = s.SaveProjectsCache(m.projects)
				s.Close()
			}
			return m, delayClearIndexingCmd()
		}

	case clearIndexingStatusMsg:
		m.indexingComplete = false
		width, height := m.responsive.GetListDimensions(m.screenW, m.screenH)
		m.list.SetSize(width, height)
		return m, nil

	case barAnimTickMsg:
		if !m.barAnimDone && m.barAnimProgress != nil {
			*m.barAnimProgress += 0.05
			if *m.barAnimProgress >= 1.0 {
				*m.barAnimProgress = 1.0
				m.barAnimDone = true
				return m, nil
			}
			return m, barAnimTickCmd()
		}
		return m, nil

	case spinner.TickMsg:
		if m.indexing {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case tea.KeyMsg:
		if m.activeView == "project" {
			newPm, cmd := m.projectPageModel.Update(msg)
			if pm, ok := newPm.(projectPageModel); ok {
				m.projectPageModel = &pm
				if pm.selectedCmd != "" {
					m.selectedCmd = pm.selectedCmd
				}
				return m, cmd
			}
			return newPm, cmd
		}

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

		case "p":
			if item, ok := m.list.SelectedItem().(projectItem); ok {
				targetPath := item.Project.Path
				store, err := index.NewStore()
				if err == nil {
					defer store.Close()
					if item.Project.Pinned {
						_ = store.UnpinProject(targetPath)
					} else {
						_ = store.PinProject(targetPath)
					}
					pinnedMap, _ := store.GetPinnedProjects()
					for i := range m.projects {
						order, pinned := pinnedMap[m.projects[i].Path]
						m.projects[i].Pinned = pinned
						m.projects[i].PinOrder = order
					}
				} else {
					for i := range m.projects {
						if m.projects[i].Path == targetPath {
							m.projects[i].Pinned = !m.projects[i].Pinned
							break
						}
					}
				}

				// Reset list filter to exit the filtering view
				m.list.ResetFilter()

				sortProjects(m.projects)

				items := buildListItems(m.projects, m.config.ShowTaskless, m.config)
				m.list.SetItems(items)

				newCursorIdx := 0
				for idx, itm := range items {
					if pi, ok := itm.(projectItem); ok && pi.Project.Path == targetPath {
						newCursorIdx = idx
						break
					}
				}
				m.list.Select(newCursorIdx)
				return m, nil
			}

		case "s":
			if !m.indexing {
				// Reset all projects to unscanned state and start fresh scan
				for i := range m.projects {
					m.projects[i].Scanned = false
					m.projects[i].CompletedTasks = 0
					m.projects[i].TotalTasks = 0
					m.projects[i].TodoCount = 0
					m.projects[i].Todos = nil
					m.projects[i].FtsFiles = nil
				}
				m.indexing = true
				m.indexingIndex = 0
				m.indexingComplete = false
				items := buildListItems(m.projects, m.config.ShowTaskless, m.config)
				m.list.SetItems(items)
				return m, tea.Batch(
					m.spinner.Tick,
					scanProjectCmd(m.scanner, 0, m.projects[0].Path),
				)
			}

		case "enter":
			if item, ok := m.list.SelectedItem().(projectItem); ok {
				m.activeView = "project"
				pm := newProjectPageModel(&item.Project, m.config, m.theme, func() (tea.Model, tea.Cmd) {
					m.activeView = "list"
					return m, nil
				}, m)
				pm.width = m.list.Width()
				pm.height = m.list.Height()
				m.projectPageModel = &pm
				return m, nil
			}

		case "o":
			if item, ok := m.list.SelectedItem().(projectItem); ok {
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

	if m.activeView == "project" && m.projectPageModel != nil {
		body += m.projectPageModel.View()
	} else {
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
			key.WithHelp("enter/o", "open project"),
		),
		key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "pin/unpin"),
		),
		key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "sync"),
		),
	}
}

func (m projectListModel) getShortHelpKeys() []key.Binding {
	return m.allHelpKeys()
}

func sortProjects(projects []Project) {
	sort.SliceStable(projects, func(i, j int) bool {
		// 1. Recurring projects come first (daily before weekly if both)
		if projects[i].IsRecurring && projects[j].IsRecurring {
			if projects[i].RecurringType == "daily" && projects[j].RecurringType != "daily" {
				return true
			}
			if projects[i].RecurringType != "daily" && projects[j].RecurringType == "daily" {
				return false
			}
			return strings.ToLower(projects[i].Name) < strings.ToLower(projects[j].Name)
		}
		if projects[i].IsRecurring && !projects[j].IsRecurring {
			return true
		}
		if !projects[i].IsRecurring && projects[j].IsRecurring {
			return false
		}

		// 2. Pinned projects come next
		if projects[i].Pinned && projects[j].Pinned {
			return projects[i].PinOrder < projects[j].PinOrder
		}
		if projects[i].Pinned && !projects[j].Pinned {
			return true
		}
		if !projects[i].Pinned && projects[j].Pinned {
			return false
		}

		// 3. Regular projects by name
		return strings.ToLower(projects[i].Name) < strings.ToLower(projects[j].Name)
	})
}

func quoteShellArg(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
