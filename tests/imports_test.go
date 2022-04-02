package tests

import (
	"bufio"
	"bytes"
	"compress/flate"
	"compress/gzip"
	"context"
	"crypto/md5"
	crand "crypto/rand"
	"crypto/sha1"
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
	"text/template"
	"time"
	"unicode/utf8"

	"github.com/gnolang/gno"
	"github.com/gnolang/gno/pkgs/crypto"
	dbm "github.com/gnolang/gno/pkgs/db"
	osm "github.com/gnolang/gno/pkgs/os"
	"github.com/gnolang/gno/pkgs/std"
	"github.com/gnolang/gno/pkgs/store/dbadapter"
	"github.com/gnolang/gno/pkgs/store/iavl"
	stypes "github.com/gnolang/gno/pkgs/store/types"
	"github.com/gnolang/gno/stdlibs"
)

// NOTE: this isn't safe.
func testStore(stdin io.Reader, stdout, stderr io.Writer, nativeLibs bool) (store gno.Store) {
	filesPath := "./files"
	if nativeLibs {
		filesPath = "./files2"
	}
	getPackage := func(pkgPath string) (pn *gno.PackageNode, pv *gno.PackageValue) {
		if pkgPath == "" {
			panic(fmt.Sprintf("invalid zero package path in testStore().pkgGetter"))
		}
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
		// TODO: if isRealm, can we panic here?
		// otherwise, built-in package value.
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
			pkg.DefineGoNativeValue("Sscanf", fmt.Sscanf)
			pkg.DefineGoNativeValue("Errorf", fmt.Errorf)
			pkg.DefineGoNativeValue("Fprintln", fmt.Fprintln)
			pkg.DefineGoNativeValue("Fprintf", fmt.Fprintf)
			pkg.DefineGoNativeValue("Fprint", fmt.Fprint)
			return pkg, pkg.NewPackage()
		case "encoding/base64":
			if nativeLibs {
				pkg := gno.NewPackageNode("base64", pkgPath, nil)
				pkg.DefineGoNativeValue("RawStdEncoding", base64.RawStdEncoding)
				pkg.DefineGoNativeValue("StdEncoding", base64.StdEncoding)
				pkg.DefineGoNativeValue("NewDecoder", base64.NewDecoder)
				return pkg, pkg.NewPackage()
			}
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
			if nativeLibs {
				pkg := gno.NewPackageNode("bufio", pkgPath, nil)
				pkg.DefineGoNativeValue("NewScanner", bufio.NewScanner)
				pkg.DefineGoNativeType(reflect.TypeOf(bufio.SplitFunc(nil)))
				return pkg, pkg.NewPackage()
			}
		case "bytes":
			if nativeLibs {
				pkg := gno.NewPackageNode("bytes", pkgPath, nil)
				pkg.DefineGoNativeValue("Equal", bytes.Equal)
				pkg.DefineGoNativeValue("Compare", bytes.Compare)
				pkg.DefineGoNativeValue("NewReader", bytes.NewReader)
				pkg.DefineGoNativeValue("NewBuffer", bytes.NewBuffer)
				pkg.DefineGoNativeValue("Repeat", bytes.Repeat)
				pkg.DefineGoNativeType(reflect.TypeOf(bytes.Buffer{}))
				return pkg, pkg.NewPackage()
			}
		case "time":
			pkg := gno.NewPackageNode("time", pkgPath, nil)
			pkg.DefineGoNativeValue("Date", time.Date)
			pkg.DefineGoNativeValue("Second", time.Second)
			pkg.DefineGoNativeValue("Minute", time.Minute)
			pkg.DefineGoNativeValue("Hour", time.Hour)
			pkg.DefineGoNativeValue("Now", func() time.Time { return time.Unix(0, 0).UTC() }) // deterministic
			pkg.DefineGoNativeValue("November", time.November)
			pkg.DefineGoNativeValue("UTC", time.UTC)
			pkg.DefineGoNativeValue("Unix", time.Unix)
			pkg.DefineGoNativeType(reflect.TypeOf(time.Time{}))
			pkg.DefineGoNativeType(reflect.TypeOf(time.Month(0)))
			pkg.DefineGoNativeType(reflect.TypeOf(time.Duration(0)))
			return pkg, pkg.NewPackage()
		case "strings":
			if nativeLibs {
				pkg := gno.NewPackageNode("strings", pkgPath, nil)
				pkg.DefineGoNativeValue("Split", strings.Split)
				pkg.DefineGoNativeValue("SplitN", strings.SplitN)
				pkg.DefineGoNativeValue("Contains", strings.Contains)
				pkg.DefineGoNativeValue("TrimSpace", strings.TrimSpace)
				pkg.DefineGoNativeValue("HasPrefix", strings.HasPrefix)
				pkg.DefineGoNativeValue("NewReader", strings.NewReader)
				pkg.DefineGoNativeValue("Index", strings.Index)
				pkg.DefineGoNativeValue("IndexRune", strings.IndexRune)
				pkg.DefineGoNativeType(reflect.TypeOf(strings.Builder{}))
				return pkg, pkg.NewPackage()
			}
		case "math":
			if nativeLibs {
				pkg := gno.NewPackageNode("math", pkgPath, nil)
				pkg.DefineGoNativeValue("Abs", math.Abs)
				pkg.DefineGoNativeValue("Cos", math.Cos)
				pkg.DefineGoNativeValue("Pi", math.Pi)
				pkg.DefineGoNativeValue("MaxFloat32", math.MaxFloat32)
				return pkg, pkg.NewPackage()
			}
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
			if nativeLibs {
				pkg := gno.NewPackageNode("sort", pkgPath, nil)
				pkg.DefineGoNativeValue("Strings", sort.Strings)
				// pkg.DefineGoNativeValue("Sort", sort.Sort) not supported
				return pkg, pkg.NewPackage()
			}
		case "flag":
			pkg := gno.NewPackageNode("flag", pkgPath, nil)
			pkg.DefineGoNativeType(reflect.TypeOf(flag.Flag{}))
			return pkg, pkg.NewPackage()
		case "io":
			if nativeLibs {
				pkg := gno.NewPackageNode("io", pkgPath, nil)
				pkg.DefineGoNativeValue("EOF", io.EOF)
				pkg.DefineGoNativeValue("ReadFull", io.ReadFull)
				pkg.DefineGoNativeType(reflect.TypeOf((*io.ReadCloser)(nil)).Elem())
				pkg.DefineGoNativeType(reflect.TypeOf((*io.Closer)(nil)).Elem())
				pkg.DefineGoNativeType(reflect.TypeOf((*io.Reader)(nil)).Elem())
				return pkg, pkg.NewPackage()
			}
		case "io/ioutil":
			if nativeLibs {
				pkg := gno.NewPackageNode("ioutil", pkgPath, nil)
				pkg.DefineGoNativeValue("NopCloser", ioutil.NopCloser)
				pkg.DefineGoNativeValue("ReadAll", ioutil.ReadAll)
				return pkg, pkg.NewPackage()
			}
		case "log":
			pkg := gno.NewPackageNode("log", pkgPath, nil)
			pkg.DefineGoNativeValue("Fatal", log.Fatal)
			return pkg, pkg.NewPackage()
		case "text/template":
			pkg := gno.NewPackageNode("template", pkgPath, nil)
			pkg.DefineGoNativeType(reflect.TypeOf(template.FuncMap{}))
			return pkg, pkg.NewPackage()
		case "unicode/utf8":
			if nativeLibs {
				pkg := gno.NewPackageNode("utf8", pkgPath, nil)
				pkg.DefineGoNativeValue("DecodeRuneInString", utf8.DecodeRuneInString)
				tv := gno.TypedValue{T: gno.UntypedRuneType} // TODO dry
				tv.SetInt32(utf8.RuneSelf)                   // ..
				pkg.Define("RuneSelf", tv)                   // ..
				return pkg, pkg.NewPackage()
			}
		case "errors":
			if nativeLibs {
				pkg := gno.NewPackageNode("errors", pkgPath, nil)
				pkg.DefineGoNativeValue("New", errors.New)
				return pkg, pkg.NewPackage()
			}
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
		// if stdlibs package...
		stdlibPath := filepath.Join("../stdlibs", pkgPath)
		if osm.DirExists(stdlibPath) {
			memPkg := gno.ReadMemPackage(stdlibPath, pkgPath)
			m2 := gno.NewMachineWithOptions(gno.MachineOptions{
				PkgPath: "test",
				Output:  stdout,
				Store:   store,
			})
			return m2.RunMemPackage(memPkg, true)
		}
		// if examples package...
		examplePath := filepath.Join("../examples", pkgPath)
		if osm.DirExists(examplePath) {
			memPkg := gno.ReadMemPackage(examplePath, pkgPath)
			m2 := gno.NewMachineWithOptions(gno.MachineOptions{
				PkgPath: "test",
				Output:  stdout,
				Store:   store,
			})
			return m2.RunMemPackage(memPkg, true)
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
	return
}

//----------------------------------------
// testInjectNatives
// analogous to stdlibs.InjectNatives, but with
// native methods suitable for the testing environment.

func testPackageInjector(store gno.Store, pn *gno.PackageNode) {
	// Also inject stdlibs native functions.
	stdlibs.InjectPackage(store, pn)
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
				isOrigin := len(m.Frames) == 3
				if !isOrigin {
					panic("invalid non-origin call")
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
				isOrigin := len(m.Frames) == 3
				res0 := gno.TypedValue{T: gno.BoolType}
				res0.SetBool(isOrigin)
				m.PushValue(res0)
			},
		)
		pn.DefineNativeOverride("GetCallerAt",
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
					panic("GetCallerAt requires positive arg")
				}
				if n >= m.NumFrames() {
					// NOTE: the last frame's LastPackage
					// is set to the original non-frame
					// package, so need this check.
					panic("frame not found")
				}
				var pkgAddr string
				if n == m.NumFrames()-1 {
					// This makes it consistent with GetOrigCaller and TestSetOrigCaller.
					ctx := m.Context.(stdlibs.ExecContext)
					pkgAddr = string(ctx.OrigCaller)
				} else {
					pkgAddr = string(m.LastCallFrame(n).LastPackage.GetPkgAddr().Bech32())
				}
				res0 := gno.Go2GnoValue(
					m.Alloc,
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
				ctx := m.Context.(stdlibs.ExecContext)
				ctx.OrigCaller = crypto.Bech32Address(addr)
				m.Context = ctx // NOTE: tramples context for testing.
			},
		)
		pn.DefineNative("TestSetTxSend",
			gno.Flds( // params
				"", "Coins",
			),
			gno.Flds( // results
			),
			func(m *gno.Machine) {
				arg0 := m.LastBlock().GetParams1().TV
				var coins std.Coins
				rv := reflect.ValueOf(&coins).Elem()
				gno.Gno2GoValue(arg0, rv)
				coins = rv.Interface().(std.Coins)
				ctx := m.Context.(stdlibs.ExecContext)
				ctx.TxSend = coins
				m.Context = ctx // NOTE: tramples context for testing.
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
