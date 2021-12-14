package tests

import (
	"bytes"
	"fmt"
	"log"
	"os"

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
	testDirs := map[string]string{} // aggregate here, dir -> pkgPath
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
				testDirs[filepath.Join(rootDir, dirPath)] = dirPath
			}
			return nil
		})
	}
	fmt.Println(testDirs)
	// For each package with testfiles (in testDirs), call Machine.TestMemPackage.
	for testDir, pkgPath := range testDirs {
		t.Run(pkgPath, func(t *testing.T) {
			runPackageTest(t, testDir, pkgPath)
		})
	}
}

func runPackageTest(t *testing.T, dir string, path string) {
	memPkg := gno.ReadMemPackage(dir, path)

	isRealm := false // XXX try true too?
	output := new(bytes.Buffer)
	store := testStore(output, isRealm)
	store.SetLogStoreOps(true)
	m := gno.NewMachineWithOptions(gno.MachineOptions{
		Package: nil,
		Output:  output,
		Store:   store,
		Context: nil,
	})
	m.TestMemPackage(memPkg)

	// Check that machine is empty.
	err := m.CheckEmpty()
	if err != nil {
		t.Log("last state: \n", m.String())
		panic(fmt.Sprintf("machine not empty after main: %v", err))
	}
}
