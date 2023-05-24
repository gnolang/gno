package tests

import (
	"bufio"
	"bytes"
	"compress/flate"
	"compress/gzip"
	"context"
	"crypto/md5" //nolint:gosec
	crand "crypto/rand"
	"crypto/sha1" //nolint:gosec
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"hash/fnv"
	"image"
	"image/color"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/big"
	"math/rand"
	"net"
	"net/url"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"text/template"
	"time"
	"unicode/utf8"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
	stypes "github.com/gnolang/gno/tm2/pkg/store/types"
)

type importMode uint64

const (
	ImportModeStdlibsOnly importMode = iota
	ImportModeStdlibsPreferred
	ImportModeNativePreferred
)

// ImportModeStdlibsOnly: use stdlibs/* only (except a few exceptions). for stdlibs/* and examples/* testing.
// ImportModeStdlibsPreferred: use stdlibs/* if present, otherwise use native. for files/tests2/*.
// ImportModeNativePreferred: do not use stdlibs/* if native registered. for files/tests/*.
// NOTE: this isn't safe, should only be used for testing.
func TestStore(rootDir, filesPath string, stdin io.Reader, stdout, stderr io.Writer, mode importMode) (store gno.Store) {
	getPackage := func(pkgPath string) (pn *gno.PackageNode, pv *gno.PackageValue) {
		if pkgPath == "" {
			panic(fmt.Sprintf("invalid zero package path in testStore().pkgGetter"))
		}
		if mode != ImportModeStdlibsOnly &&
			mode != ImportModeStdlibsPreferred &&
			mode != ImportModeNativePreferred {
			panic(fmt.Sprintf("unrecognized import mode"))
		}

		if filesPath != "" {
			// if _test package...
			const testPath = "github.com/gnolang/gno/_test/"
			if strings.HasPrefix(pkgPath, testPath) {
				baseDir := filepath.Join(filesPath, "extern", pkgPath[len(testPath):])
				memPkg := gno.ReadMemPackage(baseDir, pkgPath)
				m2 := gno.NewMachineWithOptions(gno.MachineOptions{
					PkgPath: "test",
					Output:  stdout,
					Store:   store,
				})
				// pkg := gno.NewPackageNode(gno.Name(memPkg.Name), memPkg.Path, nil)
				// pv := pkg.NewPackage()
				// m2.SetActivePackage(pv)
				return m2.RunMemPackage(memPkg, false)
			}
		}

		// if stdlibs package is preferred , try to load it first.
		if mode == ImportModeStdlibsOnly ||
			mode == ImportModeStdlibsPreferred {
			stdlibPath := filepath.Join(rootDir, "gnovm", "stdlibs", pkgPath)
			if osm.DirExists(stdlibPath) {
				memPkg := gno.ReadMemPackage(stdlibPath, pkgPath)
				m2 := gno.NewMachineWithOptions(gno.MachineOptions{
					// NOTE: see also pkgs/sdk/vm/builtins.go
					// XXX: why does this fail when just pkgPath?
					PkgPath: "gno.land/r/stdlibs/" + pkgPath,
					Output:  stdout,
					Store:   store,
				})
				return m2.RunMemPackage(memPkg, true)
			}
		}

		// if native package is allowed, return it.
		if pkgPath == "os" || // special cases even when StdlibsOnly (for tests).
			pkgPath == "fmt" || // TODO: try to minimize these exceptions over time.
			pkgPath == "log" ||
			pkgPath == "crypto/rand" ||
			pkgPath == "crypto/md5" ||
			pkgPath == "crypto/sha1" ||
			pkgPath == "encoding/base64" ||
			pkgPath == "encoding/binary" ||
			pkgPath == "encoding/json" ||
			pkgPath == "encoding/xml" ||
			pkgPath == "internal/os_test" ||
			pkgPath == "math" ||
			pkgPath == "math/big" ||
			pkgPath == "math/rand" ||
			mode == ImportModeStdlibsPreferred ||
			mode == ImportModeNativePreferred {
			switch pkgPath {
			case "os":
				pkg := gno.NewPackageNode("os", pkgPath, nil)
				pkg.DefineGoNativeValue("Stdin", stdin)
				pkg.DefineGoNativeValue("Stdout", stdout)
				pkg.DefineGoNativeValue("Stderr", stderr)
				return pkg, pkg.NewPackage()
			case "fmt":
				pkg := gno.NewPackageNode("fmt", pkgPath, nil)
				pkg.DefineGoNativeType(reflect.TypeOf((*fmt.Stringer)(nil)).Elem())
				pkg.DefineGoNativeType(reflect.TypeOf((*fmt.Formatter)(nil)).Elem())
				pkg.DefineGoNativeValue("Println", func(a ...interface{}) (n int, err error) {
					// NOTE: uncomment to debug long running tests
					fmt.Println(a...)
					res := fmt.Sprintln(a...)
					return stdout.Write([]byte(res))
				})
				pkg.DefineGoNativeValue("Printf", func(format string, a ...interface{}) (n int, err error) {
					res := fmt.Sprintf(format, a...)
					return stdout.Write([]byte(res))
				})
				pkg.DefineGoNativeValue("Print", func(a ...interface{}) (n int, err error) {
					res := fmt.Sprint(a...)
					return stdout.Write([]byte(res))
				})
				pkg.DefineGoNativeValue("Sprint", fmt.Sprint)
				pkg.DefineGoNativeValue("Sprintf", fmt.Sprintf)
				pkg.DefineGoNativeValue("Sprintln", fmt.Sprintln)
				pkg.DefineGoNativeValue("Sscanf", fmt.Sscanf)
				pkg.DefineGoNativeValue("Errorf", fmt.Errorf)
				pkg.DefineGoNativeValue("Fprintln", fmt.Fprintln)
				pkg.DefineGoNativeValue("Fprintf", fmt.Fprintf)
				pkg.DefineGoNativeValue("Fprint", fmt.Fprint)
				return pkg, pkg.NewPackage()
			case "encoding/base64":
				pkg := gno.NewPackageNode("base64", pkgPath, nil)
				pkg.DefineGoNativeValue("RawStdEncoding", base64.RawStdEncoding)
				pkg.DefineGoNativeValue("StdEncoding", base64.StdEncoding)
				pkg.DefineGoNativeValue("NewDecoder", base64.NewDecoder)
				return pkg, pkg.NewPackage()
			case "encoding/binary":
				pkg := gno.NewPackageNode("binary", pkgPath, nil)
				pkg.DefineGoNativeValue("LittleEndian", binary.LittleEndian)
				pkg.DefineGoNativeValue("BigEndian", binary.BigEndian)
				return pkg, pkg.NewPackage()
			case "encoding/json":
				pkg := gno.NewPackageNode("json", pkgPath, nil)
				pkg.DefineGoNativeValue("Unmarshal", json.Unmarshal)
				pkg.DefineGoNativeValue("Marshal", json.Marshal)
				return pkg, pkg.NewPackage()
			case "encoding/xml":
				pkg := gno.NewPackageNode("xml", pkgPath, nil)
				pkg.DefineGoNativeValue("Unmarshal", xml.Unmarshal)
				return pkg, pkg.NewPackage()
			case "internal/os_test":
				pkg := gno.NewPackageNode("os_test", pkgPath, nil)
				pkg.DefineNative("Sleep",
					gno.Flds( // params
						"d", gno.AnyT(), // NOTE: should be time.Duration
					),
					gno.Flds( // results
					),
					func(m *gno.Machine) {
						// For testing purposes here, nanoseconds are separately kept track.
						arg0 := m.LastBlock().GetParams1().TV
						d := arg0.GetInt64()
						sec := d / int64(time.Second)
						nano := d % int64(time.Second)
						ctx := m.Context.(stdlibs.ExecContext)
						ctx.Timestamp += sec
						ctx.TimestampNano += nano
						if ctx.TimestampNano >= int64(time.Second) {
							ctx.Timestamp += 1
							ctx.TimestampNano -= int64(time.Second)
						}
						m.Context = ctx
					},
				)
				return pkg, pkg.NewPackage()
			case "net":
				pkg := gno.NewPackageNode("net", pkgPath, nil)
				pkg.DefineGoNativeType(reflect.TypeOf(net.TCPAddr{}))
				pkg.DefineGoNativeValue("IPv4", net.IPv4)
				return pkg, pkg.NewPackage()
			case "net/http":
				// XXX UNSAFE
				// There's no reason why we can't replace these with safer alternatives.
				panic("just say gno")
				/*
					pkg := gno.NewPackageNode("http", pkgPath, nil)
					pkg.DefineGoNativeType(reflect.TypeOf(http.Request{}))
					pkg.DefineGoNativeValue("DefaultClient", http.DefaultClient)
					pkg.DefineGoNativeType(reflect.TypeOf(http.Client{}))
					return pkg, pkg.NewPackage()
				*/
			case "net/url":
				pkg := gno.NewPackageNode("url", pkgPath, nil)
				pkg.DefineGoNativeType(reflect.TypeOf(url.Values{}))
				return pkg, pkg.NewPackage()
			case "bufio":
				pkg := gno.NewPackageNode("bufio", pkgPath, nil)
				pkg.DefineGoNativeValue("NewScanner", bufio.NewScanner)
				pkg.DefineGoNativeType(reflect.TypeOf(bufio.SplitFunc(nil)))
				return pkg, pkg.NewPackage()
			case "bytes":
				pkg := gno.NewPackageNode("bytes", pkgPath, nil)
				pkg.DefineGoNativeValue("Equal", bytes.Equal)
				pkg.DefineGoNativeValue("Compare", bytes.Compare)
				pkg.DefineGoNativeValue("NewReader", bytes.NewReader)
				pkg.DefineGoNativeValue("NewBuffer", bytes.NewBuffer)
				pkg.DefineGoNativeValue("Repeat", bytes.Repeat)
				pkg.DefineGoNativeType(reflect.TypeOf(bytes.Buffer{}))
				return pkg, pkg.NewPackage()
			case "time":
				pkg := gno.NewPackageNode("time", pkgPath, nil)
				pkg.DefineGoNativeValue("Millisecond", time.Millisecond)
				pkg.DefineGoNativeValue("Second", time.Second)
				pkg.DefineGoNativeValue("Minute", time.Minute)
				pkg.DefineGoNativeValue("Hour", time.Hour)
				pkg.DefineGoNativeValue("Date", time.Date)
				pkg.DefineGoNativeValue("Now", func() time.Time { return time.Unix(0, 0).UTC() }) // deterministic
				pkg.DefineGoNativeValue("November", time.November)
				pkg.DefineGoNativeValue("UTC", time.UTC)
				pkg.DefineGoNativeValue("Unix", time.Unix)
				pkg.DefineGoNativeType(reflect.TypeOf(time.Time{}))
				pkg.DefineGoNativeType(reflect.TypeOf(time.Duration(0)))
				pkg.DefineGoNativeType(reflect.TypeOf(time.Month(0)))
				return pkg, pkg.NewPackage()
			case "strings":
				pkg := gno.NewPackageNode("strings", pkgPath, nil)
				pkg.DefineGoNativeValue("Split", strings.Split)
				pkg.DefineGoNativeValue("SplitN", strings.SplitN)
				pkg.DefineGoNativeValue("Contains", strings.Contains)
				pkg.DefineGoNativeValue("TrimSpace", strings.TrimSpace)
				pkg.DefineGoNativeValue("HasPrefix", strings.HasPrefix)
				pkg.DefineGoNativeValue("NewReader", strings.NewReader)
				pkg.DefineGoNativeValue("Index", strings.Index)
				pkg.DefineGoNativeValue("IndexRune", strings.IndexRune)
				pkg.DefineGoNativeValue("Join", strings.Join)
				pkg.DefineGoNativeType(reflect.TypeOf(strings.Builder{}))
				return pkg, pkg.NewPackage()
			case "math":
				pkg := gno.NewPackageNode("math", pkgPath, nil)
				pkg.DefineGoNativeValue("Abs", math.Abs)
				pkg.DefineGoNativeValue("Cos", math.Cos)
				pkg.DefineGoNativeValue("Pi", math.Pi)
				pkg.DefineGoNativeValue("MaxFloat32", math.MaxFloat32)
				pkg.DefineGoNativeValue("MaxFloat64", math.MaxFloat64)
				return pkg, pkg.NewPackage()
			case "math/rand":
				// XXX only expose for tests.
				pkg := gno.NewPackageNode("rand", pkgPath, nil)
				pkg.DefineGoNativeValue("Intn", rand.Intn)
				pkg.DefineGoNativeValue("Uint32", rand.Uint32)
				pkg.DefineGoNativeValue("Seed", rand.Seed)
				pkg.DefineGoNativeValue("New", rand.New)
				pkg.DefineGoNativeValue("NewSource", rand.NewSource)
				pkg.DefineGoNativeType(reflect.TypeOf(rand.Rand{}))
				return pkg, pkg.NewPackage()
			case "crypto/rand":
				pkg := gno.NewPackageNode("rand", pkgPath, nil)
				pkg.DefineGoNativeValue("Prime", crand.Prime)
				// for determinism:
				// pkg.DefineGoNativeValue("Reader", crand.Reader)
				pkg.DefineGoNativeValue("Reader", &dummyReader{})
				return pkg, pkg.NewPackage()
			case "crypto/md5":
				pkg := gno.NewPackageNode("md5", pkgPath, nil)
				pkg.DefineGoNativeValue("New", md5.New)
				return pkg, pkg.NewPackage()
			case "crypto/sha1":
				pkg := gno.NewPackageNode("sha1", pkgPath, nil)
				pkg.DefineGoNativeValue("New", sha1.New)
				return pkg, pkg.NewPackage()
			case "image":
				pkg := gno.NewPackageNode("image", pkgPath, nil)
				pkg.DefineGoNativeType(reflect.TypeOf(image.Point{}))
				return pkg, pkg.NewPackage()
			case "image/color":
				pkg := gno.NewPackageNode("color", pkgPath, nil)
				pkg.DefineGoNativeType(reflect.TypeOf(color.NRGBA64{}))
				return pkg, pkg.NewPackage()
			case "compress/flate":
				pkg := gno.NewPackageNode("flate", pkgPath, nil)
				pkg.DefineGoNativeValue("BestSpeed", flate.BestSpeed)
				return pkg, pkg.NewPackage()
			case "compress/gzip":
				pkg := gno.NewPackageNode("gzip", pkgPath, nil)
				pkg.DefineGoNativeType(reflect.TypeOf(gzip.Writer{}))
				pkg.DefineGoNativeValue("BestCompression", gzip.BestCompression)
				pkg.DefineGoNativeValue("BestSpeed", gzip.BestSpeed)
				return pkg, pkg.NewPackage()
			case "context":
				pkg := gno.NewPackageNode("context", pkgPath, nil)
				pkg.DefineGoNativeType(reflect.TypeOf((*context.Context)(nil)).Elem())
				pkg.DefineGoNativeValue("WithValue", context.WithValue)
				pkg.DefineGoNativeValue("Background", context.Background)
				return pkg, pkg.NewPackage()
			case "sync":
				pkg := gno.NewPackageNode("sync", pkgPath, nil)
				pkg.DefineGoNativeType(reflect.TypeOf(sync.Mutex{}))
				pkg.DefineGoNativeType(reflect.TypeOf(sync.RWMutex{}))
				pkg.DefineGoNativeType(reflect.TypeOf(sync.Pool{}))
				return pkg, pkg.NewPackage()
			case "sync/atomic":
				pkg := gno.NewPackageNode("atomic", pkgPath, nil)
				pkg.DefineGoNativeType(reflect.TypeOf(atomic.Value{}))
				return pkg, pkg.NewPackage()
			case "math/big":
				pkg := gno.NewPackageNode("big", pkgPath, nil)
				pkg.DefineGoNativeValue("NewInt", big.NewInt)
				return pkg, pkg.NewPackage()
			case "sort":
				pkg := gno.NewPackageNode("sort", pkgPath, nil)
				pkg.DefineGoNativeValue("Strings", sort.Strings)
				// pkg.DefineGoNativeValue("Sort", sort.Sort)
				return pkg, pkg.NewPackage()
			case "flag":
				pkg := gno.NewPackageNode("flag", pkgPath, nil)
				pkg.DefineGoNativeType(reflect.TypeOf(flag.Flag{}))
				return pkg, pkg.NewPackage()
			case "io":
				pkg := gno.NewPackageNode("io", pkgPath, nil)
				pkg.DefineGoNativeValue("EOF", io.EOF)
				pkg.DefineGoNativeValue("ReadFull", io.ReadFull)
				pkg.DefineGoNativeType(reflect.TypeOf((*io.ReadCloser)(nil)).Elem())
				pkg.DefineGoNativeType(reflect.TypeOf((*io.Closer)(nil)).Elem())
				pkg.DefineGoNativeType(reflect.TypeOf((*io.Reader)(nil)).Elem())
				return pkg, pkg.NewPackage()
			case "io/ioutil":
				pkg := gno.NewPackageNode("ioutil", pkgPath, nil)
				pkg.DefineGoNativeValue("NopCloser", ioutil.NopCloser)
				pkg.DefineGoNativeValue("ReadAll", ioutil.ReadAll)
				return pkg, pkg.NewPackage()
			case "log":
				pkg := gno.NewPackageNode("log", pkgPath, nil)
				pkg.DefineGoNativeValue("Fatal", log.Fatal)
				return pkg, pkg.NewPackage()
			case "text/template":
				pkg := gno.NewPackageNode("template", pkgPath, nil)
				pkg.DefineGoNativeType(reflect.TypeOf(template.FuncMap{}))
				return pkg, pkg.NewPackage()
			case "unicode/utf8":
				pkg := gno.NewPackageNode("utf8", pkgPath, nil)
				pkg.DefineGoNativeValue("DecodeRuneInString", utf8.DecodeRuneInString)
				tv := gno.TypedValue{T: gno.UntypedRuneType} // TODO dry
				tv.SetInt32(utf8.RuneSelf)                   // ..
				pkg.Define("RuneSelf", tv)                   // ..
				return pkg, pkg.NewPackage()
			case "errors":
				pkg := gno.NewPackageNode("errors", pkgPath, nil)
				pkg.DefineGoNativeValue("New", errors.New)
				return pkg, pkg.NewPackage()
			case "hash/fnv":
				pkg := gno.NewPackageNode("fnv", pkgPath, nil)
				pkg.DefineGoNativeValue("New32a", fnv.New32a)
				return pkg, pkg.NewPackage()
			/* XXX support somehow for speed. for now, generic implemented in stdlibs.
			case "internal/bytealg":
				pkg := gno.NewPackageNode("bytealg", pkgPath, nil)
				pkg.DefineGoNativeValue("Compare", bytealg.Compare)
				pkg.DefineGoNativeValue("CountString", bytealg.CountString)
				pkg.DefineGoNativeValue("Cutover", bytealg.Cutover)
				pkg.DefineGoNativeValue("Equal", bytealg.Equal)
				pkg.DefineGoNativeValue("HashStr", bytealg.HashStr)
				pkg.DefineGoNativeValue("HashStrBytes", bytealg.HashStrBytes)
				pkg.DefineGoNativeValue("HashStrRev", bytealg.HashStrRev)
				pkg.DefineGoNativeValue("HashStrRevBytes", bytealg.HashStrRevBytes)
				pkg.DefineGoNativeValue("Index", bytealg.Index)
				pkg.DefineGoNativeValue("IndexByte", bytealg.IndexByte)
				pkg.DefineGoNativeValue("IndexByteString", bytealg.IndexByteString)
				pkg.DefineGoNativeValue("IndexRabinKarp", bytealg.IndexRabinKarp)
				pkg.DefineGoNativeValue("IndexRabinKarpBytes", bytealg.IndexRabinKarpBytes)
				pkg.DefineGoNativeValue("IndexString", bytealg.IndexString)
				return pkg, pkg.NewPackage()
			*/
			default:
				// continue on...
			}
		}

		// if native package is preferred, try to load stdlibs/* as backup.
		if mode == ImportModeNativePreferred {
			stdlibPath := filepath.Join(rootDir, "gnovm", "stdlibs", pkgPath)
			if osm.DirExists(stdlibPath) {
				memPkg := gno.ReadMemPackage(stdlibPath, pkgPath)
				m2 := gno.NewMachineWithOptions(gno.MachineOptions{
					PkgPath: "test",
					Output:  stdout,
					Store:   store,
				})
				pn, pv = m2.RunMemPackage(memPkg, true)
				return
			}
		}

		// if examples package...
		examplePath := filepath.Join(rootDir, "examples", pkgPath)
		if osm.DirExists(examplePath) {
			memPkg := gno.ReadMemPackage(examplePath, pkgPath)
			m2 := gno.NewMachineWithOptions(gno.MachineOptions{
				PkgPath: "test",
				Output:  stdout,
				Store:   store,
			})
			pn, pv = m2.RunMemPackage(memPkg, true)
			return
		}
		return nil, nil
	}
	// NOTE: store is also used in closure above.
	db := dbm.NewMemDB()
	baseStore := dbadapter.StoreConstructor(db, stypes.StoreOptions{})
	iavlStore := iavl.StoreConstructor(db, stypes.StoreOptions{})
	store = gno.NewStore(nil, baseStore, iavlStore)
	store.SetPackageGetter(getPackage)
	store.SetPackageInjector(testPackageInjector)
	store.SetStrictGo2GnoMapping(false)
	// native mappings
	stdlibs.InjectNativeMappings(store)
	return
}

//----------------------------------------
// testInjectNatives
// analogous to stdlibs.InjectNatives, but with
// native methods suitable for the testing environment.

func testPackageInjector(store gno.Store, pn *gno.PackageNode) {
	// Also inject stdlibs native functions.
	stdlibs.InjectPackage(store, pn)
	isOriginCall := func(m *gno.Machine) bool {
		tname := m.Frames[0].Func.Name
		switch tname {
		case "main": // test is a _filetest
			return len(m.Frames) == 3
		case "runtest": // test is a _test
			return len(m.Frames) == 7
		}
		// support init() in _filetest
		// XXX do we need to distinguish from 'runtest'/_test?
		// XXX pretty hacky even if not.
		if strings.HasPrefix(string(tname), "init.") {
			return len(m.Frames) == 3
		}
		panic("unable to determine if test is a _test or a _filetest")
	}
	// Test specific injections:
	switch pn.PkgPath {
	case "strconv":
		// NOTE: Itoa and Atoi are already injected
		// from stdlibs.InjectNatives.
		pn.DefineGoNativeType(reflect.TypeOf(strconv.NumError{}))
		pn.DefineGoNativeValue("ParseInt", strconv.ParseInt)
	case "std":
		// NOTE: some of these are overrides.
		// Also see stdlibs/InjectPackage.
		pn.DefineNativeOverride("AssertOriginCall",
			/*
				gno.Flds( // params
				),
				gno.Flds( // results
				),
			*/
			func(m *gno.Machine) {
				if !isOriginCall(m) {
					m.Panic(typedString("invalid non-origin call"))
					return
				}
			},
		)
		pn.DefineNativeOverride("IsOriginCall",
			/*
				gno.Flds( // params
				),
				gno.Flds( // results
					"isOrigin", "bool",
				),
			*/
			func(m *gno.Machine) {
				res0 := gno.TypedValue{T: gno.BoolType}
				res0.SetBool(isOriginCall(m))
				m.PushValue(res0)
			},
		)
		pn.DefineNativeOverride("CallerAt",
			/*
				gno.Flds( // params
					"n", "int",
				),
				gno.Flds( // results
					"", "Address",
				),
			*/
			func(m *gno.Machine) {
				arg0 := m.LastBlock().GetParams1().TV
				n := arg0.GetInt()
				if n <= 0 {
					m.Panic(typedString("CallerAt requires positive arg"))
					return
				}
				if n > m.NumFrames()-1 {
					// NOTE: the last frame's LastPackage
					// is set to the original non-frame
					// package, so need this check.
					m.Panic(typedString("frame not found"))
					return
				}
				var pkgAddr string
				if n == m.NumFrames()-1 {
					// This makes it consistent with OrigCaller and TestSetOrigCaller.
					ctx := m.Context.(stdlibs.ExecContext)
					pkgAddr = string(ctx.OrigCaller)
				} else {
					pkgAddr = string(m.LastCallFrame(n).LastPackage.GetPkgAddr().Bech32())
				}
				res0 := gno.Go2GnoValue(
					m.Alloc,
					m.Store,
					reflect.ValueOf(pkgAddr),
				)
				addrT := store.GetType(gno.DeclaredTypeID("std", "Address"))
				res0.T = addrT
				m.PushValue(res0)
			},
		)
		pn.DefineNative("TestSetOrigCaller",
			gno.Flds( // params
				"", "Address",
			),
			gno.Flds( // results
			),
			func(m *gno.Machine) {
				arg0 := m.LastBlock().GetParams1().TV
				addr := arg0.GetString()
				// overwrite context
				ctx := m.Context.(stdlibs.ExecContext)
				ctx.OrigCaller = crypto.Bech32Address(addr)
				m.Context = ctx
			},
		)
		pn.DefineNative("TestSetOrigPkgAddr",
			gno.Flds( // params
				"", "Address",
			),
			gno.Flds( // results
			),
			func(m *gno.Machine) {
				arg0 := m.LastBlock().GetParams1().TV
				addr := crypto.Bech32Address(arg0.GetString())
				// overwrite context
				ctx := m.Context.(stdlibs.ExecContext)
				ctx.OrigPkgAddr = addr
				m.Context = ctx
			},
		)
		pn.DefineNative("TestSetOrigSend",
			gno.Flds( // params
				"sent", "Coins",
				"spent", "Coins",
			),
			gno.Flds( // results
			),
			func(m *gno.Machine) {
				arg0, arg1 := m.LastBlock().GetParams2()
				var sent std.Coins
				rvSent := reflect.ValueOf(&sent).Elem()
				gno.Gno2GoValue(arg0.TV, rvSent)
				sent = rvSent.Interface().(std.Coins) // needed?
				var spent std.Coins
				rvSpent := reflect.ValueOf(&spent).Elem()
				gno.Gno2GoValue(arg1.TV, rvSpent)
				spent = rvSpent.Interface().(std.Coins) // needed?
				// overwrite context.
				ctx := m.Context.(stdlibs.ExecContext)
				ctx.OrigSend = sent
				ctx.OrigSendSpent = &spent
				m.Context = ctx
			},
		)
		pn.DefineNative("TestIssueCoins",
			gno.Flds( // params
				"addr", "Address",
				"coins", "Coins",
			),
			gno.Flds( // results
			),
			func(m *gno.Machine) {
				arg0, arg1 := m.LastBlock().GetParams2()
				addr := crypto.Bech32Address(arg0.TV.GetString())
				var coins std.Coins
				rvCoins := reflect.ValueOf(&coins).Elem()
				gno.Gno2GoValue(arg1.TV, rvCoins)
				coins = rvCoins.Interface().(std.Coins) // needed?
				// overwrite context.
				ctx := m.Context.(stdlibs.ExecContext)
				banker := ctx.Banker
				for _, coin := range coins {
					banker.IssueCoin(addr, coin.Denom, coin.Amount)
				}
			},
		)
		pn.DefineNative("TestCurrentRealm",
			gno.Flds( // params
			),
			gno.Flds( // results
				"realm", "string",
			),
			func(m *gno.Machine) {
				rlmpath := m.Realm.Path
				m.PushValue(typedString(rlmpath))
			},
		)
		pn.DefineNative("TestSkipHeights",
			gno.Flds( // params
				"count", "int64",
			),
			gno.Flds( // results
			),
			func(m *gno.Machine) {
				arg0 := m.LastBlock().GetParams1().TV
				count := arg0.GetInt64()

				ctx := m.Context.(stdlibs.ExecContext)
				ctx.Height += count
				m.Context = ctx
			},
		)
		// TODO: move elsewhere.
		pn.DefineNative("ClearStoreCache",
			gno.Flds( // params
			),
			gno.Flds( // results
			),
			func(m *gno.Machine) {
				if gno.IsDebug() && testing.Verbose() {
					store.Print()
					fmt.Println("========================================")
					fmt.Println("CLEAR CACHE (RUNTIME)")
					fmt.Println("========================================")
				}
				m.Store.ClearCache()
				m.PreprocessAllFilesAndSaveBlockNodes()
				if gno.IsDebug() && testing.Verbose() {
					store.Print()
					fmt.Println("========================================")
					fmt.Println("CLEAR CACHE DONE")
					fmt.Println("========================================")
				}
			},
		)
	}
}

//----------------------------------------

type dummyReader struct{}

func (*dummyReader) Read(b []byte) (n int, err error) {
	for i := 0; i < len(b); i++ {
		b[i] = byte((100 + i) % 256)
	}
	return len(b), nil
}

//----------------------------------------

// NOTE: does not allocate; used for panics.
func typedString(s string) gno.TypedValue {
	tv := gno.TypedValue{T: gno.StringType}
	tv.V = gno.StringValue(s)
	return tv
}

type TestReport struct {
	Name    string
	Verbose bool
	Failed  bool
	Skipped bool
	Output  string
}
