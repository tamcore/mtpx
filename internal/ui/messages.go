package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/tamcore/mtpx/internal/backend"
)

// objectsMsg carries the device listing loaded by loadCmd.
type objectsMsg struct{ objects []backend.Object }

// errMsg carries an error from an asynchronous command.
type errMsg struct{ err error }

// deletedMsg reports how many files a delete action removed and how many failed.
type deletedMsg struct{ count, failed int }

// pulledMsg reports how many files a pull action copied and how many failed.
type pulledMsg struct{ count, failed int }

// loadCmd lists the device in the background.
func (m Model) loadCmd() tea.Cmd {
	return func() tea.Msg {
		objs, err := m.backend.List(m.ctx)
		if err != nil {
			return errMsg{err}
		}
		return objectsMsg{objs}
	}
}
