package wspace3

import (
	"gno.example.com/r/wspace3/subwork"
	"gno.example.com/r/wspace3/subwork/subworkpkg"
)

func Foo() {
	subworkpkg.Foo()
	subwork.Foo()
}
