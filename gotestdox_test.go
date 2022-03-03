package gotestdox_test

import (
	"testing"

	"github.com/bitfield/gotestdox"
	"github.com/google/go-cmp/cmp"
)

func TestSentence(t *testing.T) {
	t.Parallel()
	tcs := []struct {
		name, input, want string
	}{
		{
			name:  "correctly renders a well-formed test name",
			input: "TestSumCorrectlySumsInputNumbers",
			want:  "Sum correctly sums input numbers",
		},
		{
			name:  "preserves initialisms such as PDF",
			input: "TestFooGeneratesValidPDFFile",
			want:  "Foo generates valid PDF file",
		},
		{
			name:  "preserves more initialisms",
			input: "TestToValidUTF8",
			want:  "To valid UTF 8",
		},
		{
			name:  "treats numbers as word separators",
			input: "TestFooDoes8Things",
			want:  "Foo does 8 things",
		},
		{
			name:  "knows that just Test is a valid test name",
			input: "Test",
			want:  "",
		},
		{
			name:  "treats underscores as word breaks",
			input: "Test_Foo_GeneratesValidPDFFile",
			want:  "Foo generates valid PDF file",
		},
		{
			name:  "doesn't incorrectly title-case single-letter words",
			input: "TestFooDoesAThing",
			want:  "Foo does a thing",
		},
		{
			name:  "renders subtest names without the slash, and with underscores replaced by spaces",
			input: "TestSliceSink/Empty_line_between_two_existing_lines",
			want:  "Slice sink empty line between two existing lines",
		},
		{
			name:  "inserts a word break before subtest names beginning with a lowercase letter",
			input: "TestExec/go_help",
			want:  "Exec go help",
		},
		{
			name:  "is okay with test names not in the form of a sentence",
			input: "TestMatch",
			want:  "Match",
		},
		{
			name:  "treats a single underscore as marking the end of a multi-word function name",
			input: "TestFindFiles_WorksCorrectly",
			want:  "FindFiles works correctly",
		},
		{
			name:  "treats a single underscore before the first slash as marking the end of a multi-word function name",
			input: "TestFindFiles_/WorksCorrectly",
			want:  "FindFiles works correctly",
		},
		{
			name:  "handles multiple underscores, with the first marking the end of a multi-word function name",
			input: "TestFindFiles_Does_Stuff",
			want:  "FindFiles does stuff",
		},
		{
			name:  "eliminates any words containing underscores after splitting",
			input: "TestSentence/does_x,_correctly",
			want:  "Sentence does x correctly",
		},
		{
			name:  "retains hyphenated words in their original form",
			input: "TestFoo/has_well-formed_output",
			want:  "Foo has well-formed output",
		},
		{
			name:  "retains apostrophised words in their original form",
			input: "TestFoo/does_what's_required",
			want:  "Foo does what's required",
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			got := gotestdox.Sentence(tc.input)
			if tc.want != got {
				t.Errorf("%s:\ninput: %q:\nresult: %s", tc.name, tc.input, cmp.Diff(tc.want, got))
			}
		})
	}
}

func TestExtractFuncName_(t *testing.T) {
	t.Parallel()
	tcs := []struct {
		name, input, funcName, behaviour string
	}{
		{
			name:      "matches the first camel-case word if there are no slashes or underscores",
			input:     "SumCorrectlySumsInputNumbers",
			funcName:  "Sum",
			behaviour: "CorrectlySumsInputNumbers",
		},
		{
			name:      "treats a single underscore as marking the end of a multi-word function name",
			input:     "FindFiles_WorksCorrectly",
			funcName:  "FindFiles",
			behaviour: "_WorksCorrectly",
		},
		{
			name:      "treats a single underscore before the first slash as marking the end of a multi-word function name",
			input:     "FindFiles_/WorksCorrectly",
			funcName:  "FindFiles",
			behaviour: "_/WorksCorrectly",
		},
		{
			name:      "treats multiple underscores as word breaks",
			input:     "_Foo_GeneratesValidPDFFile",
			funcName:  "Foo",
			behaviour: "_GeneratesValidPDFFile",
		},
		{
			name:      "correctly extracts func name from a subtest",
			input:     "Slice/Empty_line_between_two_existing_lines",
			funcName:  "Slice",
			behaviour: "/Empty_line_between_two_existing_lines",
		},
		{
			name:      "without an underscore before a slash, treats camel case as word breaks",
			input:     "SliceSink/Empty_line_between_two_existing_lines",
			funcName:  "Slice",
			behaviour: "Sink/Empty_line_between_two_existing_lines",
		},
		{
			name:      "doesn't break if the test is named just Test",
			input:     "",
			funcName:  "",
			behaviour: "",
		},
		{
			name:      "doesn't break if the test is named just Test followed by an underscore",
			input:     "_",
			funcName:  "",
			behaviour: "",
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			funcName, behaviour := gotestdox.ExtractFuncName(tc.input)
			if tc.funcName != funcName {
				t.Errorf("%s\ninput: %q:\nfuncName: %s", tc.name, tc.input, cmp.Diff(tc.funcName, funcName))
			}
			if tc.behaviour != behaviour {
				t.Errorf("%s\ninput: %q:\nbehaviour: %s", tc.name, tc.input, cmp.Diff(tc.behaviour, behaviour))
			}
		})
	}
}

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

func TestEventString_FormatsPassEventsWithATick(t *testing.T) {
	t.Parallel()
	input := gotestdox.Event{
		Action:  "pass",
		Test:    "TestFooDoesX",
		Elapsed: 0.01,
	}
	want := " ✔ Foo does x (0.01s)"
	got := input.String()
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestEventString_FormatsFailEventsWithACross(t *testing.T) {
	t.Parallel()
	input := gotestdox.Event{
		Action:  "fail",
		Test:    "TestFooDoesX",
		Elapsed: 0.01,
	}
	want := " ✘ Foo does x (0.01s)"
	got := input.String()
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
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

func TestRelevantIsFalseForOtherEvents(t *testing.T) {
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
