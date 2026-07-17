// Package ui implements the interactive terminal UI for browsing an MTP device.
package ui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/tamcore/mtpx/internal/backend"
	"github.com/tamcore/mtpx/internal/vfs"
)

type mode int

const (
	modeBrowse mode = iota
	modeConfirm
	modeDeleting
)

// column is one Finder-style pane: a directory and the cursor within it.
type column struct {
	dir    string
	cursor int
}

// Model is the Bubble Tea model backing the column browser.
type Model struct {
	backend  backend.Backend
	ctx      context.Context
	destDir  string
	objects  []backend.Object
	cols     []column
	selected map[uint32]bool
	loading  bool
	err      error
	width    int
	height   int
	mode     mode
	message  string
	pending  []backend.Object // files awaiting delete confirmation
	queue    []backend.Object // files being deleted
	done     int
	failed   int
	log      []string
}

// removeObject returns objects without the entry with the given id.
func removeObject(objects []backend.Object, id uint32) []backend.Object {
	out := make([]backend.Object, 0, len(objects))
	for _, o := range objects {
		if o.ID != id {
			out = append(out, o)
		}
	}
	return out
}

// NewModel returns a Model that lists the device on Init. Pulled files are
// written under destDir, preserving their device paths.
func NewModel(ctx context.Context, bk backend.Backend, destDir string) Model {
	return Model{
		backend:  bk,
		ctx:      ctx,
		destDir:  destDir,
		cols:     []column{{dir: ""}},
		selected: map[uint32]bool{},
		loading:  true,
	}
}

// Init starts loading the device listing.
func (m Model) Init() tea.Cmd {
	return m.loadCmd()
}

// entriesOf returns the direct children of a directory.
func (m Model) entriesOf(dir string) []backend.Object {
	return vfs.Children(m.objects, dir)
}

// focused returns the rightmost (active) column.
func (m Model) focused() column {
	return m.cols[len(m.cols)-1]
}

// focusedEntry returns the entry under the cursor of the active column.
func (m Model) focusedEntry() (backend.Object, bool) {
	c := m.focused()
	entries := m.entriesOf(c.dir)
	if c.cursor < 0 || c.cursor >= len(entries) {
		return backend.Object{}, false
	}
	return entries[c.cursor], true
}

// targets returns the files the next action applies to: the selected files, or
// the file under the cursor when nothing is selected.
func (m Model) targets() []backend.Object {
	if len(m.selected) > 0 {
		var out []backend.Object
		for _, o := range m.objects {
			if !o.IsDir && m.selected[o.ID] {
				out = append(out, o)
			}
		}
		return out
	}
	if e, ok := m.focusedEntry(); ok && !e.IsDir {
		return []backend.Object{e}
	}
	return nil
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

// reconcile drops columns whose directory no longer exists and clamps every
// cursor into range, keeping at least the root column. Used after (re)loading.
func (m Model) reconcile() Model {
	cols := []column{m.cols[0]}
	for i := 1; i < len(m.cols); i++ {
		if o, ok := vfs.Find(m.objects, m.cols[i].dir); ok && o.IsDir {
			cols = append(cols, m.cols[i])
		} else {
			break
		}
	}
	for i := range cols {
		n := len(m.entriesOf(cols[i].dir))
		if cols[i].cursor >= n {
			cols[i].cursor = n - 1
		}
		if cols[i].cursor < 0 {
			cols[i].cursor = 0
		}
	}
	m.cols = cols
	return m
}

func setCursor(cols []column, i, cursor int) []column {
	next := append([]column(nil), cols...)
	next[i].cursor = cursor
	return next
}

func pushCol(cols []column, c column) []column {
	return append(append([]column(nil), cols...), c)
}

func popCol(cols []column) []column {
	return append([]column(nil), cols[:len(cols)-1]...)
}

// windowAround returns the [start, end) slice of n items that keeps cursor
// visible in a viewport of rows lines. rows <= 0 or rows >= n shows everything.
func windowAround(cursor, n, rows int) (int, int) {
	if n == 0 {
		return 0, 0
	}
	if rows <= 0 || rows >= n {
		return 0, n
	}
	start := 0
	if cursor >= rows {
		start = cursor - rows + 1
	}
	return start, start + rows
}

func indexOfPath(entries []backend.Object, path string) int {
	for i, e := range entries {
		if e.Path == path {
			return i
		}
	}
	return -1
}
