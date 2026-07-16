package backend

import (
	"context"
	"strings"
	"testing"
)

func TestExecRunnerRun(t *testing.T) {
	out, err := execRunner{}.Run(context.Background(), "echo", "hello")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if strings.TrimSpace(string(out)) != "hello" {
		t.Fatalf("out = %q", out)
	}
}

func TestExecRunnerRunErrorNoStderr(t *testing.T) {
	_, err := execRunner{}.Run(context.Background(), "false")
	if err == nil {
		t.Fatal("want error from `false`")
	}
	if !strings.Contains(err.Error(), "false") {
		t.Errorf("error should name the command: %v", err)
	}
}

func TestExecRunnerRunErrorWithStderr(t *testing.T) {
	_, err := execRunner{}.Run(context.Background(), "sh", "-c", "echo boom 1>&2; exit 1")
	if err == nil {
		t.Fatal("want error")
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Errorf("error should include stderr: %v", err)
	}
}

func TestExecRunnerLookPath(t *testing.T) {
	var r execRunner
	if _, err := r.LookPath("echo"); err != nil {
		t.Errorf("LookPath(echo): %v", err)
	}
	if _, err := r.LookPath("mtpx-no-such-binary-xyz"); err == nil {
		t.Error("want error for missing binary")
	}
}
