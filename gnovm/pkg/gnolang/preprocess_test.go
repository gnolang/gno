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

func TestPanicWhenAssigningBlankIdentifierToBlankIdentifier(t *testing.T) {
	pn := NewPackageNode("main", "main", nil)
	pv := pn.NewPackage()
	store := gonativeTestStore(pn, pv)
	store.SetBlockNode(pn)

	src := `package main

func main() {
	_ = _
}`

	n := MustParseFile("main.go", src)

	initStaticBlocks(store, pn, n)

	defer func() {
		err := recover()
		assert.NotNil(t, err, "Expected panic")
		errMsg := fmt.Sprint(err)
		assert.Contains(t, errMsg, "cannot use _ as value or type")
	}()

	Preprocess(store, pn, n)
}

func TestPanicWhenAssigningBlankIdentifierToVariable(t *testing.T) {
	pn := NewPackageNode("main", "main", nil)
	pv := pn.NewPackage()
	store := gonativeTestStore(pn, pv)
	store.SetBlockNode(pn)

	const src = `package main

func main() {
	a := _
}`

	n := MustParseFile("main.go", src)

	initStaticBlocks(store, pn, n)

	defer func() {
		err := recover()
		assert.NotNil(t, err, "Expected panic")
		errMsg := fmt.Sprint(err)
		assert.Contains(t, errMsg, "cannot use _ as value or type")
	}()

	Preprocess(store, pn, n)
}

func TestAssignmentsToBlankIdentifier(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		src  string
	}{
		{
			name: "ValidAssignmentToBlankIdentifier",
			src: `package main

func main() {
    _ = 1
}`,
		},
		{
			name: "AssignNilInterfaceToBlankIdentifier",
			src: `package main

type zilch interface{}

func main() {
    _ = zilch(nil)
}`,
		},
		{
			name: "AssignNilMapToBlankIdentifier",
			src: `package main

type anyMap map[string]interface{}

func main() {
    _ = anyMap(nil)
}`,
		},
		{
			name: "AssignNilStructToBlankIdentifier",
			src: `package main

type empty struct{}

func main() {
    _ = empty{}
}`,
		},
		{
			name: "AssignNilSliceToBlankIdentifier",
			src: `package main

type emptySlice []interface{}

func main() {
    _ = emptySlice(nil)
}`,
		},
		{
			name: "AssignNilFunctionToBlankIdentifier",
			src: `package main

type voidFunc func()

func main() {
    _ = voidFunc(nil)
}`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pn := NewPackageNode("main", "main", nil)
			pv := pn.NewPackage()
			store := gonativeTestStore(pn, pv)
			store.SetBlockNode(pn)

			n := MustParseFile("main.go", tt.src)

			initStaticBlocks(store, pn, n)

			res := Preprocess(store, pn, n)
			assert.NotNil(t, res)
		})
	}
}
