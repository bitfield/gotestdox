package gotestdox_test

import (
	"os"
	"testing"

	"github.com/bitfield/gotestdox"
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
