package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func mouseModel(t *testing.T, dirs ...string) Model {
	t.Helper()
	m := atPath(t, dirs...)
	m.width = 200
	m.height = 30
	return m
}

func click(x, y int) tea.MouseMsg {
	return tea.MouseMsg{Action: tea.MouseActionPress, Button: tea.MouseButtonLeft, X: x, Y: y}
}

func wheel(up bool) tea.MouseMsg {
	b := tea.MouseButtonWheelDown
	if up {
		b = tea.MouseButtonWheelUp
	}
	return tea.MouseMsg{Action: tea.MouseActionPress, Button: b}
}

func TestMouseWheel(t *testing.T) {
	m := mouseModel(t) // root: GARMIN, Music
	m, _ = update(t, m, wheel(false))
	if m.focused().cursor != 1 {
		t.Fatalf("wheel down cursor = %d", m.focused().cursor)
	}
	m, _ = update(t, m, wheel(false)) // at last, no move
	if m.focused().cursor != 1 {
		t.Fatalf("wheel down should clamp, got %d", m.focused().cursor)
	}
	m, _ = update(t, m, wheel(true))
	m, _ = update(t, m, wheel(true)) // at first, no move
	if m.focused().cursor != 0 {
		t.Fatalf("wheel up should clamp, got %d", m.focused().cursor)
	}
}

func TestMouseIgnoredInDialog(t *testing.T) {
	m := mouseModel(t)
	m.mode = modeConfirm
	m, _ = update(t, m, wheel(false))
	if m.focused().cursor != 0 {
		t.Fatal("mouse should be ignored outside browse mode")
	}
}

func TestMouseIgnoresNonPress(t *testing.T) {
	m := mouseModel(t)
	m, _ = update(t, m, tea.MouseMsg{Action: tea.MouseActionRelease, Button: tea.MouseButtonLeft, X: 5, Y: 3})
	if m.focused().cursor != 0 {
		t.Fatal("non-press mouse events should be ignored")
	}
}

func TestMouseClickActiveColumn(t *testing.T) {
	m := mouseModel(t)               // root entries: GARMIN (0), Music (1)
	m, _ = update(t, m, click(5, 3)) // column 0, row 1 -> Music
	if m.focused().cursor != 1 {
		t.Fatalf("click should move cursor to row 1, got %d", m.focused().cursor)
	}
}

func TestMouseClickPreviewDescends(t *testing.T) {
	m := mouseModel(t)                // root focused on GARMIN; preview shows GARMIN's children
	m, _ = update(t, m, click(25, 2)) // preview column, row 0 -> Activity
	if len(m.cols) != 2 || m.focused().dir != "GARMIN" || m.focused().cursor != 0 {
		t.Fatalf("preview click should descend: cols=%d dir=%q", len(m.cols), m.focused().dir)
	}
}

func TestMouseClickTrailColumnCollapses(t *testing.T) {
	m := mouseModel(t, "GARMIN")     // cols: root, GARMIN
	m, _ = update(t, m, click(5, 3)) // click root column (col 0), row 1 -> Music
	if len(m.cols) != 1 || m.focused().cursor != 1 {
		t.Fatalf("clicking a trail column should collapse deeper panes: cols=%d cursor=%d", len(m.cols), m.focused().cursor)
	}
}

func TestMouseClickHeaderIgnored(t *testing.T) {
	m := mouseModel(t)
	m, _ = update(t, m, click(5, 0)) // header row
	if m.focused().cursor != 0 {
		t.Fatal("clicking the header should do nothing")
	}
}

func TestMouseClickBelowEntriesIgnored(t *testing.T) {
	m := mouseModel(t)                // only 2 root entries
	m, _ = update(t, m, click(5, 10)) // row 8, past the entries
	if m.focused().cursor != 0 {
		t.Fatal("clicking below the entries should do nothing")
	}
}

func TestMouseClickFooterIgnored(t *testing.T) {
	m := mouseModel(t) // height 30 -> 26 list rows, so row 27 is below them
	m, _ = update(t, m, click(5, 29))
	if m.focused().cursor != 0 {
		t.Fatal("clicking below the list viewport should do nothing")
	}
}

func TestMouseClickOutsideColumnsIgnored(t *testing.T) {
	m := mouseModel(t)
	m, _ = update(t, m, click(500, 3)) // far right of any column
	if m.focused().cursor != 0 {
		t.Fatal("clicking past the columns should do nothing")
	}
}

func TestMouseClickEmptyColumnIgnored(t *testing.T) {
	m := mouseModel(t, "Music")       // Music has no children
	m, _ = update(t, m, click(25, 2)) // the (empty) active Music column
	if m.focused().cursor != 0 || len(m.cols) != 2 {
		t.Fatal("clicking an empty column should do nothing")
	}
}
