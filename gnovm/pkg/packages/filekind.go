package packages

import (
	"go/parser"
	"go/token"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/std"
)

// FileKind represent the category a gno package file falls in, can be one of:
//
// - [FileKindPackageSource] -> A *.gno file that will be included in the gnovm package
//
// - [FileKindTest] -> A *_test.gno file that will be used for testing
//
// - [FileKindXTest] -> A *_test.gno file with a package name ending in _test that will be used for blackbox testing
//
// - [FileKindFiletest] -> A *_filetest.gno file that will be used for snapshot testing
//
// - [FileKindOther] -> Any other file in the package, e.g.: *.toml, README.md, etc...
type FileKind string

const (
	FileKindUnknown       FileKind = ""
	FileKindPackageSource FileKind = "PackageSource"
	FileKindTest          FileKind = "Test"
	FileKindXTest         FileKind = "XTest"
	FileKindFiletest      FileKind = "Filetest"
	FileKindOther         FileKind = "Other"
)

func GnoFileKinds() []FileKind {
	return []FileKind{FileKindPackageSource, FileKindTest, FileKindXTest, FileKindFiletest}
}

// GetFileKind analyzes a file's name and body to get it's [FileKind], fset is optional.
// For an in-memory MemFile that carries an explicit Kind, prefer GetMemFileKind —
// it picks up new-style filetests whose Name is a bare basename.
func GetFileKind(filename string, body string, fset *token.FileSet) FileKind {
	if !strings.HasSuffix(filename, ".gno") {
		return FileKindOther
	}

	if std.IsFiletestName(filename) {
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

// GetMemFileKind returns the FileKind of a MemFile. It prefers the explicit
// std.MemFile.Kind set at read time (which carries filetest classification
// for new-style filetests whose Name is a bare basename) and falls back to
// name/body inspection via GetFileKind for legacy MemFiles with Kind unset.
func GetMemFileKind(mfile *std.MemFile, fset *token.FileSet) FileKind {
	switch mfile.Kind {
	case std.KindFiletest:
		return FileKindFiletest
	case std.KindXTest:
		return FileKindXTest
	case std.KindPackageSource:
		return FileKindPackageSource
	case std.KindOther:
		return FileKindOther
	case std.KindTest:
		// Could be Test or XTest; distinction needs body parse.
		return GetFileKind(mfile.Name, mfile.Body, fset)
	case std.KindUnknown:
		return GetFileKind(mfile.Name, mfile.Body, fset)
	default:
		return FileKindUnknown
	}
}
