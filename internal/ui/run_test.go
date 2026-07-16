package ui

import (
	"context"
	"io"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/tamcore/mtpx/internal/backend"
)

func TestRun(t *testing.T) {
	bk := &backend.FakeBackend{Objects: sampleObjects()}
	err := Run(context.Background(), bk, t.TempDir(),
		tea.WithInput(strings.NewReader("q")),
		tea.WithOutput(io.Discard))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
}
