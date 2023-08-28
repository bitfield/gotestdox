package gotestdox

import (
	"fmt"
	"os"

	"github.com/bitfield/gotestdox/pkg/gotestdox"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
)

const Usage = `gotestdox is a command-line tool for turning Go test names into readable sentences.

Usage:

	gotestdox [ARGS]

This will run 'go test -json [ARGS]' in the current directory and format the results in a readable
way. You can use any arguments that 'go test -json' accepts, including a list of packages, for
example.

If the standard input is not an interactive terminal, gotestdox will assume you want to pipe JSON
data into it. For example:

	go test -json |gotestdox

See https://github.com/bitfield/gotestdox for more information.`

// Main runs the command-line interface for gotestdox. The exit status for the
// binary is 0 if the tests passed, or 1 if the tests failed, or there was some
// error.
//
// # Colour
//
// If the program is attached to an interactive terminal, as determined by
// [github.com/mattn/go-isatty], and the NO_COLOR environment variable is not
// set, check marks will be shown in green and x's in red.
func Main() int {
	if len(os.Args) > 1 && os.Args[1] == "-h" {
		fmt.Println(Usage)
		return 0
	}
	td := gotestdox.NewTestDoxer()
	td.Fail = color.RedString(td.Fail)
	td.Pass = color.GreenString(td.Pass)
	if isatty.IsTerminal(os.Stdin.Fd()) {
		td.ExecGoTest(os.Args[1:])
	} else {
		td.Filter()
	}
	if !td.OK {
		return 1
	}
	return 0
}
