// Package cli wires the mtpx command-line interface.
package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tamcore/mtpx/internal/backend"
)

// NewRootCmd builds the mtpx command tree using the given backend.
func NewRootCmd(version, commit string, bk backend.Backend) *cobra.Command {
	root := &cobra.Command{
		Use:          "mtpx",
		Short:        "Browse and manage files on an MTP device",
		Long:         "mtpx browses and manages files on MTP devices such as Garmin watches, over USB via libmtp.",
		Version:      fmt.Sprintf("%s (%s)", version, commit),
		SilenceUsage: true,
	}
	root.AddCommand(
		newListCmd(bk),
		newPullCmd(bk),
		newPushCmd(bk),
	)
	return root
}

// Execute builds the root command with the real libmtp backend and runs it with
// the given arguments.
func Execute(version, commit string, args []string) error {
	root := NewRootCmd(version, commit, backend.NewLibmtpBackend())
	root.SetArgs(args)
	return root.Execute()
}
