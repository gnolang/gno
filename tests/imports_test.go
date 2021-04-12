package interp_test

import (
	"bufio"
	"bytes"
	"compress/flate"
	"context"
	crand "crypto/rand"
	"crypto/sha1"
	"encoding/binary"
	"encoding/xml"
	"flag"
	"fmt"
	"image"
	"image/color"
	"io"
	"io/ioutil"
	"math"
	"math/big"
	"math/rand"
	"net"
	"net/http"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gnolang/gno"
)

// NOTE: this isn't safe.
func testImporter(out io.Writer) (imp gno.Importer) {
	cache := make(map[string]*gno.PackageValue)
	imp = func(pkgPath string) (pv *gno.PackageValue) {
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
				Package:  pv,
				Output:   out,
				Importer: imp,
			})
			m2.RunFiles(fnodes...)
			return pv
		}
		// construct built-in package value.
		switch pkgPath {
		case "fmt":
			pkg := gno.NewPackageNode("fmt", "fmt", nil)
			pkg.DefineGoNativeType(reflect.TypeOf((*fmt.Stringer)(nil)).Elem())
			pkg.DefineGoNativeFunc("Println", func(a ...interface{}) (n int, err error) {
				res := fmt.Sprintln(a...)
				return out.Write([]byte(res))
			})
			pkg.DefineGoNativeFunc("Printf", func(format string, a ...interface{}) (n int, err error) {
				res := fmt.Sprintf(format, a...)
				return out.Write([]byte(res))
			})
			pkg.DefineGoNativeFunc("Sprintf", fmt.Sprintf)
			pkg.DefineGoNativeFunc("Errorf", fmt.Errorf)
			return pkg.NewPackage(nil)
		case "encoding/binary":
			pkg := gno.NewPackageNode("binary", "encoding/binary", nil)
			pkg.DefineGoNativeValue("LittleEndian", binary.LittleEndian)
			pkg.DefineGoNativeValue("BigEndian", binary.BigEndian)
			return pkg.NewPackage(nil)
		case "encoding/xml":
			pkg := gno.NewPackageNode("xml", "encoding/xml", nil)
			pkg.DefineGoNativeValue("Unmarshal", xml.Unmarshal)
			return pkg.NewPackage(nil)
		case "net":
			pkg := gno.NewPackageNode("net", "net", nil)
			pkg.DefineGoNativeType(reflect.TypeOf(net.TCPAddr{}))
			pkg.DefineGoNativeValue("IPv4", net.IPv4)
			return pkg.NewPackage(nil)
		case "net/http":
			// XXX UNSAFE
			// There's no reason why we can't replace these with safer alternatives.
			pkg := gno.NewPackageNode("http", "net/http", nil)
			pkg.DefineGoNativeType(reflect.TypeOf(http.Request{}))
			pkg.DefineGoNativeValue("DefaultClient", http.DefaultClient)
			pkg.DefineGoNativeType(reflect.TypeOf(http.Client{}))
			return pkg.NewPackage(nil)
		case "bufio":
			pkg := gno.NewPackageNode("bufio", "bufio", nil)
			pkg.DefineGoNativeValue("NewScanner", bufio.NewScanner)
			pkg.DefineGoNativeType(reflect.TypeOf(bufio.SplitFunc(nil)))
			return pkg.NewPackage(nil)
		case "bytes":
			pkg := gno.NewPackageNode("bytes", "bytes", nil)
			pkg.DefineGoNativeValue("NewReader", bytes.NewReader)
			pkg.DefineGoNativeValue("NewBuffer", bytes.NewBuffer)
			return pkg.NewPackage(nil)
		case "time":
			pkg := gno.NewPackageNode("time", "time", nil)
			pkg.DefineGoNativeValue("Date", time.Date)
			pkg.DefineGoNativeValue("Second", time.Second)
			pkg.DefineGoNativeValue("Minute", time.Minute)
			pkg.DefineGoNativeValue("Hour", time.Hour)
			pkg.DefineGoNativeType(reflect.TypeOf(time.Duration(0)))
			return pkg.NewPackage(nil)
		case "strings":
			pkg := gno.NewPackageNode("strings", "strings", nil)
			pkg.DefineGoNativeValue("SplitN", strings.SplitN)
			pkg.DefineGoNativeValue("HasPrefix", strings.HasPrefix)
			return pkg.NewPackage(nil)
		case "math":
			pkg := gno.NewPackageNode("math", "math", nil)
			pkg.DefineGoNativeValue("Abs", math.Abs)
			return pkg.NewPackage(nil)
		case "math/rand":
			pkg := gno.NewPackageNode("rand", "math/rand", nil)
			pkg.DefineGoNativeValue("Uint32", rand.Uint32)
			pkg.DefineGoNativeValue("Seed", rand.Seed)
			return pkg.NewPackage(nil)
		case "crypto/rand":
			pkg := gno.NewPackageNode("rand", "crypto/rand", nil)
			pkg.DefineGoNativeValue("Prime", crand.Prime)
			return pkg.NewPackage(nil)
		case "crypto/sha1":
			pkg := gno.NewPackageNode("sha1", "crypto/sha1", nil)
			pkg.DefineGoNativeValue("New", sha1.New)
			return pkg.NewPackage(nil)
		case "image":
			pkg := gno.NewPackageNode("image", "image", nil)
			pkg.DefineGoNativeType(reflect.TypeOf(image.Point{}))
			return pkg.NewPackage(nil)
		case "image/color":
			pkg := gno.NewPackageNode("color", "color", nil)
			pkg.DefineGoNativeType(reflect.TypeOf(color.NRGBA64{}))
			return pkg.NewPackage(nil)
		case "compress/flate":
			pkg := gno.NewPackageNode("flate", "flate", nil)
			pkg.DefineGoNativeValue("BestSpeed", flate.BestSpeed)
			return pkg.NewPackage(nil)
		case "context":
			pkg := gno.NewPackageNode("context", "context", nil)
			pkg.DefineGoNativeType(reflect.TypeOf((*context.Context)(nil)).Elem())
			pkg.DefineGoNativeValue("WithValue", context.WithValue)
			pkg.DefineGoNativeValue("Background", context.Background)
			return pkg.NewPackage(nil)
		case "strconv":
			pkg := gno.NewPackageNode("strconv", "strconv", nil)
			pkg.DefineGoNativeValue("Atoi", strconv.Atoi)
			pkg.DefineGoNativeValue("Itoa", strconv.Itoa)
			return pkg.NewPackage(nil)
		case "sync":
			pkg := gno.NewPackageNode("sync", "sync", nil)
			pkg.DefineGoNativeType(reflect.TypeOf(sync.RWMutex{}))
			pkg.DefineGoNativeType(reflect.TypeOf(sync.Pool{}))
			return pkg.NewPackage(nil)
		case "math/big":
			pkg := gno.NewPackageNode("big", "big", nil)
			pkg.DefineGoNativeValue("NewInt", big.NewInt)
			return pkg.NewPackage(nil)
		case "sort":
			pkg := gno.NewPackageNode("sort", "sort", nil)
			pkg.DefineGoNativeValue("Strings", sort.Strings)
			return pkg.NewPackage(nil)
		case "flag":
			pkg := gno.NewPackageNode("flag", "flag", nil)
			pkg.DefineGoNativeType(reflect.TypeOf(flag.Flag{}))
			return pkg.NewPackage(nil)
		default:
			panic("unknown package path " + pkgPath)
		}
	}
	return
}
