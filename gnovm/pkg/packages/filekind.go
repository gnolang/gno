package packages

import (
	"go/parser"
	"go/token"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/std"
)

// FileKind represents the category a gno package file falls in. It is an
// alias for [std.MemFileKind] — the same enum that ships on every
// [std.MemFile] — so a MemFile's Kind and a Package's per-kind file/import
// buckets share one taxonomy.
//
// JSON map keys (used by `gno list -json`) round-trip as the names defined
// below; see [std.MemFileKind.MarshalText].
type FileKind = std.MemFileKind

// Aliases for the std package's MemFileKind constants. New code should use
// the std.Kind* spellings directly; these aliases exist for callers that
// historically referred to FileKind*.
const (
	FileKindUnknown       = std.KindUnknown
	FileKindPackageSource = std.KindPackageSource
	FileKindTest          = std.KindTest
	FileKindXTest         = std.KindXTest
	FileKindFiletest      = std.KindFiletest
	FileKindOther         = std.KindOther
)

// GnoFileKinds returns the file kinds that are part of a Gno package's
// source set (excludes [FileKindOther] and [FileKindUnknown]).
func GnoFileKinds() []FileKind {
	return []FileKind{FileKindPackageSource, FileKindTest, FileKindXTest, FileKindFiletest}
}

// GetFileKind analyzes a file's name and body to derive its [FileKind].
// fset is optional. For an in-memory MemFile that carries an explicit Kind,
// prefer [GetMemFileKind] — it picks up new-style filetests whose Name is a
// bare basename (Kind=Filetest, no `_filetest.gno` suffix).
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
	if strings.HasSuffix(ast.Name.Name, "_test") {
		return FileKindXTest
	}
	return FileKindTest
}

// GetMemFileKind returns the FileKind of a MemFile. It prefers the explicit
// Kind field (which carries new-style filetests via [std.KindFiletest]) and
// falls back to name/body inspection via [GetFileKind] only when Kind is
// [FileKindUnknown] (e.g. amino-decoded from legacy storage) or when
// distinguishing Test vs. XTest needs a package-clause parse.
func GetMemFileKind(mfile *std.MemFile, fset *token.FileSet) FileKind {
	switch mfile.Kind {
	case FileKindUnknown, FileKindTest:
		// KindTest can't tell us Test vs XTest without parsing.
		return GetFileKind(mfile.Name, mfile.Body, fset)
	default:
		return mfile.Kind
	}
}
