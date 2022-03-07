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

type TestDoxer struct {
	Stdin          io.Reader
	Stdout, Stderr io.Writer
	OK             bool
}

func NewTestDoxer() *TestDoxer {
	return &TestDoxer{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
}

func (td *TestDoxer) ExecGoTest(userArgs []string) {
	args := []string{"test", "-json"}
	args = append(args, userArgs...)
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

func (td *TestDoxer) Filter() {
	td.OK = true
	var curPkg string
	scanner := bufio.NewScanner(td.Stdin)
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
			if curPkg != "" {
				fmt.Fprintln(td.Stdout, curPkg)
			}
			fmt.Fprintf(td.Stdout, "%s:\n", event.Package)
			curPkg = event.Package
		}
		fmt.Fprintln(td.Stdout, event)
	}
}

// https://cs.opensource.google/go/go/+/refs/tags/go1.17.7:src/cmd/internal/test2json/test2json.go;l=30
type Event struct {
	Action  string
	Package string
	Test    string
	Elapsed float64
}

func (e Event) String() string {
	status := color.RedString("x")
	if e.Action == "pass" {
		status = color.GreenString("✔")
	}
	return fmt.Sprintf(" %s %s (%.2fs)", status, Prettify(e.Test), e.Elapsed)
}

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

func ParseJSON(line string) (Event, error) {
	event := Event{}
	err := json.Unmarshal([]byte(line), &event)
	if err != nil {
		return Event{}, err
	}
	return event, nil
}
