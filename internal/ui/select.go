package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

const fixedLines = 2 // header + filter/blank line
const minVisible = 5

type multiSelectModel struct {
	choices    []string
	filtered   []int // indices into choices matching the filter
	cursor     int   // index into filtered
	offset     int   // scroll offset
	termHeight int   // terminal height
	selected   map[int]bool
	filtering  bool
	filter     string
	done       bool
	canceled   bool
}

func newMultiSelectModel(choices []string) multiSelectModel {
	filtered := make([]int, len(choices))
	for i := range choices {
		filtered[i] = i
	}
	return multiSelectModel{
		choices:    choices,
		filtered:   filtered,
		selected:   make(map[int]bool),
		termHeight: minVisible + fixedLines,
	}
}

func (m multiSelectModel) Init() tea.Cmd {
	return nil
}

func (m multiSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.termHeight = msg.Height
		m.adjustScroll()

	case tea.KeyMsg:
		if m.filtering {
			return m.updateFilter(msg)
		}
		return m.updateNormal(msg)
	}
	return m, nil
}

func (m multiSelectModel) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc":
		m.canceled = true
		return m, tea.Quit
	case "enter":
		m.done = true
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
			m.adjustScroll()
		}
	case "down", "j":
		if m.cursor < len(m.filtered)-1 {
			m.cursor++
			m.adjustScroll()
		}
	case " ":
		if len(m.filtered) > 0 {
			idx := m.filtered[m.cursor]
			m.selected[idx] = !m.selected[idx]
		}
	case "a":
		for _, idx := range m.filtered {
			m.selected[idx] = true
		}
	case "n":
		for _, idx := range m.filtered {
			delete(m.selected, idx)
		}
	case "/":
		m.filtering = true
	}
	return m, nil
}

func (m multiSelectModel) updateFilter(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "esc":
		m.filtering = false
		return m, nil
	case "backspace":
		if len(m.filter) > 0 {
			m.filter = m.filter[:len(m.filter)-1]
			m.applyFilter()
		}
	case "ctrl+u":
		m.filter = ""
		m.applyFilter()
	default:
		if len(msg.String()) == 1 {
			m.filter += msg.String()
			m.applyFilter()
		}
	}
	return m, nil
}

func (m *multiSelectModel) applyFilter() {
	m.filtered = m.filtered[:0]
	lower := strings.ToLower(m.filter)
	for i, choice := range m.choices {
		if m.filter == "" || strings.Contains(strings.ToLower(choice), lower) {
			m.filtered = append(m.filtered, i)
		}
	}
	m.cursor = 0
	m.offset = 0
}

func (m multiSelectModel) visibleHeight() int {
	total := len(m.filtered)
	available := m.termHeight - fixedLines
	if available < minVisible {
		available = minVisible
	}

	// If all items fit, no indicators needed
	if total <= available {
		return total
	}

	// Reserve space for scroll indicators (up to 2 lines)
	h := available - 2
	if h < minVisible {
		h = minVisible
	}
	if h > total {
		h = total
	}
	return h
}

func (m *multiSelectModel) adjustScroll() {
	h := m.visibleHeight()
	if h == 0 {
		return
	}
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+h {
		m.offset = m.cursor - h + 1
	}
}

func (m multiSelectModel) View() string {
	var b strings.Builder
	b.WriteString(Bold("Select repositories to clone:") + " (space: select, a: all, n: none, /: filter, enter: confirm, esc: cancel)\n")

	if m.filtering {
		fmt.Fprintf(&b, "Filter: %s█  %s\n", m.filter, Warn("(ctrl+u: clear, esc/enter: close filter)"))
	} else if m.filter != "" {
		b.WriteString(fmt.Sprintf("Filter: %s (%d/%d)\n", Warn(m.filter), len(m.filtered), len(m.choices)))
	} else {
		b.WriteString("\n")
	}

	if len(m.filtered) == 0 {
		b.WriteString(Warn("  No matches found\n"))
		return b.String()
	}

	end := m.offset + m.visibleHeight()
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	if m.offset > 0 {
		b.WriteString(fmt.Sprintf("  ... (%d more above)\n", m.offset))
	}

	for fi := m.offset; fi < end; fi++ {
		idx := m.filtered[fi]
		choice := m.choices[idx]
		cursor := "  "
		if m.cursor == fi {
			cursor = "> "
		}

		if m.selected[idx] {
			b.WriteString(fmt.Sprintf("%s%s\n", cursor, Info("[x] "+choice)))
		} else {
			b.WriteString(fmt.Sprintf("%s[ ] %s\n", cursor, choice))
		}
	}

	remaining := len(m.filtered) - end
	if remaining > 0 {
		b.WriteString(fmt.Sprintf("  ... (%d more below)\n", remaining))
	}

	return b.String()
}

// SelectRepos shows an interactive multi-select prompt for repository selection.
func SelectRepos(repos []string) ([]string, error) {
	m := newMultiSelectModel(repos)
	p := tea.NewProgram(m)

	result, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("prompt failed: %w", err)
	}

	final := result.(multiSelectModel)

	if final.canceled {
		return nil, nil
	}

	var selected []string
	for i, name := range final.choices {
		if final.selected[i] {
			selected = append(selected, name)
		}
	}
	return selected, nil
}
