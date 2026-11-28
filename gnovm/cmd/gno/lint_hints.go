package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// Lint hints are advisory: they are printed like any other issue, but they
// never make `gno lint` fail. They point realm authors at code whose cost or
// behaviour is easy to misjudge. Every hint is syntactic and only reports what
// it can see for certain; when in doubt it stays silent.
// See gnovm/adr/pr6536_lint_hints.md.

const lintHintConfidence = 0.8

// lintHints runs the advisory checks on the production files of a realm and
// prints one issue per finding. It never fails.
func lintHints(io commands.IO, dir string, mpkg *std.MemPackage) {
	for _, issue := range collectLintHints(dir, mpkg) {
		io.ErrPrintln(issue)
	}
}

func collectLintHints(dir string, mpkg *std.MemPackage) []gnoIssue {
	if mpkg == nil || !gno.IsRealmPath(mpkg.Path) {
		return nil
	}
	fset := token.NewFileSet()
	var files []*ast.File
	for _, mfile := range mpkg.Files {
		name := mfile.Name
		if !strings.HasSuffix(name, ".gno") ||
			strings.HasSuffix(name, "_test.gno") ||
			strings.HasSuffix(name, "_filetest.gno") {
			continue
		}
		f, err := parser.ParseFile(fset, filepath.Join(dir, name), mfile.Body, parser.SkipObjectResolution)
		if err != nil {
			// Parse errors are reported by the type checker.
			return nil
		}
		files = append(files, f)
	}

	globals := packageVars(files)
	var issues []gnoIssue
	report := func(code gnoCode, pos token.Pos, msg string) {
		issues = append(issues, gnoIssue{
			Code:       code,
			Msg:        msg,
			Confidence: lintHintConfidence,
			Location:   tryRelativizePath(fset.Position(pos).String()),
		})
	}
	for _, f := range files {
		imports := fileImports(f)
		for _, decl := range f.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Body == nil {
				continue
			}
			checkRenderLoopConcat(fd, report)
			checkUnboundedBulkCopy(fd, imports, globals, report)
			checkUnrecoverableNative(fd, imports, report)
		}
	}
	return issues
}

type hintReporter func(code gnoCode, pos token.Pos, msg string)

// ----------------------------------------
// gnoRenderLoopConcatHint: Render building its output with + or += in a loop.

func checkRenderLoopConcat(fd *ast.FuncDecl, report hintReporter) {
	if fd.Recv != nil || fd.Name.Name != "Render" || !returnsSingleString(fd.Type) {
		return
	}
	strVars := stringVars(fd)
	ast.Inspect(fd.Body, func(n ast.Node) bool {
		var body *ast.BlockStmt
		switch n := n.(type) {
		case *ast.FuncLit:
			return false
		case *ast.ForStmt:
			body = n.Body
		case *ast.RangeStmt:
			body = n.Body
		default:
			return true
		}
		if pos, ok := findStringConcat(body, strVars); ok {
			report(gnoRenderLoopConcatHint, pos,
				"Render concatenates strings inside a loop; use strings.Builder, "+
					"and paginate if the loop ranges over a collection that can grow")
		}
		// One hint per outermost loop.
		return false
	})
}

func returnsSingleString(ft *ast.FuncType) bool {
	if ft.Results == nil || len(ft.Results.List) != 1 || len(ft.Results.List[0].Names) > 1 {
		return false
	}
	return isIdent(ft.Results.List[0].Type, "string")
}

// stringVars returns the local names in fd that are certainly strings: a named
// string result, `var x string`, `var x = "..."` and `x := "..."`.
func stringVars(fd *ast.FuncDecl) map[string]bool {
	vars := map[string]bool{}
	for _, field := range fd.Type.Results.List {
		for _, name := range field.Names {
			vars[name.Name] = true
		}
	}
	ast.Inspect(fd.Body, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.FuncLit:
			return false
		case *ast.ValueSpec:
			for i, name := range n.Names {
				if isIdent(n.Type, "string") || (n.Type == nil && i < len(n.Values) && isStringExpr(n.Values[i])) {
					vars[name.Name] = true
				}
			}
		case *ast.AssignStmt:
			if n.Tok == token.DEFINE && len(n.Lhs) == len(n.Rhs) {
				for i, lhs := range n.Lhs {
					if id, ok := lhs.(*ast.Ident); ok && isStringExpr(n.Rhs[i]) {
						vars[id.Name] = true
					}
				}
			}
		}
		return true
	})
	return vars
}

// isStringExpr reports whether e is a string literal, or a + chain that
// contains one.
func isStringExpr(e ast.Expr) bool {
	switch e := e.(type) {
	case *ast.BasicLit:
		return e.Kind == token.STRING
	case *ast.ParenExpr:
		return isStringExpr(e.X)
	case *ast.BinaryExpr:
		return e.Op == token.ADD && (isStringExpr(e.X) || isStringExpr(e.Y))
	}
	return false
}

// findStringConcat finds `s += ...` or `s = s + ...` on a known string, or
// one whose right side is a string literal chain.
func findStringConcat(body *ast.BlockStmt, strVars map[string]bool) (token.Pos, bool) {
	var found token.Pos
	ast.Inspect(body, func(n ast.Node) bool {
		if found.IsValid() {
			return false
		}
		if _, ok := n.(*ast.FuncLit); ok {
			return false
		}
		as, ok := n.(*ast.AssignStmt)
		if !ok || len(as.Lhs) != 1 || len(as.Rhs) != 1 {
			return true
		}
		id, ok := as.Lhs[0].(*ast.Ident)
		if !ok {
			return true
		}
		switch as.Tok {
		case token.ADD_ASSIGN:
			if strVars[id.Name] || isStringExpr(as.Rhs[0]) {
				found = as.Pos()
			}
		case token.ASSIGN:
			if be, ok := as.Rhs[0].(*ast.BinaryExpr); ok && be.Op == token.ADD &&
				isIdent(be.X, id.Name) && (strVars[id.Name] || isStringExpr(be.Y)) {
				found = as.Pos()
			}
		}
		return true
	})
	return found, found.IsValid()
}

// ----------------------------------------
// gnoUnboundedBulkCopyHint: bulk copies sized by a caller argument, into
// package state or through strings.Repeat / bytes.Repeat.

func checkUnboundedBulkCopy(fd *ast.FuncDecl, imports map[string]string, globals map[string]bool, report hintReporter) {
	if fd.Recv != nil || !fd.Name.IsExported() {
		return
	}
	intParams, sliceParams := paramKinds(fd.Type)
	if len(intParams) == 0 && len(sliceParams) == 0 {
		return
	}
	bounded := conditionIdents(fd.Body)
	repeatPkgs := map[string]bool{}
	for _, p := range []string{"strings", "bytes"} {
		if name := imports[p]; name != "" {
			repeatPkgs[name] = true
		}
	}

	ast.Inspect(fd.Body, func(n ast.Node) bool {
		if _, ok := n.(*ast.FuncLit); ok {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		switch fun := call.Fun.(type) {
		case *ast.SelectorExpr:
			// strings.Repeat(s, n), bytes.Repeat(b, n)
			pkg, ok := fun.X.(*ast.Ident)
			if !ok || !repeatPkgs[pkg.Name] || fun.Sel.Name != "Repeat" || len(call.Args) != 2 {
				return true
			}
			if name := unconvertedIdent(call.Args[1]); intParams[name] && !bounded[name] {
				report(gnoUnboundedBulkCopyHint, call.Pos(), fmt.Sprintf(
					"%s.Repeat count comes from parameter %q with no bound checked in %s; "+
						"limit it so the size of the result stays predictable",
					pkg.Name, name, fd.Name.Name))
			}
		case *ast.Ident:
			// append(global, p...). copy is left out on purpose: it never
			// copies more than its destination already holds.
			if fun.Name != "append" || !call.Ellipsis.IsValid() || len(call.Args) != 2 {
				return true
			}
			d, ok := call.Args[0].(*ast.Ident)
			if !ok || !globals[d.Name] || isLocal(fd, d.Name) {
				return true
			}
			s, ok := call.Args[1].(*ast.Ident)
			if !ok || !sliceParams[s.Name] || bounded[s.Name] {
				return true
			}
			report(gnoUnboundedBulkCopyHint, call.Pos(), fmt.Sprintf(
				"append copies all of parameter %q into package state %q with no length checked in %s; "+
					"limit its length so the state and the cost of reading it stay bounded",
				s.Name, d.Name, fd.Name.Name))
		}
		return true
	})
}

var integerTypes = map[string]bool{
	"int": true, "int8": true, "int16": true, "int32": true, "int64": true,
	"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
}

// paramKinds splits the parameters of ft into integers and slices (strings,
// slices and variadics).
func paramKinds(ft *ast.FuncType) (ints, slices map[string]bool) {
	ints, slices = map[string]bool{}, map[string]bool{}
	for _, field := range ft.Params.List {
		var isInt, isSlice bool
		switch t := field.Type.(type) {
		case *ast.Ident:
			isInt = integerTypes[t.Name]
			isSlice = t.Name == "string"
		case *ast.ArrayType:
			isSlice = t.Len == nil
		case *ast.Ellipsis:
			isSlice = true
		}
		for _, name := range field.Names {
			if isInt {
				ints[name.Name] = true
			}
			if isSlice {
				slices[name.Name] = true
			}
		}
	}
	return
}

// conditionIdents collects every identifier used in an if, for or switch
// condition. Any of those counts as a bound.
func conditionIdents(body *ast.BlockStmt) map[string]bool {
	ids := map[string]bool{}
	collect := func(e ast.Node) {
		if e == nil {
			return
		}
		ast.Inspect(e, func(n ast.Node) bool {
			if id, ok := n.(*ast.Ident); ok {
				ids[id.Name] = true
			}
			return true
		})
	}
	ast.Inspect(body, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.IfStmt:
			collect(n.Cond)
		case *ast.ForStmt:
			if n.Cond != nil {
				collect(n.Cond)
			}
		case *ast.SwitchStmt:
			if n.Tag != nil {
				collect(n.Tag)
			}
			for _, stmt := range n.Body.List {
				if cc, ok := stmt.(*ast.CaseClause); ok {
					for _, e := range cc.List {
						collect(e)
					}
				}
			}
		}
		return true
	})
	return ids
}

// unconvertedIdent returns the name of e, looking through one integer
// conversion such as int(n).
func unconvertedIdent(e ast.Expr) string {
	if call, ok := e.(*ast.CallExpr); ok && len(call.Args) == 1 {
		if fn, ok := call.Fun.(*ast.Ident); ok && integerTypes[fn.Name] {
			e = call.Args[0]
		}
	}
	if id, ok := e.(*ast.Ident); ok {
		return id.Name
	}
	return ""
}

// packageVars returns the names of package-level variables.
func packageVars(files []*ast.File) map[string]bool {
	vars := map[string]bool{}
	for _, f := range files {
		for _, decl := range f.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.VAR {
				continue
			}
			for _, spec := range gd.Specs {
				for _, name := range spec.(*ast.ValueSpec).Names {
					vars[name.Name] = true
				}
			}
		}
	}
	return vars
}

// isLocal reports whether name is a parameter, result or local declared in
// fd, which would shadow a package variable of the same name.
func isLocal(fd *ast.FuncDecl, name string) bool {
	for _, fl := range []*ast.FieldList{fd.Type.Params, fd.Type.Results} {
		if fl == nil {
			continue
		}
		for _, field := range fl.List {
			for _, n := range field.Names {
				if n.Name == name {
					return true
				}
			}
		}
	}
	local := false
	ast.Inspect(fd.Body, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.ValueSpec:
			for _, id := range n.Names {
				local = local || id.Name == name
			}
		case *ast.AssignStmt:
			if n.Tok == token.DEFINE {
				for _, lhs := range n.Lhs {
					local = local || isIdent(lhs, name)
				}
			}
		case *ast.RangeStmt:
			if n.Tok == token.DEFINE {
				local = local || isIdent(n.Key, name) || isIdent(n.Value, name)
			}
		}
		return !local
	})
	return local
}

// ----------------------------------------
// gnoUnrecoverableNativeHint: banker operations in a function that relies on
// a deferred recover().
//
// On chain, the banker behind chain/banker reports a failed send, issue or
// remove (for example an insufficient balance) as a Go panic raised by the
// native binding, not as a Gno panic. Machine.Run only turns *Exception
// values into Gno panics (see Machine.runOnce), so the failure ends the
// transaction and a deferred recover() in the caller never runs on it.

var bankerFailingMethods = map[string]bool{
	"SendCoins":  true,
	"IssueCoin":  true,
	"RemoveCoin": true,
}

func checkUnrecoverableNative(fd *ast.FuncDecl, imports map[string]string, report hintReporter) {
	if imports["chain/banker"] == "" || !defersRecover(fd.Body) {
		return
	}
	ast.Inspect(fd.Body, func(n ast.Node) bool {
		if _, ok := n.(*ast.FuncLit); ok {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || !bankerFailingMethods[sel.Sel.Name] {
			return true
		}
		report(gnoUnrecoverableNativeHint, call.Pos(), fmt.Sprintf(
			"a failed banker %s (for example an insufficient balance) ends the transaction "+
				"and is not caught by the recover() deferred in %s; check the precondition before calling it",
			sel.Sel.Name, fd.Name.Name))
		return true
	})
}

// defersRecover reports whether body directly defers a function literal
// that calls recover().
func defersRecover(body *ast.BlockStmt) bool {
	for _, stmt := range body.List {
		ds, ok := stmt.(*ast.DeferStmt)
		if !ok {
			continue
		}
		fl, ok := ds.Call.Fun.(*ast.FuncLit)
		if !ok {
			continue
		}
		found := false
		ast.Inspect(fl.Body, func(n ast.Node) bool {
			if call, ok := n.(*ast.CallExpr); ok && isIdent(call.Fun, "recover") && len(call.Args) == 0 {
				found = true
			}
			return !found
		})
		if found {
			return true
		}
	}
	return false
}

// ----------------------------------------
// helpers

// fileImports maps an import path to the name it is bound to in f.
func fileImports(f *ast.File) map[string]string {
	m := map[string]string{}
	for _, imp := range f.Imports {
		p, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			continue
		}
		name := p[strings.LastIndex(p, "/")+1:]
		if imp.Name != nil {
			name = imp.Name.Name
		}
		if name == "_" || name == "." {
			continue
		}
		m[p] = name
	}
	return m
}

func isIdent(e ast.Expr, name string) bool {
	id, ok := e.(*ast.Ident)
	return ok && id.Name == name
}
