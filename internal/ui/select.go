package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

const fixedLines = 2 // header + filter/blank line
const minVisible = 5

type multiSelectModel struct {
	header     string   // header message
	choices    []string // display labels (may include tags like [fork])
	values     []string // actual values to return
	dimmed     map[int]bool
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

func newMultiSelectModel(header string, choices, values []string, dimmed map[int]bool) multiSelectModel {
	filtered := make([]int, len(choices))
	for i := range choices {
		filtered[i] = i
	}
	m := multiSelectModel{
		header:     header,
		choices:    choices,
		values:     values,
		dimmed:     dimmed,
		filtered:   filtered,
		selected:   make(map[int]bool),
		termHeight: minVisible + fixedLines,
	}
	m.skipToSelectable()
	return m
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
		m.moveCursor(-1)
	case "down", "j":
		m.moveCursor(1)
	case " ":
		if len(m.filtered) > 0 {
			idx := m.filtered[m.cursor]
			if !m.dimmed[idx] {
				m.selected[idx] = !m.selected[idx]
			}
		}
	case "a":
		for _, idx := range m.filtered {
			if !m.dimmed[idx] {
				m.selected[idx] = true
			}
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

// moveCursor moves the cursor by delta, skipping dimmed items.
func (m *multiSelectModel) moveCursor(delta int) {
	for {
		next := m.cursor + delta
		if next < 0 || next >= len(m.filtered) {
			return
		}
		m.cursor = next
		m.adjustScroll()
		if !m.dimmed[m.filtered[m.cursor]] {
			return
		}
	}
}

// skipToSelectable moves the cursor to the next selectable item from current position.
func (m *multiSelectModel) skipToSelectable() {
	if len(m.filtered) == 0 {
		return
	}
	// Try forward first
	for i := m.cursor; i < len(m.filtered); i++ {
		if !m.dimmed[m.filtered[i]] {
			m.cursor = i
			m.adjustScroll()
			return
		}
	}
	// Try backward
	for i := m.cursor - 1; i >= 0; i-- {
		if !m.dimmed[m.filtered[i]] {
			m.cursor = i
			m.adjustScroll()
			return
		}
	}
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
	m.skipToSelectable()
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
	b.WriteString(Bold(m.header) + " (space: select, a: all, n: none, /: filter, enter: confirm, esc: cancel)\n")

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
		} else if m.dimmed[idx] {
			b.WriteString(fmt.Sprintf("%s%s\n", cursor, Dim("[ ] "+choice)))
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

// RepoOption represents a repository with display metadata.
type RepoOption struct {
	Name   string
	Fork   bool
	Cloned bool
}

// SelectRepos shows an interactive multi-select prompt for repository selection.
func SelectRepos(repos []RepoOption) ([]string, error) {
	choices := make([]string, len(repos))
	values := make([]string, len(repos))
	dimmed := make(map[int]bool)
	for i, r := range repos {
		label := r.Name
		var tags []string
		if r.Fork {
			tags = append(tags, Warn("[fork]"))
		}
		if r.Cloned {
			tags = append(tags, "[cloned]")
			dimmed[i] = true
		}
		if len(tags) > 0 {
			label += " " + strings.Join(tags, " ")
		}
		choices[i] = label
		values[i] = r.Name
	}

	m := newMultiSelectModel("Select repositories to clone:", choices, values, dimmed)
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
	for i := range final.choices {
		if final.selected[i] {
			selected = append(selected, final.values[i])
		}
	}
	return selected, nil
}

// SelectItems shows an interactive multi-select prompt with a custom header.
func SelectItems(header string, items []string) ([]string, error) {
	m := newMultiSelectModel(header, items, items, nil)
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
	for i := range final.choices {
		if final.selected[i] {
			selected = append(selected, final.values[i])
		}
	}
	return selected, nil
}
