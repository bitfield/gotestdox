package gotestdox

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
)

func ExecGoTest() bool {
	args := []string{"test", "-json"}
	args = append(args, os.Args[1:]...)
	cmd := exec.Command("go", args...)
	input, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	allOK := Filter(input, os.Stdout)
	if err := cmd.Wait(); err != nil {
		allOK = false
	}
	return allOK
}

func Filter(input io.Reader, output io.Writer) bool {
	allOK := true
	var curPkg string
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		event, err := ParseJSON(scanner.Text())
		if err != nil {
			continue
		}
		if event.Action == "fail" {
			allOK = false
		}
		if !event.Relevant() {
			continue
		}
		if event.Package != curPkg {
			if curPkg != "" {
				fmt.Fprintln(output)
			}
			fmt.Fprintf(output, "%s:\n", event.Package)
			curPkg = event.Package
		}
		fmt.Fprintln(output, event)
	}
	return allOK
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
