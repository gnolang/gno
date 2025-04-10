// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fix

import (
	"fmt"
	"go/ast"
	"go/token"
	"slices"
	"strconv"

	"golang.org/x/tools/go/ast/astutil"
)

// Fix is an individual fix provided by this package.
type Fix struct {
	Name     string
	Date     string // date that fix was introduced, in YYYY-MM-DD format
	F        func(*ast.File) bool
	Desc     string
	Disabled bool // whether this fix should be disabled by default
}

var Fixes = []Fix{
	{
		Name: "stdsplit",
		Date: "2025-04-10",
		Desc: "rewrites imports and symbols of the std package into the new packages and symbols",
		F:    stdsplit,
	},
}

// imports reports whether f imports path.
func imports(f *ast.File, path string) bool {
	return importSpec(f, path) != nil
}

// importSpec returns the import spec if f imports path,
// or nil otherwise.
func importSpec(f *ast.File, path string) *ast.ImportSpec {
	for _, s := range f.Imports {
		if importPath(s) == path {
			return s
		}
	}
	return nil
}

var importPathCache = map[*ast.ImportSpec]string{}

// importPath returns the unquoted import path of s,
// or "" if the path is not properly quoted.
func importPath(s *ast.ImportSpec) string {
	if cached, ok := importPathCache[s]; ok {
		return cached
	}
	t, err := strconv.Unquote(s.Path.Value)
	if err == nil {
		importPathCache[s] = t
		return t
	}
	return ""
}

// matchLen returns the length of the longest prefix shared by x and y.
func matchLen(x, y string) int {
	i := 0
	for i < len(x) && i < len(y) && x[i] == y[i] {
		i++
	}
	return i
}

// addImport adds the import path to the file f, if absent.
func addImport(f *ast.File, ipath, name string) (added bool) {
	if imports(f, ipath) {
		return false
	}

	newImport := &ast.ImportSpec{
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: strconv.Quote(ipath),
		},
	}
	if name != "" {
		newImport.Name = &ast.Ident{
			Name: name,
		}
	}

	// Find an import decl to add to.
	var (
		bestMatch  = -1
		lastImport = -1
		impDecl    *ast.GenDecl
		impIndex   = -1
	)
	for i, decl := range f.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if ok && gen.Tok == token.IMPORT {
			lastImport = i

			// Compute longest shared prefix with imports in this block.
			for j, spec := range gen.Specs {
				impspec := spec.(*ast.ImportSpec)
				n := matchLen(importPath(impspec), ipath)
				if n > bestMatch {
					bestMatch = n
					impDecl = gen
					impIndex = j
				}
			}
		}
	}

	// If no import decl found, add one after the last import.
	if impDecl == nil {
		impDecl = &ast.GenDecl{
			Tok: token.IMPORT,
		}
		f.Decls = append(f.Decls, nil)
		copy(f.Decls[lastImport+2:], f.Decls[lastImport+1:])
		f.Decls[lastImport+1] = impDecl
	}

	// Ensure the import decl has parentheses, if needed.
	if len(impDecl.Specs) > 0 && !impDecl.Lparen.IsValid() {
		impDecl.Lparen = impDecl.Pos()
	}

	insertAt := impIndex + 1
	if insertAt == 0 {
		insertAt = len(impDecl.Specs)
	}
	impDecl.Specs = slices.Insert(impDecl.Specs, insertAt, ast.Spec(newImport))
	if insertAt > 0 {
		// Assign same position as the previous import,
		// so that the sorter sees it as being in the same block.
		prev := impDecl.Specs[insertAt-1]
		newImport.Path.ValuePos = prev.Pos()
		newImport.EndPos = prev.Pos()
	}

	f.Imports = append(f.Imports, newImport)
	return true
}

// deleteImport deletes the import path from the file f, if present.
func deleteImport(f *ast.File, path string) (deleted bool) {
	oldImport := importSpec(f, path)

	// Find the import node that imports path, if any.
	for i, decl := range f.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.IMPORT {
			continue
		}
		for j, spec := range gen.Specs {
			impspec := spec.(*ast.ImportSpec)
			if oldImport != impspec {
				continue
			}

			// We found an import spec that imports path.
			// Delete it.
			deleted = true
			gen.Specs = slices.Delete(gen.Specs, j, j+1)

			// If this was the last import spec in this decl,
			// delete the decl, too.
			if len(gen.Specs) == 0 {
				f.Decls = slices.Delete(f.Decls, i, i+1)
			} else if len(gen.Specs) == 1 {
				gen.Lparen = token.NoPos // drop parens
			}
			if j > 0 {
				// We deleted an entry but now there will be
				// a blank line-sized hole where the import was.
				// Close the hole by making the previous
				// import appear to "end" where this one did.
				gen.Specs[j-1].(*ast.ImportSpec).EndPos = impspec.End()
			}
			break
		}
	}

	// Delete it from f.Imports.
	for i, imp := range f.Imports {
		if imp == oldImport {
			copy(f.Imports[i:], f.Imports[i+1:])
			f.Imports = f.Imports[:len(f.Imports)-1]
			break
		}
	}

	return
}

// apply is a modified version of astutil.Apply that additionally keeps track of
// the scopes of the file, to resolve any name to its declaration.
//
// Callers are responsible for tracking identifiers of any ImportSpec.
func apply(f ast.Node, pre, post func(*astutil.Cursor, scopes) bool) ast.Node {
	sc := scopes{scope{}}
	return astutil.Apply(
		f,
		func(c *astutil.Cursor) bool {
			if pre != nil && !pre(c, sc) {
				return false
			}

			n := c.Node()

			// This contains the logic for handling scopes.
			switch n := n.(type) {
			case *ast.TypeSpec:
				sc.declare(n.Name.Name, n)
			case *ast.ValueSpec:
				for _, name := range n.Names {
					sc.declare(name.Name, n)
				}
			case *ast.AssignStmt:
				if n.Tok == token.DEFINE {
					for _, name := range n.Lhs {
						// only declare if it doesn't exist in the last scope,
						// := allows the LHS to contain already defined values
						// which are then simply assigned instead of declared.
						name := name.(*ast.Ident).Name
						if _, ok := sc[len(sc)-1][name]; !ok {
							sc.declare(name, n)
						}
					}
				}
			case *ast.FuncDecl:
				name := n.Name.Name
				if n.Recv != nil && len(n.Recv.List) > 0 {
					tp := recvType(n.Recv.List[0].Type)
					if tp != nil {
						name = tp.Name + "." + name
					}
				}
				if name != "init" {
					sc.declare(name, n)
				}
				sc.push()
				sc.declareList(n.Recv)
				sc.declareList(n.Type.Params)
				sc.declareList(n.Type.Results)
			case *ast.FuncLit:
				sc.push()
				sc.declareList(n.Type.Params)
				sc.declareList(n.Type.Results)
			case *ast.RangeStmt:
				sc.push()
				if n.Tok == token.DEFINE {
					if id, ok := n.Key.(*ast.Ident); ok {
						sc.declare(id.Name, n)
					}
					if id, ok := n.Value.(*ast.Ident); ok {
						sc.declare(id.Name, n)
					}
				}
			case *ast.BlockStmt,
				*ast.IfStmt,
				*ast.SwitchStmt,
				*ast.TypeSwitchStmt,
				*ast.CaseClause,
				*ast.CommClause,
				*ast.ForStmt,
				*ast.SelectStmt:
				sc.push()
			}
			return true
		},
		func(c *astutil.Cursor) bool {
			if post != nil && !post(c, sc) {
				return false
			}

			n := c.Node()
			switch n.(type) {
			case *ast.FuncDecl,
				*ast.BlockStmt,
				*ast.FuncLit,
				*ast.IfStmt,
				*ast.SwitchStmt,
				*ast.TypeSwitchStmt,
				*ast.CaseClause,
				*ast.CommClause,
				*ast.ForStmt,
				*ast.SelectStmt,
				*ast.RangeStmt:
				sc.pop()
			}
			return true
		},
	)
}

type scope map[string]ast.Node

type scopes []scope

func (s scopes) lookup(name string) ast.Node {
	for _, scope := range slices.Backward(s) {
		if stmt, ok := scope[name]; ok {
			return stmt
		}
	}
	return nil
}

func (s *scopes) push() {
	*s = append(*s, scope{})
}

func (s *scopes) pop() {
	*s = (*s)[:len(*s)-1]
}

func (s scopes) declare(name string, stmt ast.Node) {
	if name == "_" {
		return
	}
	sc := s[len(s)-1]
	if _, ok := sc[name]; ok {
		panic(fmt.Sprintf("duplicate declaration of ident %q", name))
	}
	sc[name] = stmt
}

func (s scopes) declareList(fl *ast.FieldList) {
	if fl == nil {
		return
	}
	for _, field := range fl.List {
		for _, name := range field.Names {
			s.declare(name.Name, field)
		}
	}
}
