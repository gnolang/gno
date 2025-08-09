package home

import (
	"std"
	"strings"
)

type myInterface interface {
	Render(path string) string
	Foo() error
}

var currentImpl myInterface

func SetImpl(impl myInterface) {
	// registration only accepts a nested package path
	if !strings.HasPrefix(std.PrevRealm().PkgPath(), "gno.land/r/x/manfred_upgrade_patterns/upgrade_e/") {
		panic("could only add implementation from a nested package")
	}
	currentImpl = impl
}

func Render(path string) string {
	assertImplIsDefined()
	return currentImpl.Render(path)
}

func Foo() error {
	assertImplIsDefined()
	return currentImpl.Foo()
}

func Bar() error {
	// doing some extra logic here
	err := currentImpl.Foo()
	// doing some more extra logic here
	return err
}

func assertImplIsDefined() {
	if currentImpl == nil {
		panic("no implementation")
	}
}
