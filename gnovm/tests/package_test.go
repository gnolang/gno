package tests

import (
	"bytes"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs"
)

func TestStdlibs(t *testing.T) {
	t.Parallel()

	// NOTE: this test only works using _test.gno files;
	// filetests are not meant to be used for testing standard libraries.
	// The examples directory is tested directly using `gno test`u

	// find all packages with *_test.gno files.
	srcs := stdlibs.EmbeddedSources()
	rootDirs, err := fs.ReadDir(srcs, ".")
	require.NoError(t, err)
	pkgPaths := []string{}
	for _, rootDir := range rootDirs {
		fileSystem, err := fs.Sub(srcs, rootDir.Name())
		require.NoError(t, err)
		fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				log.Fatal(err)
			}
			if d.IsDir() {
				return nil
			}

			path = filepath.Join(rootDir.Name(), path)
			if strings.HasSuffix(path, "_test.gno") {
				dirPath := filepath.Dir(path)
				if slices.Contains(pkgPaths, dirPath) {
					// already exists.
				} else {
					pkgPaths = append(pkgPaths, dirPath)
				}
			}
			return nil
		})
	}
	// For each package with testfiles (in testPaths), call Machine.TestMemPackage.
	for _, pkgPath := range pkgPaths {
		t.Run(pkgPath, func(t *testing.T) {
			t.Parallel()
			runPackageTest(t, pkgPath)
		})
	}
}

func runPackageTest(t *testing.T, pkgPath string) {
	t.Helper()

	srcs := stdlibs.EmbeddedSources()
	memPkg := gno.ReadMemPackageFromFS(srcs, pkgPath, pkgPath)
	require.False(t, memPkg.IsEmpty())

	stdin := new(bytes.Buffer)
	// stdout := new(bytes.Buffer)
	stdout := os.Stdout
	stderr := new(bytes.Buffer)
	rootDir := filepath.Join("..", "..")
	store := TestStore(rootDir, pkgPath, stdin, stdout, stderr, ImportModeStdlibsOnly)
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
