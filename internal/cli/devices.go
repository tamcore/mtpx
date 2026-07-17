package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tamcore/mtpx/internal/backend"
)

func newDevicesCmd(bk backend.Backend) *cobra.Command {
	var asJSON bool
	cmd := &cobra.Command{
		Use:     "devices",
		Aliases: []string{"device"},
		Short:   "List attached MTP devices",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if cmd.Flags().Changed("wait") {
				raw, _ := cmd.Flags().GetString("wait")
				if wait, err := parseWait(raw); err == nil && wait > 0 {
					_ = ensureDevice(cmd.Context(), bk, wait, devicePollInterval)
				}
			}
			return runDevices(cmd.Context(), cmd.OutOrStdout(), bk, asJSON)
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "output as JSON")
	return cmd
}

func runDevices(ctx context.Context, w io.Writer, bk backend.Backend, asJSON bool) error {
	devices, err := bk.Detect(ctx)
	if err != nil {
		return err
	}
	if asJSON {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(devices)
	}
	if len(devices) == 0 {
		fmt.Fprintln(w, "no MTP devices found")
		return nil
	}
	for _, d := range devices {
		name := strings.TrimSpace(d.Vendor + " " + d.Product)
		if name == "" {
			name = "unknown device"
		}
		fmt.Fprintf(w, "%s (%04x:%04x) bus %d dev %d", name, d.VendorID, d.ProductID, d.Bus, d.Dev)
		if d.Serial != "" {
			fmt.Fprintf(w, " SN:%s", d.Serial)
		}
		fmt.Fprintln(w)
	}
	return nil
}
