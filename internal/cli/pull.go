package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/tamcore/mtpx/internal/backend"
	"github.com/tamcore/mtpx/internal/vfs"
)

func newPullCmd(bk backend.Backend) *cobra.Command {
	return &cobra.Command{
		Use:   "pull <remote> <local>",
		Short: "Copy a file from the device to local disk",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireDevice(cmd, bk); err != nil {
				return err
			}
			return runPull(cmd.Context(), cmd.OutOrStdout(), bk, args[0], args[1])
		},
	}
}

func runPull(ctx context.Context, w io.Writer, bk backend.Backend, remote, local string) error {
	objects, err := bk.List(ctx)
	if err != nil {
		return err
	}
	o, ok := vfs.Find(objects, remote)
	if !ok {
		return fmt.Errorf("not found on device: %s", remote)
	}
	if o.IsDir {
		return fmt.Errorf("is a directory: %s", remote)
	}

	dest := local
	if info, statErr := os.Stat(local); statErr == nil && info.IsDir() {
		dest = filepath.Join(local, o.Name)
	}
	if err := bk.Get(ctx, o.ID, dest); err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "pulled %s -> %s\n", remote, dest)
	return err
}
