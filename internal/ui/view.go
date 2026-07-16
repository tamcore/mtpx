package ui

import (
	"fmt"
	"strings"
)

// reservedRows is the number of lines the header and footer occupy, subtracted
// from the terminal height when deciding how many entries fit.
const reservedRows = 4

// View renders the current model state.
func (m Model) View() string {
	var b strings.Builder
	loc := m.cwd
	if loc == "" {
		loc = "/"
	}
	fmt.Fprintf(&b, "mtpx — %s\n\n", loc)

	if m.loading {
		b.WriteString("loading…\n")
		return b.String()
	}
	if m.err != nil {
		fmt.Fprintf(&b, "error: %v\n", m.err)
		return b.String()
	}

	entries := m.entries()
	if len(entries) == 0 {
		b.WriteString("  (empty)\n")
	}
	start, end := m.visibleRange(len(entries))
	for i := start; i < end; i++ {
		e := entries[i]
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}
		mark := " "
		if m.selected[e.ID] {
			mark = "x"
		}
		name := e.Name
		if e.IsDir {
			name += "/"
		}
		fmt.Fprintf(&b, "%s[%s] %s\n", cursor, mark, name)
	}

	fmt.Fprintf(&b, "\n%d selected · ↑/↓ move · enter open · ⌫ up · space select · r refresh · q quit\n", len(m.selected))
	return b.String()
}
