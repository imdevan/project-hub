package todo

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/super-todo/internal/adapters/editor"
	"github.com/super-todo/internal/domain"
	"github.com/super-todo/internal/ui"
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
			// Toggle between tabs (0: TODOs, 1: Tasks)
			if m.activeTab == 0 {
				m.activeTab = 1
			} else {
				m.activeTab = 0
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
				if m.taskSelected < len(m.project.FtsFiles)-1 {
					m.taskSelected++
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
				fts := m.project.FtsFiles[m.taskSelected]
				relPath = fts.Path
				targetPath = filepath.Join(m.project.Path, relPath)
			}

			if m.config.ExitWithAlias != "" {
				tmpl, err := template.New("exit_with_alias").Funcs(template.FuncMap{
					"quote": quoteShellArg,
				}).Parse(m.config.ExitWithAlias)
				if err != nil {
					m.message = fmt.Sprintf("✗ template error: %s", err)
					return m, nil
				}

				var buf bytes.Buffer
				data := map[string]string{
					"Project":     quoteShellArg(m.project.Path),
					"ProjectPath": quoteShellArg(m.project.Path),
					"TaskFile":    quoteShellArg(relPath),
					"Line":        strconv.Itoa(lineNum),
				}
				if err := tmpl.Execute(&buf, data); err != nil {
					m.message = fmt.Sprintf("✗ template execution error: %s", err)
					return m, nil
				}

				m.selectedCmd = buf.String()
				
				if pm, ok := m.parentListModel.(projectListModel); ok {
					pm.selectedCmd = m.selectedCmd
					m.parentListModel = pm
				}

				return m, tea.Quit
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
	
	nameLine := nameStr + strings.Repeat(" ", spaces) + locationStr
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
	statsLine := fmt.Sprintf("(%d/%d) tasks complete   %d todos", completed, total, m.project.TodoCount)
	sections = append(sections, statsStyle.Render(statsLine))

	sections = append(sections, "") // blank line separator

	// 3. Tabs bar
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

	// 4. Render active tab items
	availableHeight := m.height - 10
	if availableHeight < 5 {
		availableHeight = 5
	}

	type renderedBlock struct {
		index int
		block string
		lines int
	}

	var blocks []renderedBlock
	if m.activeTab == 0 {
		for i, todo := range m.project.Todos {
			block := m.renderTodoBlock(todo, i == m.todoSelected)
			blocks = append(blocks, renderedBlock{
				index: i,
				block: block,
				lines: 2,
			})
		}
	} else {
		for i, fts := range m.project.FtsFiles {
			block := m.renderTaskBlock(fts, i == m.taskSelected)
			blocks = append(blocks, renderedBlock{
				index: i,
				block: block,
				lines: 3,
			})
		}
	}

	// Calculate scrolling
	selected := m.todoSelected
	scrollOffset := m.todoScroll
	if m.activeTab == 1 {
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

	if m.activeTab == 0 {
		m.todoScroll = scrollOffset
	} else {
		m.taskScroll = scrollOffset
	}

	// Render visible viewport
	if len(blocks) == 0 {
		if m.activeTab == 0 {
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
			sections = append(sections, "")
			totalLinesRendered += blocks[i].lines + 1
		}
	}

	// Help footer & Pagination
	footerStyle := lipgloss.NewStyle().Foreground(m.theme.Muted)
	helpText := "j/k/↑/↓: navigate • tab/h/l: switch tab • o/enter: open • esc: back"
	
	var paginationText string
	if m.activeTab == 0 {
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
	var lines []string
	indicator := "  "
	if isSelected {
		indicator = lipgloss.NewStyle().Foreground(m.theme.Primary).Render("│ ")
	}

	titleStyle := lipgloss.NewStyle().Bold(true)
	if isSelected {
		titleStyle = titleStyle.Foreground(m.theme.ColorHighlightName)
	} else {
		titleStyle = titleStyle.Foreground(m.theme.ColorName)
	}

	locationStyle := lipgloss.NewStyle().Foreground(m.theme.Muted)

	lines = append(lines, fmt.Sprintf("%s%s", indicator, titleStyle.Render(todo.Text)))
	lines = append(lines, fmt.Sprintf("%s%s", indicator, locationStyle.Render(todo.File)))

	return strings.Join(lines, "\n")
}

func (m projectPageModel) renderTaskBlock(fts FtsFileDetail, isSelected bool) string {
	var lines []string
	indicator := "  "
	if isSelected {
		indicator = lipgloss.NewStyle().Foreground(m.theme.Primary).Render("│ ")
	}

	titleStyle := lipgloss.NewStyle().Bold(true)
	if isSelected {
		titleStyle = titleStyle.Foreground(m.theme.ColorHighlightName)
	} else {
		titleStyle = titleStyle.Foreground(m.theme.ColorName)
	}

	lines = append(lines, fmt.Sprintf("%s%s", indicator, titleStyle.Render(fts.Path)))

	barW := m.width - 16
	if barW < 20 {
		barW = 20
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
	lines = append(lines, fmt.Sprintf("%s  %s", indicator, fBarStr))

	statsStr := lipgloss.NewStyle().Foreground(m.theme.ColorTasks).Render(fmt.Sprintf("(%d/%d) tasks complete", fts.CompletedTasks, fts.TotalTasks))
	lines = append(lines, fmt.Sprintf("%s  %s", indicator, statsStr))

	return strings.Join(lines, "\n")
}
