package interp_test

import (
	"bytes"
	"fmt"

	//"go/build"
	"go/parser"
	"go/token"
	"io/ioutil"
	"path/filepath"
	rtdb "runtime/debug"
	"strings"
	"testing"

	"github.com/gnolang/gno"
)

func TestFile(t *testing.T) {
	filePath := "./files/str.go"
	runCheck(t, filePath)

	baseDir := filepath.Join(".", "files")
	files, err := ioutil.ReadDir(baseDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".go" {
			continue
		}
		file := file
		t.Run(file.Name(), func(t *testing.T) {
			runCheck(t, filepath.Join(baseDir, file.Name()))
		})
	}
}

func runCheck(t *testing.T, path string) {
	pkgPath, goPath, resWanted, errWanted, rops := wantedFromComment(path)
	if pkgPath == "" {
		pkgPath = "main"
	}
	realmer := testRealmer(pkgPath) // may be nil.
	pkgName := defaultPkgName(pkgPath)
	if goPath != "" {
		// See original Yaegi repo;
		// used to import the files in the goPath
		// to be imported in testfiles.
		panic("TODO")
	}
	pn := gno.NewPackageNode(pkgName, pkgPath, &gno.FileSet{})
	pv := pn.NewPackage(realmer)

	var output = new(bytes.Buffer)
	m := gno.NewMachineWithOptions(gno.MachineOptions{
		Package:  pv,
		Output:   output,
		Importer: testImporter(output),
	})
	// TODO support stdlib groups, but make testing safe;
	// e.g. not be able to make network connections.
	// interp.New(interp.Options{GoPath: goPath, Stdout: &stdout, Stderr: &stderr})
	// m.Use(interp.Symbols)
	// m.Use(stdlib.Symbols)
	// m.Use(unsafe.Symbols)
	bz, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
	{ // Validate result, errors, etc.
		var pnc interface{}
		func() {
			defer func() {
				if r := recover(); r != nil {
					pnc = r
					if errWanted == "" {
						// unexpected: print stack.
						rtdb.PrintStack()
					}
				}
			}()
			n := gno.MustParseFile(path, string(bz))
			m.RunFiles(n)
			if rops != "" {
				// clear rlm.ropslog from init funtion(s).
				rlm := pv.GetRealm()
				rlm.SetLogRealmOps(true) // resets.
			}
			m.RunMain()
		}()
		// check errors
		if errWanted != "" {
			if pnc == nil {
				panic(fmt.Sprintf("got nil error, want: %q", errWanted))
			}
			err := strings.TrimSpace(fmt.Sprintf("%v", pnc))
			if !strings.Contains(err, errWanted) {
				panic(fmt.Sprintf("got %q, want: %q", err, errWanted))
			}
		} else {
			if pnc != nil {
				panic(fmt.Sprintf("got unexpected error: %v", pnc))
			}
		}
		// check result
		res := strings.TrimSpace(output.String())
		if resWanted != "" {
			if res != resWanted {
				// panic so tests immediately fail (for now).
				panic(fmt.Sprintf("got:\n%s\n\nwant:\n%s\n", res, resWanted))
			}
		} else {
			if res != "" {
				panic(fmt.Sprintf("got unexpected output: %s", res))
			}
		}
		// check realm ops
		if rops != "" {
			rlm := pv.GetRealm()
			if rlm == nil {
				panic("expected realm but got none")
			}
			rops2 := rlm.SprintRealmOps()
			if rops != rops2 {
				panic(fmt.Sprintf("got:\n%s\n\nwant:\n%s\n", rops2, rops))
			}
		}
	}

	// Check that machine is empty.
	err = m.CheckEmpty()
	if err != nil {
		t.Log("last state: \n", m.String())
		panic(fmt.Sprintf("machine not empty after main: %v", err))
	}
}

func wantedFromComment(p string) (pkgPath, goPath, res, err, rops string) {
	fset := token.NewFileSet()
	f, err2 := parser.ParseFile(fset, p, nil, parser.ParseComments)
	if err2 != nil {
		panic(err2)
	}
	if len(f.Comments) == 0 {
		return
	}
	for _, comments := range f.Comments {
		text := comments.Text()
		if strings.HasPrefix(text, "PKGPATH:") {
			line := strings.SplitN(text, "\n", 2)[0]
			pkgPath = strings.TrimSpace(strings.TrimPrefix(line, "PKGPATH:"))
		} else if strings.HasPrefix(text, "GOPATH:") {
			line := strings.SplitN(text, "\n", 2)[0]
			goPath = strings.TrimSpace(strings.TrimPrefix(line, "GOPATH:"))
		} else if strings.HasPrefix(text, "Output:\n") {
			res = strings.TrimPrefix(text, "Output:\n")
			res = strings.TrimSpace(res)
		} else if strings.HasPrefix(text, "Error:\n") {
			err = strings.TrimPrefix(text, "Error:\n")
			err = strings.TrimSpace(err)
		} else if strings.HasPrefix(text, "Realm:\n") {
			rops = strings.TrimPrefix(text, "Realm:\n")
			rops = strings.TrimSpace(rops)
		} else {
			// ignore unexpected.
		}
	}
	return
}

func defaultPkgName(gopkgPath string) gno.Name {
	parts := strings.Split(gopkgPath, "/")
	last := parts[len(parts)-1]
	parts = strings.Split(last, "-")
	name := parts[len(parts)-1]
	name = strings.ToLower(name)
	return gno.Name(name)
}

func testRealmer(testPkgPath string) gno.Realmer {
	if gno.IsRealmPath(testPkgPath) {
		// Start blank test realm.
		rlm := gno.NewRealm(testPkgPath)
		// Store realm ops in rlm.
		rlm.SetLogRealmOps(true)
		return gno.Realmer(func(pkgPath string) *gno.Realm {
			if pkgPath == testPkgPath {
				return rlm
			} else {
				panic("should not happen")
			}
		})
	} else {
		// shouldn't need to request a realm.
		return nil
	}
}
