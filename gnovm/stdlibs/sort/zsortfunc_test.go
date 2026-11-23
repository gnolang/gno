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

var update = flag.Bool("update", false, "rewrite zsortfunc.gno from sort.gno")

const zsortfuncHeader = `// Code generated from sort.gno by zsortfunc_test.go; DO NOT EDIT.
// Regenerate with: go test ./gnovm/stdlibs/sort -run TestZSortFunc -update

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

`

// TestZSortFunc checks that zsortfunc.gno is what genZSortFunc makes of the
// current sort.gno, so the copy Slice and SliceStable run cannot drift from
// the one Sort and Stable run.
func TestZSortFunc(t *testing.T) {
	src, err := genZSortFunc("sort.gno")
	if err != nil {
		t.Fatal(err)
	}
	if *update {
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
		t.Fatal("zsortfunc.gno is stale; run: go test ./gnovm/stdlibs/sort -run TestZSortFunc -update")
	}
}

// genZSortFunc does what Go's sort/genzfunc.go does to sort.go: it keeps the
// unexported funcs taking a "data Interface", renames each with a "_func"
// suffix, and makes it take a "data lessSwap". Unlike Go's, it also turns
// data.Swap(a, b) into sliceSwap(data.x, a, b), and it renames only calls to
// the funcs it keeps, since sort.gno also calls int() and uint().
func genZSortFunc(path string) ([]byte, error) {
	fset := token.NewFileSet()
	af, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return nil, err
	}
	af.Doc, af.Imports, af.Comments = nil, nil, nil

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
		typ.Name = "lessSwap"
		kept[fd.Name.Name] = true
		decls = append(decls, fd)
	}
	af.Decls = decls

	var walkErr error
	ast.Inspect(af, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.FuncDecl:
			n.Name.Name += "_func"
		case *ast.CallExpr:
			switch fun := n.Fun.(type) {
			case *ast.Ident:
				if kept[fun.Name] {
					fun.Name += "_func"
				}
			case *ast.SelectorExpr:
				if x, ok := fun.X.(*ast.Ident); ok && x.Name == "data" {
					switch fun.Sel.Name {
					case "Less":
					case "Swap":
						n.Fun = ast.NewIdent("sliceSwap")
						n.Args = append([]ast.Expr{&ast.SelectorExpr{X: ast.NewIdent("data"), Sel: ast.NewIdent("x")}}, n.Args...)
					default:
						walkErr = fmt.Errorf("data.%s has no lessSwap counterpart", fun.Sel.Name)
					}
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
	// Drop the blank lines the removed comments leave, then label each func.
	src := regexp.MustCompile(`\n{2,}`).ReplaceAll(out.Bytes(), []byte("\n"))
	src = regexp.MustCompile(`\nfunc (\w+)_func`).ReplaceAll(src, []byte("\n\n// Auto-generated variant of sort.gno:$1\nfunc ${1}_func"))
	src, err = format.Source(append([]byte(zsortfuncHeader), src...))
	if err != nil {
		return nil, err
	}
	return src, nil
}
