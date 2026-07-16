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

// defaultPurgeFolders are the Garmin folders cleared when no --folders are given.
var defaultPurgeFolders = []string{"Activity", "Workouts", "Courses", "PaceBands"}

func newPurgeCmd(bk backend.Backend) *cobra.Command {
	var folders []string
	var backupDir string
	var dryRun, yes bool

	cmd := &cobra.Command{
		Use:   "purge",
		Short: "Delete all files in the target folders (default: Activity, Workouts, Courses, PaceBands)",
		Long: "purge deletes every file (never a folder) inside the target folders. " +
			"By default it targets the Garmin Activity, Workouts, Courses and PaceBands folders. " +
			"Use --backup to copy the files to local disk before deleting them.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireDevice(cmd, bk); err != nil {
				return err
			}
			return runPurge(cmd.Context(), cmd.OutOrStdout(), cmd.InOrStdin(), bk,
				folders, backupDir, dryRun, yes)
		},
	}
	cmd.Flags().StringSliceVar(&folders, "folders", nil,
		"folder names or paths to purge (default: Activity,Workouts,Courses,PaceBands)")
	cmd.Flags().StringVar(&backupDir, "backup", "",
		"copy files into this local directory before deleting them")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "list what would be deleted without deleting")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "skip the confirmation prompt")
	return cmd
}

func runPurge(ctx context.Context, w io.Writer, in io.Reader, bk backend.Backend,
	folders []string, backupDir string, dryRun, yes bool) error {
	if len(folders) == 0 {
		folders = defaultPurgeFolders
	}

	objects, err := bk.List(ctx)
	if err != nil {
		return err
	}

	files := collectPurgeFiles(w, objects, folders)
	if len(files) == 0 {
		fmt.Fprintln(w, "no files to purge")
		return nil
	}

	if dryRun {
		for _, f := range files {
			fmt.Fprintf(w, "would delete %s\n", f.Path)
		}
		fmt.Fprintf(w, "%d file(s) would be deleted\n", len(files))
		return nil
	}

	if !yes {
		prompt := fmt.Sprintf("Delete %d file(s)?", len(files))
		if backupDir != "" {
			prompt = fmt.Sprintf("Back up to %s and delete %d file(s)?", backupDir, len(files))
		}
		ok, err := confirm(in, w, prompt)
		if err != nil {
			return err
		}
		if !ok {
			fmt.Fprintln(w, "aborted")
			return nil
		}
	}

	if backupDir != "" {
		if err := backupFiles(ctx, w, bk, files, backupDir); err != nil {
			return fmt.Errorf("backup failed, nothing deleted: %w", err)
		}
	}

	return deleteObjects(ctx, w, bk, files)
}

// collectPurgeFiles returns the files (deduplicated) under every directory that
// matches one of the requested folders by full path or by name. Missing folders
// produce a warning.
func collectPurgeFiles(w io.Writer, objects []backend.Object, folders []string) []backend.Object {
	seen := make(map[uint32]bool)
	var files []backend.Object
	for _, name := range folders {
		var matched bool
		for _, o := range objects {
			if !o.IsDir || (o.Path != name && o.Name != name) {
				continue
			}
			matched = true
			for _, f := range vfs.FilesUnder(objects, o.Path) {
				if !seen[f.ID] {
					seen[f.ID] = true
					files = append(files, f)
				}
			}
		}
		if !matched {
			fmt.Fprintf(w, "warning: folder %q not found on device\n", name)
		}
	}
	return files
}

// backupFiles copies every file to backupDir, preserving its device path. It
// stops at the first error so the caller can abort before deleting anything.
func backupFiles(ctx context.Context, w io.Writer, bk backend.Backend, files []backend.Object, backupDir string) error {
	for _, f := range files {
		dest := filepath.Join(backupDir, filepath.FromSlash(f.Path))
		if err := os.MkdirAll(filepath.Dir(dest), 0o750); err != nil {
			return fmt.Errorf("create backup dir for %s: %w", f.Path, err)
		}
		if err := bk.Get(ctx, f.ID, dest); err != nil {
			return fmt.Errorf("back up %s: %w", f.Path, err)
		}
		fmt.Fprintf(w, "backed up %s\n", f.Path)
	}
	return nil
}
