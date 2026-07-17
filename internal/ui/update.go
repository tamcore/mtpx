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
	case deletedOneMsg:
		return m.handleDeleted(msg)
	case pulledMsg:
		if msg.failed > 0 {
			m.message = fmt.Sprintf("pulled %d, %d failed", msg.count, msg.failed)
		} else {
			m.message = fmt.Sprintf("pulled %d file(s) to %s", msg.count, m.destDir)
		}
		return m, nil
	case tea.KeyMsg:
		switch m.mode {
		case modeConfirm:
			return m.handleConfirm(msg)
		case modeDeleting:
			return m.handleDeleting(msg)
		default:
			return m.handleKey(msg)
		}
	}
	return m, nil
}

// handleDeleted records one file's result and either continues the queue or
// finishes, dropping deleted files from the cached listing so no re-scan runs.
func (m Model) handleDeleted(msg deletedOneMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.failed++
		m.log = append(m.log, "failed  "+msg.obj.Path+": "+msg.err.Error())
	} else {
		m.done++
		m.log = append(m.log, "deleted "+msg.obj.Path)
		m.objects = removeObject(m.objects, msg.obj.ID)
	}
	if next := msg.index + 1; next < len(m.queue) {
		return m, m.deleteAtCmd(next)
	}
	m.mode = modeBrowse
	m.selected = map[uint32]bool{}
	if m.failed > 0 {
		m.message = fmt.Sprintf("deleted %d, %d failed", m.done, m.failed)
	} else {
		m.message = fmt.Sprintf("deleted %d file(s)", m.done)
	}
	m.queue = nil
	return m.reconcile(), nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		m.mode = modeDeleting
		m.queue = m.pending
		m.pending = nil
		m.done, m.failed = 0, 0
		m.log = nil
		m.message = ""
		return m, m.deleteAtCmd(0)
	case "ctrl+c":
		return m, tea.Quit
	case "n", "N", "esc":
		m.mode = modeBrowse
		m.pending = nil
		m.message = "cancelled"
	}
	return m, nil
}

func (m Model) handleDeleting(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}
	return m, nil
}
