package main

import (
	"os"

	"github.com/tamcore/mtpx/internal/cli"
)

var (
	version = "dev"
	commit  = "none"
)

func main() {
	if err := cli.Execute(version, commit, os.Args[1:]); err != nil {
		os.Exit(1)
	}
}
