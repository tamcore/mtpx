package ui

import (
	"fmt"
	"strings"

	"github.com/tamcore/mtpx/internal/backend"
)

const (
	colWidth     = 22
	reservedRows = 4
	defaultRows  = 12
)

// renderCol is one pane prepared for display.
type renderCol struct {
	entries []backend.Object
	cursor  int // highlighted index, -1 for none (preview pane)
	active  bool
}

// View renders the current model state as Finder-style columns.
func (m Model) View() string {
	var b strings.Builder
	loc := m.focused().dir
	if loc == "" {
		loc = "/"
	}
	if m.mode == modeSearch || m.filter != "" {
		fmt.Fprintf(&b, "mtpx — %s   [search: %s]\n\n", loc, m.filter)
	} else {
		fmt.Fprintf(&b, "mtpx — %s\n\n", loc)
	}

	if m.loading {
		b.WriteString("loading…\n")
		return b.String()
	}
	if m.err != nil {
		fmt.Fprintf(&b, "error: %v\n", m.err)
		return b.String()
	}

	switch m.mode {
	case modeConfirm:
		b.WriteString(m.confirmView())
	case modeDeleting:
		b.WriteString(m.deletingView())
	default:
		b.WriteString(m.columnsView())
		fmt.Fprintf(&b, "\n%d selected · ↑/↓ move · →/enter open · ←/⌫ back · space select · s search · c copy · d delete · r refresh · q quit\n", len(m.selected))
		if m.message != "" {
			fmt.Fprintf(&b, "%s\n", m.message)
		}
	}
	return b.String()
}

// confirmView lists the files queued for deletion and the prompt.
func (m Model) confirmView() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Delete %d file(s)?\n\n", len(m.pending))
	limit := m.rows()
	for i, o := range m.pending {
		if i >= limit {
			fmt.Fprintf(&b, "  … and %d more\n", len(m.pending)-limit)
			break
		}
		fmt.Fprintf(&b, "  %s\n", o.Path)
	}
	b.WriteString("\n[y] delete   [n]/esc cancel\n")
	return b.String()
}

// deletingView shows delete progress and the most recent per-file results.
func (m Model) deletingView() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Deleting… %d/%d\n\n", m.done+m.failed, len(m.queue))
	limit := m.rows()
	start := 0
	if len(m.log) > limit {
		start = len(m.log) - limit
	}
	for _, line := range m.log[start:] {
		fmt.Fprintf(&b, "  %s\n", line)
	}
	return b.String()
}

// rows is the number of list lines each column shows.
func (m Model) rows() int {
	if m.height <= 0 {
		return defaultRows
	}
	if r := m.height - reservedRows; r > 0 {
		return r
	}
	return 1
}

// renderColumns builds one pane per open directory, plus a preview pane for the
// highlighted folder.
func (m Model) renderColumns() []renderCol {
	var rcs []renderCol
	for i, c := range m.cols {
		if i == len(m.cols)-1 {
			rcs = append(rcs, renderCol{entries: m.focusedEntries(), cursor: c.cursor, active: true})
			continue
		}
		entries := m.entriesOf(c.dir)
		rcs = append(rcs, renderCol{entries: entries, cursor: indexOfPath(entries, m.cols[i+1].dir)})
	}
	if e, ok := m.focusedEntry(); ok && e.IsDir {
		rcs = append(rcs, renderCol{entries: m.entriesOf(e.Path), cursor: -1})
	}
	return rcs
}

// visibleColumns keeps the rightmost panes that fit the terminal width.
func visibleColumns(rcs []renderCol, width int) []renderCol {
	if width <= 0 {
		return rcs
	}
	k := width / colWidth
	if k < 1 {
		k = 1
	}
	if len(rcs) <= k {
		return rcs
	}
	return rcs[len(rcs)-k:]
}

func (m Model) columnsView() string {
	rcs := visibleColumns(m.renderColumns(), m.width)
	rows := m.rows()
	cells := make([][]string, len(rcs))
	for i, rc := range rcs {
		cells[i] = m.columnLines(rc, rows)
	}
	var b strings.Builder
	for r := 0; r < rows; r++ {
		for _, col := range cells {
			b.WriteString(col[r])
		}
		b.WriteString("\n")
	}
	return b.String()
}

func (m Model) columnLines(rc renderCol, rows int) []string {
	lines := make([]string, rows)
	for i := range lines {
		lines[i] = pad("", colWidth)
	}
	if len(rc.entries) == 0 {
		lines[0] = pad("  (empty)", colWidth)
		return lines
	}
	start, end := windowAround(rc.cursor, len(rc.entries), rows)
	r := 0
	for i := start; i < end; i++ {
		e := rc.entries[i]
		point := " "
		if i == rc.cursor {
			if rc.active {
				point = ">"
			} else {
				point = "·"
			}
		}
		sel := " "
		if m.selected[e.ID] {
			sel = "x"
		}
		name := e.Name
		if e.IsDir {
			name += "/"
		}
		lines[r] = pad(fmt.Sprintf("%s%s %s", point, sel, name), colWidth)
		r++
	}
	return lines
}

// pad truncates or space-pads s to exactly w runes.
func pad(s string, w int) string {
	r := []rune(s)
	if len(r) >= w {
		return string(r[:w-1]) + " "
	}
	return s + strings.Repeat(" ", w-len(r))
}
