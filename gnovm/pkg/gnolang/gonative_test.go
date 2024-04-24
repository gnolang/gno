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

func TestCollectInterfaceMethods(t *testing.T) {
	ift := &InterfaceType{
		Methods: []FieldType{
			{
				Name: "Foo",
				Type: &FuncType{
					Params:  nil,
					Results: nil,
				},
			},
			{
				Name: "Bar",
				Type: &FuncType{
					Params: []FieldType{
						{Type: IntType},
						{Type: StringType},
					},
					Results: []FieldType{
						{Type: BoolType},
					},
				},
			},
		},
	}

	tests := []struct {
		name             string
		expectparamNums  int
		expectResultNums int
	}{
		{
			name:             "Foo",
			expectparamNums:  0,
			expectResultNums: 0,
		},
		{
			name:             "Bar",
			expectparamNums:  2,
			expectResultNums: 1,
		},
	}

	methods, err := collectInterfaceMethods(ift)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(methods) != 2 {
		t.Fatalf("expected 1 method, got %d", len(methods))
	}

	for i, tt := range tests {
		if methods[i].Name != tt.name {
			t.Fatalf("expected method name %s, got %s", tt.name, methods[i].Name)
		}

		if methods[i].Type.NumIn() != tt.expectparamNums {
			t.Fatalf("expected %d input arguments, got %d", tt.expectparamNums, methods[i].Type.NumIn())
		}

		if methods[i].Type.NumOut() != tt.expectResultNums {
			t.Fatalf("expected %d output arguments, got %d", tt.expectResultNums, methods[i].Type.NumOut())
		}
	}
}

func TestGno2GoType_Interface(t *testing.T) {
	emptyInterfaceType := &InterfaceType{
		Methods: []FieldType{},
	}

	result := gno2GoType(emptyInterfaceType)
	expectedType := createEmptyInterfaceType()

	if result != expectedType {
		t.Errorf("expected empty interface type, but got: %v", result)
	}

	// test interface with methods
	interfaceType := &InterfaceType{
		Methods: []FieldType{
			{
				Name: "Method1",
				Type: &FuncType{
					Params:  nil,
					Results: nil,
				},
			},
			{
				Name: "Method2",
				Type: &FuncType{
					Params: []FieldType{
						{Type: IntType},
						{Type: StringType},
					},
					Results: []FieldType{
						{Type: BoolType},
					},
				},
			},
		},
	}

	// test nil interface type
	result = gno2GoType(interfaceType)
	expectedNumMethods := 2
	if result.NumField() != expectedNumMethods {
		t.Errorf("expected %d methods, but got %d", expectedNumMethods, result.NumMethod())
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic for nil  InterfaceType, but got none")
		}
	}()
	gno2GoType(nil)
}

func TestGno2GoTypeMatches_InterfaceType(t *testing.T) {
	emptyInterfaceType := &InterfaceType{}
	nonEmptyInterfaceType := &InterfaceType{
		Methods: []FieldType{
			{
				Name: "Method1",
				Type: &FuncType{
					Params:  nil,
					Results: nil,
				},
			},
			{
				Name: "Method2",
				Type: &FuncType{
					Params: []FieldType{
						{Type: IntType},
					},
					Results: []FieldType{
						{Type: StringType},
					},
				},
			},
		},
	}

	// generate Go interface type
	type nonEmptyGoInterface interface {
		Method1()
		Method2(int) string
	}

	tests := []struct {
		name      string
		gnoIF     Type
		goIF      reflect.Type
		expectFit bool
	}{
		{
			name:      "empty interface",
			gnoIF:     emptyInterfaceType,
			goIF:      reflect.TypeOf((*interface{})(nil)).Elem(),
			expectFit: true,
		},
		{
			name:      "non-empty interface",
			gnoIF:     nonEmptyInterfaceType,
			goIF:      reflect.TypeOf((*nonEmptyGoInterface)(nil)).Elem(),
			expectFit: true,
		},
		{
			name:  "method count mismatch",
			gnoIF: emptyInterfaceType,
			goIF:  reflect.TypeOf((*nonEmptyGoInterface)(nil)).Elem(),
		},
		{
			name:  "mismatched method parameter type",
			gnoIF: nonEmptyInterfaceType,
			goIF: reflect.TypeOf((*interface {
				Method1()
				Method2(string) string
			})(nil)).Elem(),
		},
		{
			name:  "mismatched method return value",
			gnoIF: nonEmptyInterfaceType,
			goIF: reflect.TypeOf((*interface {
				Method1()
				Method2(int) int
			})(nil)).Elem(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fit := gno2GoTypeMatches(tt.gnoIF, tt.goIF)
			if fit != tt.expectFit {
				t.Errorf("expected fit %v, but got %v", tt.expectFit, fit)
			}
		})
	}
}
