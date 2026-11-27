package sort

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"testing"
)

var updateGoldenTests = flag.Bool("update-golden-tests", false, "rewrite zsortfunc.gno from sort.gno")

const zsortfuncHeader = `// Code generated from sort.gno by zsortfunc_test.go; DO NOT EDIT.
// Regenerate with: go test ./gnovm/stdlibs/sort -run TestZSortFunc -update-golden-tests

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

`

// TestZSortFunc checks that zsortfunc.gno matches what genZSortFunc makes of
// sort.gno, so the routines Slice and SliceStable run cannot drift from the
// ones Sort and Stable run. With -update-golden-tests it rewrites the file.
func TestZSortFunc(t *testing.T) {
	src, err := genZSortFunc("sort.gno")
	if err != nil {
		t.Fatal(err)
	}
	if *updateGoldenTests {
		if err := os.WriteFile("zsortfunc.gno", src, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	got, err := os.ReadFile("zsortfunc.gno")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, src) {
		t.Fatal("zsortfunc.gno is stale; run: go test ./gnovm/stdlibs/sort -run TestZSortFunc -update-golden-tests")
	}
}

// genZSortFunc makes zsortfunc.gno from sort.gno as Go's sort/genzfunc.go
// makes zfuncversion.go. It keeps the unexported funcs that take a
// "data Interface", adds "_func" to their names, and makes them take a
// "data *lessSwap". Unlike genzfunc.go, it rewrites data.Swap(a, b) to
// sliceSwap(data.x, a, b), and it renames only calls to the funcs it keeps,
// because sort.gno also calls int() and uint().
func genZSortFunc(path string) ([]byte, error) {
	fset := token.NewFileSet()
	af, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return nil, err
	}
	af.Comments = nil

	kept := map[string]bool{}
	var decls []ast.Decl
	for _, d := range af.Decls {
		fd, ok := d.(*ast.FuncDecl)
		if !ok || fd.Recv != nil || fd.Name.IsExported() || len(fd.Type.Params.List) == 0 {
			continue
		}
		arg0 := fd.Type.Params.List[0]
		typ, ok := arg0.Type.(*ast.Ident)
		if !ok || len(arg0.Names) != 1 || arg0.Names[0].Name != "data" || typ.Name != "Interface" {
			continue
		}
		arg0.Type = &ast.StarExpr{X: ast.NewIdent("lessSwap")}
		kept[fd.Name.Name] = true
		fd.Name.Name += "_func"
		decls = append(decls, fd)
	}
	af.Decls = decls

	var walkErr error
	ast.Inspect(af, func(n ast.Node) bool {
		ce, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		switch fun := ce.Fun.(type) {
		case *ast.Ident:
			if kept[fun.Name] {
				fun.Name += "_func"
			}
		case *ast.SelectorExpr:
			if x, ok := fun.X.(*ast.Ident); ok && x.Name == "data" {
				switch fun.Sel.Name {
				case "Less":
				case "Swap":
					ce.Fun = ast.NewIdent("sliceSwap")
					ce.Args = append([]ast.Expr{&ast.SelectorExpr{X: ast.NewIdent("data"), Sel: ast.NewIdent("x")}}, ce.Args...)
				default:
					walkErr = fmt.Errorf("data.%s has no lessSwap counterpart", fun.Sel.Name)
				}
			}
		}
		return true
	})
	if walkErr != nil {
		return nil, walkErr
	}

	var out bytes.Buffer
	if err := format.Node(&out, fset, af); err != nil {
		return nil, err
	}
	// Drop every blank line, then put a blank line and a label before each func.
	src := regexp.MustCompile(`\n{2,}`).ReplaceAll(out.Bytes(), []byte("\n"))
	src = regexp.MustCompile(`\nfunc (\w+)_func`).ReplaceAll(src, []byte("\n\n// Auto-generated variant of sort.gno:$1\nfunc ${1}_func"))
	return format.Source(append([]byte(zsortfuncHeader), src...))
}
