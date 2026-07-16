package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/tamcore/mtpx/internal/backend"
	"github.com/tamcore/mtpx/internal/vfs"
)

func newListCmd(bk backend.Backend) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:   "list [path]",
		Short: "List the contents of a directory on the device",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireDevice(cmd, bk); err != nil {
				return err
			}
			var path string
			if len(args) == 1 {
				path = args[0]
			}
			return runList(cmd.Context(), cmd.OutOrStdout(), bk, path, asJSON)
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "output as JSON")
	return cmd
}

func runList(ctx context.Context, w io.Writer, bk backend.Backend, path string, asJSON bool) error {
	objects, err := bk.List(ctx)
	if err != nil {
		return err
	}
	if path != "" {
		o, ok := vfs.Find(objects, path)
		if !ok || !o.IsDir {
			return fmt.Errorf("not a directory: %s", path)
		}
	}

	children := vfs.Children(objects, path)
	if asJSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(children)
	}
	for _, c := range children {
		var line string
		if c.IsDir {
			line = fmt.Sprintf("%-4s  %12s  %s/", "dir", "-", c.Name)
		} else {
			line = fmt.Sprintf("%-4s  %12d  %s", "file", c.Size, c.Name)
		}
		if _, err := io.WriteString(w, line+"\n"); err != nil {
			return err
		}
	}
	return nil
}
