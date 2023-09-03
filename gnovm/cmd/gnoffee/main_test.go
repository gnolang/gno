package main

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/jaekwon/testify/require"
	"github.com/rogpeppe/go-internal/testscript"
)

func TestTest(t *testing.T) {
	testscript.Run(t, setupTestScript(t, "testdata"))
}

func setupTestScript(t *testing.T, txtarDir string) testscript.Params {
	t.Helper()
	// Get root location of github.com/gnolang/gno
	goModPath, err := exec.Command("go", "env", "GOMOD").CombinedOutput()
	require.NoError(t, err)
	rootDir := filepath.Dir(string(goModPath))
	// Build a fresh gno binary in a temp directory
	gnoffeeBin := filepath.Join(t.TempDir(), "gnoffee")
	err = exec.Command("go", "build", "-o", gnoffeeBin, filepath.Join(rootDir, "gnovm", "cmd", "gnoffee")).Run()
	require.NoError(t, err)
	// Define script params
	return testscript.Params{
		Setup: func(env *testscript.Env) error {
			return nil
		},
		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){
			// add a custom "gnoffee" command so txtar files can easily execute "gno"
			// without knowing where is the binary or how it is executed.
			"gnoffee": func(ts *testscript.TestScript, neg bool, args []string) {
				err := ts.Exec(gnoffeeBin, args...)
				if err != nil {
					ts.Logf("[%v]\n", err)
					if !neg {
						ts.Fatalf("unexpected gnoffee command failure")
					}
				} else {
					if neg {
						ts.Fatalf("unexpected gnoffee command success")
					}
				}
			},
		},
		Dir: txtarDir,
	}
}
