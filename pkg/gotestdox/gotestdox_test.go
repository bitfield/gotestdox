package gotestdox_test

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/bitfield/gotestdox/pkg/gotestdox"

	"github.com/google/go-cmp/cmp"
)

func TestParseJSON_ReturnsValidDataForValidJSON(t *testing.T) {
	t.Parallel()
	input := `{"Time":"2022-02-28T15:53:43.532326Z","Action":"pass","Package":"github.com/bitfield/script","Test":"TestFindFilesInNonexistentPathReturnsError","Elapsed":0.12}`
	want := gotestdox.Event{
		Action:  "pass",
		Package: "github.com/bitfield/script",
		Test:    "TestFindFilesInNonexistentPathReturnsError",
		Elapsed: 0.12,
		Status:  "✔",
	}

	td := gotestdox.NewTestDoxer()
	got, err := td.ParseJSON(input)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestParseJSON_ErrorsOnInvalidJSON(t *testing.T) {
	t.Parallel()
	input := `invalid`
	td := gotestdox.NewTestDoxer()
	_, err := td.ParseJSON(input)
	if err == nil {
		t.Error("want error")
	}
}

func TestEventString_FormatsPassAndFailEventsDifferently(t *testing.T) {
	t.Parallel()
	pass := gotestdox.Event{
		Action: "pass",
		Status: "✔",
		Test:   "TestFooDoesX",
	}.String()
	fail := gotestdox.Event{
		Action: "fail",
		Status: "x",
		Test:   "TestFooDoesX",
	}.String()
	if pass == fail {
		t.Errorf("both pass and fail events formatted as %q", pass)
	}
}

func TestIsFuzzFail_IsTrueForFuzzFailEvents(t *testing.T) {
	t.Parallel()
	event := gotestdox.Event{
		Action: "fail",
		Test:   "FuzzBar",
	}
	if !event.IsFuzzFail() {
		t.Errorf("false for %q event on %q", event.Action, event.Test)
	}
}

func TestIsFuzzFail_IsFalseForNonFuzzFailEvents(t *testing.T) {
	t.Parallel()
	tcs := []gotestdox.Event{
		{
			Action: "pass",
			Test:   "FuzzBar",
		},
		{
			Action: "fail",
			Test:   "TestFooDoesX",
		},
	}
	for _, event := range tcs {
		if event.IsFuzzFail() {
			t.Errorf("true for %q event on %q", event.Action, event.Test)
		}
	}
}

func TestIsTestResult_IsTrueForTestPassOrFailEvents(t *testing.T) {
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
		if !event.IsTestResult() {
			t.Errorf("false for %q event on %q", event.Action, event.Test)
		}
	}
}

func TestIsTestResult_IsFalseForNonTestPassFailEvents(t *testing.T) {
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
			Action: "pass",
			Test:   "FuzzBar",
		},
		{
			Action: "run",
			Test:   "TestFooDoesX",
		},
	}
	for _, event := range tcs {
		if event.IsTestResult() {
			t.Errorf("true for %q event on %q", event.Action, event.Test)
		}
	}
}

func TestIsPackageResult_IsTrueForPackageResultEvents(t *testing.T) {
	t.Parallel()
	tcs := []gotestdox.Event{
		{
			Action: "pass",
			Test:   "",
		},
		{
			Action: "fail",
			Test:   "",
		},
	}
	for _, event := range tcs {
		if !event.IsPackageResult() {
			t.Errorf("false for package result event %#v", event)
		}
	}
}

func TestIsPackageResult_IsFalseForNonPackageResultEvents(t *testing.T) {
	t.Parallel()
	tcs := []gotestdox.Event{
		{
			Action: "pass",
			Test:   "TestSomething",
		},
		{
			Action: "fail",
			Test:   "TestSomething",
		},
		{
			Action: "output",
			Test:   "",
		},
	}
	for _, event := range tcs {
		if event.IsPackageResult() {
			t.Errorf("true for non package result event %#v", event)
		}
	}
}

func TestNewTestDoxer_ReturnsTestdoxerWithStandardIOStreams(t *testing.T) {
	t.Parallel()
	td := gotestdox.NewTestDoxer()
	if td.Stdin != os.Stdin {
		t.Error("want stdin os.Stdin")
	}
	if td.Stdout != os.Stdout {
		t.Error("want stdout os.Stdout")
	}
	if td.Stderr != os.Stderr {
		t.Error("want stderr os.Stderr")
	}
}

func TestExecGoTest_SetsOKToFalseWhenCommandErrors(t *testing.T) {
	t.Parallel()
	td := gotestdox.TestDoxer{
		Stdout: io.Discard,
		Stderr: io.Discard,
	}
	td.ExecGoTest([]string{"bogus"})
	if td.OK {
		t.Error("want not ok")
	}
}

func ExampleTestDoxer_Filter() {
	input := `{"Action":"pass","Package":"demo","Test":"TestItWorks"}
	{"Action":"pass","Package":"demo","Elapsed":0}`
	td := gotestdox.NewTestDoxer()
	td.Stdin = strings.NewReader(input)
	td.Filter()
	// Output:
	// demo:
	//  ✔ It works (0.00s)
}

func ExampleEvent_String() {
	event := gotestdox.Event{
		Action:   "pass",
		Status:   "✔",
		Sentence: "It works",
	}
	fmt.Println(event.String())
	// Output:
	// ✔ It works (0.00s)
}

func ExampleEvent_IsTestResult_true() {
	event := gotestdox.Event{
		Action: "pass",
		Test:   "TestItWorks",
	}
	fmt.Println(event.IsTestResult())
	// Output:
	// true
}

func ExampleEvent_IsTestResult_false() {
	event := gotestdox.Event{
		Action: "fail",
		Test:   "ExampleEventsShouldBeIgnored",
	}
	fmt.Println(event.IsTestResult())
	// Output:
	// false
}

func ExampleTestDoxer_ParseJSON() {
	input := `{"Action":"pass","Package":"demo","Test":"TestItWorks","Output":"","Elapsed":0.2}`
	td := gotestdox.NewTestDoxer()
	event, err := td.ParseJSON(input)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%#v\n", event)
	// Output:
	// gotestdox.Event{Action:"pass", Package:"demo", Test:"TestItWorks", Sentence:"", Output:"", Elapsed:0.2, Status:"✔"}
}
