package gotestdox_test

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/bitfield/gotestdox"
	"github.com/fatih/color"
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

func TestFilterSetsOKToTrueIfThereAreNoTestFailures(t *testing.T) {
	t.Parallel()
	data, err := os.Open("testdata/passing_tests.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer data.Close()
	td := gotestdox.TestDoxer{
		Stdin:  data,
		Stdout: io.Discard,
		Stderr: io.Discard,
	}
	td.Filter()
	if !td.OK {
		t.Error("not OK")
	}
}

func TestFilterSetsOKToFalseIfAnyTestFails(t *testing.T) {
	t.Parallel()
	data, err := os.Open("testdata/failing_tests.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer data.Close()
	td := gotestdox.TestDoxer{
		Stdin:  data,
		Stdout: io.Discard,
		Stderr: io.Discard,
	}
	td.Filter()
	if td.OK {
		t.Error("got OK")
	}
}

func TestFilterSkipsIrrelevantEvents(t *testing.T) {
	t.Parallel()
	data, err := os.Open("testdata/passing_tests.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer data.Close()
	buf := &bytes.Buffer{}
	td := gotestdox.TestDoxer{
		Stdin:  data,
		Stdout: buf,
		Stderr: io.Discard,
	}
	td.Filter()
	if strings.Contains(buf.String(), "Example") {
		t.Error("irrelevant event (Example)")
	}
}

func TestFilterKeepsTrackOfCurrentPackage(t *testing.T) {
	t.Parallel()
	data, err := os.Open("testdata/packages.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer data.Close()
	buf := &bytes.Buffer{}
	td := gotestdox.TestDoxer{
		Stdin:  data,
		Stdout: buf,
		Stderr: io.Discard,
	}
	td.Filter()
	if strings.Count(buf.String(), "a:") > 1 {
		t.Error("want package a to be shown only once")
	}
}

func TestFilterOrdersTestsByPrettifiedName(t *testing.T) {
	data, err := os.Open("testdata/unordered_tests.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer data.Close()
	buf := &bytes.Buffer{}
	td := gotestdox.TestDoxer{
		Stdin:  data,
		Stdout: buf,
		Stderr: io.Discard,
	}
	color.NoColor = true
	td.Filter()
	want := "p:\n x A (0.00s)\n ✔ B (0.00s)\n ✔ C (0.00s)\n"
	got := buf.String()
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestFilterSetsOKToFalseOnParsingError(t *testing.T) {
	t.Parallel()
	td := gotestdox.TestDoxer{
		Stdin:  strings.NewReader("invalid"),
		Stdout: io.Discard,
		Stderr: io.Discard,
	}
	td.Filter()
	if td.OK {
		t.Error("got OK")
	}
}

func TestExecGoTest_SetsOKToTrueWhenTestsPass(t *testing.T) {
	t.Parallel()
	path := newTempTestPath(t, passingTest)
	td := gotestdox.TestDoxer{
		Stdin:  strings.NewReader("invalid"),
		Stdout: io.Discard,
		Stderr: io.Discard,
	}
	td.ExecGoTest([]string{path})
	if !td.OK {
		t.Error("want ok")
	}
}

func TestExecGoTest_SetsOKToFalseWhenTestsFail(t *testing.T) {
	t.Parallel()
	path := newTempTestPath(t, failingTest)
	td := gotestdox.TestDoxer{
		Stdin:  strings.NewReader("invalid"),
		Stdout: io.Discard,
		Stderr: io.Discard,
	}
	td.ExecGoTest([]string{path})
	if td.OK {
		t.Error("want not ok")
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

var (
	preamble    = "package dummy\nimport \"testing\"\nfunc TestDummy(t *testing.T)"
	passingTest = preamble + "{}"
	failingTest = preamble + "{t.FailNow()}"
)

func newTempTestPath(t *testing.T, data string) (path string) {
	t.Helper()
	testDir := t.TempDir()
	err := os.WriteFile(testDir+"/go.mod", []byte("module dummy"), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	path = testDir + "/dummy_test.go"
	err = os.WriteFile(path, []byte(data), os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func ExampleTestDoxer_Filter() {
	td := gotestdox.NewTestDoxer()
	data, err := os.Open("testdata/small_passing_tests.txt")
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
		Action: "pass",
		Test:   "TestEventString_FormatsPassAndFailEventsDifferently",
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
	// gotestdox.Event{Action:"pass", Package:"github.com/bitfield/gotestdox", Test:"TestRelevantIsTrueForTestPassOrFailEvents", Elapsed:0}
}
