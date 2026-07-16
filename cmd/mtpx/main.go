package main

import "fmt"

var (
	version = "dev"
	commit  = "none"
)

func main() {
	fmt.Printf("mtpx %s (%s)\n", version, commit)
}
