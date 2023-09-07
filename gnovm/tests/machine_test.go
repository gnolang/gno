package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func TestMachineTestMemPackage(t *testing.T) {
	matchFunc := func(pat, str string) (bool, error) { return true, nil }

	tests := []struct {
		name          string
		path          string
		shouldSucceed bool
	}{
		{
			name:          "TestSuccess",
			path:          "testdata/TestMemPackage/success",
			shouldSucceed: true,
		},
		{
			name:          "TestFail",
			path:          "testdata/TestMemPackage/fail",
			shouldSucceed: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// NOTE: Because the purpose of this test is to ensure testing.T.Failed()
			// returns true if a gno test is failing, and because we don't want this
			// to affect the current testing.T, we are creating an other one thanks
			// to testing.RunTests() function.
			testing.RunTests(matchFunc, []testing.InternalTest{
				{
					Name: tt.name,
					F: func(t2 *testing.T) { //nolint:thelper
						rootDir := filepath.Join("..", "..")
						store := TestStore(rootDir, "test", os.Stdin, os.Stdout, os.Stderr, ImportModeStdlibsOnly)
						store.SetLogStoreOps(true)
						m := gno.NewMachineWithOptions(gno.MachineOptions{
							PkgPath: "test",
							Output:  os.Stdout,
							Store:   store,
							Context: nil,
						})
						memPkg := gno.ReadMemPackage(tt.path, "test")

						m.TestMemPackage(t2, memPkg)

						if tt.shouldSucceed {
							assert.False(t, t2.Failed(), "test %q should have succeed", tt.name)
						} else {
							assert.True(t, t2.Failed(), "test %q should have failed", tt.name)
						}
					},
				},
			})
		})
	}
}
