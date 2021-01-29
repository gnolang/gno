package interp_test

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"image"
	"io"
	"math"
	"net"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/gnolang/gno"
)

// NOTE: this isn't safe.
func testImporter(out io.Writer) gno.Importer {
	return func(pkgPath string) *gno.PackageValue {
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
			return pkg.NewPackage(nil)
		case "bytes":
			pkg := gno.NewPackageNode("bytes", "bytes", nil)
			pkg.DefineGoNativeValue("NewReader", bytes.NewReader)
			return pkg.NewPackage(nil)
		case "time":
			pkg := gno.NewPackageNode("time", "time", nil)
			pkg.DefineGoNativeValue("Second", time.Second)
			return pkg.NewPackage(nil)
		case "strings":
			pkg := gno.NewPackageNode("strings", "strings", nil)
			pkg.DefineGoNativeValue("SplitN", strings.SplitN)
			pkg.DefineGoNativeValue("HasPrefix", strings.HasPrefix)
			return pkg.NewPackage(nil)
		case "crypto/sha1":
			pkg := gno.NewPackageNode("sha1", "crypto/sha1", nil)
			pkg.DefineGoNativeValue("New", sha1.New)
			return pkg.NewPackage(nil)
		case "math":
			pkg := gno.NewPackageNode("math", "math", nil)
			pkg.DefineGoNativeValue("Abs", math.Abs)
			return pkg.NewPackage(nil)
		case "image":
			pkg := gno.NewPackageNode("image", "image", nil)
			pkg.DefineGoNativeType(reflect.TypeOf(image.Point{}))
			return pkg.NewPackage(nil)
		default:
			panic("unknown package path " + pkgPath)
		}
	}
}
