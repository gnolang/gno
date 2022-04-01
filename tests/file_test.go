package tests

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"

	//"go/build"
	"go/parser"
	"go/token"
	"io/ioutil"
	"path/filepath"
	rtdb "runtime/debug"
	"strings"
	"testing"

	"github.com/gnolang/gno"
	"github.com/gnolang/gno/pkgs/crypto"
	"github.com/gnolang/gno/pkgs/sdk/testutils"
	"github.com/gnolang/gno/pkgs/std"
	"github.com/gnolang/gno/stdlibs"
)

func TestFileStr(t *testing.T) {
	filePath := "./files/str.go"
	runFileTest(t, filePath, true)
}

// Bootstrapping test files from tests/files/*.go,
// which primarily uses native stdlib shims.
func TestFiles1(t *testing.T) {
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
			runFileTest(t, filepath.Join(baseDir, file.Name()), true)
		})
	}
}

// Like TestFiles1(), but with more full-gno stdlib packages.
func TestFiles2(t *testing.T) {
	baseDir := filepath.Join(".", "files2")
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
			runFileTest(t, filepath.Join(baseDir, file.Name()), false)
		})
	}
}

func runFileTest(t *testing.T, path string, nativeLibs bool) {
	pkgPath, resWanted, errWanted, rops, maxAlloc, send := wantedFromComment(path)
	if pkgPath == "" {
		pkgPath = "main"
	}
	pkgName := defaultPkgName(pkgPath)
	if send == "" {
		send = "" // send nothing.
	}
	stdin := new(bytes.Buffer)
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	store := testStore(stdin, stdout, stderr, nativeLibs)
	store.SetLogStoreOps(true)
	caller := testutils.TestBech32Address("testaddr____________")
	txSend := std.MustParseCoins(send)
	pkgCoins := std.MustParseCoins("200gnot") // >= txSend.
	pkgAddr := testutils.TestBech32Address("packageaddr_________")
	banker := newtestBanker(pkgAddr, pkgCoins)
	ctx := stdlibs.ExecContext{
		ChainID:     "testchain",
		Height:      123,
		Msg:         nil,
		OrigCaller:  caller,
		OrigPkgAddr: pkgAddr,
		TxSend:      txSend,
		TxSendSpent: new(std.Coins),
		Banker:      banker,
	}
	m := gno.NewMachineWithOptions(gno.MachineOptions{
		PkgPath:       "", // set later.
		Output:        stdout,
		Store:         store,
		Context:       ctx,
		MaxAllocBytes: maxAlloc,
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
						// unexpected: print output.
						fmt.Println("OUTPUT:\n", stdout.String())
						// print stack.
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
			if gno.IsDebug() && testing.Verbose() {
				t.Log("========================================")
				t.Log("RUN FILES & INIT")
				t.Log("========================================")
			}
			if !gno.IsRealmPath(pkgPath) {
				// simple case.
				pn := gno.NewPackageNode(pkgName, pkgPath, &gno.FileSet{})
				pv := pn.NewPackage()
				store.SetBlockNode(pn)
				store.SetCachePackage(pv)
				m.SetActivePackage(pv)
				n := gno.MustParseFile(path, string(bz)) // "main.go", string(bz))
				m.RunFiles(n)
				if gno.IsDebug() && testing.Verbose() {
					t.Log("========================================")
					t.Log("RUN MAIN")
					t.Log("========================================")
				}
				m.RunMain()
				if gno.IsDebug() && testing.Verbose() {
					t.Log("========================================")
					t.Log("RUN MAIN END")
					t.Log("========================================")
				}
			} else {
				// realm case.
				gno.DisableDebug() // until main call.
				// save package using realm crawl procedure.
				memPkg := &std.MemPackage{
					Name: string(pkgName),
					Path: pkgPath,
					Files: []*std.MemFile{
						{
							Name: "main.go", // dontcare
							Body: string(bz),
						},
					},
				}
				m.RunMemPackage(memPkg, true)
				// reconstruct machine and clear store cache.
				// whether package is realm or not, since non-realm
				// may call realm packages too.
				if gno.IsDebug() && testing.Verbose() {
					t.Log("========================================")
					t.Log("CLEAR STORE CACHE")
					t.Log("========================================")
				}
				store.ClearCache()
				/*
					m = gno.NewMachineWithOptions(gno.MachineOptions{
						PkgPath:       "",
						Output:        stdout,
						Store:         store,
						Context:       ctx,
						MaxAllocBytes: maxAlloc,
					})
				*/
				if gno.IsDebug() && testing.Verbose() {
					store.Print()
					t.Log("========================================")
					t.Log("PREPROCESS ALL FILES")
					t.Log("========================================")
				}
				m.PreprocessAllFilesAndSaveBlockNodes()
				if gno.IsDebug() && testing.Verbose() {
					t.Log("========================================")
					t.Log("RUN MAIN")
					t.Log("========================================")
					store.Print()
				}
				pv2 := store.GetPackage(pkgPath, false)
				m.SetActivePackage(pv2)
				gno.EnableDebug()
				if rops != "" {
					// clear store.opslog from init funtion(s),
					// and PreprocessAllFilesAndSaveBlockNodes().
					store.SetLogStoreOps(true) // resets.
				}
				m.RunMain()
				if gno.IsDebug() && testing.Verbose() {
					t.Log("========================================")
					t.Log("RUN MAIN END")
					t.Log("========================================")
				}
			}
		}()
		// check errors
		if errWanted != "" {
			if pnc == nil {
				panic(fmt.Sprintf("got nil error, want: %q", errWanted))
			}
			errstr := ""
			if tv, ok := pnc.(*gno.TypedValue); ok {
				errstr = tv.Sprint(m)
			} else {
				errstr = strings.TrimSpace(fmt.Sprintf("%v", pnc))
			}
			if !strings.Contains(errstr, errWanted) {
				panic(fmt.Sprintf("got %q, want: %q", errstr, errWanted))
			}
			// NOTE: ignores any gno.GetDebugErrors().
			gno.ClearDebugErrors()
			return // nothing more to do.
		} else {
			if pnc != nil {
				if tv, ok := pnc.(*gno.TypedValue); ok {
					panic(fmt.Sprintf("got unexpected error: %s", tv.Sprint(m)))
				} else { // TODO: does this happen?
					panic(fmt.Sprintf("got unexpected error: %v", pnc))
				}
			}
			if gno.HasDebugErrors() {
				panic(fmt.Sprintf("got unexpected debug error(s): %v", gno.GetDebugErrors()))
			}
		}
		// check result
		res := strings.TrimSpace(stdout.String())
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
			rops2 := strings.TrimSpace(store.SprintStoreOps())
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

func wantedFromComment(p string) (pkgPath, res, err, rops string, maxAlloc int64, send string) {
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
			// NOTE: this should be the first comment in a file test.
			lines := strings.SplitN(text, "\n", 2)
			line0 := lines[0]
			pkgPath = strings.TrimSpace(strings.TrimPrefix(line0, "PKGPATH:"))
			// This comment block can contain additional lines beneath PKGPATH.
			for _, line := range lines[1:] {
				if strings.HasPrefix(line, "SEND:") {
					send = strings.TrimSpace(strings.TrimPrefix(line, "SEND:"))
				}
			}
		} else if strings.HasPrefix(text, "MAXALLOC:") {
			line := strings.SplitN(text, "\n", 2)[0]
			maxstr := strings.TrimSpace(strings.TrimPrefix(line, "MAXALLOC:"))
			maxint, err := strconv.Atoi(maxstr)
			if err != nil {
				panic(fmt.Sprintf("invalid maxalloc amount: %v", maxstr))
			}
			maxAlloc = int64(maxint)
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

//----------------------------------------
// testBanker

type testBanker struct {
	coinTable map[crypto.Bech32Address]std.Coins
}

func newtestBanker(args ...interface{}) *testBanker {
	return &testBanker{
		coinTable: make(map[crypto.Bech32Address]std.Coins),
	}
}

func (tb *testBanker) GetCoins(addr crypto.Bech32Address, dst *std.Coins) {
	coins, exists := tb.coinTable[addr]
	if !exists {
		*dst = nil
	} else {
		*dst = coins
	}
}

func (tb *testBanker) SendCoins(from, to crypto.Bech32Address, amt std.Coins) {
	fcoins, fexists := tb.coinTable[from]
	if !fexists {
		panic(fmt.Sprintf(
			"source address %s does not exist",
			from.String()))
	}
	if !fcoins.IsAllGTE(amt) {
		panic(fmt.Sprintf(
			"source address %s has %s; cannot send %s",
			from.String(), fcoins, amt))
	}
	// First, subtract from 'from'.
	frest := fcoins.Sub(amt)
	tb.setCoins(from, frest)
	// Second, add to 'to'.
	// NOTE: even works when from==to, due to 2-step isolation.
	tcoins, _ := tb.coinTable[to]
	tsum := tcoins.Add(amt)
	tb.setCoins(to, tsum)
}

func (tb *testBanker) setCoins(addr crypto.Bech32Address, amt std.Coins) {
	coins, _ := tb.coinTable[addr]
	sum := coins.Add(amt)
	tb.coinTable[addr] = sum
}

func (tb *testBanker) TotalCoin(denom string) int64 {
	panic("not yet implemented")
}

func (tb *testBanker) IssueCoin(addr crypto.Bech32Address, denom string, amt int64) {
	coins, _ := tb.coinTable[addr]
	sum := coins.Add(std.Coins{{denom, amt}})
	tb.setCoins(addr, sum)
}

func (tb *testBanker) RemoveCoin(addr crypto.Bech32Address, denom string, amt int64) {
	coins, _ := tb.coinTable[addr]
	rest := coins.Sub(std.Coins{{denom, amt}})
	tb.setCoins(addr, rest)
}
