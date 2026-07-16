package cli

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/tamcore/mtpx/internal/backend"
)

func newPushCmd(bk backend.Backend) *cobra.Command {
	return &cobra.Command{
		Use:   "push <local> <remote>",
		Short: "Copy a local file to the device",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPush(cmd.Context(), cmd.OutOrStdout(), bk, args[0], args[1])
		},
	}
}

func runPush(ctx context.Context, w io.Writer, bk backend.Backend, local, remote string) error {
	info, err := os.Stat(local)
	if err != nil {
		return fmt.Errorf("local file: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("not a file: %s", local)
	}
	if _, err := bk.Put(ctx, local, remote); err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "pushed %s -> %s\n", local, remote)
	return err
}
