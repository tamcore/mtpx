package cli

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tamcore/mtpx/internal/backend"
)

func pullBackend() *backend.FakeBackend {
	return &backend.FakeBackend{
		Devices: testDev(),
		Objects: []backend.Object{
			{ID: 1, Path: "GARMIN", Name: "GARMIN", IsDir: true},
			{ID: 2, Path: "GARMIN/Activity", Name: "Activity", IsDir: true},
			{ID: 3, Path: "GARMIN/Activity/a.fit", Name: "a.fit", Size: 3},
		},
		Contents: map[uint32][]byte{3: []byte("abc")},
	}
}

func TestPullToFile(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "out.fit")
	out, err := runCmd(t, pullBackend(), "pull", "GARMIN/Activity/a.fit", dest)
	if err != nil {
		t.Fatalf("pull: %v", err)
	}
	if !strings.Contains(out, "pulled") {
		t.Errorf("output = %q", out)
	}
	b, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("read dest: %v", err)
	}
	if string(b) != "abc" {
		t.Errorf("content = %q", b)
	}
}

func TestPullToDir(t *testing.T) {
	dir := t.TempDir()
	if _, err := runCmd(t, pullBackend(), "pull", "GARMIN/Activity/a.fit", dir); err != nil {
		t.Fatalf("pull: %v", err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "a.fit"))
	if err != nil {
		t.Fatalf("read dest: %v", err)
	}
	if string(b) != "abc" {
		t.Errorf("content = %q", b)
	}
}

func TestPullNotFound(t *testing.T) {
	if _, err := runCmd(t, pullBackend(), "pull", "GARMIN/nope.fit", t.TempDir()); err == nil {
		t.Fatal("want error for unknown remote")
	}
}

func TestPullDirectory(t *testing.T) {
	if _, err := runCmd(t, pullBackend(), "pull", "GARMIN/Activity", t.TempDir()); err == nil {
		t.Fatal("want error pulling a directory")
	}
}

func TestPullListError(t *testing.T) {
	bk := &backend.FakeBackend{Devices: testDev(), ListErr: errors.New("read error")}
	if _, err := runCmd(t, bk, "pull", "x", t.TempDir()); err == nil {
		t.Fatal("want list error")
	}
}

func TestPullGetError(t *testing.T) {
	bk := pullBackend()
	bk.GetErr = errors.New("io fail")
	dest := filepath.Join(t.TempDir(), "out.fit")
	if _, err := runCmd(t, bk, "pull", "GARMIN/Activity/a.fit", dest); err == nil {
		t.Fatal("want get error")
	}
}
