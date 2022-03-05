package main

import (
	"os"

	"github.com/bitfield/gotestdox"
	"github.com/mattn/go-isatty"
)

func main() {
	allOK := true
	if isatty.IsTerminal(os.Stdin.Fd()) {
		allOK = gotestdox.ExecGoTest()
	} else {
		allOK = gotestdox.Filter(os.Stdin)
	}
	if !allOK {
		os.Exit(1)
	}
}
