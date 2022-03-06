package gotestdox_test

import (
	"io"
	"os"
	"testing"

	"github.com/bitfield/gotestdox"
	"github.com/google/go-cmp/cmp"
)

func TestParseJSON_CorrectlyParsesASingleGoTestJSONOutputLine(t *testing.T) {
	t.Parallel()
	input := `{"Time":"2022-02-28T15:53:43.532326Z","Action":"pass","Package":"github.com/bitfield/script","Test":"TestFindFilesInNonexistentPathReturnsError","Elapsed":0.12}`
	want := gotestdox.Event{
		Action:  "pass",
		Package: "github.com/bitfield/script",
		Test:    "TestFindFilesInNonexistentPathReturnsError",
		Elapsed: 0.12,
	}

	got, err := gotestdox.ParseJSON(input)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestEventString_FormatsPassAndFailEventsDifferently(t *testing.T) {
	t.Parallel()
	pass := gotestdox.Event{
		Action: "pass",
		Test:   "TestFooDoesX",
	}.String()
	fail := gotestdox.Event{
		Action: "fail",
		Test:   "TestFooDoesX",
	}.String()
	if pass == fail {
		t.Errorf("both pass and fail events formatted as %q", pass)
	}
}

func TestRelevantIsTrueForTestPassOrFailEvents(t *testing.T) {
	t.Parallel()
	tcs := []gotestdox.Event{
		{
			Action: "pass",
			Test:   "TestFooDoesX",
		},
		{
			Action: "fail",
			Test:   "TestFooDoesX",
		},
	}
	for _, event := range tcs {
		relevant := event.Relevant()
		if !relevant {
			t.Errorf("false for relevant event %q on %q", event.Action, event.Test)
		}
	}
}

func TestRelevantIsFalseForNonPassFailEvents(t *testing.T) {
	t.Parallel()
	tcs := []gotestdox.Event{
		{
			Action: "pass",
			Test:   "ExampleFooDoesX",
		},
		{
			Action: "fail",
			Test:   "BenchmarkFooDoesX",
		},
		{
			Action: "pass",
			Test:   "",
		},
		{
			Action: "fail",
			Test:   "",
		},
		{
			Action: "run",
			Test:   "TestFooDoesX",
		},
	}
	for _, event := range tcs {
		relevant := event.Relevant()
		if relevant {
			t.Errorf("true for irrelevant event %q on %q", event.Action, event.Test)
		}
	}
}

func TestFilterReturnsOKIfThereAreNoTestFailures(t *testing.T) {
	t.Parallel()
	f, err := os.Open("testdata/passing_tests.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	allOK := gotestdox.Filter(f, io.Discard)
	if !allOK {
		t.Error("not OK")
	}
}

func TestFilterReturnsNotOKIfAnyTestFails(t *testing.T) {
	t.Parallel()
	f, err := os.Open("testdata/failing_tests.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	allOK := gotestdox.Filter(f, io.Discard)
	if allOK {
		t.Error("got OK")
	}
}
