package backend

import (
	"context"
	"fmt"
	"strconv"
)

const (
	toolDetect   = "mtp-detect"
	toolFolders  = "mtp-folders"
	toolFiles    = "mtp-files"
	toolGetfile  = "mtp-getfile"
	toolSendfile = "mtp-sendfile"
	toolDelfile  = "mtp-delfile"
)

// libmtpTools lists every external tool LibmtpBackend depends on.
var libmtpTools = []string{toolDetect, toolFolders, toolFiles, toolGetfile, toolSendfile, toolDelfile}

// LibmtpBackend is a Backend backed by the libmtp mtp-* command line tools.
type LibmtpBackend struct {
	runner Runner
}

var _ Backend = (*LibmtpBackend)(nil)

// NewLibmtpBackend returns a Backend that shells out to the libmtp tools.
func NewLibmtpBackend() *LibmtpBackend {
	return &LibmtpBackend{runner: execRunner{}}
}

// EnsureAvailable reports an error if any required libmtp tool is missing from
// PATH, with a hint on how to install it.
func (b *LibmtpBackend) EnsureAvailable() error {
	for _, tool := range libmtpTools {
		if _, err := b.runner.LookPath(tool); err != nil {
			return fmt.Errorf("libmtp tool %q not found in PATH; install libmtp "+
				"(macOS: brew install libmtp, Debian/Ubuntu: apt install libmtp-runtime)", tool)
		}
	}
	return nil
}

func (b *LibmtpBackend) Detect(ctx context.Context) ([]Device, error) {
	out, err := b.runner.Run(ctx, toolDetect)
	if err != nil {
		return nil, fmt.Errorf("detect devices: %w", err)
	}
	return parseDetect(out), nil
}

func (b *LibmtpBackend) List(ctx context.Context) ([]Object, error) {
	folderOut, err := b.runner.Run(ctx, toolFolders)
	if err != nil {
		return nil, fmt.Errorf("list folders: %w", err)
	}
	fileOut, err := b.runner.Run(ctx, toolFiles)
	if err != nil {
		return nil, fmt.Errorf("list files: %w", err)
	}
	return buildTree(parseFolders(folderOut), parseFiles(fileOut)), nil
}

func (b *LibmtpBackend) Get(ctx context.Context, id uint32, destPath string) error {
	if _, err := b.runner.Run(ctx, toolGetfile, formatID(id), destPath); err != nil {
		return fmt.Errorf("get object %d: %w", id, err)
	}
	return nil
}

func (b *LibmtpBackend) Put(ctx context.Context, localPath, remotePath string) (uint32, error) {
	if _, err := b.runner.Run(ctx, toolSendfile, localPath, remotePath); err != nil {
		return 0, fmt.Errorf("put %q: %w", localPath, err)
	}
	// mtp-sendfile does not reliably report the new object id; callers refresh
	// the listing to discover it.
	return 0, nil
}

func (b *LibmtpBackend) Delete(ctx context.Context, id uint32) error {
	if _, err := b.runner.Run(ctx, toolDelfile, "-n", formatID(id)); err != nil {
		return fmt.Errorf("delete object %d: %w", id, err)
	}
	return nil
}

func formatID(id uint32) string {
	return strconv.FormatUint(uint64(id), 10)
}
