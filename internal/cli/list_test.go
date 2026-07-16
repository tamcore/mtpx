package cli

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/tamcore/mtpx/internal/backend"
)

type errWriter struct{}

func (errWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }

func sampleBackend() *backend.FakeBackend {
	return &backend.FakeBackend{Devices: testDev(), Objects: []backend.Object{
		{ID: 1, Path: "GARMIN", Name: "GARMIN", IsDir: true},
		{ID: 2, Path: "GARMIN/Activity", Name: "Activity", IsDir: true},
		{ID: 3, Path: "GARMIN/Activity/a.fit", Name: "a.fit", Size: 1234},
	}}
}

func TestListRoot(t *testing.T) {
	out, err := runCmd(t, sampleBackend(), "list")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(out, "GARMIN") {
		t.Fatalf("output = %q", out)
	}
}

func TestListSubdir(t *testing.T) {
	out, err := runCmd(t, sampleBackend(), "list", "GARMIN")
	if err != nil {
		t.Fatalf("list GARMIN: %v", err)
	}
	if !strings.Contains(out, "Activity") || !strings.Contains(out, "dir") {
		t.Fatalf("output = %q", out)
	}
}

func TestListFileEntry(t *testing.T) {
	out, err := runCmd(t, sampleBackend(), "list", "GARMIN/Activity")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(out, "a.fit") || !strings.Contains(out, "1234") {
		t.Fatalf("output = %q", out)
	}
}

func TestListJSON(t *testing.T) {
	out, err := runCmd(t, sampleBackend(), "list", "--json")
	if err != nil {
		t.Fatalf("list --json: %v", err)
	}
	var objs []backend.Object
	if err := json.Unmarshal([]byte(out), &objs); err != nil {
		t.Fatalf("invalid JSON %q: %v", out, err)
	}
	if len(objs) != 1 || objs[0].Name != "GARMIN" {
		t.Fatalf("json objects = %+v", objs)
	}
}

func TestListNotADirectory(t *testing.T) {
	if _, err := runCmd(t, sampleBackend(), "list", "GARMIN/Activity/a.fit"); err == nil {
		t.Fatal("want error listing a file")
	}
}

func TestListUnknownPath(t *testing.T) {
	if _, err := runCmd(t, sampleBackend(), "list", "nope"); err == nil {
		t.Fatal("want error for unknown path")
	}
}

func TestListBackendError(t *testing.T) {
	bk := &backend.FakeBackend{Devices: testDev(), ListErr: errors.New("read error")}
	if _, err := runCmd(t, bk, "list"); err == nil {
		t.Fatal("want backend error")
	}
}

func TestListWriteError(t *testing.T) {
	if err := runList(context.Background(), errWriter{}, sampleBackend(), "", false); err == nil {
		t.Fatal("want write error")
	}
}
