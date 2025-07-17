package callee

import (
	"github.com/gnolang/gno/gnovm/stdlibs/std"
	"gno.land/p/demo/avl"
)

var (
	localVar         string
	customVarPointer = &Custom{
		caller: "Pointer!",
		tree:   avl.NewTree(),
	}
)

type Custom struct {
	caller string
	tree   avl.Tree
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
