package main

import (
	"errors"
	"fmt"
	"go/ast"
	"go/doc"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	slashpath "path"
	"path/filepath"
	"slices"
	"strings"
)

func main() {
	gr, err := gitRoot()
	if err != nil {
		panic(err)
	}

	ffs, err := foreignFunctions(gr)
	if err != nil {
		panic(err)
	}
	fmt.Println(ffs)
}

func gitRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		joined := filepath.Join(dir, ".git")
		if _, err := os.Stat(joined); err == nil {
			return dir, nil
		}
		newdir := filepath.Dir(dir)
		if dir == newdir {
			return "", errors.New("not a git repository")
		}
		dir = newdir
	}
}

var ffTemplate = strings.ReplaceAll(`### func %s.%s

'''go
package %s // import %q

%s
'''

%s
`, "'", "`")

func foreignFunctions(gr string) (string, error) {
	joined := filepath.Join(gr, "gnovm", "stdlibs")
	if _, err := os.Stat(joined); err != nil {
		return "", fmt.Errorf("cannot open stdlibs directory: %w", err)
	}
	var result strings.Builder
	fset := token.NewFileSet()
	err := filepath.WalkDir(joined, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".gno") {
			return nil
		}
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return err
		}
		hasBodyless := slices.ContainsFunc(file.Decls, func(decl ast.Decl) bool {
			if decl, ok := decl.(*ast.FuncDecl); ok {
				return decl.Body == nil
			}
			return false
		})
		if !hasBodyless {
			return nil
		}
		importPath, err := filepath.Rel(joined, filepath.Dir(path))
		if err != nil {
			return err
		}
		importPath = filepath.ToSlash(importPath)
		ap, _ := ast.NewPackage(fset, map[string]*ast.File{path: file}, nil, nil)
		if err != nil {
			return err
		}
		pkg := doc.New(ap, importPath, doc.AllDecls|doc.PreserveAST)
		for _, fn := range pkg.Funcs {
			if fn.Decl.Body == nil {
				md := pkg.Markdown(fn.Doc)
				signature := mustFormatNode(fset, fn.Decl)
				fmt.Fprintf(&result, ffTemplate,
					importPath, fn.Name,
					slashpath.Base(importPath), importPath,
					signature, md,
				)
			}
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("error walking stdlibs directory: %w", err)
	}
	return result.String(), nil
}

// mustFormatNode calls format.Node and returns the result as a string.
// Panic on error, which shouldn't happen since the node is a valid AST from go/parser.
// If source is true and the optional ast.File is given, then use it to get internal comments.
func mustFormatNode(fset *token.FileSet, node *ast.FuncDecl) string {
	// Omit the Doc and Body so that it's not in the signature
	node = &ast.FuncDecl{
		Recv: node.Recv,
		Name: node.Name,
		Type: node.Type,
	}

	buf := new(strings.Builder)
	if err := format.Node(buf, fset, node); err != nil {
		panic("Error in format.Node: " + err.Error())
	}
	return buf.String()
}
