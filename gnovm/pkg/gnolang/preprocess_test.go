package gnolang

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPreprocess_BinaryExpressionOneNative(t *testing.T) {
	pn := NewPackageNode("time", "time", nil)
	pn.DefineGoNativeValue("Millisecond", time.Millisecond)
	pn.DefineGoNativeValue("Second", time.Second)
	pn.DefineGoNativeType(reflect.TypeOf(time.Duration(0)))
	pv := pn.NewPackage()
	store := gonativeTestStore(pn, pv)
	store.SetBlockNode(pn)

	const src = `package main
	import "time"
func main() {
	var a int64 = 2
	println(time.Second * a)

}`
	n := MustParseFile("main.go", src)

	defer func() {
		err := recover()
		assert.Contains(t, fmt.Sprint(err), "incompatible operands in binary expression")
	}()
	initStaticBlocks(store, pn, n)
	Preprocess(store, pn, n)
}

func TestPreprocess_BinaryExpressionBothNative(t *testing.T) {
	pn := NewPackageNode("time", "time", nil)
	pn.DefineGoNativeValue("March", time.March)
	pn.DefineGoNativeValue("Wednesday", time.Wednesday)
	pn.DefineGoNativeType(reflect.TypeOf(time.Month(0)))
	pn.DefineGoNativeType(reflect.TypeOf(time.Weekday(0)))
	pv := pn.NewPackage()
	store := gonativeTestStore(pn, pv)
	store.SetBlockNode(pn)

	const src = `package main
	import "time"
func main() {
	println(time.March * time.Wednesday)

}`
	n := MustParseFile("main.go", src)

	defer func() {
		err := recover()
		assert.Contains(t, fmt.Sprint(err), "incompatible operands in binary expression")
	}()
	initStaticBlocks(store, pn, n)
	Preprocess(store, pn, n)
}

func TestEvalStaticTypeOf_MultipleValues(t *testing.T) {
	pn := NewPackageNode("main", "main", nil)
	pv := pn.NewPackage()
	store := gonativeTestStore(pn, pv)
	store.SetBlockNode(pn)

	const src = `package main
func multipleReturns() (int, string) {
	return 1, "hello"
}
func main() {
	x := multipleReturns()
}`
	n := MustParseFile("main.go", src)

	initStaticBlocks(store, pn, n)

	defer func() {
		err := recover()
		assert.NotNil(t, err, "Expected panic")
		errMsg := fmt.Sprint(err)
		assert.Contains(t, errMsg, "multipleReturns() (2 values) used as single value", "Unexpected error message")
		assert.Contains(t, errMsg, "Hint: Ensure the function returns a single value, or use multiple assignment", "Missing hint in error message")
	}()

	Preprocess(store, pn, n)
}

func TestEvalStaticTypeOf_NoValue(t *testing.T) {
	pn := NewPackageNode("main", "main", nil)
	pv := pn.NewPackage()
	store := gonativeTestStore(pn, pv)
	store.SetBlockNode(pn)

	const src = `package main

func main() {
	n := f()
}

func f() {
	println("hello!")
}
`
	n := MustParseFile("main.go", src)

	initStaticBlocks(store, pn, n)

	defer func() {
		err := recover()
		assert.NotNil(t, err, "Expected panic")
		errMsg := fmt.Sprint(err)
		assert.Contains(t, errMsg, "f() (no value) used as value", "Unexpected error message")
		assert.Contains(t, errMsg, "Hint: Ensure the function returns a single value, or use multiple assignment", "Missing hint in error message")
	}()

	Preprocess(store, pn, n)
}
