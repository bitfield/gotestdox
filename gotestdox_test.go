package gotestdox_test

import (
	"fmt"
	"io"
	"log"
	"os"
	"testing"

	"github.com/bitfield/gotestdox"
	"github.com/fatih/color"
	"github.com/google/go-cmp/cmp"
	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	os.Exit(testscript.RunMain(m, map[string]func() int{
		"gotestdox": gotestdox.Main,
	}))
}

func TestGotestdoxProducesCorrectOutputWhen(t *testing.T) {
	t.Parallel()
	testscript.Run(t, testscript.Params{
		Dir: "testdata/script",
	})
}

func TestParseJSON_ReturnsValidDataForValidJSON(t *testing.T) {
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

func TestParseJSON_ErrorsOnInvalidJSON(t *testing.T) {
	t.Parallel()
	input := `invalid`
	_, err := gotestdox.ParseJSON(input)
	if err == nil {
		t.Error("want error")
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

func TestRelevantIsFalseForNonTestPassFailEvents(t *testing.T) {
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
	td := gotestdox.NewTestDoxer()
	data, err := os.Open("testdata/example.json")
	if err != nil {
		panic(err)
	}
	defer data.Close()
	td.Stdin = data
	color.NoColor = true
	td.Filter()
	// Output:
	// github.com/bitfield/gotestdox:
	//  ✔ Relevant is true for test pass or fail events (0.00s)
}

func ExampleEvent_String() {
	event := gotestdox.Event{
		Action:   "pass",
		Sentence: "EventString formats pass and fail events differently",
	}
	color.NoColor = true
	fmt.Println(event.String())
	// Output:
	// ✔ EventString formats pass and fail events differently (0.00s)
}

func ExampleEvent_Relevant_true() {
	event := gotestdox.Event{
		Action: "pass",
		Test:   "TestFoo",
	}
	fmt.Println(event.Relevant())
	// Output:
	// true
}

func ExampleEvent_Relevant_false() {
	event := gotestdox.Event{
		Action: "fail",
		Test:   "ExampleFoo",
	}
	fmt.Println(event.Relevant())
	// Output:
	// false
}

func ExampleParseJSON() {
	input := `{"Time":"2022-03-05T11:33:08.167467Z","Action":"pass","Package":"github.com/bitfield/gotestdox","Test":"TestRelevantIsTrueForTestPassOrFailEvents","Elapsed":0}`
	event, err := gotestdox.ParseJSON(input)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%#v\n", event)
	// Output:
	// gotestdox.Event{Action:"pass", Package:"github.com/bitfield/gotestdox", Test:"TestRelevantIsTrueForTestPassOrFailEvents", Sentence:"", Elapsed:0}
}
