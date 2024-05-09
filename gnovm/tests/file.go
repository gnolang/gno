package tests

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"regexp"
	rtdb "runtime/debug"
	"strconv"
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/pmezard/go-difflib/difflib"
)

type loggerFunc func(args ...interface{})

func TestMachine(store gno.Store, stdout io.Writer, pkgPath string) *gno.Machine {
	// default values
	var (
		send     std.Coins
		maxAlloc int64
	)

	return testMachineCustom(store, pkgPath, stdout, maxAlloc, send)
}

func testMachineCustom(store gno.Store, pkgPath string, stdout io.Writer, maxAlloc int64, send std.Coins) *gno.Machine {
	ctx := testContext(pkgPath, send)
	m := gno.NewMachineWithOptions(gno.MachineOptions{
		PkgPath:       "", // set later.
		Output:        stdout,
		Store:         store,
		Context:       ctx,
		MaxAllocBytes: maxAlloc,
	})
	return m
}

func testContext(pkgPath string, send std.Coins) stdlibs.ExecContext {
	// FIXME: create a better package to manage this, with custom constructors
	pkgAddr := gno.DerivePkgAddr(pkgPath) // the addr of the pkgPath called.
	caller := gno.DerivePkgAddr("user1.gno")

	pkgCoins := std.MustParseCoins("200000000ugnot").Add(send) // >= send.
	banker := newTestBanker(pkgAddr.Bech32(), pkgCoins)
	ctx := stdlibs.ExecContext{
		ChainID:       "dev",
		Height:        123,
		Timestamp:     1234567890,
		Msg:           nil,
		OrigCaller:    caller.Bech32(),
		OrigPkgAddr:   pkgAddr.Bech32(),
		OrigSend:      send,
		OrigSendSpent: new(std.Coins),
		Banker:        banker,
		EventLogger:   sdk.NewEventLogger(),
	}
	return ctx
}

type runFileTestOptions struct {
	nativeLibs bool
	logger     loggerFunc
	syncWanted bool
}

// RunFileTestOptions specify changing options in [RunFileTest], deviating
// from the zero value.
type RunFileTestOption func(*runFileTestOptions)

// WithNativeLibs enables using go native libraries (ie, [ImportModeNativePreferred])
// instead of using stdlibs/*.
func WithNativeLibs() RunFileTestOption {
	return func(r *runFileTestOptions) { r.nativeLibs = true }
}

// WithLoggerFunc sets a logging function for [RunFileTest].
func WithLoggerFunc(f func(args ...interface{})) RunFileTestOption {
	return func(r *runFileTestOptions) { r.logger = f }
}

// WithSyncWanted sets the syncWanted flag to true.
// It rewrites tests files so that the values of Output: and of Realm:
// comments match the actual output or realm state after the test.
func WithSyncWanted(v bool) RunFileTestOption {
	return func(r *runFileTestOptions) { r.syncWanted = v }
}

// RunFileTest executes the filetest at the given path, using rootDir as
// the directory where to find the "stdlibs" directory.
func RunFileTest(rootDir string, path string, opts ...RunFileTestOption) error {
	var f runFileTestOptions
	for _, opt := range opts {
		opt(&f)
	}

	directives, pkgPath, resWanted, errWanted, rops, maxAlloc, send := wantedFromComment(path)
	if pkgPath == "" {
		pkgPath = "main"
	}
	pkgName := DefaultPkgName(pkgPath)
	stdin := new(bytes.Buffer)
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	mode := ImportModeStdlibsPreferred
	if f.nativeLibs {
		mode = ImportModeNativePreferred
	}
	store := TestStore(rootDir, "./files", stdin, stdout, stderr, mode)
	store.SetLogStoreOps(true)
	m := testMachineCustom(store, pkgPath, stdout, maxAlloc, send)

	// TODO support stdlib groups, but make testing safe;
	// e.g. not be able to make network connections.
	// interp.New(interp.Options{GoPath: goPath, Stdout: &stdout, Stderr: &stderr})
	// m.Use(interp.Symbols)
	// m.Use(stdlib.Symbols)
	// m.Use(unsafe.Symbols)
	bz, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	{ // Validate result, errors, etc.
		var pnc interface{}
		func() {
			defer func() {
				if r := recover(); r != nil {
					// print output.
					fmt.Println("OUTPUT:\n", stdout.String())
					// print stack if unexpected error.
					pnc = r
					if errWanted == "" {
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
			if f.logger != nil {
				f.logger("========================================")
				f.logger("RUN FILES & INIT")
				f.logger("========================================")
			}
			if !gno.IsRealmPath(pkgPath) {
				// simple case.
				pn := gno.NewPackageNode(pkgName, pkgPath, &gno.FileSet{})
				pv := pn.NewPackage()
				store.SetBlockNode(pn)
				store.SetCachePackage(pv)
				m.SetActivePackage(pv)
				n := gno.MustParseFile(path, string(bz)) // "main.gno", string(bz))
				m.RunFiles(n)
				if f.logger != nil {
					f.logger("========================================")
					f.logger("RUN MAIN")
					f.logger("========================================")
				}
				m.RunMain()
				if f.logger != nil {
					f.logger("========================================")
					f.logger("RUN MAIN END")
					f.logger("========================================")
				}
			} else {
				// realm case.
				store.SetStrictGo2GnoMapping(true) // in gno.land, natives must be registered.
				gno.DisableDebug()                 // until main call.
				// save package using realm crawl procedure.
				memPkg := &std.MemPackage{
					Name: string(pkgName),
					Path: pkgPath,
					Files: []*std.MemFile{
						{
							Name: "main.gno", // dontcare
							Body: string(bz),
						},
					},
				}
				// run decls and init functions.
				m.RunMemPackage(memPkg, true)
				// reconstruct machine and clear store cache.
				// whether package is realm or not, since non-realm
				// may call realm packages too.
				if f.logger != nil {
					f.logger("========================================")
					f.logger("CLEAR STORE CACHE")
					f.logger("========================================")
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
				if f.logger != nil {
					store.Print()
					f.logger("========================================")
					f.logger("PREPROCESS ALL FILES")
					f.logger("========================================")
				}
				m.PreprocessAllFilesAndSaveBlockNodes()
				if f.logger != nil {
					f.logger("========================================")
					f.logger("RUN MAIN")
					f.logger("========================================")
					store.Print()
				}
				pv2 := store.GetPackage(pkgPath, false)
				m.SetActivePackage(pv2)
				gno.EnableDebug()
				if rops != "" {
					// clear store.opslog from init function(s),
					// and PreprocessAllFilesAndSaveBlockNodes().
					store.SetLogStoreOps(true) // resets.
				}
				m.RunMain()
				if f.logger != nil {
					f.logger("========================================")
					f.logger("RUN MAIN END")
					f.logger("========================================")
				}
			}
		}()

		for _, directive := range directives {
			switch directive {
			case "Error":
				// errWanted given
				if errWanted != "" {
					if pnc == nil {
						panic(fmt.Sprintf("fail on %s: got nil error, want: %q", path, errWanted))
					}

					errstr := ""
					switch v := pnc.(type) {
					case *gno.TypedValue:
						errstr = v.Sprint(m)
					case *gno.PreprocessError:
						errstr = v.Unwrap().Error()
					default:
						errstr = strings.TrimSpace(fmt.Sprintf("%v", pnc))
					}
					if errstr != errWanted {
						panic(fmt.Sprintf("fail on %s: got %q, want: %q", path, errstr, errWanted))
					}

					// NOTE: ignores any gno.GetDebugErrors().
					gno.ClearDebugErrors()
					return nil // nothing more to do.
				} else {
					// record errors when errWanted is empty and pnc not nil
					if pnc != nil {
						errstr := ""
						if tv, ok := pnc.(*gno.TypedValue); ok {
							errstr = tv.Sprint(m)
						} else {
							errstr = strings.TrimSpace(fmt.Sprintf("%v", pnc))
						}
						// check tip line, write to file
						ctl := fmt.Sprintf(
							errstr +
								"\n*** CHECK THE ERR MESSAGES ABOVE, MAKE SURE IT'S WHAT YOU EXPECTED, " +
								"DELETE THIS LINE AND RUN TEST AGAIN ***",
						)
						replaceWantedInPlace(path, "Error", ctl)
						panic(fmt.Sprintf("fail on %s: err recorded, check the message and run test again", path))
					}
					// check gno debug errors when errWanted is empty, pnc is nil
					if gno.HasDebugErrors() {
						panic(fmt.Sprintf("fail on %s: got unexpected debug error(s): %v", path, gno.GetDebugErrors()))
					}
					// pnc is nil, errWanted empty, no gno debug errors
					return nil
				}
			case "Output":
				// panic if got unexpected error
				if pnc != nil {
					if tv, ok := pnc.(*gno.TypedValue); ok {
						panic(fmt.Sprintf("fail on %s: got unexpected error: %s", path, tv.Sprint(m)))
					} else { // happens on 'unknown import path ...'
						panic(fmt.Sprintf("fail on %s: got unexpected error: %v", path, pnc))
					}
				}
				// check result
				res := strings.TrimSpace(stdout.String())
				res = trimTrailingSpaces(res)
				if res != resWanted {
					if f.syncWanted {
						// write output to file.
						replaceWantedInPlace(path, "Output", res)
					} else {
						// panic so tests immediately fail (for now).
						if resWanted == "" {
							panic(fmt.Sprintf("fail on %s: got unexpected output: %s", path, res))
						} else {
							diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
								A:        difflib.SplitLines(resWanted),
								B:        difflib.SplitLines(res),
								FromFile: "Expected",
								FromDate: "",
								ToFile:   "Actual",
								ToDate:   "",
								Context:  1,
							})
							panic(fmt.Sprintf("fail on %s: diff:\n%s\n", path, diff))
						}
					}
				}
			case "Realm":
				// panic if got unexpected error
				if pnc != nil {
					if tv, ok := pnc.(*gno.TypedValue); ok {
						panic(fmt.Sprintf("fail on %s: got unexpected error: %s", path, tv.Sprint(m)))
					} else { // TODO: does this happen?
						panic(fmt.Sprintf("fail on %s: got unexpected error: %v", path, pnc))
					}
				}
				// check realm ops
				if rops != "" {
					rops2 := strings.TrimSpace(store.SprintStoreOps())
					if rops != rops2 {
						if f.syncWanted {
							// write output to file.
							replaceWantedInPlace(path, "Realm", rops2)
						} else {
							diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
								A:        difflib.SplitLines(rops),
								B:        difflib.SplitLines(rops2),
								FromFile: "Expected",
								FromDate: "",
								ToFile:   "Actual",
								ToDate:   "",
								Context:  1,
							})
							panic(fmt.Sprintf("fail on %s: diff:\n%s\n", path, diff))
						}
					}
				}
			default:
				return nil
			}
		}
	}

	// Check that machine is empty.
	err = m.CheckEmpty()
	if err != nil {
		if f.logger != nil {
			f.logger("last state: \n", m.String())
		}
		panic(fmt.Sprintf("fail on %s: machine not empty after main: %v", path, err))
	}
	return nil
}

func wantedFromComment(p string) (directives []string, pkgPath, res, err, rops string, maxAlloc int64, send std.Coins) {
	fset := token.NewFileSet()
	f, err2 := parser.ParseFile(fset, p, nil, parser.ParseComments)
	if err2 != nil {
		panic(err2)
	}
	if len(f.Comments) == 0 {
		return
	}
	for _, comments := range f.Comments {
		text := readComments(comments)
		if strings.HasPrefix(text, "PKGPATH:") {
			line := strings.SplitN(text, "\n", 2)[0]
			pkgPath = strings.TrimSpace(strings.TrimPrefix(line, "PKGPATH:"))
		} else if strings.HasPrefix(text, "MAXALLOC:") {
			line := strings.SplitN(text, "\n", 2)[0]
			maxstr := strings.TrimSpace(strings.TrimPrefix(line, "MAXALLOC:"))
			maxint, err := strconv.Atoi(maxstr)
			if err != nil {
				panic(fmt.Sprintf("invalid maxalloc amount: %v", maxstr))
			}
			maxAlloc = int64(maxint)
		} else if strings.HasPrefix(text, "SEND:") {
			line := strings.SplitN(text, "\n", 2)[0]
			sendstr := strings.TrimSpace(strings.TrimPrefix(line, "SEND:"))
			send = std.MustParseCoins(sendstr)
		} else if strings.HasPrefix(text, "Output:\n") {
			res = strings.TrimPrefix(text, "Output:\n")
			res = strings.TrimSpace(res)
			directives = append(directives, "Output")
		} else if strings.HasPrefix(text, "Error:\n") {
			err = strings.TrimPrefix(text, "Error:\n")
			err = strings.TrimSpace(err)
			// XXX temporary until we support line:column.
			// If error starts with line:column, trim it.
			re := regexp.MustCompile(`^[0-9]+:[0-9]+: `)
			err = re.ReplaceAllString(err, "")
			directives = append(directives, "Error")
		} else if strings.HasPrefix(text, "Realm:\n") {
			rops = strings.TrimPrefix(text, "Realm:\n")
			rops = strings.TrimSpace(rops)
			directives = append(directives, "Realm")
		} else {
			// ignore unexpected.
		}
	}
	return
}

// readComments returns //-style comments from cg, but without truncating empty
// lines like cg.Text().
func readComments(cg *ast.CommentGroup) string {
	var b strings.Builder
	for _, c := range cg.List {
		if len(c.Text) < 2 || c.Text[:2] != "//" {
			// ignore no //-style comment
			break
		}
		s := strings.TrimPrefix(c.Text[2:], " ")
		b.WriteString(s + "\n")
	}
	return b.String()
}

// Replace comment in file with given output given directive.
func replaceWantedInPlace(path string, directive string, output string) {
	bz := osm.MustReadFile(path)
	body := string(bz)
	lines := strings.Split(body, "\n")
	isReplacing := false
	wroteDirective := false
	newlines := []string(nil)
	for _, line := range lines {
		if line == "// "+directive+":" {
			if wroteDirective {
				isReplacing = true
				continue
			} else {
				wroteDirective = true
				isReplacing = true
				newlines = append(newlines, "// "+directive+":")
				outlines := strings.Split(output, "\n")
				for _, outline := range outlines {
					newlines = append(newlines,
						strings.TrimRight("// "+outline, " "))
				}
				continue
			}
		} else if isReplacing {
			if strings.HasPrefix(line, "//") {
				continue
			} else {
				isReplacing = false
			}
		}
		newlines = append(newlines, line)
	}
	osm.MustWriteFile(path, []byte(strings.Join(newlines, "\n")), 0o644)
}

func DefaultPkgName(gopkgPath string) gno.Name {
	parts := strings.Split(gopkgPath, "/")
	last := parts[len(parts)-1]
	parts = strings.Split(last, "-")
	name := parts[len(parts)-1]
	name = strings.ToLower(name)
	return gno.Name(name)
}

// go comments strip trailing spaces.
func trimTrailingSpaces(result string) string {
	lines := strings.Split(result, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	return strings.Join(lines, "\n")
}

// ----------------------------------------
// testBanker

type testBanker struct {
	coinTable map[crypto.Bech32Address]std.Coins
}

func newTestBanker(args ...interface{}) *testBanker {
	coinTable := make(map[crypto.Bech32Address]std.Coins)
	if len(args)%2 != 0 {
		panic("newTestBanker requires even number of arguments; addr followed by coins")
	}
	for i := 0; i < len(args); i += 2 {
		addr := args[i].(crypto.Bech32Address)
		amount := args[i+1].(std.Coins)
		coinTable[addr] = amount
	}
	return &testBanker{
		coinTable: coinTable,
	}
}

func (tb *testBanker) GetCoins(addr crypto.Bech32Address) (dst std.Coins) {
	return tb.coinTable[addr]
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
	tb.coinTable[from] = frest
	// Second, add to 'to'.
	// NOTE: even works when from==to, due to 2-step isolation.
	tcoins, _ := tb.coinTable[to]
	tsum := tcoins.Add(amt)
	tb.coinTable[to] = tsum
}

func (tb *testBanker) TotalCoin(denom string) int64 {
	panic("not yet implemented")
}

func (tb *testBanker) IssueCoin(addr crypto.Bech32Address, denom string, amt int64) {
	coins, _ := tb.coinTable[addr]
	sum := coins.Add(std.Coins{{denom, amt}})
	tb.coinTable[addr] = sum
}

func (tb *testBanker) RemoveCoin(addr crypto.Bech32Address, denom string, amt int64) {
	coins, _ := tb.coinTable[addr]
	rest := coins.Sub(std.Coins{{denom, amt}})
	tb.coinTable[addr] = rest
}
