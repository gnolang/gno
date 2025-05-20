package test_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/pkg/test"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

func TestBasic(t *testing.T) {
	rootDir := gnoenv.RootDir()
	mockOut := bytes.NewBufferString("")
	mockErr := bytes.NewBufferString("")

	io := commands.NewTestIO()
	io.SetOut(commands.WriteNopCloser(mockOut))
	io.SetErr(commands.WriteNopCloser(mockErr))
	opts := test.NewTestOptions(rootDir, io.Out(), io.Err())
	opts.Verbose = true
	if rootDir != opts.RootDir {
		t.Errorf("rootDir was %#v, (expected %#v)", opts.RootDir, rootDir)
	}

	subPkgs, err := gnomod.SubPkgsFromPaths([]string{"../../tests/integ/test/basic"})
	if err != nil {
		t.Errorf("list sub packages error: %#v", err.Error())
		return
	}
	if len(subPkgs) != 1 {
		t.Errorf("only expected one element: %#v", subPkgs)
		return
	}
	if len(subPkgs[0].TestGnoFiles) == 0 && len(subPkgs[0].FiletestGnoFiles) == 0 {
		t.Errorf("no test files found")
		return
	}
	// Determine gnoPkgPath by reading gno.mod
	modfile, _ := gnomod.ParseAt(subPkgs[0].Dir)
	if modfile == nil {
		t.Error("Unable to read package path from gno.mod or gno root directory")
		return
	}
	gnoPkgPath := modfile.Module.Mod.Path
	memPkg := gno.MustReadMemPackage(subPkgs[0].Dir, gnoPkgPath)
	err = test.Test(memPkg, subPkgs[0].Dir, opts)
	if err == nil {
		t.Error("Expected non-nil error. Got nil!")
	} else if err.Error() != "failed: \"TestBasic\"" {
		t.Errorf("results in error: %#v", err.Error())
	}
	stdErrStr := string(mockErr.Bytes())
	if !strings.Contains(stdErrStr, "PASS: TestBasic/greater_than_one") {
		t.Errorf("Expected test TestBasic/greater_than_one to pass, got: %#v", stdErrStr)
	}
	if !strings.Contains(stdErrStr, "FAIL: TestBasic/less_than_one") {
		t.Errorf("Expected test TestBasic/less_than_one to fail, got: %#v", stdErrStr)
	}
}
