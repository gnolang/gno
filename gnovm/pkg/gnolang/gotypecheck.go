package gnolang

import (
	"github.com/gnolang/gno/gnovm"
	"github.com/gnolang/gno/gnovm/pkg/internal/typecheck"
)

// MemPackageGetter implements the GetMemPackage() method. It is a subset of
// [Store], separated for ease of testing.
type MemPackageGetter interface {
	GetMemPackage(path string) *gnovm.MemPackage
}

// TypeCheckMemPackage performs type validation and checking on the given
// mempkg. To retrieve dependencies, it uses getter.
//
// The syntax checking is performed entirely using Go's go/types package.
//
// If format is true, the code will be automatically updated with the
// formatted source code.
func TypeCheckMemPackage(mempkg *gnovm.MemPackage, getter MemPackageGetter, format bool) error {
	return typecheck.CheckMemPackage(mempkg, getter, false, format)
}

// TypeCheckMemPackageTest performs the same type checks as [TypeCheckMemPackage],
// but allows re-declarations.
//
// Note: like TypeCheckMemPackage, this function ignores tests and filetests.
func TypeCheckMemPackageTest(mempkg *gnovm.MemPackage, getter MemPackageGetter) error {
	return typecheck.CheckMemPackage(mempkg, getter, true, false)
}
