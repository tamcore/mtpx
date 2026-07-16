package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/tamcore/mtpx/internal/backend"
	"github.com/tamcore/mtpx/internal/vfs"
)

func newDeleteCmd(bk backend.Backend) *cobra.Command {
	var yes, dryRun bool
	cmd := &cobra.Command{
		Use:     "delete <path>...",
		Aliases: []string{"rm"},
		Short:   "Delete one or more files from the device",
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDelete(cmd.Context(), cmd.OutOrStdout(), cmd.InOrStdin(), bk, args, yes, dryRun)
		},
	}
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "skip the confirmation prompt")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "list what would be deleted without deleting")
	return cmd
}

func runDelete(ctx context.Context, w io.Writer, in io.Reader, bk backend.Backend, paths []string, yes, dryRun bool) error {
	objects, err := bk.List(ctx)
	if err != nil {
		return err
	}

	targets := make([]backend.Object, 0, len(paths))
	for _, p := range paths {
		o, ok := vfs.Find(objects, p)
		if !ok {
			return fmt.Errorf("not found on device: %s", p)
		}
		if o.IsDir {
			return fmt.Errorf("refusing to delete a directory: %s", p)
		}
		targets = append(targets, o)
	}

	if dryRun {
		for _, o := range targets {
			fmt.Fprintf(w, "would delete %s\n", o.Path)
		}
		return nil
	}

	if !yes {
		ok, err := confirm(in, w, fmt.Sprintf("Delete %d file(s)?", len(targets)))
		if err != nil {
			return err
		}
		if !ok {
			fmt.Fprintln(w, "aborted")
			return nil
		}
	}

	return deleteObjects(ctx, w, bk, targets)
}

// deleteObjects deletes every target, reporting per-file success or failure, and
// returns an error if any deletion failed. Shared by delete and purge.
func deleteObjects(ctx context.Context, w io.Writer, bk backend.Backend, targets []backend.Object) error {
	var failed int
	for _, o := range targets {
		if err := bk.Delete(ctx, o.ID); err != nil {
			fmt.Fprintf(w, "failed to delete %s: %v\n", o.Path, err)
			failed++
			continue
		}
		fmt.Fprintf(w, "deleted %s\n", o.Path)
	}
	if failed > 0 {
		return fmt.Errorf("%d of %d deletions failed", failed, len(targets))
	}
	return nil
}
