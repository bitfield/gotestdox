package testgox

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fatih/camelcase"
)

func Sentence(name string) string {
	name = strings.TrimPrefix(name, "Test")
	if len(name) < 1 {
		return ""
	}
	// Slashes indicate a subtest
	name = strings.ReplaceAll(name, "/", "_")
	input := camelcase.Split(name)
	output := make([]string, 0, len(input))
	for _, w := range input {
		if w == "_" {
			continue
		}
		// Lowercase words, except initialisms and single-letter words
		if w != strings.ToUpper(w) || len(w) == 1 {
			w = strings.ToLower(w)
		}
		output = append(output, w)
	}
	// Capitalise the first word in the sentence
	output[0] = strings.Title(output[0])
	return strings.Join(output, " ")
}

// https://cs.opensource.google/go/go/+/refs/tags/go1.17.7:src/cmd/internal/test2json/test2json.go;l=30
type Event struct {
	Action  string
	Package string
	Test    string
	Elapsed float64
}

func (e Event) String() string {
	status := "✘"
	if e.Action == "pass" {
		status = "✔"
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
