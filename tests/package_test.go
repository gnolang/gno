package tests

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestPackages(t *testing.T) {
	// find all packages with *_test.gno files.
	rootDirs := []string{
		filepath.Join("..", "stdlibs"),
		filepath.Join("..", "examples"),
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
			if strings.HasSuffix(path, "_test.gno") {
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
			err := RunPackageTest(t, testDir, pkgPath)
			if err != nil {
				panic(fmt.Sprintf("machine not empty after main: %v", err))
			}
		})
	}
}
