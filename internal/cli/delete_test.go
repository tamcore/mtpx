package cli

import (
	"bytes"
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/tamcore/mtpx/internal/backend"
)

func deleteBackend() *backend.FakeBackend {
	return &backend.FakeBackend{Devices: testDev(), Objects: []backend.Object{
		{ID: 1, Path: "GARMIN", Name: "GARMIN", IsDir: true},
		{ID: 2, Path: "GARMIN/Activity", Name: "Activity", IsDir: true},
		{ID: 3, Path: "GARMIN/Activity/a.fit", Name: "a.fit"},
		{ID: 4, Path: "GARMIN/Activity/b.fit", Name: "b.fit"},
	}}
}

func TestDeleteYes(t *testing.T) {
	bk := deleteBackend()
	out, err := runCmd(t, bk, "delete", "--yes", "GARMIN/Activity/a.fit")
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if !strings.Contains(out, "deleted") {
		t.Errorf("output = %q", out)
	}
	if !slices.Equal(bk.Deleted, []uint32{3}) {
		t.Errorf("deleted = %v, want [3]", bk.Deleted)
	}
}

func TestDeleteMultiple(t *testing.T) {
	bk := deleteBackend()
	if _, err := runCmd(t, bk, "delete", "-y", "GARMIN/Activity/a.fit", "GARMIN/Activity/b.fit"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if !slices.Equal(bk.Deleted, []uint32{3, 4}) {
		t.Errorf("deleted = %v, want [3 4]", bk.Deleted)
	}
}

func TestDeleteConfirmYes(t *testing.T) {
	bk := deleteBackend()
	if _, err := runCmdIn(t, bk, "y\n", "delete", "GARMIN/Activity/a.fit"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if len(bk.Deleted) != 1 {
		t.Errorf("deleted = %v", bk.Deleted)
	}
}

func TestDeleteConfirmNo(t *testing.T) {
	bk := deleteBackend()
	out, err := runCmdIn(t, bk, "n\n", "delete", "GARMIN/Activity/a.fit")
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if !strings.Contains(out, "aborted") {
		t.Errorf("output = %q", out)
	}
	if len(bk.Deleted) != 0 {
		t.Errorf("nothing should be deleted, got %v", bk.Deleted)
	}
}

func TestDeleteDryRun(t *testing.T) {
	bk := deleteBackend()
	out, err := runCmd(t, bk, "delete", "--dry-run", "GARMIN/Activity/a.fit")
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if !strings.Contains(out, "would delete") {
		t.Errorf("output = %q", out)
	}
	if len(bk.Deleted) != 0 {
		t.Errorf("dry run should not delete, got %v", bk.Deleted)
	}
}

func TestDeleteNotFound(t *testing.T) {
	if _, err := runCmd(t, deleteBackend(), "delete", "-y", "GARMIN/nope"); err == nil {
		t.Fatal("want error for unknown path")
	}
}

func TestDeleteDirectory(t *testing.T) {
	if _, err := runCmd(t, deleteBackend(), "delete", "-y", "GARMIN/Activity"); err == nil {
		t.Fatal("want error deleting a directory")
	}
}

func TestDeleteListError(t *testing.T) {
	bk := &backend.FakeBackend{Devices: testDev(), ListErr: errors.New("read error")}
	if _, err := runCmd(t, bk, "delete", "-y", "x"); err == nil {
		t.Fatal("want list error")
	}
}

func TestDeleteFailure(t *testing.T) {
	bk := deleteBackend()
	bk.DeleteErr = errors.New("device busy")
	out, err := runCmd(t, bk, "delete", "-y", "GARMIN/Activity/a.fit")
	if err == nil {
		t.Fatal("want deletion failure error")
	}
	if !strings.Contains(out, "failed to delete") {
		t.Errorf("output = %q", out)
	}
}

func TestDeleteConfirmReadError(t *testing.T) {
	bk := deleteBackend()
	err := runDelete(context.Background(), &bytes.Buffer{}, errReader{}, bk,
		[]string{"GARMIN/Activity/a.fit"}, false, false)
	if err == nil {
		t.Fatal("want confirm read error")
	}
}
