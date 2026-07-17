package cli

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/tamcore/mtpx/internal/backend"
)

const devicePollInterval = time.Second

// parseWait interprets a --wait value: a bare number is seconds, otherwise it is
// a Go duration string (e.g. "500ms", "2m").
func parseWait(s string) (time.Duration, error) {
	if s == "" {
		return 0, nil
	}
	if n, err := strconv.Atoi(s); err == nil {
		return time.Duration(n) * time.Second, nil
	}
	return time.ParseDuration(s)
}

var errNoDevice = errors.New(
	"no MTP device found; connect the device, put it in MTP mode, and close other MTP programs (retry with --wait)")

// ensureDevice returns nil once at least one device is detected, re-checking
// every interval until timeout elapses. A timeout of zero checks once and fails
// immediately if no device is present.
func ensureDevice(ctx context.Context, bk backend.Backend, timeout, interval time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		devices, err := bk.Detect(ctx)
		if err != nil {
			return err
		}
		if len(devices) > 0 {
			return nil
		}
		if timeout <= 0 || !time.Now().Before(deadline) {
			return errNoDevice
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
		}
	}
}

// requireDevice blocks until a device is available, honoring the --wait flag,
// and returns an error if none appears in time.
func requireDevice(cmd *cobra.Command, bk backend.Backend) error {
	raw, _ := cmd.Flags().GetString("wait")
	wait, err := parseWait(raw)
	if err != nil {
		return fmt.Errorf("invalid --wait %q: %w", raw, err)
	}
	return ensureDevice(cmd.Context(), bk, wait, devicePollInterval)
}
