package ui

import (
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/tamcore/mtpx/internal/backend"
)

// deleteCmd removes every target from the device in the background.
func (m Model) deleteCmd(targets []backend.Object) tea.Cmd {
	bk, ctx := m.backend, m.ctx
	return func() tea.Msg {
		var failed int
		for _, o := range targets {
			if err := bk.Delete(ctx, o.ID); err != nil {
				failed++
			}
		}
		return deletedMsg{count: len(targets) - failed, failed: failed}
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
