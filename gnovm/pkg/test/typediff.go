package test

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"slices"
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

// The types check of the corpus ledger: for every declared name
// (variables, constants, functions, parameters, results, receivers) the
// static type Gno's preprocess records is compared with go/types' object
// type on the same source. Both sides are spelt in one canonical form
// (typeString / goTypeString), and a mismatch is a `GnoTypeDiffers:` line.

// typeDiff is one declared name the two checkers type differently.
type typeDiff struct {
	line, col int // machine coordinates (the rescue's prepended line included)
	name      string
	gno, goT  string
}

// gnoDecl is a declared name with Gno's static type. col is 0 when the
// declaring NameExpr carries no span (a var/const or function name, a
// parameter): the line then comes from the declaration node.
type gnoDecl struct {
	name      string
	line, col int
	typ       gno.Type
}

// goDecl is a go/types object declared at a column of some line.
type goDecl struct {
	col int
	typ string
}

type nameLine struct {
	name string
	line int
}

// newTypeInfo allocates the go/types maps the check reads.
func newTypeInfo() *types.Info {
	return &types.Info{Defs: map[*ast.Ident]types.Object{}}
}

// gnoDeclaredNames lists the names declared in pn (package level) and in
// every block of fn, each with the static type its block records; names
// without one (a reserved slot the preprocess never typed) are skipped,
// as are imports, type names and type-switch bindings.
func gnoDeclaredNames(pn *gno.PackageNode, fn *gno.FileNode) (out []gnoDecl) {
	defer func() {
		if r := recover(); r != nil { // a half-preprocessed tree: no check
			out = nil
		}
	}()
	collect := func(bn gno.BlockNode) {
		sb := bn.GetStaticBlock()
		for i, ns := range sb.GetNameSources() {
			if i >= len(sb.Types) || sb.Types[i] == nil || ns.NameExpr == nil {
				continue
			}
			name := strings.TrimSuffix(string(ns.NameExpr.Name), ".loopvar")
			if name == "_" || strings.Contains(name, ".") {
				continue
			}
			if ns.Type == gno.NSFuncDecl && (name == "main" || name == "init") {
				continue // go/types sees them rewritten for Gno 0.9 (prepareGoGno0p9)
			}
			line, col := declPos(ns)
			if line <= 0 {
				continue
			}
			out = append(out, gnoDecl{name: name, line: line, col: col, typ: sb.Types[i]})
		}
	}
	collect(pn)
	gno.Transcribe(fn, func(_ []gno.Node, _ gno.TransField, _ int, n gno.Node, stage gno.TransStage) (gno.Node, gno.TransCtrl) {
		if stage == gno.TRANS_ENTER {
			if bn, ok := n.(gno.BlockNode); ok {
				collect(bn)
			}
		}
		return n, gno.TRANS_CONTINUE
	})
	return out
}

// declPos is where a name is declared: the NameExpr's own span for a
// `:=` or range name (the only ones Go2Gno positions on the identifier),
// else the line of its declaration node, column unknown.
func declPos(ns gno.NameSource) (line, col int) {
	p := ns.NameExpr.GetSpan().Pos
	switch ns.Type {
	case gno.NSDefine, gno.NSRangeKey, gno.NSRangeValue:
		return p.Line, p.Column
	}
	switch o := ns.Origin.(type) {
	case *gno.ValueDecl:
		return o.GetSpan().Pos.Line, 0
	case *gno.FuncDecl:
		if ns.Type == gno.NSFuncReceiver {
			return o.Recv.GetSpan().Pos.Line, 0
		}
		if ns.Type == gno.NSFuncDecl {
			return o.GetSpan().Pos.Line, 0
		}
	case *gno.FuncTypeExpr:
		switch {
		case ns.Type == gno.NSFuncParam && ns.Index >= 0 && ns.Index < len(o.Params):
			return o.Params[ns.Index].GetSpan().Pos.Line, 0
		case ns.Type == gno.NSFuncResult && ns.Index >= 0 && ns.Index < len(o.Results):
			return o.Results[ns.Index].GetSpan().Pos.Line, 0
		}
	}
	return 0, 0
}

// goDeclaredNames indexes go/types' declared objects by name and line.
func goDeclaredNames(fset *token.FileSet, info *types.Info, pkgPath string) map[nameLine][]goDecl {
	out := map[nameLine][]goDecl{}
	for id, obj := range info.Defs {
		if obj == nil || id.Name == "_" {
			continue
		}
		switch o := obj.(type) {
		case *types.Const, *types.Func:
		case *types.Var:
			if o.IsField() {
				continue
			}
		default:
			continue
		}
		pos := fset.Position(id.Pos())
		k := nameLine{id.Name, pos.Line}
		out[k] = append(out[k], goDecl{col: pos.Column, typ: goTypeString(obj.Type(), pkgPath)})
	}
	for _, ds := range out {
		slices.SortFunc(ds, func(a, b goDecl) int { return a.col - b.col })
	}
	return out
}

// compareDeclaredTypes pairs each Gno declaration with the go/types
// object of the same name on the same line (same column when the Gno
// side knows it; else the line's objects must agree) and lists the
// names whose canonical types differ, in source order.
func compareDeclaredTypes(decls []gnoDecl, goDecls map[nameLine][]goDecl, pkgPath string) []typeDiff {
	var diffs []typeDiff
	seen := map[nameLine]bool{}
	for _, d := range decls {
		k := nameLine{d.name, d.line}
		goT, ok := goTypeAt(goDecls[k], d.col)
		if !ok || goT == "" {
			continue
		}
		gnoT := typeString(d.typ, pkgPath)
		if gnoT == "" || gnoT == goT {
			continue
		}
		if dk := (nameLine{fmt.Sprintf("%s@%d", d.name, d.col), d.line}); seen[dk] {
			continue // an if/switch init name, copied into each clause block
		} else {
			seen[dk] = true
		}
		diffs = append(diffs, typeDiff{line: d.line, col: d.col, name: d.name, gno: gnoT, goT: goT})
	}
	slices.SortStableFunc(diffs, func(a, b typeDiff) int {
		if a.line != b.line {
			return a.line - b.line
		}
		if a.col != b.col {
			return a.col - b.col
		}
		return strings.Compare(a.name, b.name)
	})
	return diffs
}

func goTypeAt(ds []goDecl, col int) (string, bool) {
	if len(ds) == 0 {
		return "", false
	}
	if col > 0 {
		for _, d := range ds {
			if d.col == col {
				return d.typ, true
			}
		}
		return "", false
	}
	for _, d := range ds[1:] {
		if d.typ != ds[0].typ {
			return "", false // two scopes declare the name on one line
		}
	}
	return ds[0].typ, true
}

// formatTypeDiffBlock renders the `GnoTypeDiffers:` block in body
// coordinates; lines outside the body (the rescue's synthetic main) are
// dropped. typeDiffLines lists the lines it names, for the verdict.
func formatTypeDiffBlock(diffs []typeDiff, prependedLines, bodyLines int) string {
	var sb strings.Builder
	for _, d := range diffs {
		l := d.line - prependedLines
		if l < 1 || l > bodyLines {
			continue
		}
		fmt.Fprintf(&sb, "line %d: %s gno %s, go %s\n", l, d.name, d.gno, d.goT)
	}
	return strings.TrimRight(sb.String(), "\n")
}

func typeDiffLines(diffs []typeDiff, prependedLines, bodyLines int) []int {
	var ls []int
	for _, d := range diffs {
		if l := d.line - prependedLines; l >= 1 && l <= bodyLines && !slices.Contains(ls, l) {
			ls = append(ls, l)
		}
	}
	return ls
}

// typeString spells a Gno type in the canonical form shared with
// goTypeString: rune and byte by their kinds, untyped kinds as
// `untyped int|rune|float|string|bool`, a declared type by its bare name
// (qualified by package name outside pkgPath), function types by
// signature, structs and interfaces by their literal. "" means the type
// has no Go counterpart to compare (a type value, a uverse generic).
func typeString(t gno.Type, pkgPath string) string {
	return typeStr(t, pkgPath, 0)
}

// maxTypeDepth bounds the spelling of a cyclic type (an invalid recursive
// interface in an errorcheck file): past it the type is not compared.
const maxTypeDepth = 32

func typeStr(t gno.Type, pkgPath string, depth int) string {
	if depth > maxTypeDepth {
		return ""
	}
	depth++
	switch t := t.(type) {
	case gno.PrimitiveType:
		switch t {
		case gno.InvalidType:
			return ""
		case gno.UntypedBoolType:
			return "untyped bool"
		case gno.UntypedStringType:
			return "untyped string"
		case gno.UntypedRuneType:
			return "untyped rune"
		case gno.UntypedBigintType:
			return "untyped int"
		case gno.UntypedBigdecType:
			return "untyped float"
		case gno.DataByteType:
			return "uint8"
		}
		return t.String()
	case *gno.PointerType:
		return prefix("*", typeStr(t.Elt, pkgPath, depth))
	case *gno.SliceType:
		return prefix("[]", typeStr(t.Elt, pkgPath, depth))
	case *gno.ArrayType:
		return prefix(fmt.Sprintf("[%d]", t.Len), typeStr(t.Elt, pkgPath, depth))
	case *gno.MapType:
		k, v := typeStr(t.Key, pkgPath, depth), typeStr(t.Value, pkgPath, depth)
		if k == "" || v == "" {
			return ""
		}
		return "map[" + k + "]" + v
	case *gno.ChanType:
		dir := "chan "
		switch t.Dir {
		case gno.SEND:
			dir = "chan<- "
		case gno.RECV:
			dir = "<-chan "
		}
		return prefix(dir, typeStr(t.Elt, pkgPath, depth))
	case *gno.FuncType:
		return gnoSignature(t, pkgPath, depth)
	case *gno.StructType:
		parts := make([]string, 0, len(t.Fields))
		for _, f := range t.Fields {
			ft := typeStr(f.Type, pkgPath, depth)
			if ft == "" {
				return ""
			}
			parts = append(parts, fieldString(string(f.Name), ft))
		}
		return "struct{" + strings.Join(parts, "; ") + "}"
	case *gno.InterfaceType:
		if t.Generic != "" {
			return ""
		}
		ms := slices.Clone(t.Methods)
		slices.SortStableFunc(ms, func(a, b gno.FieldType) int { return strings.Compare(string(a.Name), string(b.Name)) })
		parts := make([]string, 0, len(ms))
		for _, m := range ms {
			ft, ok := m.Type.(*gno.FuncType)
			if !ok {
				return ""
			}
			sig := gnoSignature(ft, pkgPath, depth)
			if sig == "" {
				return ""
			}
			parts = append(parts, string(m.Name)+strings.TrimPrefix(sig, "func"))
		}
		return "interface{" + strings.Join(parts, "; ") + "}"
	case *gno.DeclaredType:
		return qualify(t.PkgPath, string(t.Name), pkgPath)
	}
	return ""
}

func gnoSignature(ft *gno.FuncType, pkgPath string, depth int) string {
	params := make([]string, 0, len(ft.Params))
	for i, p := range ft.Params {
		s := typeStr(p.Type, pkgPath, depth)
		if st, ok := p.Type.(*gno.SliceType); ok && st.Vrd && i == len(ft.Params)-1 {
			s = prefix("...", typeStr(st.Elt, pkgPath, depth))
		}
		if s == "" {
			return ""
		}
		params = append(params, s)
	}
	results := make([]string, 0, len(ft.Results))
	for _, r := range ft.Results {
		s := typeStr(r.Type, pkgPath, depth)
		if s == "" {
			return ""
		}
		results = append(results, s)
	}
	return signature(params, results)
}

// goTypeString is typeString for a go/types type; "" for a type Gno has
// no counterpart of (an invalid type, a type parameter).
func goTypeString(t types.Type, pkgPath string) string {
	return goTypeStr(t, pkgPath, 0)
}

func goTypeStr(t types.Type, pkgPath string, depth int) string {
	if depth > maxTypeDepth {
		return ""
	}
	depth++
	switch t := t.(type) {
	case *types.Basic:
		switch t.Kind() {
		case types.Invalid:
			return ""
		case types.UntypedBool:
			return "untyped bool"
		case types.UntypedInt:
			return "untyped int"
		case types.UntypedRune:
			return "untyped rune"
		case types.UntypedFloat:
			return "untyped float"
		case types.UntypedComplex:
			return "untyped complex"
		case types.UntypedString:
			return "untyped string"
		case types.UntypedNil:
			return "untyped nil"
		case types.UnsafePointer:
			return "unsafe.Pointer"
		}
		return types.Typ[t.Kind()].Name() // rune → int32, byte → uint8
	case *types.Pointer:
		return prefix("*", goTypeStr(t.Elem(), pkgPath, depth))
	case *types.Slice:
		return prefix("[]", goTypeStr(t.Elem(), pkgPath, depth))
	case *types.Array:
		return prefix(fmt.Sprintf("[%d]", t.Len()), goTypeStr(t.Elem(), pkgPath, depth))
	case *types.Map:
		k, v := goTypeStr(t.Key(), pkgPath, depth), goTypeStr(t.Elem(), pkgPath, depth)
		if k == "" || v == "" {
			return ""
		}
		return "map[" + k + "]" + v
	case *types.Chan:
		dir := "chan "
		switch t.Dir() {
		case types.SendOnly:
			dir = "chan<- "
		case types.RecvOnly:
			dir = "<-chan "
		}
		return prefix(dir, goTypeStr(t.Elem(), pkgPath, depth))
	case *types.Signature:
		return goSignature(t, pkgPath, depth)
	case *types.Struct:
		parts := make([]string, 0, t.NumFields())
		for i := range t.NumFields() {
			f := t.Field(i)
			ft := goTypeStr(f.Type(), pkgPath, depth)
			if ft == "" {
				return ""
			}
			parts = append(parts, fieldString(f.Name(), ft))
		}
		return "struct{" + strings.Join(parts, "; ") + "}"
	case *types.Interface:
		if !t.IsMethodSet() {
			return "" // a constraint interface
		}
		parts := make([]string, 0, t.NumMethods())
		for i := range t.NumMethods() { // sorted by name already
			m := t.Method(i)
			sig := goSignature(m.Type().(*types.Signature), pkgPath, depth)
			if sig == "" {
				return ""
			}
			parts = append(parts, m.Name()+strings.TrimPrefix(sig, "func"))
		}
		return "interface{" + strings.Join(parts, "; ") + "}"
	case *types.Named:
		if t.TypeArgs() != nil && t.TypeArgs().Len() > 0 {
			return ""
		}
		obj := t.Obj()
		if obj.Pkg() == nil {
			return obj.Name()
		}
		return qualify(obj.Pkg().Path(), obj.Name(), pkgPath)
	case *types.Alias:
		return goTypeStr(types.Unalias(t), pkgPath, depth)
	}
	return ""
}

func goSignature(sig *types.Signature, pkgPath string, depth int) string {
	if sig.TypeParams() != nil && sig.TypeParams().Len() > 0 {
		return ""
	}
	params := make([]string, 0, sig.Params().Len())
	for i := range sig.Params().Len() {
		pt := sig.Params().At(i).Type()
		s := goTypeStr(pt, pkgPath, depth)
		if sig.Variadic() && i == sig.Params().Len()-1 {
			if st, ok := pt.(*types.Slice); ok {
				s = prefix("...", goTypeStr(st.Elem(), pkgPath, depth))
			}
		}
		if s == "" {
			return ""
		}
		params = append(params, s)
	}
	results := make([]string, 0, sig.Results().Len())
	for i := range sig.Results().Len() {
		s := goTypeStr(sig.Results().At(i).Type(), pkgPath, depth)
		if s == "" {
			return ""
		}
		results = append(results, s)
	}
	return signature(params, results)
}

// signature renders `func(P1, P2) R` / `func(P) (R1, R2)` / `func()`.
func signature(params, results []string) string {
	s := "func(" + strings.Join(params, ", ") + ")"
	switch len(results) {
	case 0:
	case 1:
		s += " " + results[0]
	default:
		s += " (" + strings.Join(results, ", ") + ")"
	}
	return s
}

func fieldString(name, typ string) string {
	if name == "" {
		return typ
	}
	return name + " " + typ
}

// prefix wraps a composed type, propagating an incomparable element.
func prefix(p, elem string) string {
	if elem == "" {
		return ""
	}
	return p + elem
}

// qualify spells a declared type: bare in the package under check and
// in the universe, `pkg.T` elsewhere (by package name, as both checkers
// print it).
func qualify(typePkg, name, pkgPath string) string {
	if typePkg == "" || typePkg == pkgPath || typePkg == ".uverse" {
		return name
	}
	return typePkg[strings.LastIndexByte(typePkg, '/')+1:] + "." + name
}
