package packages

import (
	"fmt"
	"go/parser"
	"go/token"
	"strings"
)

type FileKind uint

const (
	FileKindUnknown = FileKind(iota)
	FileKindCompiled
	FileKindTest
	FileKindXtest
	FileKindFiletest
)

func GetFileKind(filename string, body string) (FileKind, error) {
	if !strings.HasSuffix(filename, ".gno") {
		return FileKindUnknown, fmt.Errorf("%q is not a gno file", filename)
	}

	if strings.HasSuffix(filename, "_filetest.gno") {
		return FileKindFiletest, nil
	}

	if !strings.HasSuffix(filename, "_test.gno") {
		return FileKindCompiled, nil
	}

	fset := token.NewFileSet()
	ast, err := parser.ParseFile(fset, filename, body, parser.PackageClauseOnly)
	if err != nil {
		return FileKindUnknown, err
	}
	packageName := ast.Name.Name

	if strings.HasSuffix(packageName, "_test") {
		return FileKindXtest, nil
	}
	return FileKindTest, nil
}
