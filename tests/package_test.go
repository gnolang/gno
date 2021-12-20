package tests

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"sort"

	//"go/build"

	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gnolang/gno"
)

func TestPackages(t *testing.T) {
	// find all packages with *_test.go files.
	rootDirs := []string{
		filepath.Join("..", "examples"),
		filepath.Join("..", "stdlibs"),
	}
	testDirs := map[string]string{} // aggregate here, pkgPath -> dir
	pkgPaths := []string{}
	for _, rootDir := range rootDirs {
		fileSystem := os.DirFS(rootDir)
		fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				log.Fatal(err)
			}
			if d.IsDir() {
				return nil
			}
			if strings.HasSuffix(path, "_test.go") {
				dirPath := filepath.Dir(path)
				if _, exists := testDirs[dirPath]; exists {
					// already exists.
				} else {
					testDirs[dirPath] = filepath.Join(rootDir, dirPath)
					pkgPaths = append(pkgPaths, dirPath)
				}
			}
			return nil
		})
	}
	// Sort pkgPaths for determinism.
	sort.Strings(pkgPaths)
	// For each package with testfiles (in testDirs), call Machine.TestMemPackage.
	for _, pkgPath := range pkgPaths {
		testDir := testDirs[pkgPath]
		t.Run(pkgPath, func(t *testing.T) {
			runPackageTest(t, testDir, pkgPath)
		})
	}
}

func runPackageTest(t *testing.T, dir string, path string) {
	memPkg := gno.ReadMemPackage(dir, path)

	isRealm := false // XXX try true too?
	stdin := new(bytes.Buffer)
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	store := testStore(stdin, stdout, stderr, isRealm, false)
	store.SetLogStoreOps(true)
	m := gno.NewMachineWithOptions(gno.MachineOptions{
		Package: nil,
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
