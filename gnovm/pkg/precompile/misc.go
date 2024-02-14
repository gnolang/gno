package precompile

import (
	"fmt"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"go/ast"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func IsGnoFile(f fs.DirEntry) bool {
	name := f.Name()
	return !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".gno") && !f.IsDir()
}

// TODO: move to misc
// GetPrecompileFilenameAndTags returns the filename and tags for precompiled files.
func GetPrecompileFilenameAndTags(gnoFilePath string) (targetFilename, tags string) {
	nameNoExtension := strings.TrimSuffix(filepath.Base(gnoFilePath), ".gno")
	switch {
	case strings.HasSuffix(gnoFilePath, "_filetest.gno"):
		tags = "gno && filetest"
		targetFilename = "." + nameNoExtension + ".gno.gen.go"
	case strings.HasSuffix(gnoFilePath, "_test.gno"):
		tags = "gno && test"
		targetFilename = "." + nameNoExtension + ".gno.gen_test.go"
	default:
		tags = "gno"
		targetFilename = nameNoExtension + ".gno.gen.go"
	}
	return
}

func GnoFilesFromArgs(args []string) ([]string, error) {
	fmt.Println("---GnoFilesFromArgs, args: ", args)
	paths := []string{}
	for _, arg := range args {
		info, err := os.Stat(arg)
		if err != nil {
			return nil, fmt.Errorf("invalid file or package path: %w", err)
		}
		if !info.IsDir() {
			curpath := arg
			paths = append(paths, curpath)
		} else {
			err = filepath.WalkDir(arg, func(curpath string, f fs.DirEntry, err error) error {
				if err != nil {
					return fmt.Errorf("%s: walk dir: %w", arg, err)
				}

				if !IsGnoFile(f) {
					return nil // skip
				}
				paths = append(paths, curpath)
				return nil
			})
			if err != nil {
				return nil, err
			}
		}
	}
	return paths, nil
}

func GnoPackagesFromArgs(args []string) ([]string, error) {
	paths := []string{}
	for _, arg := range args {
		info, err := os.Stat(arg)
		if err != nil {
			return nil, fmt.Errorf("invalid file or package path: %w", err)
		}
		if !info.IsDir() {
			paths = append(paths, arg)
		} else {
			// if the passed arg is a dir, then we'll recursively walk the dir
			// and look for directories containing at least one .gno file.

			visited := map[string]bool{} // used to run the builder only once per folder.
			err = filepath.WalkDir(arg, func(curpath string, f fs.DirEntry, err error) error {
				if err != nil {
					return fmt.Errorf("%s: walk dir: %w", arg, err)
				}
				if f.IsDir() {
					return nil // skip
				}
				if !IsGnoFile(f) {
					return nil // skip
				}

				parentDir := filepath.Dir(curpath)
				if _, found := visited[parentDir]; found {
					return nil
				}
				visited[parentDir] = true

				pkg := parentDir
				if !filepath.IsAbs(parentDir) {
					// cannot use path.Join or filepath.Join, because we need
					// to ensure that ./ is the prefix to pass to go build.
					// if not absolute.
					pkg = "./" + parentDir
				}

				paths = append(paths, pkg)
				return nil
			})
			if err != nil {
				return nil, err
			}
		}
	}
	return paths, nil
}

func getParentDirectory(cwd string, generations int) (string, error) {
	if generations <= 0 {
		return cwd, nil
	}

	parentDir := filepath.Dir(cwd)
	return getParentDirectory(parentDir, generations-1)
}

// getPathsFromImportSpec derive and returns ImportPaths
// without ImportPrefix from *ast.ImportSpec
func getPathsFromImportSpec(importSpec []*ast.ImportSpec) (importPaths []ImportPath) {
	// Get the current working directory
	//cwd, err := os.Getwd()
	//if err != nil {
	//	fmt.Println("Error:", err)
	//	return
	//}
	//
	//// Specify the number of generations
	//generations := 1
	//
	//// Get the parent directory recursively
	//rootAbsDir, err := getParentDirectory(cwd, generations)
	//if err != nil {
	//	fmt.Println("Error:", err)
	//	return
	//}
	//
	//fmt.Printf("rootAbsDir (generations=%d): %s\n", generations, rootAbsDir)

	for _, i := range importSpec {
		path := i.Path.Value[1 : len(i.Path.Value)-1] // trim leading and trailing `"`
		fmt.Println("---getPathsFromImportSpec:", path)
		abs, err := filepath.Abs("")
		if err != nil {
			panic(err)
		}
		fmt.Println("---dir of path: ", filepath.Dir(path))
		fmt.Println("---abs dir is: ", abs)

		if strings.HasPrefix(path, testPath) {
			res := strings.TrimPrefix(path, testPath)
			importPaths = append(importPaths, ImportPath(abs+"/files/extern/"+res)) // build abs path in for go exec
		} else if strings.HasPrefix(path, ImportPrefix) {
			res := strings.TrimPrefix(path, ImportPrefix)
			importPaths = append(importPaths, ImportPath(gnoenv.RootDir()+res)) // build abs path in for go exec
		}
	}
	return
}

// resolvePath joins the output dir with relative pkg path
// e.g
// Output Dir: Temp/gno-precompile
// Pkg Path: ../example/gno.land/p/pkg
// Returns -> Temp/gno-precompile/example/gno.land/p/pkg
func resolvePath(output string, path ImportPath) (string, error) {
	fmt.Println("---resolvePath, output: , path: ", output, path)
	absOutput, err := filepath.Abs(output)
	if err != nil {
		return "", err
	}
	//absOutput = "/Users/maxwell/workbench/blockchain/gno/gno_forks/gno"
	absPkgPath, err := filepath.Abs(string(path))
	if err != nil {
		return "", err
	}

	fmt.Println("---absOutput: ", absOutput)
	fmt.Println("---absPkgPath: ", absPkgPath)
	fmt.Println("---root: ", gnoenv.RootDir())
	pkgPath := strings.TrimPrefix(absPkgPath, gnoenv.RootDir())

	fmt.Println("---pkgPath: ", pkgPath)
	//return filepath.Join(absOutput, pkgPath), nil
	return filepath.Join(gnoenv.RootDir(), pkgPath), nil // build abs path for to exec
}

// writeDirFile write file to the path and also create
// directory if needed. with:
// Dir perm -> 0755; File perm -> 0o644
func writeDirFile(pathWithName string, data []byte) error {
	path := filepath.Dir(pathWithName)

	// Create Dir if not exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, 0o755)
	}

	return os.WriteFile(pathWithName, data, 0o644)
}

func CleanGeneratedFiles(dir string) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		// Ignore if not a generated file
		if !strings.HasSuffix(path, ".gno.gen.go") && !strings.HasSuffix(path, ".gno.gen_test.go") {
			return nil
		}
		if err := os.Remove(path); err != nil {
			return err
		}

		return nil
	})
}
