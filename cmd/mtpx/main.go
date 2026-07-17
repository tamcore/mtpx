package main

import (
	"context"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/tamcore/mtpx/internal/backend"
	"github.com/tamcore/mtpx/internal/cli"
	"github.com/tamcore/mtpx/internal/ui"
)

var (
	version = "dev"
	commit  = "none"
)

func main() {
	launch := func(ctx context.Context, bk backend.Backend, dest string) error {
		return ui.Run(ctx, bk, dest, tea.WithAltScreen(), tea.WithMouseCellMotion())
	}
	if err := cli.Execute(version, commit, os.Args[1:], launch); err != nil {
		os.Exit(1)
	}
}
