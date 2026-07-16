package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Update handles messages and returns the next model state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case objectsMsg:
		m.objects = msg.objects
		m.loading = false
		m.err = nil
		return m.reconcile(), nil
	case errMsg:
		m.loading = false
		m.err = msg.err
		return m, nil
	case deletedMsg:
		m.selected = map[uint32]bool{}
		if msg.failed > 0 {
			m.message = fmt.Sprintf("deleted %d, %d failed", msg.count, msg.failed)
		} else {
			m.message = fmt.Sprintf("deleted %d file(s)", msg.count)
		}
		m.loading = true
		return m, m.loadCmd()
	case pulledMsg:
		if msg.failed > 0 {
			m.message = fmt.Sprintf("pulled %d, %d failed", msg.count, msg.failed)
		} else {
			m.message = fmt.Sprintf("pulled %d file(s) to %s", msg.count, m.destDir)
		}
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.mode == modeConfirm {
		return m.handleConfirm(msg)
	}
	last := len(m.cols) - 1
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		if m.cols[last].cursor > 0 {
			m.cols = setCursor(m.cols, last, m.cols[last].cursor-1)
		}
	case "down", "j":
		if m.cols[last].cursor < len(m.entriesOf(m.cols[last].dir))-1 {
			m.cols = setCursor(m.cols, last, m.cols[last].cursor+1)
		}
	case "enter", "right", "l":
		if e, ok := m.focusedEntry(); ok && e.IsDir {
			m.cols = pushCol(m.cols, column{dir: e.Path})
		}
	case "backspace", "left", "h":
		if len(m.cols) > 1 {
			m.cols = popCol(m.cols)
		}
	case " ":
		if e, ok := m.focusedEntry(); ok && !e.IsDir {
			m = m.toggle(e.ID)
		}
	case "r":
		m.loading = true
		return m, m.loadCmd()
	case "d":
		targets := m.targets()
		if len(targets) == 0 {
			m.message = "no files selected"
			return m, nil
		}
		m.mode = modeConfirm
		m.pending = targets
		m.message = ""
	case "c":
		targets := m.targets()
		if len(targets) == 0 {
			m.message = "no files selected"
			return m, nil
		}
		m.message = "pulling…"
		return m, m.pullCmd(targets)
	}
	return m, nil
}

func (m Model) handleConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		targets := m.pending
		m.mode = modeBrowse
		m.pending = nil
		m.message = ""
		m.loading = true
		return m, m.deleteCmd(targets)
	case "ctrl+c":
		return m, tea.Quit
	case "n", "N", "esc":
		m.mode = modeBrowse
		m.pending = nil
		m.message = "cancelled"
	}
	return m, nil
}
