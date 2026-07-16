package cli

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/tamcore/mtpx/internal/backend"
)

func purgeBackend() *backend.FakeBackend {
	return &backend.FakeBackend{
		Objects: []backend.Object{
			{ID: 1, Path: "GARMIN", Name: "GARMIN", IsDir: true},
			{ID: 2, Path: "GARMIN/Activity", Name: "Activity", IsDir: true},
			{ID: 3, Path: "GARMIN/Activity/a.fit", Name: "a.fit", Size: 3},
			{ID: 4, Path: "GARMIN/Activity/b.fit", Name: "b.fit", Size: 3},
			{ID: 5, Path: "GARMIN/Workouts", Name: "Workouts", IsDir: true},
			{ID: 6, Path: "GARMIN/Workouts/w.fit", Name: "w.fit", Size: 3},
			{ID: 7, Path: "GARMIN/Courses", Name: "Courses", IsDir: true},
			{ID: 8, Path: "GARMIN/PaceBands", Name: "PaceBands", IsDir: true},
			{ID: 9, Path: "GARMIN/gmaptz.img", Name: "gmaptz.img", Size: 3},
		},
		Contents: map[uint32][]byte{3: []byte("aaa"), 4: []byte("bbb"), 6: []byte("www")},
	}
}

func deletedSorted(bk *backend.FakeBackend) []uint32 {
	got := slices.Clone(bk.Deleted)
	slices.Sort(got)
	return got
}

func TestPurgeDefault(t *testing.T) {
	bk := purgeBackend()
	if _, err := runCmd(t, bk, "purge", "--yes"); err != nil {
		t.Fatalf("purge: %v", err)
	}
	if !slices.Equal(deletedSorted(bk), []uint32{3, 4, 6}) {
		t.Fatalf("deleted = %v, want [3 4 6] (folders and gmaptz.img must survive)", deletedSorted(bk))
	}
}

func TestPurgeCustomFolders(t *testing.T) {
	bk := purgeBackend()
	if _, err := runCmd(t, bk, "purge", "-y", "--folders", "Activity"); err != nil {
		t.Fatalf("purge: %v", err)
	}
	if !slices.Equal(deletedSorted(bk), []uint32{3, 4}) {
		t.Fatalf("deleted = %v, want [3 4]", deletedSorted(bk))
	}
}

func TestPurgeDedup(t *testing.T) {
	bk := purgeBackend()
	// GARMIN matches everything under it; Activity matches a subset already seen.
	if _, err := runCmd(t, bk, "purge", "-y", "--folders", "GARMIN,Activity"); err != nil {
		t.Fatalf("purge: %v", err)
	}
	if !slices.Equal(deletedSorted(bk), []uint32{3, 4, 6, 9}) {
		t.Fatalf("deleted = %v, want [3 4 6 9] (deduplicated)", deletedSorted(bk))
	}
}

func TestPurgeDryRun(t *testing.T) {
	bk := purgeBackend()
	out, err := runCmd(t, bk, "purge", "--dry-run")
	if err != nil {
		t.Fatalf("purge: %v", err)
	}
	if !strings.Contains(out, "would delete") || !strings.Contains(out, "3 file(s) would be deleted") {
		t.Errorf("output = %q", out)
	}
	if len(bk.Deleted) != 0 {
		t.Errorf("dry run deleted %v", bk.Deleted)
	}
}

func TestPurgeBackup(t *testing.T) {
	bk := purgeBackend()
	dir := t.TempDir()
	if _, err := runCmd(t, bk, "purge", "-y", "--backup", dir); err != nil {
		t.Fatalf("purge: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "GARMIN", "Activity", "a.fit"))
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if string(b) != "aaa" {
		t.Errorf("backup content = %q", b)
	}
	if !slices.Equal(deletedSorted(bk), []uint32{3, 4, 6}) {
		t.Errorf("deleted = %v", deletedSorted(bk))
	}
}

func TestPurgeBackupGetError(t *testing.T) {
	bk := purgeBackend()
	bk.GetErr = errors.New("read fail")
	if _, err := runCmd(t, bk, "purge", "-y", "--backup", t.TempDir()); err == nil {
		t.Fatal("want backup error")
	}
	if len(bk.Deleted) != 0 {
		t.Errorf("nothing should be deleted after backup failure, got %v", bk.Deleted)
	}
}

func TestPurgeBackupMkdirError(t *testing.T) {
	bk := purgeBackend()
	blocker := filepath.Join(t.TempDir(), "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	// blocker is a file, so creating directories beneath it fails.
	if _, err := runCmd(t, bk, "purge", "-y", "--backup", blocker); err == nil {
		t.Fatal("want mkdir error")
	}
	if len(bk.Deleted) != 0 {
		t.Errorf("nothing should be deleted, got %v", bk.Deleted)
	}
}

func TestPurgeConfirmYes(t *testing.T) {
	bk := purgeBackend()
	if _, err := runCmdIn(t, bk, "y\n", "purge"); err != nil {
		t.Fatalf("purge: %v", err)
	}
	if len(bk.Deleted) != 3 {
		t.Errorf("deleted = %v", bk.Deleted)
	}
}

func TestPurgeConfirmNo(t *testing.T) {
	bk := purgeBackend()
	out, err := runCmdIn(t, bk, "n\n", "purge")
	if err != nil {
		t.Fatalf("purge: %v", err)
	}
	if !strings.Contains(out, "aborted") {
		t.Errorf("output = %q", out)
	}
	if len(bk.Deleted) != 0 {
		t.Errorf("aborted purge deleted %v", bk.Deleted)
	}
}

func TestPurgeConfirmBackupPrompt(t *testing.T) {
	bk := purgeBackend()
	dir := t.TempDir()
	out, err := runCmdIn(t, bk, "y\n", "purge", "--backup", dir)
	if err != nil {
		t.Fatalf("purge: %v", err)
	}
	if !strings.Contains(out, "Back up to") {
		t.Errorf("expected backup prompt, got %q", out)
	}
	if len(bk.Deleted) != 3 {
		t.Errorf("deleted = %v", bk.Deleted)
	}
}

func TestPurgeNoFiles(t *testing.T) {
	bk := purgeBackend()
	out, err := runCmd(t, bk, "purge", "-y", "--folders", "Courses")
	if err != nil {
		t.Fatalf("purge: %v", err)
	}
	if !strings.Contains(out, "no files to purge") {
		t.Errorf("output = %q", out)
	}
}

func TestPurgeFolderNotFound(t *testing.T) {
	bk := purgeBackend()
	out, err := runCmd(t, bk, "purge", "-y", "--folders", "Nonexistent")
	if err != nil {
		t.Fatalf("purge: %v", err)
	}
	if !strings.Contains(out, "not found on device") {
		t.Errorf("output = %q", out)
	}
}

func TestPurgeListError(t *testing.T) {
	bk := &backend.FakeBackend{ListErr: errors.New("no device")}
	if _, err := runCmd(t, bk, "purge", "-y"); err == nil {
		t.Fatal("want list error")
	}
}

func TestPurgeDeleteFailure(t *testing.T) {
	bk := purgeBackend()
	bk.DeleteErr = errors.New("device busy")
	if _, err := runCmd(t, bk, "purge", "-y"); err == nil {
		t.Fatal("want delete failure error")
	}
}

func TestPurgeConfirmReadError(t *testing.T) {
	bk := purgeBackend()
	err := runPurge(context.Background(), &bytes.Buffer{}, errReader{}, bk, nil, "", false, false)
	if err == nil {
		t.Fatal("want confirm read error")
	}
}
