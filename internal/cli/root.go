// Package cli wires the mtpx command-line interface.
package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tamcore/mtpx/internal/backend"
)

// Launcher starts the interactive TUI. It is injected so tests can stub it and
// so the cli package stays decoupled from the ui package.
type Launcher func(ctx context.Context, bk backend.Backend, destDir string) error

// NewRootCmd builds the mtpx command tree. Running it with no subcommand invokes
// launch (the interactive browser).
func NewRootCmd(version, commit string, bk backend.Backend, launch Launcher) *cobra.Command {
	var dest string
	root := &cobra.Command{
		Use:          "mtpx",
		Short:        "Browse and manage files on an MTP device",
		Long:         "mtpx browses and manages files on MTP devices such as Garmin watches, over USB via libmtp.",
		Version:      fmt.Sprintf("%s (%s)", version, commit),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return launch(cmd.Context(), bk, dest)
		},
	}
	root.PersistentFlags().StringVar(&dest, "dest", ".",
		"destination directory for files pulled in the TUI")
	root.AddCommand(
		newListCmd(bk),
		newPullCmd(bk),
		newDeleteCmd(bk),
		newPurgeCmd(bk),
	)
	return root
}

// Execute builds the root command with the real libmtp backend and the given
// TUI launcher, then runs it with args.
func Execute(version, commit string, args []string, launch Launcher) error {
	root := NewRootCmd(version, commit, backend.NewLibmtpBackend(), launch)
	root.SetArgs(args)
	return root.Execute()
}
