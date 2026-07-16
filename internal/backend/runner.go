package backend

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Runner runs external commands. It is injected into LibmtpBackend so tests can
// supply recorded fixture output instead of executing the real libmtp tools.
type Runner interface {
	// Run executes name with args and returns its standard output. Standard
	// error is folded into the returned error on failure.
	Run(ctx context.Context, name string, args ...string) ([]byte, error)
	// LookPath reports whether name is an executable in PATH.
	LookPath(name string) (string, error)
}

// execRunner is the production Runner backed by os/exec.
type execRunner struct{}

func (execRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			return stdout.Bytes(), fmt.Errorf("%s: %w", name, err)
		}
		return stdout.Bytes(), fmt.Errorf("%s: %w: %s", name, err, msg)
	}
	return stdout.Bytes(), nil
}

func (execRunner) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}
