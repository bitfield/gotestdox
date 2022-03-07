package main

import (
	"os"

	"github.com/bitfield/gotestdox"
	"github.com/mattn/go-isatty"
)

func main() {
	td := gotestdox.NewTestDoxer()
	if isatty.IsTerminal(os.Stdin.Fd()) {
		td.ExecGoTest(os.Args[1:])
	} else {
		td.Filter()
	}
	if !td.OK {
		os.Exit(1)
	}
}
