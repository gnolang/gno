// integration_test.go - Not Gno, and ignored onchain
// This file is for advanced contract testing not managed purely with .gno files.
// It enables multi-transaction integration tests and features currently unsupported by Gno.
// We may replace or deprecate it in the future as Gno evolves. Testing purposes only.
package banktest

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs"
	"github.com/gnolang/gno/gnovm/tests"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func TestSimpleFlow(t *testing.T) {
	var (
		mode    = tests.ImportModeStdlibsOnly
		rootDir = filepath.Join("..", "..", "..", "..", "..")
		stdin   = os.Stdin
		stdout  = os.Stdout
		stderr  = os.Stderr
		store   = tests.TestStore(rootDir, "", stdin, stdout, stderr, mode)
	)
	store.SetStrictGo2GnoMapping(true) // natives must be registered
	gnolang.DisableDebug()             // until main call
	m := tests.TestMachine(store, stdout, "main")
	pkgName := "main"
	pkgPath := "gno.land/r/demo/banktest_test"
	/*
		pkgPath := "./"
		memPkg := gnolang.ReadMemPackage(".", ".")
		files := &gnolang.FileSet{}
		for _, mfile := range memPkg.Files {
			if !strings.HasSuffix(mfile.Name, ".gno") {
				continue
			}
			if strings.HasSuffix(mfile.Name, "_filetest.gno") {
				continue
			}
			n, err := gnolang.ParseFile(mfile.Name, mfile.Body)
			if err != nil {
				t.Fatalf("parsing %q: %v", mfile.Name, err)
			}
			files.AddFiles(n)
		}
		// TODO: use temporary stdout, stderr for assertion
		m.RunMemPackage(memPkg, true)
	*/

	// main1
	if true {
		ctx := m.Context.(stdlibs.ExecContext)
		ctx.OrigSend = std.MustParseCoins("1234500000ugnot")
		m.Context = ctx
		println("main1")
		memPkg := &std.MemPackage{
			Name: pkgName,
			Path: pkgPath,
			Files: []*std.MemFile{
				{
					Name: "main1.gno",
					Body: main1,
				},
			},
		}
		m.RunMemPackage(memPkg, true)
		store.ClearCache()
		m.PreprocessAllFilesAndSaveBlockNodes()
		// store.Print() // debug
		pv2 := store.GetPackage(pkgPath, false)
		m.SetActivePackage(pv2)
		gnolang.EnableDebug()
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("main1 panic: %v\n%s\n", r, m.String())
				panic(r)
			}
		}()
		m.RunStatement(gnolang.S(gnolang.Call(gnolang.X("main1"))))
		m.CheckEmpty()
	}
	// main2
	if true {
		ctx := m.Context.(stdlibs.ExecContext)
		ctx.OrigSend = std.MustParseCoins("12345ugnot")
		ctx.Height++
		ctx.Timestamp++
		m.Context = ctx

		println("main2")
		memPkg := &std.MemPackage{
			Name: pkgName,
			Path: pkgPath,
			Files: []*std.MemFile{
				{
					Name: "main2.gno",
					Body: main2,
				},
			},
		}
		m.RunMemPackage(memPkg, true)
		store.ClearCache()
		m.PreprocessAllFilesAndSaveBlockNodes()
		// store.Print() // debug
		//pv2 := store.GetPackage(pkgPath, false)
		//m.SetActivePackage(pv2)
		gnolang.EnableDebug()
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("main2 panic: %v\n%s\n", r, m.String())
				panic(r)
			}
		}()
		m.RunStatement(gnolang.S(gnolang.Call(gnolang.X("main2"))))
		m.CheckEmpty()
	}
	println("end")

}

const main1 = `package main

import "gno.land/r/demo/banktest"
import "std"

func main1() {
	banktest.Deposit("ugnot", 0)
	println(banktest.Render(""))
	/* std.TestSetCaller */
	printStats("main1")
}

func printStats(name string) {
	println("stats", name, std.GetHeight(), std.GetOrigCaller())
}
`

const main2 = `package main

import "gno.land/r/demo/banktest"
import "std"

func main2() {
	banktest.Deposit("ugnot", 0)
	/* std.TestSetCaller... */
	println(banktest.Render(""))
	printStats2("main2")
}

func printStats2(name string) {
	println("stats", name, std.GetHeight(), std.GetOrigCaller())
}
`
