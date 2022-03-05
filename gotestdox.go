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
	"unicode"
	"unicode/utf8"

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
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		event, err := ParseJSON(scanner.Text())
		if err != nil {
			continue
		}
		if event.Action == "fail" {
			allOK = false
		}
		if event.Relevant() {
			fmt.Fprintln(output, event)
		}
	}
	return allOK
}

// Based on https://github.com/fatih/camelcase, used under MIT licence
func SplitCamelCase(src string) (entries []string) {
	// don't split invalid utf8
	if !utf8.ValidString(src) {
		return []string{src}
	}
	entries = []string{}
	var runes [][]rune
	lastClass := 0
	class := 0
	// split into fields based on class of unicode character
	for _, r := range src {
		switch {
		case unicode.IsLower(r):
			class = 1
		case unicode.IsUpper(r):
			class = 2
		case r == '-' || r == '\'':
			class = lastClass
		default:
			class = 3
		}
		if class == lastClass {
			runes[len(runes)-1] = append(runes[len(runes)-1], r)
		} else {
			runes = append(runes, []rune{r})
		}
		lastClass = class
	}
	// handle upper case -> lower case sequences, e.g.
	// "PDFL", "oader" -> "PDF", "Loader"
	for i := 0; i < len(runes)-1; i++ {
		if unicode.IsUpper(runes[i][0]) && unicode.IsLower(runes[i+1][0]) {
			runes[i+1] = append([]rune{runes[i][len(runes[i])-1]}, runes[i+1]...)
			runes[i] = runes[i][:len(runes[i])-1]
		}
	}
	// construct []string from results
	for _, s := range runes {
		if len(s) > 0 {
			entries = append(entries, string(s))
		}
	}
	return
}

func Sentence(name string) string {
	name = strings.TrimPrefix(name, "Test")
	funcName, behaviour := ExtractFuncName(name)
	output := []string{funcName}
	// Slashes indicate a subtest
	behaviour = strings.ReplaceAll(behaviour, "/", "_")
	input := SplitCamelCase(behaviour)
	for _, w := range input {
		if strings.Contains(w, "_") {
			continue
		}
		// Lowercase words, except all-caps words
		if w != strings.ToUpper(w) || len(w) == 1 {
			w = strings.ToLower(w)
		}
		output = append(output, w)
	}
	// Capitalise the first word in the sentence
	if len(output) > 0 {
		output[0] = strings.Title(output[0])
	}
	return strings.Join(output, " ")
}

func ExtractFuncName(t string) (funcName, behaviour string) {
	var parts []string
	// Trim any leading underscore
	t = strings.TrimPrefix(t, "_")
	// An underscore in a name without slashes marks the end of a multi-word
	// function name
	if strings.Contains(t, "_") && !strings.Contains(t, "/") {
		parts = strings.SplitN(t, "_", 2)
		// An underscore that precedes the first slash also marks the end of the
		// function name
	} else if strings.Contains(t, "/") && strings.Index(t, "_") < strings.Index(t, "/") {
		parts = strings.SplitN(t, "_", 2)
		// Otherwise, the function name is the first camel-case word
	} else {
		parts = SplitCamelCase(t)
	}
	if len(parts) > 0 {
		funcName = parts[0]
	}
	return funcName, t[len(funcName):]
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
	return fmt.Sprintf(" %s %s (%.2fs)", status, Sentence(e.Test), e.Elapsed)
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
