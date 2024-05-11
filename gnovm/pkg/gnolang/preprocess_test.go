package gnolang

import (
	"bytes"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPrepocessBinaryExpressionPrimaryAndNative(t *testing.T) {
	t.Parallel()

	out := new(bytes.Buffer)
	pkg := NewPackageNode("time", "time", nil)
	pkg.DefineGoNativeValue("Millisecond", time.Millisecond)
	pkg.DefineGoNativeValue("Second", time.Second)
	pkg.DefineGoNativeType(reflect.TypeOf(time.Duration(0)))
	pv := pkg.NewPackage()
	store := gonativeTestStore(pkg, pv)

	m := NewMachineWithOptions(MachineOptions{
		PkgPath: "main",
		Output:  out,
		Store:   store,
	})

	c := `package main
	import "time"
func main() {
	var a int64 = 2
	println(time.Second * a)
	
}`
	n := MustParseFile("main.go", c)
	assert.Panics(t, func() { m.RunFiles(n) })
}

func TestPrepocessBinaryExpressionNativeAndNative(t *testing.T) {
	t.Parallel()

	out := new(bytes.Buffer)
	pkg := NewPackageNode("time", "time", nil)
	pkg.DefineGoNativeValue("March", time.March)
	pkg.DefineGoNativeValue("Wednesday", time.Wednesday)
	pkg.DefineGoNativeType(reflect.TypeOf(time.Month(0)))
	pkg.DefineGoNativeType(reflect.TypeOf(time.Weekday(0)))
	pv := pkg.NewPackage()
	store := gonativeTestStore(pkg, pv)

	m := NewMachineWithOptions(MachineOptions{
		PkgPath: "main",
		Output:  out,
		Store:   store,
	})

	c := `package main
	import "time"
func main() {
	println(time.March * time.Wednesday)
	
}`
	n := MustParseFile("main.go", c)
	assert.Panics(t, func() { m.RunFiles(n) })
}
