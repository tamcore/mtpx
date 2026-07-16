package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/tamcore/mtpx/internal/backend"
)

// runCmd builds the root command with bk, runs it with args, and returns the
// combined output.
func runCmd(t *testing.T, bk backend.Backend, args ...string) (string, error) {
	t.Helper()
	root := NewRootCmd("1.2.3", "abc123", bk)
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	err := root.Execute()
	return buf.String(), err
}

func TestRootVersion(t *testing.T) {
	out, err := runCmd(t, &backend.FakeBackend{}, "--version")
	if err != nil {
		t.Fatalf("--version: %v", err)
	}
	if !strings.Contains(out, "1.2.3") || !strings.Contains(out, "abc123") {
		t.Fatalf("version output = %q", out)
	}
}

func TestRootNoArgsShowsHelp(t *testing.T) {
	out, err := runCmd(t, &backend.FakeBackend{})
	if err != nil {
		t.Fatalf("no args: %v", err)
	}
	if !strings.Contains(out, "list") {
		t.Fatalf("help should mention subcommands, got %q", out)
	}
}

func TestExecuteHelp(t *testing.T) {
	// Execute constructs the real libmtp backend, but --help performs no device
	// I/O, so this stays hermetic.
	if err := Execute("v", "c", []string{"--help"}); err != nil {
		t.Fatalf("Execute --help: %v", err)
	}
}
