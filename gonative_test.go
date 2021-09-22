package gno

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/jaekwon/testify/assert"
)

type Foo struct {
	A int
	B int32
	C int64
	D string
}

func TestGoNativeDefine(t *testing.T) {
	// Create package foo and define Foo.
	pkg := NewPackageNode("foo", "test.foo", nil)
	rt := reflect.TypeOf(Foo{})
	pkg.DefineGoNativeType(rt)
	nt := pkg.GetValueRef(nil, Name("Foo")).GetType().(*nativeType)
	assert.Equal(t, nt.Type, rt)
	path := pkg.GetPathForName(nil, Name("Foo"))
	assert.Equal(t, path.Depth, uint8(1))
	assert.Equal(t, path.Index, uint16(0))
	pv := pkg.NewPackage(nil)
	nt = pv.GetPointerTo(nil, path).TV.GetType().(*nativeType)
	assert.Equal(t, nt.Type, rt)

	// Import above package and evaluate foo.Foo.
	m := NewMachineWithOptions(MachineOptions{
		Store: TestStore{
			GetPackageFn: (func(pkgPath string) *PackageValue {
				switch pkgPath {
				case "test.foo":
					return pv
				default:
					panic("unknown package path " + pkgPath)
				}
			}),
		},
	})
	m.RunDeclaration(ImportD("foo", "test.foo"))
	tv := m.Eval(Sel(Nx("foo"), "Foo"))
	assert.Equal(t, tv.V.(TypeValue).Type, nt)
}

func TestGoNativeDefine2(t *testing.T) {
	// Create package foo and define Foo.
	pkg := NewPackageNode("foo", "test.foo", nil)
	rt := reflect.TypeOf(Foo{})
	pkg.DefineGoNativeType(rt)
	pv := pkg.NewPackage(nil)

	// Import above package and run file.
	out := new(bytes.Buffer)
	m := NewMachineWithOptions(MachineOptions{
		Output: out,
		Store: TestStore{
			GetPackageFn: (func(pkgPath string) *PackageValue {
				switch pkgPath {
				case "test.foo":
					return pv
				default:
					panic("unknown package path " + pkgPath)
				}
			}),
		},
	})

	c := `package main
import foo "test.foo"
func main() {
	f := foo.Foo{A:1}
	println("A:", f.A)
	println("B:", f.B)
	println("C:", f.C)
	println("D:", f.D)
}`
	n := MustParseFile("main.go", c)
	m.RunFiles(n)
	m.RunMain()
	assert.Equal(t, string(out.Bytes()), `A: 1
B: 0
C: 0
D: 
`)
}

func TestGoNativeDefine3(t *testing.T) {
	// Create package foo and define Foo.
	out := new(bytes.Buffer)
	pkg := NewPackageNode("foo", "test.foo", nil)
	pkg.DefineGoNativeType(reflect.TypeOf(Foo{}))
	pkg.DefineGoNativeFunc("printFoo", func(f Foo) {
		out.Write([]byte(fmt.Sprintf("A: %v\n", f.A)))
		out.Write([]byte(fmt.Sprintf("B: %v\n", f.B)))
		out.Write([]byte(fmt.Sprintf("C: %v\n", f.C)))
		out.Write([]byte(fmt.Sprintf("D: %v\n", f.D)))
	})
	pv := pkg.NewPackage(nil)

	// Import above package and run file.
	m := NewMachineWithOptions(MachineOptions{
		Output: out,
		Store: TestStore{
			GetPackageFn: (func(pkgPath string) *PackageValue {
				switch pkgPath {
				case "test.foo":
					return pv
				default:
					panic("unknown package path " + pkgPath)
				}
			}),
		},
	})

	c := `package main
import foo "test.foo"
func main() {
	f := foo.Foo{A:1}
	foo.printFoo(f)
}`
	n := MustParseFile("main.go", c)
	m.RunFiles(n)
	m.RunMain()
	assert.Equal(t, string(out.Bytes()), `A: 1
B: 0
C: 0
D: 
`)
}
