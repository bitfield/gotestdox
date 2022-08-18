package gotestdox

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
)

// TestDoxer holds the state and config associated with a particular invocation
// of 'go test'.
type TestDoxer struct {
	Stdin          io.Reader
	Stdout, Stderr io.Writer
	OK             bool
	SplitErrors    bool
}

// NewTestDoxer returns a *TestDoxer configured with the default I/O streams:
// stdin, stdout, stderr.
func NewTestDoxer() *TestDoxer {
	return &TestDoxer{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
}

// ExecGoTest runs the 'go test -json' command, with any extra args supplied by
// the user, and consumes its output. Any errors are reported to the TestDoxer's
// configured Stderr stream, including the full command line that was run. If
// all tests passed, the TestDoxer's OK field will be true. If there was a test
// failure, or 'go test' returned some error, OK will be false.
func (td *TestDoxer) ExecGoTest(userArgs []string) {
	args := []string{"test", "-json"}

	for _, arg := range userArgs {
		if arg == "split-errors" {
			td.SplitErrors = true
			continue
		}

		args = append(args, arg)
	}

	cmd := exec.Command("go", args...)
	goTestOutput, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintln(td.Stderr, cmd.Args, err)
		return
	}
	cmd.Stderr = td.Stderr
	if err := cmd.Start(); err != nil {
		fmt.Fprintln(td.Stderr, cmd.Args, err)
		return
	}
	td.Stdin = goTestOutput
	td.Filter()
	if err := cmd.Wait(); err != nil {
		td.OK = false
		fmt.Fprintln(td.Stderr, cmd.Args, err)
		return
	}
}

// Filter reads from the TestDoxer's Stdin stream, line by line, processing JSON
// records emitted by 'go test -json'. For each Go package it sees records
// about, it will print the full name of the package to Stdout, followed by a
// line giving the pass/fail status and the prettified name of each test. If all
// tests passed, the TestDoxer's OK field will be true at the end. If not, or if
// there was a parsing error, it will be false. Errors will be reported to
// Stderr.
func (td *TestDoxer) Filter() {
	td.OK = true
	var curPkg string
	scanner := bufio.NewScanner(td.Stdin)

	runnedTests := make(map[string][]Event)
	failedTests := make(map[string][]Event)

	for scanner.Scan() {
		event, err := ParseJSON(scanner.Text())
		if err != nil {
			td.OK = false
			fmt.Fprintln(td.Stderr, err)
			return
		}
		if event.Action == "fail" {
			td.OK = false
		}
		if !event.Relevant() {
			continue
		}
		if event.Package != curPkg {
			curPkg = event.Package
		}

		if td.SplitErrors && event.Action == "fail" {
			failedTests[event.Package] = append(failedTests[event.Package], event)
		} else {
			runnedTests[event.Package] = append(runnedTests[event.Package], event)
		}
	}
	for k, v := range runnedTests {
		fmt.Fprintln(td.Stdout, k)
		for _, event := range v {
			fmt.Fprintln(td.Stdout, event)
		}
	}

	if td.SplitErrors && len(failedTests) > 0 {
		fmt.Fprintln(td.Stdout, "\nFailed tests")
		for k, v := range failedTests {
			fmt.Fprintln(td.Stdout, k)
			for _, event := range v {
				fmt.Fprintln(td.Stdout, event)
			}
		}
	}
}

// Event represents a Go test event as recorded by the 'go test -json' command.
// It does not attempt to unmarshal all the data, only those fields it needs to
// know about. The struct definition is based on that used by Go to create the
// JSON in the first place:
//
// https://cs.opensource.google/go/go/+/refs/tags/go1.17.7:src/cmd/internal/test2json/test2json.go;l=30
type Event struct {
	Action  string
	Package string
	Test    string
	Elapsed float64
}

// String formats a test Event for display. If the test passed, it will be
// prefixed by a check mark emoji, or an 'x' if it failed. If os.Stdin is a
// terminal (as determined by the 'isatty' library), and the 'NO_COLOR'
// environment variable is not set, check marks will be shown in green and x's
// in red.
//
// The name of the test will be filtered through Prettify to turn it into a
// sentence, and finally the elapsed time will be shown in parentheses, to 2
// decimal places.
func (e Event) String() string {
	status := color.RedString("x")
	if e.Action == "pass" {
		status = color.GreenString("✔")
	}
	return fmt.Sprintf(" %s %s (%.2fs)", status, Prettify(e.Test), e.Elapsed)
}

// Relevant determines whether or not the test event is one that we are
// interested in (namely, a pass or fail event on a test). Events on non-tests
// are ignored, and all events on tests other than pass or fail events are also
// ignored.
func (e Event) Relevant() bool {
	// Events on non-tests are irrelevant
	if !strings.HasPrefix(e.Test, "Test") {
		return false
	}
	if e.Action == "pass" || e.Action == "fail" {
		return true
	}
	return false
}

// ParseJSON takes a string representing a single JSON test record as emitted by
// 'go test -json', and attempts to parse it into an Event struct, returning any
// parsing error encountered.
func ParseJSON(line string) (Event, error) {
	event := Event{}
	err := json.Unmarshal([]byte(line), &event)
	if err != nil {
		return Event{}, fmt.Errorf("parsing JSON: %w\ninput: %s", err, line)
	}
	return event, nil
}
