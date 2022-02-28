package testgox_test

import (
	"testing"

	"github.com/bitfield/testgox"
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
			name:  "preserves initialisms such as 'PDF'",
			input: "TestFooGeneratesValidPDFFile",
			want:  "Foo generates valid PDF file",
		},
		{
			name:  "preserves more initialisms",
			input: "TestToValidUTF8",
			want:  "To valid UTF 8",
		},
		{
			name:  "knows that just 'Test' is a valid test name",
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
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			got := testgox.Sentence(tc.input)
			if tc.want != got {
				t.Errorf("%s:\ninput: %q:\nresult: %s", tc.name, tc.input, cmp.Diff(tc.want, got))
			}
		})
	}
}

func TestParseJSONCorrectlyParsesASingleGoTestJSONOutputLine(t *testing.T) {
	t.Parallel()
	input := `{"Time":"2022-02-28T15:53:43.532326Z","Action":"pass","Package":"github.com/bitfield/script","Test":"TestFindFilesInNonexistentPathReturnsError","Elapsed":0.12}`
	want := testgox.Event{
		Action:  "pass",
		Package: "github.com/bitfield/script",
		Test:    "TestFindFilesInNonexistentPathReturnsError",
		Elapsed: 0.12,
	}

	got, err := testgox.ParseJSON(input)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestEventStringFormatsPassEventsWithATick(t *testing.T) {
	t.Parallel()
	input := testgox.Event{
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

func TestEventStringFormatsFailEventsWithACross(t *testing.T) {
	t.Parallel()
	input := testgox.Event{
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
	tcs := []testgox.Event{
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
	tcs := []testgox.Event{
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
