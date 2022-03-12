package gotestdox_test

import (
	"fmt"
	"testing"

	"github.com/bitfield/gotestdox"
	"github.com/google/go-cmp/cmp"
)

func TestPrettify(t *testing.T) {
	t.Parallel()
	for _, tc := range Cases {
		t.Run(tc.name, func(t *testing.T) {
			got := gotestdox.Prettify(tc.input)
			if tc.want != got {
				t.Errorf("%s:\ninput: %q:\nresult: %s", tc.name, tc.input, cmp.Diff(tc.want, got))
			}
		})
	}
}

func BenchmarkPrettify(b *testing.B) {
	input := "TestParseJSON_CorrectlyParsesASingleGoTestJSONOutputLine"
	for i := 0; i < b.N; i++ {
		_ = gotestdox.Prettify(input)
	}
}

func ExamplePrettify() {
	input := "TestFoo/has_well-formed_output"
	fmt.Println(gotestdox.Prettify(input))
	// Output:
	// Foo has well-formed output
}

func ExamplePrettify_underscoreHint() {
	input := "TestHandleInput_ClosesInputAfterReading"
	fmt.Println(gotestdox.Prettify(input))
	// Output:
	// HandleInput closes input after reading
}

var Cases = []struct {
	name, input, want string
}{
	{
		name:  "accepts a single-letter test name",
		input: "TestS",
		want:  "S",
	},
	{
		name:  "accepts a single-word test name",
		input: "TestSum",
		want:  "Sum",
	},
	{
		name:  "replaces camel-case transitions with spaces",
		input: "TestSumCorrectlySumsInputNumbers",
		want:  "Sum correctly sums input numbers",
	},
	{
		name:  "preserves capitalisation of initialisms such as PDF",
		input: "TestFooGeneratesValidPDFFile",
		want:  "Foo generates valid PDF file",
	},
	{
		name:  "does not hang when name ends with initialism",
		input: "TestFooGeneratesValidPDF",
		want:  "Foo generates valid PDF",
	},
	{
		name:  "preserves capitalisation of initialism when it is the first word",
		input: "TestJSONSucks",
		want:  "JSON sucks",
	},
	{
		name:  "preserves capitalisation of two-letter initialisms such as OK",
		input: "TestFilterReturnsOKIfThereAreNoTestFailures",
		want:  "Filter returns OK if there are no test failures",
	},
	{
		name:  "preserves longer all-caps words",
		input: "TestCategoryTrimsLEADINGSpacesFromValidCategory",
		want:  "Category trims LEADING spaces from valid category",
	},
	{
		name:  "treats numbers as word separators",
		input: "TestFooDoes8Things",
		want:  "Foo does 8 things",
	},
	{
		name:  "keeps a trailing digit as part of an initialism",
		input: "TestFooGeneratesUTF8Correctly",
		want:  "Foo generates UTF8 correctly",
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
		name:  "treats consecutive underscores as a single word break",
		input: "Test_Foo__Works",
		want:  "Foo works",
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
		name:  "treats a single underscore as marking the end of a multiword function name",
		input: "TestFindFiles_WorksCorrectly",
		want:  "FindFiles works correctly",
	},
	{
		name:  "retains capitalisation of initialisms in a multiword function name",
		input: "TestParseJSON_CorrectlyParsesASingleGoTestJSONOutputLine",
		want:  "ParseJSON correctly parses a single go test JSON output line",
	},
	{
		name:  "treats a single underscore before the first slash as marking the end of a multiword function name",
		input: "TestFindFiles_/WorksCorrectly",
		want:  "FindFiles works correctly",
	},
	{
		name:  "handles multiple underscores, with the first marking the end of a multiword function name",
		input: "TestFindFiles_Does_Stuff",
		want:  "FindFiles does stuff",
	},
	{
		name:  "does not treat an underscore in a subtest name as marking the end of a multiword function name",
		input: "TestCallingTheFunction/Does_Stuff",
		want:  "Calling the function does stuff",
	},
	{
		name:  "eliminates any words containing underscores after splitting",
		input: "TestSentence/does_x,_correctly",
		want:  "Sentence does x, correctly",
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
	{
		name:  "does not erase the final digit in words that end with a digit",
		input: "TestExtractFiles/Truncated_bzip2_which_will_return_an_error",
		want:  "Extract files truncated bzip 2 which will return an error",
	},
	// {
	// 	name:  "recognises a dash followed by a digit as a negative number",
	// 	input: "TestColumnSelects/column_-1_of_input",
	// 	want:  "Column selects column -1 of input",
	// },
	// {
	// 	name:  "keeps numbers within a hyphenated word",
	// 	input: "TestReadExtended/nyc-taxi-data-100k.csv",
	// 	want:  "Read extended nyc-taxi-data-100k.csv",
	// },
	// {
	// 	name:  "keeps together hyphenated words with initial capitals",
	// 	input: "TestListObjectsVersionedFolders/Erasure-Test",
	// 	want:  "List objects versioned folders erasure-test",
	// },
	// {
	// 	name:  "keeps together hyphenated words with initialisms",
	// 	input: "TestListObjects/FS-Test71",
	// 	want:  "List objects FS-test 71",
	// },
	// {
	// 	name:  "keeps together digits in numbers that are standalone words",
	// 	input: "TestLex11",
	// 	want:  "Lex 11",
	// },
	// {
	// 	name:  "handles a test with no name, but with subtests",
	// 	input: "Test/default/issue12839",
	// 	want:  "Default issue 12839",
	// },
	// {
	// 	name:  "does not break words when a digit follows an = sign",
	// 	input: "TestUniformFactorial/n=3",
	// 	want:  "Uniform factorial n=3",
	// },
	// {
	// 	name:  "preserves initialisms containing digits",
	// 	input: "TestS390XOperandParser",
	// 	want:  "S390X operand parser",
	// },
	// {
	// 	name:  "preserves initialisms containing digits with two or more leading alpha characters",
	// 	input: "TestBC35A",
	// 	want:  "BC35A",
	// },
}
