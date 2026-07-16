package ui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/tamcore/mtpx/internal/backend"
)

// Run starts the interactive browser program and blocks until it exits. Callers
// pass tea options such as tea.WithAltScreen() for a real terminal, or
// tea.WithInput/tea.WithOutput in tests.
func Run(ctx context.Context, bk backend.Backend, destDir string, opts ...tea.ProgramOption) error {
	_, err := tea.NewProgram(NewModel(ctx, bk, destDir), opts...).Run()
	return err
}
