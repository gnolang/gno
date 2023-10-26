package tests

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	// "go/build"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnoutil"
	"github.com/jaekwon/testify/require"
)

func TestPackages(t *testing.T) {
	// find all packages with *_test.gno files.
	rootDirs := []string{
		filepath.Join("..", "stdlibs"),
		filepath.Join("..", "..", "examples"),
	}
	dirs, err := gnoutil.Match([]string{
		filepath.Join(rootDirs[0], "..."),
		filepath.Join(rootDirs[1], "..."),
	}, gnoutil.MatchPackages("*_test.gno"))
	if err != nil {
		t.Fatal(err)
	}
	paths := make([]string, len(dirs))
	for i, pkg := range dirs {
		for _, pref := range rootDirs {
			pkg = strings.TrimPrefix(pkg, pref)
		}
		pkg = strings.ReplaceAll(pkg, string(os.PathSeparator), "/")
		paths[i] = pkg
	}

	// For each package with testfiles (in testDirs), call Machine.TestMemPackage.
	for i, pkgPath := range paths {
		t.Run(pkgPath, func(t *testing.T) {
			runPackageTest(t, dirs[i], pkgPath)
		})
	}
}

func runPackageTest(t *testing.T, dir string, path string) {
	t.Helper()

	memPkg := gno.ReadMemPackage(dir, path)
	require.False(t, memPkg.IsEmpty())

	stdin := new(bytes.Buffer)
	// stdout := new(bytes.Buffer)
	stdout := os.Stdout
	stderr := new(bytes.Buffer)
	rootDir := filepath.Join("..", "..")
	store := TestStore(rootDir, path, stdin, stdout, stderr, ImportModeStdlibsOnly)
	store.SetLogStoreOps(true)
	m := gno.NewMachineWithOptions(gno.MachineOptions{
		PkgPath: "test",
		Output:  stdout,
		Store:   store,
		Context: nil,
	})
	m.TestMemPackage(t, memPkg)

	// Check that machine is empty.
	err := m.CheckEmpty()
	if err != nil {
		t.Log("last state: \n", m.String())
		panic(fmt.Sprintf("machine not empty after main: %v", err))
	}
}
