package doc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	assert, require := assert.New(t), require.New(t)

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
	require.NoError(err)
	require.NotNil(pkg)

	assert.Equal(pkgPath, pkg.ImportPath)
	assert.Equal("example", pkg.Name)
	assert.Equal("Package example is an example package.\n", pkg.Doc)

	assert.Len(pkg.Filenames, 1)
	assert.Len(pkg.Consts, 1)
	assert.Len(pkg.Vars, 2)
	assert.Len(pkg.Funcs, 5)
	require.Len(pkg.Types, 2)

	myInterfaceType := pkg.Types[0]
	assert.Equal("MyInterface", myInterfaceType.Name)
	assert.Equal("MyInterface", myInterfaceType.ID)
	assert.Equal("A public interface.\n", myInterfaceType.Doc)
	assert.Len(myInterfaceType.Vars, 0)
	assert.Len(myInterfaceType.Consts, 0)
	assert.Len(myInterfaceType.Funcs, 0)
	assert.Len(myInterfaceType.Methods, 0)
	assert.Equal("type MyInterface interface {\n\tMyMethod() string\n}", myInterfaceType.Definition)

	myTypeType := pkg.Types[1]
	assert.Equal("MyType", myTypeType.Name)
	assert.Equal("MyType", myTypeType.ID)
	assert.Equal("A public type.\n", myTypeType.Doc)
	assert.Len(myTypeType.Vars, 0)
	assert.Len(myTypeType.Consts, 0)

	require.Len(myTypeType.Funcs, 1)
	assert.Equal("NewMyType", myTypeType.Funcs[0].Name)
	assert.Equal("NewMyType", myTypeType.Funcs[0].ID)
	assert.Equal("A function that returns MyType.\n", myTypeType.Funcs[0].Doc)
	assert.Len(myTypeType.Funcs[0].Params, 1)
	assert.Len(myTypeType.Funcs[0].Returns, 1)
	assert.Len(myTypeType.Funcs[0].Recv, 0)

	require.Len(myTypeType.Methods, 2)
	assert.Equal("NonPointerMethod", myTypeType.Methods[0].Name)
	assert.Equal("MyType.NonPointerMethod", myTypeType.Methods[0].ID)
	assert.Equal("A method without a pointer.\n", myTypeType.Methods[0].Doc)
	assert.Len(myTypeType.Methods[0].Params, 0)
	assert.Len(myTypeType.Methods[0].Returns, 0)
	assert.Len(myTypeType.Methods[0].Recv, 1)

	assert.Equal("PointerMethod", myTypeType.Methods[1].Name)
	assert.Equal("*MyType.PointerMethod", myTypeType.Methods[1].ID)
	assert.Equal("A method with a pointer.\n", myTypeType.Methods[1].Doc)
	assert.Len(myTypeType.Methods[1].Params, 0)
	assert.Len(myTypeType.Methods[1].Returns, 0)
	assert.Len(myTypeType.Methods[1].Recv, 1)

	assert.Equal("type MyType struct {\n\tName string // Name is a public field\n\t// contains filtered or unexported fields\n}", myTypeType.Definition)
}
