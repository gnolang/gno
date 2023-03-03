package doc

import (
	"testing"
)

func TestNew(t *testing.T) {
	files := map[string]string{
		"example.gno": `
// Package example is an example package.
package example
		
// A private variable.
var private string = "I'm private"
		
// A public variable.
var Public string = "I'm public"
		
// A public grouped variable.
var (
	Grouped1 string = "I'm Grouped1"
	Grouped2 string = "I'm Grouped2"
)
		
// A private constant.
const privateConst string = "I'm a private constant"
		
// A public constant.
const PublicConst string = "I'm a public constant"
		
// A private grouped constant.
const (
	groupedConst1 string = "I'm grouped const 1"
	groupedConst2 string = "I'm grouped const 2"
)
		
// A public type.
type MyType struct {
	Name string // Name is a public field
	age int // age is a private field
}
		
// A method with a pointer.
func (mt *MyType) PointerMethod() {}
		
// A method without a pointer.
func (mt MyType) NonPointerMethod() {}

// A function that returns MyType.
func NewMyType(name string) *MyType {
	return &MyType{Name: name}
}
		
// A function that takes a MyType as a parameter.
func UseMyType(mt *MyType) {}

// A private type.
type myPrivateType struct {}

// A public method with a private type.
func (mPT *myPrivateType) PublicMethod() {}
		
// A public interface.
type MyInterface interface {
	MyMethod() string
}
		
// A function that takes various types as parameters.
func ComplexFunction(s string, i int, f float64, b bool, a []string, fn func(), mt *MyType, iface MyInterface, mt2 MyType, fn2 func(string, int) (int, string)) {}
		
// A function that returns multiple values.
func MultipleReturnValues() (string, int) {
	return "gno", 42
}
		
// A function with named parameters and named return values.
func NamedParameters(firstParam int, secondParam string) (firstReturn string, secondReturn int) {
	return "gno", 42
}
		
// A function with unnamed parameters and unnamed return values.
func UnnamedParameters(int, string) (string, int) {
	return "gno", 42
}
`,
	}
	pkgPath := "gno.land/p/demo/example"
	pkg, err := New(pkgPath, files)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if pkg.ImportPath != pkgPath {
		t.Errorf("package import path: got %q, want %q", pkg.ImportPath, pkgPath)
	}

	pkgName := "example"
	if pkg.Name != pkgName {
		t.Errorf("package name: got %q, want %q", pkg.Name, pkgName)
	}

	pkgDoc := "Package example is an example package.\n"
	if pkg.Doc != pkgDoc {
		t.Errorf("package doc: got %q, want %q", pkg.Doc, pkgDoc)
	}

	if len(pkg.Filenames) != 1 {
		t.Errorf("package filenames: got %d, want 1 file", len(pkg.Filenames))
	}

	if len(pkg.Consts) != 1 {
		t.Errorf("package consts: got %d, want 1 const", len(pkg.Consts))
	}

	if len(pkg.Vars) != 2 {
		t.Errorf("package vars: got %d, want 2 vars", len(pkg.Vars))
	}

	if len(pkg.Funcs) != 5 {
		t.Errorf("package funcs: got %d, want 5 functions", len(pkg.Funcs))
	}

	if len(pkg.Types) != 2 {
		t.Errorf("package types: got %d, want 2 types", len(pkg.Types))
	} else {
		myInterfaceType := pkg.Types[0]
		if want := "MyInterface"; myInterfaceType.Name != want {
			t.Errorf("package type name: got %q, want %q", myInterfaceType.Name, want)
		}
		if want := "MyInterface"; myInterfaceType.ID != want {
			t.Errorf("package type id: got %q, want %q", myInterfaceType.Name, want)
		}
		if want := "A public interface.\n"; myInterfaceType.Doc != want {
			t.Errorf("package type doc: got %q, want %q", myInterfaceType.Doc, want)
		}
		if want := 0; len(myInterfaceType.Funcs) != want {
			t.Errorf("package type funcs: got %d, want %d", len(myInterfaceType.Funcs), want)
		}
		if want := 0; len(myInterfaceType.Methods) != want {
			t.Errorf("package type methods: got %d, want %d", len(myInterfaceType.Methods), want)
		}
		if want := 0; len(myInterfaceType.Vars) != want {
			t.Errorf("package type vars: got %d, want %d", len(myInterfaceType.Vars), want)
		}
		if want := 0; len(myInterfaceType.Consts) != want {
			t.Errorf("package type vars: got %d, want %d", len(myInterfaceType.Consts), want)
		}
		if want := "type MyInterface interface {\n\tMyMethod() string\n}"; myInterfaceType.Definition != want {
			t.Errorf("package type definition: got %q, want %q", myInterfaceType.Definition, want)
		}

		myTypeType := pkg.Types[1]
		if want := "MyType"; myTypeType.Name != want {
			t.Errorf("package type name: got %q, want %q", myTypeType.Name, want)
		}
		if want := "MyType"; myTypeType.ID != want {
			t.Errorf("package type id: got %q, want %q", myTypeType.Name, want)
		}
		if want := "A public type.\n"; myTypeType.Doc != want {
			t.Errorf("package type doc: got %q, want %q", myTypeType.Doc, want)
		}
		if want := 2; len(myTypeType.Methods) != want {
			t.Errorf("package type methods: got %d, want %d", len(myTypeType.Methods), want)
		}
		if want := 1; len(myTypeType.Funcs) != want {
			t.Errorf("package type funcs: got %d, want %d", len(myTypeType.Funcs), want)
		}
		if want := 0; len(myTypeType.Vars) != want {
			t.Errorf("package type vars: got %d, want %d", len(myTypeType.Vars), want)
		}
		if want := 0; len(myTypeType.Consts) != want {
			t.Errorf("package type consts: got %d, want %d", len(myTypeType.Consts), want)
		}
		if want := "type MyType struct {\n\tName string // Name is a public field\n\t// contains filtered or unexported fields\n}"; myTypeType.Definition != want {
			t.Errorf("package type definition: got %q, want %q", myTypeType.Definition, want)
		}
	}
}
