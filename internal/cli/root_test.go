package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/tamcore/mtpx/internal/backend"
)

func noopLaunch(context.Context, backend.Backend, string) error { return nil }

// runCmd builds the root command with bk and a no-op launcher, runs it with
// args, and returns the combined output. Standard input is empty.
func runCmd(t *testing.T, bk backend.Backend, args ...string) (string, error) {
	t.Helper()
	return runCmdIn(t, bk, "", args...)
}

// runCmdIn is like runCmd but feeds stdin from input.
func runCmdIn(t *testing.T, bk backend.Backend, input string, args ...string) (string, error) {
	t.Helper()
	root := NewRootCmd("1.2.3", "abc123", bk, noopLaunch)
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetIn(strings.NewReader(input))
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

func TestRootLaunchesTUI(t *testing.T) {
	var called bool
	var gotDest string
	launch := func(_ context.Context, _ backend.Backend, dest string) error {
		called = true
		gotDest = dest
		return nil
	}
	root := NewRootCmd("v", "c", &backend.FakeBackend{Devices: testDev()}, launch)
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !called {
		t.Fatal("running with no args should launch the TUI")
	}
	if gotDest != "." {
		t.Fatalf("default dest = %q, want \".\"", gotDest)
	}
}

func TestRootDestFlag(t *testing.T) {
	var gotDest string
	launch := func(_ context.Context, _ backend.Backend, dest string) error {
		gotDest = dest
		return nil
	}
	root := NewRootCmd("v", "c", &backend.FakeBackend{Devices: testDev()}, launch)
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"--dest", "/tmp/pulls"})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if gotDest != "/tmp/pulls" {
		t.Fatalf("dest = %q", gotDest)
	}
}

func TestExecuteHelp(t *testing.T) {
	// Execute constructs the real libmtp backend, but --help performs no device
	// I/O and does not launch the TUI, so this stays hermetic.
	if err := Execute("v", "c", []string{"--help"}, noopLaunch); err != nil {
		t.Fatalf("Execute --help: %v", err)
	}
}
