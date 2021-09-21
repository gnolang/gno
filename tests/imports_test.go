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
)

// NOTE: this isn't safe.
func testStore(out io.Writer) (store gno.Store) {
	cache := make(map[string]*gno.PackageValue)
	getPackage := func(pkgPath string) (pv *gno.PackageValue) {
		// look up cache.
		if pv, exists := cache[pkgPath]; exists {
			if pv == nil {
				panic(fmt.Sprintf(
					"import cycle detected: %q",
					pkgPath))
			}
			return pv
		}
		// set entry to detect import cycles.
		cache[pkgPath] = nil
		// defer: save to cache.
		defer func() {
			cache[pkgPath] = pv
		}()
		// construct test package value.
		const testPath = "github.com/gnolang/gno/_test/"
		if strings.HasPrefix(pkgPath, testPath) {
			baseDir := filepath.Join("./files/extern", pkgPath[len(testPath):])
			pkgName := defaultPkgName(pkgPath)
			files, err := ioutil.ReadDir(baseDir)
			if err != nil {
				panic(err)
			}
			fnodes := []*gno.FileNode{}
			for i, file := range files {
				if filepath.Ext(file.Name()) != ".go" {
					continue
				}
				fpath := filepath.Join(baseDir, file.Name())
				fnode := gno.MustReadFile(fpath)
				if i == 0 {
					pkgName = fnode.PkgName
				} else if fnode.PkgName != pkgName {
					panic(fmt.Sprintf(
						"expected package name %q but got %v",
						pkgName,
						fnode.PkgName))
				}
				fnodes = append(fnodes, fnode)
			}
			pkg := gno.NewPackageNode(pkgName, pkgPath, nil)
			pv := pkg.NewPackage(nil)
			m2 := gno.NewMachineWithOptions(gno.MachineOptions{
				Package: pv,
				Output:  out,
				Store:   store,
			})
			m2.RunFiles(fnodes...)
			return pv
		}
		// construct built-in package value.
		switch pkgPath {
		case "fmt":
			pkg := gno.NewPackageNode("fmt", pkgPath, nil)
			pkg.DefineGoNativeType(reflect.TypeOf((*fmt.Stringer)(nil)).Elem())
			pkg.DefineGoNativeType(reflect.TypeOf((*fmt.Formatter)(nil)).Elem())
			pkg.DefineGoNativeFunc("Println", func(a ...interface{}) (n int, err error) {
				// NOTE: uncomment to debug long running tests
				fmt.Println(a...)
				res := fmt.Sprintln(a...)
				return out.Write([]byte(res))
			})
			pkg.DefineGoNativeFunc("Printf", func(format string, a ...interface{}) (n int, err error) {
				res := fmt.Sprintf(format, a...)
				return out.Write([]byte(res))
			})
			pkg.DefineGoNativeFunc("Print", func(a ...interface{}) (n int, err error) {
				res := fmt.Sprint(a...)
				return out.Write([]byte(res))
			})
			pkg.DefineGoNativeFunc("Sprintf", fmt.Sprintf)
			pkg.DefineGoNativeFunc("Sscanf", fmt.Sscanf)
			pkg.DefineGoNativeFunc("Errorf", fmt.Errorf)
			return pkg.NewPackage(nil)
		case "encoding/base64":
			pkg := gno.NewPackageNode("base64", pkgPath, nil)
			pkg.DefineGoNativeValue("RawStdEncoding", base64.RawStdEncoding)
			return pkg.NewPackage(nil)
		case "encoding/binary":
			pkg := gno.NewPackageNode("binary", pkgPath, nil)
			pkg.DefineGoNativeValue("LittleEndian", binary.LittleEndian)
			pkg.DefineGoNativeValue("BigEndian", binary.BigEndian)
			return pkg.NewPackage(nil)
		case "encoding/json":
			pkg := gno.NewPackageNode("json", pkgPath, nil)
			pkg.DefineGoNativeValue("Unmarshal", json.Unmarshal)
			pkg.DefineGoNativeValue("Marshal", json.Marshal)
			return pkg.NewPackage(nil)
		case "encoding/xml":
			pkg := gno.NewPackageNode("xml", pkgPath, nil)
			pkg.DefineGoNativeValue("Unmarshal", xml.Unmarshal)
			return pkg.NewPackage(nil)
		case "net":
			pkg := gno.NewPackageNode("net", pkgPath, nil)
			pkg.DefineGoNativeType(reflect.TypeOf(net.TCPAddr{}))
			pkg.DefineGoNativeValue("IPv4", net.IPv4)
			return pkg.NewPackage(nil)
		case "net/http":
			// XXX UNSAFE
			// There's no reason why we can't replace these with safer alternatives.
			panic("just say gno")
			/*
				pkg := gno.NewPackageNode("http", pkgPath, nil)
				pkg.DefineGoNativeType(reflect.TypeOf(http.Request{}))
				pkg.DefineGoNativeValue("DefaultClient", http.DefaultClient)
				pkg.DefineGoNativeType(reflect.TypeOf(http.Client{}))
				return pkg.NewPackage(nil)
			*/
		case "net/url":
			pkg := gno.NewPackageNode("url", pkgPath, nil)
			pkg.DefineGoNativeType(reflect.TypeOf(url.Values{}))
			return pkg.NewPackage(nil)
		case "bufio":
			pkg := gno.NewPackageNode("bufio", pkgPath, nil)
			pkg.DefineGoNativeValue("NewScanner", bufio.NewScanner)
			pkg.DefineGoNativeType(reflect.TypeOf(bufio.SplitFunc(nil)))
			return pkg.NewPackage(nil)
		case "bytes":
			pkg := gno.NewPackageNode("bytes", pkgPath, nil)
			pkg.DefineGoNativeValue("Equal", bytes.Equal)
			pkg.DefineGoNativeValue("Compare", bytes.Compare)
			pkg.DefineGoNativeValue("NewReader", bytes.NewReader)
			pkg.DefineGoNativeValue("NewBuffer", bytes.NewBuffer)
			pkg.DefineGoNativeType(reflect.TypeOf(bytes.Buffer{}))
			return pkg.NewPackage(nil)
		case "time":
			pkg := gno.NewPackageNode("time", pkgPath, nil)
			pkg.DefineGoNativeValue("Date", time.Date)
			pkg.DefineGoNativeValue("Second", time.Second)
			pkg.DefineGoNativeValue("Minute", time.Minute)
			pkg.DefineGoNativeValue("Hour", time.Hour)
			pkg.DefineGoNativeValue("Now", func() time.Time { return time.Unix(0, 0) }) // deterministic
			pkg.DefineGoNativeValue("November", time.November)
			pkg.DefineGoNativeValue("UTC", time.UTC)
			pkg.DefineGoNativeValue("Unix", time.Unix)
			pkg.DefineGoNativeType(reflect.TypeOf(time.Time{}))
			pkg.DefineGoNativeType(reflect.TypeOf(time.Month(0)))
			pkg.DefineGoNativeType(reflect.TypeOf(time.Duration(0)))
			return pkg.NewPackage(nil)
		case "strings":
			pkg := gno.NewPackageNode("strings", pkgPath, nil)
			pkg.DefineGoNativeValue("SplitN", strings.SplitN)
			pkg.DefineGoNativeValue("Contains", strings.Contains)
			pkg.DefineGoNativeValue("TrimSpace", strings.TrimSpace)
			pkg.DefineGoNativeValue("HasPrefix", strings.HasPrefix)
			pkg.DefineGoNativeValue("NewReader", strings.NewReader)
			return pkg.NewPackage(nil)
		case "math":
			pkg := gno.NewPackageNode("math", pkgPath, nil)
			pkg.DefineGoNativeValue("Abs", math.Abs)
			pkg.DefineGoNativeValue("Cos", math.Cos)
			pkg.DefineGoNativeValue("Pi", math.Pi)
			pkg.DefineGoNativeValue("MaxFloat32", math.MaxFloat32)
			return pkg.NewPackage(nil)
		case "math/rand":
			pkg := gno.NewPackageNode("rand", pkgPath, nil)
			pkg.DefineGoNativeValue("Uint32", rand.Uint32)
			pkg.DefineGoNativeValue("Seed", rand.Seed)
			return pkg.NewPackage(nil)
		case "crypto/rand":
			pkg := gno.NewPackageNode("rand", pkgPath, nil)
			pkg.DefineGoNativeValue("Prime", crand.Prime)
			// for determinism:
			// pkg.DefineGoNativeValue("Reader", crand.Reader)
			pkg.DefineGoNativeValue("Reader", &dummyReader{})
			return pkg.NewPackage(nil)
		case "crypto/md5":
			pkg := gno.NewPackageNode("md5", pkgPath, nil)
			pkg.DefineGoNativeValue("New", md5.New)
			return pkg.NewPackage(nil)
		case "crypto/sha1":
			pkg := gno.NewPackageNode("sha1", pkgPath, nil)
			pkg.DefineGoNativeValue("New", sha1.New)
			return pkg.NewPackage(nil)
		case "image":
			pkg := gno.NewPackageNode("image", pkgPath, nil)
			pkg.DefineGoNativeType(reflect.TypeOf(image.Point{}))
			return pkg.NewPackage(nil)
		case "image/color":
			pkg := gno.NewPackageNode("color", pkgPath, nil)
			pkg.DefineGoNativeType(reflect.TypeOf(color.NRGBA64{}))
			return pkg.NewPackage(nil)
		case "compress/flate":
			pkg := gno.NewPackageNode("flate", pkgPath, nil)
			pkg.DefineGoNativeValue("BestSpeed", flate.BestSpeed)
			return pkg.NewPackage(nil)
		case "compress/gzip":
			pkg := gno.NewPackageNode("gzip", pkgPath, nil)
			pkg.DefineGoNativeType(reflect.TypeOf(gzip.Writer{}))
			pkg.DefineGoNativeValue("BestCompression", gzip.BestCompression)
			pkg.DefineGoNativeValue("BestSpeed", gzip.BestSpeed)
			return pkg.NewPackage(nil)
		case "context":
			pkg := gno.NewPackageNode("context", pkgPath, nil)
			pkg.DefineGoNativeType(reflect.TypeOf((*context.Context)(nil)).Elem())
			pkg.DefineGoNativeValue("WithValue", context.WithValue)
			pkg.DefineGoNativeValue("Background", context.Background)
			return pkg.NewPackage(nil)
		case "strconv":
			pkg := gno.NewPackageNode("strconv", pkgPath, nil)
			pkg.DefineGoNativeType(reflect.TypeOf(strconv.NumError{}))
			pkg.DefineGoNativeValue("Atoi", strconv.Atoi)
			pkg.DefineGoNativeValue("Itoa", strconv.Itoa)
			pkg.DefineGoNativeValue("ParseInt", strconv.ParseInt)
			return pkg.NewPackage(nil)
		case "sync":
			pkg := gno.NewPackageNode("sync", pkgPath, nil)
			pkg.DefineGoNativeType(reflect.TypeOf(sync.Mutex{}))
			pkg.DefineGoNativeType(reflect.TypeOf(sync.RWMutex{}))
			pkg.DefineGoNativeType(reflect.TypeOf(sync.Pool{}))
			return pkg.NewPackage(nil)
		case "sync/atomic":
			pkg := gno.NewPackageNode("atomic", pkgPath, nil)
			pkg.DefineGoNativeType(reflect.TypeOf(atomic.Value{}))
			return pkg.NewPackage(nil)
		case "math/big":
			pkg := gno.NewPackageNode("big", pkgPath, nil)
			pkg.DefineGoNativeValue("NewInt", big.NewInt)
			return pkg.NewPackage(nil)
		case "sort":
			pkg := gno.NewPackageNode("sort", pkgPath, nil)
			pkg.DefineGoNativeValue("Strings", sort.Strings)
			return pkg.NewPackage(nil)
		case "flag":
			pkg := gno.NewPackageNode("flag", pkgPath, nil)
			pkg.DefineGoNativeType(reflect.TypeOf(flag.Flag{}))
			return pkg.NewPackage(nil)
		case "io":
			pkg := gno.NewPackageNode("io", pkgPath, nil)
			pkg.DefineGoNativeValue("EOF", io.EOF)
			pkg.DefineGoNativeValue("ReadFull", io.ReadFull)
			pkg.DefineGoNativeType(reflect.TypeOf((*io.ReadCloser)(nil)).Elem())
			pkg.DefineGoNativeType(reflect.TypeOf((*io.Closer)(nil)).Elem())
			pkg.DefineGoNativeType(reflect.TypeOf((*io.Reader)(nil)).Elem())
			return pkg.NewPackage(nil)
		case "io/ioutil":
			pkg := gno.NewPackageNode("ioutil", pkgPath, nil)
			pkg.DefineGoNativeValue("NopCloser", ioutil.NopCloser)
			pkg.DefineGoNativeValue("ReadAll", ioutil.ReadAll)
			return pkg.NewPackage(nil)
		case "log":
			pkg := gno.NewPackageNode("log", pkgPath, nil)
			pkg.DefineGoNativeValue("Fatal", log.Fatal)
			return pkg.NewPackage(nil)
		case "text/template":
			pkg := gno.NewPackageNode("template", pkgPath, nil)
			pkg.DefineGoNativeType(reflect.TypeOf(template.FuncMap{}))
			return pkg.NewPackage(nil)
		case "unicode/utf8":
			pkg := gno.NewPackageNode("utf8", pkgPath, nil)
			pkg.DefineGoNativeValue("DecodeRuneInString", utf8.DecodeRuneInString)
			tv := gno.TypedValue{T: gno.UntypedRuneType} // TODO dry
			tv.SetInt32(utf8.RuneSelf)                   // ..
			pkg.Define("RuneSelf", tv)                   // ..
			return pkg.NewPackage(nil)
		case "errors":
			pkg := gno.NewPackageNode("errors", pkgPath, nil)
			pkg.DefineGoNativeValue("New", errors.New)
			return pkg.NewPackage(nil)
		case "hash/fnv":
			pkg := gno.NewPackageNode("fnv", pkgPath, nil)
			pkg.DefineGoNativeValue("New32a", fnv.New32a)
			return pkg.NewPackage(nil)
		default:
			panic("unknown package path " + pkgPath)
		}
	}
	tstore := gno.TestStore{
		GetPackageFn: getPackage,
	}
	cstore := gno.NewCacheStore(tstore)
	return cstore
}

//----------------------------------------

type dummyReader struct{}

func (*dummyReader) Read(b []byte) (n int, err error) {
	for i := 0; i < len(b); i++ {
		b[i] = byte((100 + i) % 256)
	}
	return len(b), nil
}
