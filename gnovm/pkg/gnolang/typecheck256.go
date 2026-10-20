package gnolang

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"path"
	"sort"
	"strconv"
)

// Tools that turn Gno into Go (gno tool transpile) need to know which code
// uses int256 or uint256, which Go lacks. These helpers answer from go/types
// information (TypeCheckOptions.Info), not from identifier spelling: a type is
// 256-bit if it is, or is built from, the int256 or uint256 shim of
// gnobuiltins/gno0p9.

// is256GoType reports whether t is int256 or uint256, or a pointer, slice,
// array, map, channel, function or unnamed struct type built from one. A
// named type is not looked through: its declaration is checked where it is
// written.
func is256GoType(t types.Type) bool {
	switch t := types.Unalias(t).(type) {
	case *types.Named:
		obj := t.Obj()
		return obj.Pkg() != nil && obj.Pkg().Path() == "gnobuiltins/gno0p9" &&
			(obj.Name() == "int256" || obj.Name() == "uint256")
	case *types.Pointer:
		return is256GoType(t.Elem())
	case *types.Slice:
		return is256GoType(t.Elem())
	case *types.Array:
		return is256GoType(t.Elem())
	case *types.Chan:
		return is256GoType(t.Elem())
	case *types.Map:
		return is256GoType(t.Key()) || is256GoType(t.Elem())
	case *types.Struct:
		for i := range t.NumFields() {
			if is256GoType(t.Field(i).Type()) {
				return true
			}
		}
	case *types.Tuple:
		for i := range t.Len() {
			if is256GoType(t.At(i).Type()) {
				return true
			}
		}
	case *types.Signature:
		return is256GoType(t.Params()) || is256GoType(t.Results())
	}
	return false
}

// Files256 returns the base names of the files, among those info describes,
// in which some expression or declared object has a 256-bit type (leaving
// out the .gnobuiltins.gno file the type checker injects).
func Files256(info *types.Info, fset *token.FileSet) []string {
	files := map[string]bool{}
	note := func(pos token.Pos, t types.Type) {
		if !pos.IsValid() || t == nil || !is256GoType(t) {
			return
		}
		// The injected builtins file declares the shims themselves.
		if name := path.Base(fset.Position(pos).Filename); name != ".gnobuiltins.gno" {
			files[name] = true
		}
	}
	for expr, tv := range info.Types {
		note(expr.Pos(), tv.Type)
	}
	for id, obj := range info.Defs {
		if obj != nil {
			note(id.Pos(), obj.Type())
		}
	}
	return sortedKeys(files)
}

// RefsInto256Files returns, for a package whose files in leftOut are left out
// of its Go form, the uses in its other files of objects declared in a
// left-out file, as "file: name" entries. An empty result means leaving those
// files out cannot change what the other files refer to.
func RefsInto256Files(info *types.Info, fset *token.FileSet, leftOut map[string]bool) []string {
	refs := map[string]bool{}
	for id, obj := range info.Uses {
		if obj == nil || !obj.Pos().IsValid() {
			continue
		}
		declFile := path.Base(fset.Position(obj.Pos()).Filename)
		useFile := path.Base(fset.Position(id.Pos()).Filename)
		if leftOut[declFile] && !leftOut[useFile] && useFile != ".gnobuiltins.gno" && isPackageLevel(obj) {
			refs[useFile+": "+obj.Name()] = true
		}
	}
	return sortedKeys(refs)
}

// isPackageLevel reports whether obj is declared at package scope, or is a
// method or field (which a file can only reach through a package-level
// type).
func isPackageLevel(obj types.Object) bool {
	if obj.Pkg() == nil {
		return false
	}
	if obj.Parent() == obj.Pkg().Scope() {
		return true
	}
	if v, ok := obj.(*types.Var); ok && v.IsField() {
		return true
	}
	if f, ok := obj.(*types.Func); ok {
		return f.Type().(*types.Signature).Recv() != nil
	}
	return false
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// InitEffects256 returns what, in the file named name (with source src) of a
// package type-checked into info, would run when the package is initialized
// and could affect the rest of the package: init functions, and calls (other
// than conversions) in package-level variable initializers. A file left out
// of a package's Go form must have none, or the rest of the package could
// behave differently without it.
func InitEffects256(info *types.Info, fset *token.FileSet, name, src string) ([]string, error) {
	f, err := parser.ParseFile(token.NewFileSet(), name, src, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}
	var effects []string
	type span struct{ from, to int }
	var varSpans []span
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *ast.FuncDecl:
			if d.Recv == nil && d.Name.Name == "init" {
				effects = append(effects, "func init")
			}
		case *ast.GenDecl:
			if d.Tok == token.VAR {
				varSpans = append(varSpans, span{int(d.Pos()) - int(f.FileStart), int(d.End()) - int(f.FileStart)})
			}
		}
	}
	for expr := range info.Types {
		call, ok := expr.(*ast.CallExpr)
		if !ok {
			continue
		}
		pos := fset.Position(call.Pos())
		if path.Base(pos.Filename) != name {
			continue
		}
		if fun, ok := info.Types[call.Fun]; ok && fun.IsType() {
			continue // a conversion
		}
		for _, sp := range varSpans {
			if pos.Offset >= sp.from && pos.Offset < sp.to {
				effects = append(effects, "call in a variable initializer at line "+strconv.Itoa(pos.Line))
				break
			}
		}
	}
	sort.Strings(effects)
	return effects, nil
}
