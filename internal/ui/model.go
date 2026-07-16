// Package ui implements the interactive terminal UI for browsing an MTP device.
package ui

import (
	"context"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/tamcore/mtpx/internal/backend"
	"github.com/tamcore/mtpx/internal/vfs"
)

// Model is the Bubble Tea model backing the browser.
type Model struct {
	backend  backend.Backend
	ctx      context.Context
	objects  []backend.Object
	cwd      string
	cursor   int
	selected map[uint32]bool
	loading  bool
	err      error
	height   int
}

// NewModel returns a Model that lists the device on Init.
func NewModel(ctx context.Context, bk backend.Backend) Model {
	return Model{
		backend:  bk,
		ctx:      ctx,
		selected: map[uint32]bool{},
		loading:  true,
	}
}

// Init starts loading the device listing.
func (m Model) Init() tea.Cmd {
	return m.loadCmd()
}

// entries returns the direct children of the current directory.
func (m Model) entries() []backend.Object {
	return vfs.Children(m.objects, m.cwd)
}

// current returns the entry under the cursor.
func (m Model) current() (backend.Object, bool) {
	entries := m.entries()
	if m.cursor < 0 || m.cursor >= len(entries) {
		return backend.Object{}, false
	}
	return entries[m.cursor], true
}

// toggle returns a copy of the model with id's selection flipped.
func (m Model) toggle(id uint32) Model {
	next := make(map[uint32]bool, len(m.selected))
	for k, v := range m.selected {
		next[k] = v
	}
	if next[id] {
		delete(next, id)
	} else {
		next[id] = true
	}
	m.selected = next
	return m
}

// clampCursor keeps the cursor within the current entry list.
func (m Model) clampCursor() Model {
	n := len(m.entries())
	if m.cursor >= n {
		m.cursor = n - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	return m
}

// visibleRange returns the [start, end) slice of entries to render so the cursor
// stays on screen for a list of n entries. Height 0 (size unknown) shows all.
func (m Model) visibleRange(n int) (int, int) {
	if n == 0 {
		return 0, 0
	}
	rows := m.height - reservedRows
	if m.height == 0 || rows >= n {
		return 0, n
	}
	if rows < 1 {
		rows = 1
	}
	start := 0
	if m.cursor >= rows {
		start = m.cursor - rows + 1
	}
	return start, start + rows
}

func parentDir(p string) string {
	if i := strings.LastIndex(p, "/"); i >= 0 {
		return p[:i]
	}
	return ""
}
