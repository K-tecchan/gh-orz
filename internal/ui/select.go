package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

const headerLines = 2 // header + blank line
const minVisible = 5

type multiSelectModel struct {
	choices  []string
	cursor   int
	offset   int // scroll offset
	height   int // visible area height (items)
	selected map[int]bool
	done     bool
	canceled bool
}

func newMultiSelectModel(choices []string) multiSelectModel {
	return multiSelectModel{
		choices:  choices,
		selected: make(map[int]bool),
		height:   len(choices), // will be adjusted on first WindowSizeMsg
	}
}

func (m multiSelectModel) Init() tea.Cmd {
	return nil
}

func (m multiSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		available := msg.Height - headerLines
		if available < minVisible {
			available = minVisible
		}
		if available > len(m.choices) {
			available = len(m.choices)
		}
		m.height = available
		m.adjustScroll()

	case tea.KeyMsg:
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
			if m.cursor < len(m.choices)-1 {
				m.cursor++
				m.adjustScroll()
			}
		case " ":
			m.selected[m.cursor] = !m.selected[m.cursor]
		case "a":
			for i := range m.choices {
				m.selected[i] = true
			}
		case "n":
			m.selected = make(map[int]bool)
		}
	}
	return m, nil
}

func (m *multiSelectModel) adjustScroll() {
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+m.height {
		m.offset = m.cursor - m.height + 1
	}
}

func (m multiSelectModel) View() string {
	var b strings.Builder
	b.WriteString(Bold("Select repositories to clone:") + " (space: select, a: all, n: none, enter: confirm, esc: cancel)\n\n")

	end := m.offset + m.height
	if end > len(m.choices) {
		end = len(m.choices)
	}

	if m.offset > 0 {
		b.WriteString(fmt.Sprintf("  ... (%d more above)\n", m.offset))
	}

	for i := m.offset; i < end; i++ {
		choice := m.choices[i]
		cursor := "  "
		if m.cursor == i {
			cursor = "> "
		}

		if m.selected[i] {
			b.WriteString(fmt.Sprintf("%s%s\n", cursor, Info("[x] "+choice)))
		} else {
			b.WriteString(fmt.Sprintf("%s[ ] %s\n", cursor, choice))
		}
	}

	remaining := len(m.choices) - end
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
