package ui

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/project-hub/internal/adapters/tty"
	"github.com/project-hub/internal/importer"
)

// GitJobStyle returns the theme style for a git job status.
func GitJobStyle(theme Theme, a importer.Action) lipgloss.Style {
	color := map[importer.Action]lipgloss.Color{
		importer.Cloned:   theme.Success,
		importer.Pulled:   theme.Primary,
		importer.UpToDate: theme.Muted,
		importer.Ahead:    theme.Secondary,
		importer.Dirty:    theme.Flags,
		importer.Skipped:  theme.Muted,
		importer.Failed:   theme.Error,
	}[a]
	return lipgloss.NewStyle().Foreground(color)
}

// FormatGitJob renders one result as a status column, the label, and indented muted detail lines.
func FormatGitJob(theme Theme, label string, res importer.Result) string {
	width := 0
	for _, a := range importer.Actions {
		width = max(width, len(a)+1)
	}
	indent := "\n" + strings.Repeat(" ", width)
	line := GitJobStyle(theme, res.Action).Render(fmt.Sprintf("%-*s", width, res.Action)) + label
	if res.Detail != "" {
		line += indent + lipgloss.NewStyle().Foreground(theme.Muted).Render(strings.ReplaceAll(res.Detail, "\n", indent))
	}
	return line
}

// RunGitJobs runs run(0..n-1) on up to jobs workers and prints each result to out in index
// order as soon as it and every earlier job have finished. On a terminal it shows an inline
// spinner with the repositories currently in progress; ctrl+c stops starting new jobs and
// waits for running ones to finish. It returns every job's result, indexed like labels.
func RunGitJobs(out io.Writer, theme Theme, verb string, labels []string, jobs int, run func(i int) importer.Result) ([]importer.Result, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	events := importer.RunParallel(ctx, len(labels), jobs, run)

	if !tty.IsTerminal(os.Stdout.Fd()) {
		p := newOrderedPrinter(theme, labels)
		for ev := range events {
			if ev.Done {
				if s := p.add(ev.Index, ev.Result); s != "" {
					fmt.Fprintln(out, s)
				}
			}
		}
		return p.results(), nil
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.Primary)
	m := gitJobsModel{
		spinner: s,
		theme:   theme,
		verb:    verb,
		labels:  labels,
		events:  events,
		cancel:  cancel,
		running: map[int]bool{},
		printer: newOrderedPrinter(theme, labels),
	}
	final, err := tea.NewProgram(m, tty.GetProgramOptions()...).Run()
	if err != nil {
		return nil, err
	}
	fm := final.(gitJobsModel)
	if fm.stopping {
		return fm.printer.results(), fmt.Errorf("%s interrupted", verb)
	}
	return fm.printer.results(), nil
}

// orderedPrinter buffers out-of-order results and releases them in index order.
type orderedPrinter struct {
	theme  Theme
	labels []string
	res    []*importer.Result
	next   int
}

func newOrderedPrinter(theme Theme, labels []string) *orderedPrinter {
	return &orderedPrinter{theme: theme, labels: labels, res: make([]*importer.Result, len(labels))}
}

// add records result i and returns the lines that are now ready to print (possibly "").
func (p *orderedPrinter) add(i int, r importer.Result) string {
	p.res[i] = &r
	var lines []string
	for p.next < len(p.res) && p.res[p.next] != nil {
		lines = append(lines, FormatGitJob(p.theme, p.labels[p.next], *p.res[p.next]))
		p.next++
	}
	return strings.Join(lines, "\n")
}

func (p *orderedPrinter) results() []importer.Result {
	out := make([]importer.Result, len(p.res))
	for i, r := range p.res {
		if r != nil {
			out[i] = *r
		}
	}
	return out
}

type gitJobEventMsg importer.Event
type gitJobsDoneMsg struct{}

type gitJobsModel struct {
	spinner  spinner.Model
	theme    Theme
	verb     string
	labels   []string
	events   <-chan importer.Event
	cancel   context.CancelFunc
	running  map[int]bool
	printer  *orderedPrinter
	done     int
	stopping bool
	finished bool
}

func (m gitJobsModel) waitEvent() tea.Msg {
	ev, ok := <-m.events
	if !ok {
		return gitJobsDoneMsg{}
	}
	return gitJobEventMsg(ev)
}

func (m gitJobsModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.waitEvent)
}

func (m gitJobsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case gitJobEventMsg:
		if !msg.Done {
			m.running[msg.Index] = true
			return m, m.waitEvent
		}
		delete(m.running, msg.Index)
		m.done++
		if lines := m.printer.add(msg.Index, msg.Result); lines != "" {
			// Sequence keeps prints ordered: the next event is only read after this print is queued.
			return m, tea.Sequence(tea.Println(lines), m.waitEvent)
		}
		return m, m.waitEvent
	case gitJobsDoneMsg:
		m.finished = true
		return m, tea.Quit
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC && !m.stopping {
			m.stopping = true
			m.cancel()
		}
		return m, nil
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m gitJobsModel) View() string {
	if m.finished {
		return ""
	}
	muted := lipgloss.NewStyle().Foreground(m.theme.Muted)
	header := fmt.Sprintf("%s %s %d/%d", m.spinner.View(), m.verb, m.done, len(m.labels))
	if m.stopping {
		header += muted.Render("  stopping: waiting for running jobs to finish")
	} else {
		header += muted.Render("  ctrl+c to stop")
	}
	idx := make([]int, 0, len(m.running))
	for i := range m.running {
		idx = append(idx, i)
	}
	sort.Ints(idx)
	lines := []string{header}
	for _, i := range idx {
		lines = append(lines, muted.Render("  "+m.labels[i]))
	}
	return strings.Join(lines, "\n")
}
