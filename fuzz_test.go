//go:build go1.18

package gotestdox_test

import (
	"strings"
	"testing"
	"unicode"

	"github.com/bitfield/gotestdox"
)

func FuzzPrettify(f *testing.F) {
	for _, tc := range Cases {
		f.Add(tc.input)
	}
	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > 0 && unicode.IsLower([]rune(input)[0]) {
			t.Skip()
		}
		got := gotestdox.Prettify(input)
		if got == "" {
			t.Skip()
		}
		if strings.ContainsRune(got, '_') {
			t.Errorf("%q: contains underscore %q", input, got)
		}
		if strings.ContainsRune(got, '/') {
			t.Errorf("%q: contains slash %q", input, got)
		}
	})
}
