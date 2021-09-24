package tests

import (
	"bytes"
	"fmt"
	"regexp"

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

func TestFileStr(t *testing.T) {
	filePath := "./files/str.go"
	runCheck(t, filePath)
}

func TestFiles(t *testing.T) {
	baseDir := filepath.Join(".", "files")
	files, err := ioutil.ReadDir(baseDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".go" {
			continue
		}
		if testing.Short() && strings.Contains(file.Name(), "_long") {
			t.Log(fmt.Sprintf("skipping test %s in short mode.", file.Name()))
			continue
		}
		file := file
		t.Run(file.Name(), func(t *testing.T) {
			runCheck(t, filepath.Join(baseDir, file.Name()))
		})
	}
}

func runCheck(t *testing.T, path string) {
	pkgPath, resWanted, errWanted, rops := wantedFromComment(path)
	if pkgPath == "" {
		pkgPath = "main"
	}
	rlm := testRealm(pkgPath) // may be nil.
	pkgName := defaultPkgName(pkgPath)
	pn := gno.NewPackageNode(pkgName, pkgPath, &gno.FileSet{})
	pv := pn.NewPackage(rlm)

	var output = new(bytes.Buffer)
	m := gno.NewMachineWithOptions(gno.MachineOptions{
		Package: pv,
		Output:  output,
		Store:   testStore(output),
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
					err := strings.TrimSpace(fmt.Sprintf("%v", pnc))
					if !strings.Contains(err, errWanted) {
						// error didn't match: print stack
						// NOTE: will fail testcase later.
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
			// NOTE: ignores any gno.GetDebugErrors().
			gno.ClearDebugErrors()
			return // nothing more to do.
		} else {
			if pnc != nil {
				panic(fmt.Sprintf("got unexpected error: %v", pnc))
			}
			if gno.HasDebugErrors() {
				panic(fmt.Sprintf("got unexpected debug error(s): %v", gno.GetDebugErrors()))
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
			rops2 := strings.TrimSpace(rlm.SprintRealmOps())
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

func wantedFromComment(p string) (pkgPath, res, err, rops string) {
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
			goPath := strings.TrimSpace(strings.TrimPrefix(line, "GOPATH:"))
			panic(fmt.Sprintf(
				"GOPATH directive not supported -- move %s to extern",
				goPath))
		} else if strings.HasPrefix(text, "Output:\n") {
			res = strings.TrimPrefix(text, "Output:\n")
			res = strings.TrimSpace(res)
		} else if strings.HasPrefix(text, "Error:\n") {
			err = strings.TrimPrefix(text, "Error:\n")
			err = strings.TrimSpace(err)
			// XXX temporary until we support line:column.
			// If error starts with line:column, trim it.
			re := regexp.MustCompile(`^[0-9]+:[0-9]+: `)
			err = re.ReplaceAllString(err, "")
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

func testRealm(testPkgPath string) *gno.Realm {
	if gno.IsRealmPath(testPkgPath) {
		// Start blank test realm.
		rlm := gno.NewRealm(testPkgPath)
		// Store realm ops in rlm.
		rlm.SetLogRealmOps(true)
		return rlm
	} else {
		// shouldn't need to request a realm.
		return nil
	}
}
