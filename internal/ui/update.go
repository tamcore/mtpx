package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles messages and returns the next model state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		return m, nil
	case objectsMsg:
		m.objects = msg.objects
		m.loading = false
		m.err = nil
		return m.clampCursor(), nil
	case errMsg:
		m.loading = false
		m.err = msg.err
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.entries())-1 {
			m.cursor++
		}
	case "enter", "right", "l":
		if e, ok := m.current(); ok && e.IsDir {
			m.cwd = e.Path
			m.cursor = 0
		}
	case "backspace", "left", "h":
		if m.cwd != "" {
			m.cwd = parentDir(m.cwd)
			m.cursor = 0
		}
	case " ":
		if e, ok := m.current(); ok && !e.IsDir {
			m = m.toggle(e.ID)
		}
	case "r":
		m.loading = true
		return m, m.loadCmd()
	}
	return m, nil
}
