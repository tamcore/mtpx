package cli

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// confirm prints prompt and reads a yes/no answer from in. Only "y" or "yes"
// (case-insensitive) count as yes; anything else, including EOF, is no.
func confirm(in io.Reader, w io.Writer, prompt string) (bool, error) {
	fmt.Fprintf(w, "%s [y/N]: ", prompt)
	line, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}
	switch strings.TrimSpace(strings.ToLower(line)) {
	case "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}
