package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// headerRows is the mouse-Y offset of the first pane row. The view prints a
// title and a blank line above the panes, but clicks land one row higher than
// that implies, so the effective offset is one.
const headerRows = 1

// handleMouse handles wheel scrolling and left clicks while browsing.
func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if m.mode != modeBrowse || msg.Action != tea.MouseActionPress {
		return m, nil
	}
	last := len(m.cols) - 1
	switch msg.Button {
	case tea.MouseButtonWheelUp:
		if m.cols[last].cursor > 0 {
			m.cols = setCursor(m.cols, last, m.cols[last].cursor-1)
		}
	case tea.MouseButtonWheelDown:
		if m.cols[last].cursor < len(m.focusedEntries())-1 {
			m.cols = setCursor(m.cols, last, m.cols[last].cursor+1)
		}
	case tea.MouseButtonLeft:
		m = m.handleClick(msg.X, msg.Y)
	}
	return m, nil
}

// handleClick focuses the clicked pane and entry. Clicking the preview pane
// descends into the focused folder and selects the clicked child.
func (m Model) handleClick(x, y int) Model {
	ri, ei, ok := m.clickTarget(x, y)
	if !ok {
		return m
	}
	if ri < len(m.cols) {
		next := append([]column(nil), m.cols[:ri+1]...)
		next[ri].cursor = ei
		m.cols = next
		m.filter = ""
		return m
	}
	// ri == len(m.cols): the preview pane, which only renders when the focused
	// entry is a folder, so descending into it is always valid.
	e, _ := m.focusedEntry()
	m.cols = pushCol(m.cols, column{dir: e.Path})
	m.cols = setCursor(m.cols, len(m.cols)-1, ei)
	m.filter = ""
	return m
}

// clickTarget maps a screen position to a rendered pane index and entry index.
func (m Model) clickTarget(x, y int) (ri, ei int, ok bool) {
	row := y - headerRows
	rows := m.rows()
	if row < 0 || row >= rows {
		return 0, 0, false
	}
	rcs := m.renderColumns()
	visible := visibleColumns(rcs, m.width)
	vc := x / colWidth
	if vc < 0 || vc >= len(visible) {
		return 0, 0, false
	}
	ri = len(rcs) - len(visible) + vc
	rc := rcs[ri]
	if len(rc.entries) == 0 {
		return 0, 0, false
	}
	start, _ := windowAround(rc.cursor, len(rc.entries), rows)
	if ei = start + row; ei >= len(rc.entries) {
		return 0, 0, false
	}
	return ri, ei, true
}
