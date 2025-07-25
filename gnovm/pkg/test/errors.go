package test

import "fmt"

// XXX use it; this isn't used yet.
type TestImportError struct {
	PkgPath string
}

func (err TestImportError) Error() string {
	return fmt.Sprintf("unknown package path %q", err.PkgPath)
}

func (err TestImportError) String() string {
	return fmt.Sprintf("TestImportError(%q)", err.Error())
}
