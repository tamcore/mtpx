package cli

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("read failed") }

func TestConfirmYes(t *testing.T) {
	for _, in := range []string{"y\n", "Y\n", "yes\n", "YES\n", "  yes  \n"} {
		ok, err := confirm(strings.NewReader(in), &bytes.Buffer{}, "?")
		if err != nil || !ok {
			t.Fatalf("input %q: ok=%v err=%v", in, ok, err)
		}
	}
}

func TestConfirmNo(t *testing.T) {
	for _, in := range []string{"n\n", "no\n", "\n", "maybe\n", ""} {
		ok, err := confirm(strings.NewReader(in), &bytes.Buffer{}, "?")
		if err != nil || ok {
			t.Fatalf("input %q: ok=%v err=%v", in, ok, err)
		}
	}
}

func TestConfirmReadError(t *testing.T) {
	if _, err := confirm(errReader{}, &bytes.Buffer{}, "?"); err == nil {
		t.Fatal("want read error")
	}
}

func TestConfirmPromptWritten(t *testing.T) {
	buf := &bytes.Buffer{}
	if _, err := confirm(strings.NewReader("y\n"), buf, "Delete 3 file(s)?"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "Delete 3 file(s)?") {
		t.Fatalf("prompt = %q", buf.String())
	}
}
