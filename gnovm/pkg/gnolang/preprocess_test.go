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

func TestIsNamedConversion(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		xt       Type
		t        Type
		expected bool
		panic    bool
	}{
		{
			name:     "both nil types",
			xt:       nil,
			t:        nil,
			expected: false,
			panic:    true,
		},
		{
			name:     "both named",
			xt:       &DeclaredType{Name: "MyInt1", Base: IntType},
			t:        &DeclaredType{Name: "MyInt2", Base: IntType},
			expected: false,
		},
		{
			name:     "both unnamed",
			xt:       IntType,
			t:        IntType,
			expected: false,
		},
		{
			name:     "t is interface",
			xt:       &DeclaredType{Name: "MyInt", Base: IntType},
			t:        &InterfaceType{},
			expected: false,
		},
		{
			name:     "xt is TypeType",
			xt:       &TypeType{},
			t:        IntType,
			expected: false,
		},
		{
			name:     "t is TypeType",
			xt:       IntType,
			t:        &TypeType{},
			expected: false,
		},
		{
			name: "assign int to blank identifier",
			xt:   IntType,
			t:    &DeclaredType{Name: "_", Base: IntType},
		},
		{
			name:     "assign nil interface to blank identifier",
			xt:       &InterfaceType{},
			t:        &DeclaredType{Name: "_", Base: &InterfaceType{}},
			expected: true,
		},
		{
			name:     "assign nil map to blank identifier",
			xt:       &MapType{Key: StringType, Value: &InterfaceType{}},
			t:        &DeclaredType{Name: "_", Base: &MapType{Key: StringType, Value: &InterfaceType{}}},
			expected: true,
		},
		{
			name:     "assign empty struct to blank identifier",
			xt:       &StructType{},
			t:        &DeclaredType{Name: "_", Base: &StructType{}},
			expected: true,
		},
		{
			name:     "assign nil slice to blank identifier",
			xt:       &SliceType{Elt: &InterfaceType{}},
			t:        &DeclaredType{Name: "_", Base: &SliceType{Elt: &InterfaceType{}}},
			expected: true,
		},
		{
			name:     "assign nil function to blank identifier",
			xt:       &FuncType{},
			t:        &DeclaredType{Name: "_", Base: &FuncType{}},
			expected: true,
		},
		{
			name:     "xt is nil, t is not nil",
			xt:       nil,
			t:        IntType,
			expected: false,
		},
		{
			name:     "xt is nil, t is named type",
			xt:       nil,
			t:        &DeclaredType{Name: "MyInt", Base: IntType},
			expected: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.panic {
				defer func() {
					r := recover()
					if r == nil {
						t.Errorf("Expected panic, but none occurred")
					}
					if r != "cannot use _ as value or type" {
						t.Errorf("Expected panic, but none occurred")
					}
				}()
			}

			result := isNamedConversion(tt.xt, tt.t)
			if result != tt.expected {
				t.Errorf("Expected result %v, but got %v", tt.expected, result)
			}
		})
	}
}
