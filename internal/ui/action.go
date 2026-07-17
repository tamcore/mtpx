package ui

import (
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/tamcore/mtpx/internal/backend"
)

// deleteAtCmd deletes the queued file at index i and reports the result, so the
// caller can stream progress one file at a time.
func (m Model) deleteAtCmd(i int) tea.Cmd {
	bk, ctx, obj := m.backend, m.ctx, m.queue[i]
	return func() tea.Msg {
		return deletedOneMsg{index: i, obj: obj, err: bk.Delete(ctx, obj.ID)}
	}
}

// pullCmd copies every target to destDir, preserving its device path.
func (m Model) pullCmd(targets []backend.Object) tea.Cmd {
	bk, ctx, dest := m.backend, m.ctx, m.destDir
	return func() tea.Msg {
		var failed int
		for _, o := range targets {
			out := filepath.Join(dest, filepath.FromSlash(o.Path))
			if err := os.MkdirAll(filepath.Dir(out), 0o750); err != nil {
				failed++
				continue
			}
			if err := bk.Get(ctx, o.ID, out); err != nil {
				failed++
			}
		}
		return pulledMsg{count: len(targets) - failed, failed: failed}
	}
}
