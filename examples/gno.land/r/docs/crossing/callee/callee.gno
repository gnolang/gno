package callee

import (
	"std"

	"gno.land/p/demo/avl"
)

var (
	localVar         string
	CustomVarPointer = &Custom{
		caller: "Pointer!",
		tree:   avl.NewTree(),
	}
)

type Custom struct {
	caller string
	tree   *avl.Tree
}

func Caller() string {
	return std.PreviousRealm().PkgPath()
}

func GetCustom(_ realm) *Custom {
	return CustomVarPointer
}

func (c *Custom) UpdateCaller(s string) string {
	localVar = s
	return localVar
}

// UpdateLocalVarCrossing is callable
func UpdateLocalVarCrossing(_ realm) {
	localVar = std.PreviousRealm().PkgPath() // will be the caller realm
}

// UpdateLocalVar is not callable
// It will panic with external realm error
func UpdateLocalVar() {
	localVar = std.PreviousRealm().PkgPath()
}

func UpdateCustomPointer() {

}

func UpdateCustomPointerCrossing() {

}

func UpdateCustomNonPointer() {

}

func UpdateCustomNonPointerCrossing() {

}
