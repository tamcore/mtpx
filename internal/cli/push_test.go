package cli

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tamcore/mtpx/internal/backend"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "in.fit")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp: %v", err)
	}
	return p
}

func TestPushFile(t *testing.T) {
	bk := &backend.FakeBackend{}
	out, err := runCmd(t, bk, "push", writeTemp(t, "hi"), "GARMIN/NewFiles/in.fit")
	if err != nil {
		t.Fatalf("push: %v", err)
	}
	if !strings.Contains(out, "pushed") {
		t.Errorf("output = %q", out)
	}
	objs, err := bk.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	var found bool
	for _, o := range objs {
		if o.Path == "GARMIN/NewFiles/in.fit" {
			found = true
		}
	}
	if !found {
		t.Fatalf("pushed object not registered: %+v", objs)
	}
}

func TestPushMissingLocal(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "nope")
	if _, err := runCmd(t, &backend.FakeBackend{}, "push", missing, "a/b"); err == nil {
		t.Fatal("want error for missing local file")
	}
}

func TestPushDirectory(t *testing.T) {
	if _, err := runCmd(t, &backend.FakeBackend{}, "push", t.TempDir(), "a/b"); err == nil {
		t.Fatal("want error pushing a directory")
	}
}

func TestPushPutError(t *testing.T) {
	bk := &backend.FakeBackend{PutErr: errors.New("send fail")}
	if _, err := runCmd(t, bk, "push", writeTemp(t, "x"), "a/b"); err == nil {
		t.Fatal("want put error")
	}
}
