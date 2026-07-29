package todo

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/super-todo/internal/adapters/editor"
	"github.com/super-todo/internal/adapters/icon"
	"github.com/super-todo/internal/domain"
	"github.com/super-todo/internal/ui"
	"github.com/super-todo/internal/utils"
)

type projectPageModel struct {
	project         *Project
	config          domain.Config
	theme           ui.Theme
	width           int
	height          int
	activeTab       int // 0: TODOs, 1: Tasks
	todoSelected    int
	todoScroll      int
	taskSelected    int
	taskScroll      int
	message         string
	selectedCmd     string
	onBack          func() (tea.Model, tea.Cmd)
	parentListModel tea.Model
}

func newProjectPageModel(proj *Project, cfg domain.Config, theme ui.Theme, onBack func() (tea.Model, tea.Cmd), parent tea.Model) projectPageModel {
	activeTab := 0
	if len(proj.Todos) == 0 && len(proj.FtsFiles) > 0 {
		activeTab = 1
	}

	return projectPageModel{
		project:         proj,
		config:          cfg,
		theme:           theme,
		onBack:          onBack,
		parentListModel: parent,
		activeTab:       activeTab,
	}
}

func (m projectPageModel) Init() tea.Cmd {
	return nil
}

func (m projectPageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.parentListModel != nil {
			m.parentListModel, _ = m.parentListModel.Update(msg)
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "esc", "backspace":
			if m.onBack != nil {
				return m.onBack()
			}
			return m, nil

		case "tab", "right", "left":
			// Toggle between tabs (0: TODOs, 1: Tasks) for regular projects
			if !m.project.IsRecurring {
				if m.activeTab == 0 {
					m.activeTab = 1
				} else {
					m.activeTab = 0
				}
			}
			return m, nil

		case "up", "k":
			if m.activeTab == 0 {
				if m.todoSelected > 0 {
					m.todoSelected--
				}
			} else {
				if m.taskSelected > 0 {
					m.taskSelected--
				}
			}
			return m, nil

		case "down", "j":
			if m.activeTab == 0 {
				if m.todoSelected < len(m.project.Todos)-1 {
					m.todoSelected++
				}
			} else {
				maxItems := 0
				if m.project.IsRecurring {
					if len(m.project.FtsFiles) > 0 {
						for _, feat := range m.project.FtsFiles[0].Features {
							maxItems += len(feat.Tasks)
						}
					}
				} else {
					maxItems = len(m.project.FtsFiles)
				}
				if m.taskSelected < maxItems-1 {
					m.taskSelected++
				}
			}
			return m, nil

		case " ":
			// Toggle checkbox for recurring tasks via space key
			if m.project.IsRecurring && len(m.project.FtsFiles) > 0 {
				var allTasks []TaskItem
				for _, feat := range m.project.FtsFiles[0].Features {
					allTasks = append(allTasks, feat.Tasks...)
				}
				if m.taskSelected >= 0 && m.taskSelected < len(allTasks) {
					task := allTasks[m.taskSelected]
					if err := ToggleTaskInFile(m.project.Path, task.Line); err == nil {
						// Rescan file content
						completed, total, features, err := ParseFTSDetailed(m.project.Path, m.config)
						if err == nil {
							m.project.CompletedTasks = completed
							m.project.TotalTasks = total
							m.project.FtsFiles[0].CompletedTasks = completed
							m.project.FtsFiles[0].TotalTasks = total
							m.project.FtsFiles[0].Features = features
						}
					}
				}
			}
			return m, nil

		case "enter", "o":
			relPath := "."
			targetPath := m.project.Path
			lineNum := 1

			if m.activeTab == 0 {
				if len(m.project.Todos) == 0 {
					return m, nil
				}
				todo := m.project.Todos[m.todoSelected]
				relPath = todo.File
				targetPath = filepath.Join(m.project.Path, relPath)
				lineNum = todo.Line
			} else {
				if len(m.project.FtsFiles) == 0 {
					return m, nil
				}
				if m.project.IsRecurring {
					targetPath = m.project.Path
					var allTasks []TaskItem
					for _, feat := range m.project.FtsFiles[0].Features {
						allTasks = append(allTasks, feat.Tasks...)
					}
					if m.taskSelected >= 0 && m.taskSelected < len(allTasks) {
						lineNum = allTasks[m.taskSelected].Line
					}
				} else {
					fts := m.project.FtsFiles[m.taskSelected]
					relPath = fts.Path
					targetPath = filepath.Join(m.project.Path, relPath)
				}
			}



			editorAdapter := editor.New(m.config.Editor)
			c, err := editorAdapter.BuildCmd(targetPath, lineNum)
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

	return m, nil
}

func (m projectPageModel) View() string {
	var sections []string

	// 1. Back Header
	headerStyle := lipgloss.NewStyle().Foreground(m.theme.Primary).Bold(true)
	sections = append(sections, headerStyle.Render("← Back to Projects (esc)"))
	sections = append(sections, "")

	// 2. Project Metadata block
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(m.theme.ColorTitle)
	pathStyle := lipgloss.NewStyle().Foreground(m.theme.Muted)

	nameStr := titleStyle.Render(m.project.Name)
	locationStr := pathStyle.Render(m.project.Path)
	
	metaWidth := m.width - 4
	if metaWidth < 20 {
		metaWidth = 20
	}
	
	spaces := metaWidth - lipgloss.Width(nameStr) - lipgloss.Width(locationStr)
	if spaces < 2 {
		spaces = 2
	}
	
	nameLine := nameStr
	if !m.project.IsRecurring {
		nameLine = nameStr + strings.Repeat(" ", spaces) + locationStr
	}
	sections = append(sections, nameLine)

	// Progress bar
	completed := m.project.CompletedTasks
	total := m.project.TotalTasks
	barWidth := m.width - 4
	if barWidth < 20 {
		barWidth = 20
	}
	
	filled := 0
	if total > 0 {
		filled = (completed * barWidth) / total
	}
	
	fgStyle := lipgloss.NewStyle().Foreground(m.theme.ColorProgressFg)
	bgStyle := lipgloss.NewStyle().Foreground(m.theme.ColorProgressBg)
	filledPart := fgStyle.Render(strings.Repeat("█", filled))
	emptyPart := bgStyle.Render(strings.Repeat("░", barWidth-filled))
	barStr := filledPart + emptyPart
	sections = append(sections, barStr)

	statsStyle := lipgloss.NewStyle().Foreground(m.theme.ColorTasks)
	var statsLine string
	if m.width < 60 {
		statsLine = fmt.Sprintf("(%d/%d)", completed, total)
		if !m.project.IsRecurring {
			statsLine = fmt.Sprintf("(%d/%d)   +%d", completed, total, m.project.TodoCount)
		}
	} else {
		statsLine = fmt.Sprintf("(%d/%d) tasks complete", completed, total)
		if !m.project.IsRecurring {
			statsLine = fmt.Sprintf("(%d/%d) tasks complete   %d todos", completed, total, m.project.TodoCount)
		}
	}
	sections = append(sections, statsStyle.Render(statsLine))

	sections = append(sections, "") // blank line separator

	// 3. Tabs bar (only for regular projects)
	if !m.project.IsRecurring {
		todoLabel := fmt.Sprintf("TODOs (%d)", len(m.project.Todos))
		taskLabel := fmt.Sprintf("Tasks (%d)", len(m.project.FtsFiles))
		var tabs []string
		
		if m.activeTab == 0 {
			tabs = append(tabs, lipgloss.NewStyle().Foreground(m.theme.ColorHighlightName).Bold(true).Underline(true).Render(todoLabel))
			tabs = append(tabs, lipgloss.NewStyle().Foreground(m.theme.Muted).Render(taskLabel))
		} else {
			tabs = append(tabs, lipgloss.NewStyle().Foreground(m.theme.Muted).Render(todoLabel))
			tabs = append(tabs, lipgloss.NewStyle().Foreground(m.theme.ColorHighlightName).Bold(true).Underline(true).Render(taskLabel))
		}
		
		sections = append(sections, "  "+strings.Join(tabs, "    "))
		sections = append(sections, "") // blank line separator
	}

	// 4. Render active tab items or recurring task list
	availableHeight := m.height - 10
	if m.project.IsRecurring {
		availableHeight = m.height - 8
	}
	if availableHeight < 2 {
		availableHeight = 2
	}

	type renderedBlock struct {
		index int
		block string
		lines int
	}

	var blocks []renderedBlock
	var totalTaskItems int

	if m.project.IsRecurring {
		if len(m.project.FtsFiles) > 0 {
			taskIdx := 0
			for _, feat := range m.project.FtsFiles[0].Features {
				for _, task := range feat.Tasks {
					block := m.renderRecurringTaskItem(task, taskIdx == m.taskSelected)
					blockLines := strings.Split(block, "\n")
					blocks = append(blocks, renderedBlock{
						index: taskIdx,
						block: block,
						lines: len(blockLines),
					})
					taskIdx++
				}
			}
			totalTaskItems = taskIdx
		}
	} else if m.activeTab == 0 {
		for i, todo := range m.project.Todos {
			block := m.renderTodoBlock(todo, i == m.todoSelected)
			blockLines := strings.Split(block, "\n")
			blocks = append(blocks, renderedBlock{
				index: i,
				block: block,
				lines: len(blockLines),
			})
		}
	} else {
		for i, fts := range m.project.FtsFiles {
			block := m.renderTaskBlock(fts, i == m.taskSelected)
			blockLines := strings.Split(block, "\n")
			blocks = append(blocks, renderedBlock{
				index: i,
				block: block,
				lines: len(blockLines),
			})
		}
	}

	// Calculate scrolling
	selected := m.todoSelected
	scrollOffset := m.todoScroll
	if m.activeTab == 1 || m.project.IsRecurring {
		selected = m.taskSelected
		scrollOffset = m.taskScroll
	}

	fitsInViewport := func(offset, selected int) bool {
		linesRendered := 0
		for i := offset; i < len(blocks); i++ {
			if linesRendered+blocks[i].lines > availableHeight && i > offset {
				return false
			}
			if i == selected {
				return true
			}
			linesRendered += blocks[i].lines + 1
		}
		return false
	}

	if selected < scrollOffset {
		scrollOffset = selected
	}
	
	for !fitsInViewport(scrollOffset, selected) && scrollOffset < selected {
		scrollOffset++
	}

	if m.activeTab == 0 && !m.project.IsRecurring {
		m.todoScroll = scrollOffset
	} else {
		m.taskScroll = scrollOffset
	}

	// Render visible viewport
	if len(blocks) == 0 {
		if m.project.IsRecurring {
			sections = append(sections, lipgloss.NewStyle().Foreground(m.theme.Muted).Render("  No tasks found in this file."))
		} else if m.activeTab == 0 {
			sections = append(sections, lipgloss.NewStyle().Foreground(m.theme.Muted).Render("  No TODOs found in this project."))
		} else {
			sections = append(sections, lipgloss.NewStyle().Foreground(m.theme.Muted).Render("  No task files found in this project."))
		}
	} else {
		totalLinesRendered := 0
		for i := scrollOffset; i < len(blocks); i++ {
			if totalLinesRendered+blocks[i].lines > availableHeight && i > scrollOffset {
				break
			}
			sections = append(sections, blocks[i].block)
			totalLinesRendered += blocks[i].lines
		}
	}

	// Help footer & Pagination
	footerStyle := lipgloss.NewStyle().Foreground(m.theme.Muted)
	helpText := "j/k/↑/↓: navigate • tab/h/l: switch tab • o/enter: open • esc: back"
	if m.project.IsRecurring {
		helpText = "j/k/↑/↓: navigate • space: toggle • o/enter: open • esc: back"
	}
	
	var paginationText string
	if m.project.IsRecurring {
		paginationText = fmt.Sprintf("item %d of %d", m.taskSelected+1, totalTaskItems)
		if totalTaskItems == 0 {
			paginationText = "item 0 of 0"
		}
	} else if m.activeTab == 0 {
		paginationText = fmt.Sprintf("item %d of %d", m.todoSelected+1, len(m.project.Todos))
		if len(m.project.Todos) == 0 {
			paginationText = "item 0 of 0"
		}
	} else {
		paginationText = fmt.Sprintf("item %d of %d", m.taskSelected+1, len(m.project.FtsFiles))
		if len(m.project.FtsFiles) == 0 {
			paginationText = "item 0 of 0"
		}
	}
	
	footerWidth := m.width
	spaces = footerWidth - lipgloss.Width(helpText) - lipgloss.Width(paginationText)
	if spaces < 2 {
		spaces = 2
	}
	footerLine := helpText + strings.Repeat(" ", spaces) + paginationText
	sections = append(sections, footerStyle.Render(footerLine))

	// Combine sections into a single text block
	content := strings.Join(sections, "\n")
	
	// Split into individual lines to strictly enforce height constraints
	rawLines := strings.Split(content, "\n")
	
	var finalLines []string
	if len(rawLines) > m.height {
		footerLine := rawLines[len(rawLines)-1]
		finalLines = append([]string{}, rawLines[:m.height-1]...)
		finalLines = append(finalLines, footerLine)
	} else {
		finalLines = append([]string{}, rawLines...)
		for len(finalLines) < m.height {
			if len(finalLines) > 0 {
				footerLine := finalLines[len(finalLines)-1]
				finalLines = append(finalLines[:len(finalLines)-1], "", footerLine)
			} else {
				finalLines = append(finalLines, "")
			}
		}
	}
	
	return strings.Join(finalLines, "\n")
}

func (m projectPageModel) renderTodoBlock(todo TodoComment, isSelected bool) string {
	indicator := "  "
	if isSelected {
		indicator = lipgloss.NewStyle().Foreground(m.theme.Primary).Render("│ ")
	}

	contentWidth := m.width - 4
	if contentWidth < 10 {
		contentWidth = 10
	}

	titleStyle := lipgloss.NewStyle().Bold(true)
	if isSelected {
		titleStyle = titleStyle.Foreground(m.theme.ColorHighlightName)
	} else {
		titleStyle = titleStyle.Foreground(m.theme.ColorName)
	}

	locationStyle := lipgloss.NewStyle().Foreground(m.theme.Muted)

	wrappedTitle := utils.WrapText(todo.Text, contentWidth)
	wrappedLocation := utils.WrapText(todo.File, contentWidth)

	var lines []string
	for _, line := range wrappedTitle {
		lines = append(lines, indicator+titleStyle.Render(line))
	}
	for _, line := range wrappedLocation {
		lines = append(lines, indicator+locationStyle.Render(line))
	}

	return strings.Join(lines, "\n")
}

func (m projectPageModel) renderTaskBlock(fts FtsFileDetail, isSelected bool) string {
	indicator := "  "
	if isSelected {
		indicator = lipgloss.NewStyle().Foreground(m.theme.Primary).Render("│ ")
	}

	contentWidth := m.width - 4
	if contentWidth < 10 {
		contentWidth = 10
	}

	titleStyle := lipgloss.NewStyle().Bold(true)
	if isSelected {
		titleStyle = titleStyle.Foreground(m.theme.ColorHighlightName)
	} else {
		titleStyle = titleStyle.Foreground(m.theme.ColorName)
	}

	wrappedPath := utils.WrapText(fts.Path, contentWidth)

	barW := contentWidth - 12
	if barW < 10 {
		barW = 10
	}
	if barW > 60 {
		barW = 60
	}
	
	fFilled := 0
	if fts.TotalTasks > 0 {
		fFilled = (fts.CompletedTasks * barW) / fts.TotalTasks
	}
	
	fFgStyle := lipgloss.NewStyle().Foreground(m.theme.ColorProgressFg)
	fBgStyle := lipgloss.NewStyle().Foreground(m.theme.ColorProgressBg)
	if isSelected {
		fFgStyle = lipgloss.NewStyle().Foreground(m.theme.ColorHighlightProgressFg)
		fBgStyle = lipgloss.NewStyle().Foreground(m.theme.ColorHighlightProgressBg)
	}
	
	fFilledPart := fFgStyle.Render(strings.Repeat("█", fFilled))
	fEmptyPart := fBgStyle.Render(strings.Repeat("░", barW-fFilled))
	fBarStr := fFilledPart + fEmptyPart

	statsStr := lipgloss.NewStyle().Foreground(m.theme.ColorTasks).Render(fmt.Sprintf("(%d/%d) tasks complete", fts.CompletedTasks, fts.TotalTasks))

	var lines []string
	for _, line := range wrappedPath {
		lines = append(lines, indicator+titleStyle.Render(line))
	}
	lines = append(lines, indicator+"  "+fBarStr)
	lines = append(lines, indicator+"  "+statsStr)

	return strings.Join(lines, "\n")
}

func (m projectPageModel) renderRecurringTaskItem(task TaskItem, isSelected bool) string {
	indicator := "  "
	if isSelected {
		indicator = lipgloss.NewStyle().Foreground(m.theme.Primary).Render("│ ")
	}

	checkbox := icon.CheckboxUnchecked.String() + " "
	checkStyle := lipgloss.NewStyle().Foreground(m.theme.Muted)
	textStyle := lipgloss.NewStyle().Foreground(m.theme.ColorName)

	if task.Completed {
		checkbox = icon.CheckboxChecked.String() + " "
		checkStyle = lipgloss.NewStyle().Foreground(m.theme.Success)
		textStyle = textStyle.Strikethrough(true).Foreground(m.theme.Muted)
	}

	if isSelected {
		textStyle = textStyle.Foreground(m.theme.ColorHighlightName).Bold(true)
	}

	contentWidth := m.width - 8
	if contentWidth < 10 {
		contentWidth = 10
	}

	wrappedText := utils.WrapText(task.Text, contentWidth)
	var lines []string
	for i, line := range wrappedText {
		if i == 0 {
			lines = append(lines, indicator+checkStyle.Render(checkbox)+textStyle.Render(line))
		} else {
			lines = append(lines, indicator+"    "+textStyle.Render(line))
		}
	}

	return strings.Join(lines, "\n")
}
