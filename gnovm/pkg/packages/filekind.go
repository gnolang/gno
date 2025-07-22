package packages

import (
	"go/parser"
	"go/token"
	"strings"
)

// FileKind represent the category a gno source file falls in, can be one of:
//
// - [FileKindPackageSource] -> A *.gno file that will be included in the gnovm package
//
// - [FileKindTest] -> A *_test.gno file that will be used for testing
//
// - [FileKindXTest] -> A *_test.gno file with a package name ending in _test that will be used for blackbox testing
//
// - [FileKindFiletest] -> A *_filetest.gno file that will be used for snapshot testing
type FileKind string

const (
	FileKindUnknown       FileKind = ""
	FileKindPackageSource          = "PackageSource"
	FileKindTest                   = "Test"
	FileKindXTest                  = "XTest"
	FileKindFiletest               = "Filetest"
	FileKindOther                  = "Other"
)

func GnoFileKinds() []FileKind {
	return []FileKind{FileKindPackageSource, FileKindTest, FileKindXTest, FileKindFiletest}
}

// GetFileKind analyzes a file's name and body to get it's [FileKind], fset is optional
func GetFileKind(filename string, body string, fset *token.FileSet) FileKind {
	if !strings.HasSuffix(filename, ".gno") {
		return FileKindOther
	}

	if strings.HasSuffix(filename, "_filetest.gno") {
		return FileKindFiletest
	}

	if !strings.HasSuffix(filename, "_test.gno") {
		return FileKindPackageSource
	}

	if fset == nil {
		fset = token.NewFileSet()
	}
	ast, err := parser.ParseFile(fset, filename, body, parser.PackageClauseOnly)
	if err != nil {
		return FileKindTest
	}
	packageName := ast.Name.Name

	if strings.HasSuffix(packageName, "_test") {
		return FileKindXTest
	}
	return FileKindTest
}
