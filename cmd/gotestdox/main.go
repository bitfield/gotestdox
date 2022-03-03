package main

import (
	"os"

	"github.com/bitfield/gotestdox"
	"github.com/mattn/go-isatty"
)

func main() {
	if isatty.IsTerminal(os.Stdin.Fd()) {
		gotestdox.ExecGoTest()
	} else {
		gotestdox.Filter(os.Stdin)
	}
}
