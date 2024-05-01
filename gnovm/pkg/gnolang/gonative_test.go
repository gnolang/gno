package gnolang

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/stretchr/testify/assert"
)

// args is an even number of elements,
// the even index items are package nodes,
// and the odd index items are corresponding package values.
func gonativeTestStore(args ...interface{}) Store {
	store := NewStore(nil, nil, nil)
	store.SetPackageGetter(func(pkgPath string) (*PackageNode, *PackageValue) {
		for i := 0; i < len(args)/2; i++ {
			pn := args[i*2].(*PackageNode)
			pv := args[i*2+1].(*PackageValue)
			if pkgPath == pv.PkgPath {
				return pn, pv
			}
		}
		return nil, nil
	})
	store.SetStrictGo2GnoMapping(false)
	return store
}

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
	nt := pkg.GetValueRef(nil, Name("Foo")).GetType().(*NativeType)
	assert.Equal(t, rt, nt.Type)
	path := pkg.GetPathForName(nil, Name("Foo"))
	assert.Equal(t, uint8(1), path.Depth)
	assert.Equal(t, uint16(0), path.Index)
	pv := pkg.NewPackage()
	nt = pv.GetBlock(nil).GetPointerTo(nil, path).TV.GetType().(*NativeType)
	assert.Equal(t, rt, nt.Type)
	store := gonativeTestStore(pkg, pv)

	// Import above package and evaluate foo.Foo.
	m := NewMachineWithOptions(MachineOptions{
		PkgPath: "test",
		Store:   store,
	})
	m.RunDeclaration(ImportD("foo", "test.foo"))
	tvs := m.Eval(Sel(Nx("foo"), "Foo"))
	assert.Equal(t, 1, len(tvs))
	assert.Equal(t, nt, tvs[0].V.(TypeValue).Type)
}

func TestGoNativeDefine2(t *testing.T) {
	// Create package foo and define Foo.
	pkg := NewPackageNode("foo", "test.foo", nil)
	rt := reflect.TypeOf(Foo{})
	pkg.DefineGoNativeType(rt)
	pv := pkg.NewPackage()
	store := gonativeTestStore(pkg, pv)

	// Import above package and run file.
	out := new(bytes.Buffer)
	m := NewMachineWithOptions(MachineOptions{
		PkgPath: "main",
		Output:  out,
		Store:   store,
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
	assert.Equal(t, `A: 1
B: 0
C: 0
D: 
`, out.String())
}

func TestGoNativeDefine3(t *testing.T) {
	t.Parallel()

	// Create package foo and define Foo.
	out := new(bytes.Buffer)
	pkg := NewPackageNode("foo", "test.foo", nil)
	pkg.DefineGoNativeType(reflect.TypeOf(Foo{}))
	pkg.DefineGoNativeValue("PrintFoo", func(f Foo) {
		out.Write([]byte(fmt.Sprintf("A: %v\n", f.A)))
		out.Write([]byte(fmt.Sprintf("B: %v\n", f.B)))
		out.Write([]byte(fmt.Sprintf("C: %v\n", f.C)))
		out.Write([]byte(fmt.Sprintf("D: %v\n", f.D)))
	})
	pv := pkg.NewPackage()
	store := gonativeTestStore(pkg, pv)

	// Import above package and run file.
	m := NewMachineWithOptions(MachineOptions{
		PkgPath: "main",
		Output:  out,
		Store:   store,
	})

	c := `package main
import foo "test.foo"
func main() {
	f := foo.Foo{A:1}
	foo.PrintFoo(f)
}`
	n := MustParseFile("main.go", c)
	m.RunFiles(n)
	m.RunMain()
	assert.Equal(t, `A: 1
B: 0
C: 0
D: 
`, out.String())
}

func TestCrypto(t *testing.T) {
	t.Parallel()

	addr := crypto.Address{}
	store := gonativeTestStore()
	tv := Go2GnoValue(nilAllocator, store, reflect.ValueOf(addr))
	assert.Equal(t,
		`(array[0x0000000000000000000000000000000000000000] github.com/gnolang/gno/tm2/pkg/crypto.Address)`,
		tv.String())
}
