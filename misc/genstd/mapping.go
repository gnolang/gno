package main

import (
	"fmt"
	"go/ast"
	"go/types"
	"path"
	"strconv"
)

const gnoPackagePath = "github.com/gnolang/gno/gnovm/pkg/gnolang"

type mapping struct {
	GnoImportPath string // time
	GnoFunc       string // now
	GoImportPath  string // github.com/gnolang/gno/gnovm/stdlibs/time
	GoFunc        string // X_now
	Params        []mappingType
	Results       []mappingType
	MachineParam  bool

	gnoImports []*ast.ImportSpec
	goImports  []*ast.ImportSpec
}

type mappingType struct {
	Type ast.Expr

	// IsTypedValue is set to true if the parameter or result in go is of type
	// gno.TypedValue. This prevents the generated code from performing
	// Go2Gno/Gno2Go reflection-based conversion.
	IsTypedValue bool
}

func (mt mappingType) GoQualifiedName() string {
	return types.ExprString(mt.Type)
}

func (mt mappingType) GnoType() string {
	return types.ExprString(mt.Type)
}

func linkFunctions(pkgs []*pkgData) []mapping {
	var mappings []mapping
	for _, pkg := range pkgs {
		for _, gb := range pkg.gnoBodyless {
			nameWant := gb.Name.Name
			if !gb.Name.IsExported() {
				nameWant = "X_" + nameWant
			}
			fn := findFuncByName(pkg.goExported, nameWant)
			if fn.FuncDecl == nil {
				panic(
					fmt.Errorf("package %q: no matching go function declaration (%q) exists for function %q",
						pkg.importPath, nameWant, gb.Name.Name),
				)
			}
			mp := mapping{
				GnoImportPath: pkg.importPath,
				GnoFunc:       gb.Name.Name,
				GoImportPath:  "github.com/gnolang/gno/" + relPath() + "/" + pkg.importPath,
				GoFunc:        fn.Name.Name,

				gnoImports: gb.imports,
				goImports:  fn.imports,
			}
			if !mp.signaturesMatch(gb, fn) {
				panic(
					fmt.Errorf("package %q: signature of gno function %s doesn't match signature of go function %s",
						pkg.importPath, gb.Name.Name, fn.Name.Name),
				)
			}
			mp.loadParamsResults(gb, fn)
			mappings = append(mappings, mp)
		}
	}
	return mappings
}

func findFuncByName(fns []funcDecl, name string) funcDecl {
	for _, fn := range fns {
		if fn.Name.Name == name {
			return fn
		}
	}
	return funcDecl{}
}

func (m *mapping) loadParamsResults(gnof, gof funcDecl) {
	// initialise with lengths
	m.Params = make([]mappingType, 0, gnof.Type.Params.NumFields())
	m.Results = make([]mappingType, 0, gnof.Type.Results.NumFields())

	gofpl := gof.Type.Params.List
	if m.MachineParam {
		// skip machine parameter
		gofpl = gofpl[1:]
	}
	if gnof.Type.Params != nil {
		m._loadParamsResults(&m.Params, gnof.Type.Params.List, gofpl)
	}
	if gnof.Type.Results != nil {
		m._loadParamsResults(&m.Results, gnof.Type.Results.List, gof.Type.Results.List)
	}
}

func (m *mapping) _loadParamsResults(dst *[]mappingType, gnol, gol []*ast.Field) {
	iterFields(gnol, gol, func(gnoe, goe ast.Expr) error {
		if m.isTypedValue(goe) {
			*dst = append(*dst, mappingType{Type: gnoe, IsTypedValue: true})
		} else {
			*dst = append(*dst, mappingType{Type: gnoe})
		}
		return nil
	})
}

// isGnoMachine checks whether field is of type *gno.Machine,
// and it has at most 1 name.
func (m *mapping) isGnoMachine(field *ast.Field) bool {
	if len(field.Names) > 1 {
		return false
	}

	return m.isGnoType(field.Type, true, "Machine")
}

// isTypedValue checks whether e is type gno.TypedValue.
func (m *mapping) isTypedValue(e ast.Expr) bool {
	return m.isGnoType(e, false, "TypedValue")
}

func (m *mapping) isGnoType(e ast.Expr, star bool, typeName string) bool {
	if star {
		px, ok := e.(*ast.StarExpr)
		if !ok {
			return false
		}
		e = px.X
	}

	sx, ok := e.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	imp := resolveSelectorImport(m.goImports, sx)
	return imp == gnoPackagePath && sx.Sel.Name == typeName
}

// iterFields iterates over gnol and gol, calling callback for each matching
// parameter. iterFields assumes the caller already checked for the "true" number
// of parameters in the two arrays to be equal (can be checked using
// (*ast.FieldList).NumFields()).
//
// If callback returns an error, iterFields returns that error immediately.
// No errors are otherwise generated.
func iterFields(gnol, gol []*ast.Field, callback func(gnoType, goType ast.Expr) error) error {
	var goIdx, goNameIdx int

	for _, l := range gnol {
		n := len(l.Names)
		if n == 0 {
			n = 1
		}
		gnoe := l.Type
		for i := 0; i < n; i++ {
			goe := gol[goIdx].Type

			if err := callback(gnoe, goe); err != nil {
				return err
			}

			goNameIdx++
			if goNameIdx >= len(gol[goIdx].Names) {
				goIdx++
				goNameIdx = 0
			}
		}
	}
	return nil
}

type typeMismatchError struct {
	gnoe, goe ast.Expr
}

func (te *typeMismatchError) Error() string {
	return fmt.Sprintf("typesEqual: gno type %q does not match go type %q",
		types.ExprString(te.gnoe), types.ExprString(te.goe))
}

// typesEqual ensures that the given gnoe and goe, expected to represent
// expressions to identify types, are equal.
func (m *mapping) typesEqual(gnoe, goe ast.Expr) error {
	// If a type assertion fails, like in the below
	// goe, ok := ..., then goe will be set to a zero value, and might
	// lead to nil pointer dereferences. Setting up the mismatch error
	// here avoids that.
	mismatch := typeMismatchError{gnoe, goe}

	switch gnoe := gnoe.(type) {
	// We're working with a subset of all expressions:
	// https://go.dev/ref/spec#Type

	case *ast.Ident:
		goi, ok := goe.(*ast.Ident)
		switch {
		case !ok || gnoe.Name != goi.Name:
			return &mismatch
		case !isBuiltin(gnoe.Name):
			return fmt.Errorf("typesEqual: usage of non-builtin type %q", gnoe.Name)
		default:
			return nil
		}
	case *ast.StarExpr:
		goe, ok := goe.(*ast.StarExpr)
		if !ok {
			return &mismatch
		}
		if err := m.typesEqual(gnoe.X, goe.X); err != nil {
			return err
		}
		return nil
	case *ast.ArrayType:
		goe, ok := goe.(*ast.ArrayType)
		if !ok || !basicLitsEqual(gnoe.Len, goe.Len) {
			return &mismatch
		}
		if err := m.typesEqual(gnoe.Elt, goe.Elt); err != nil {
			return err
		}
		return nil

	case *ast.StructType,
		*ast.FuncType,
		*ast.InterfaceType,
		*ast.MapType,
		*ast.Ellipsis,
		*ast.SelectorExpr:
		// TODO
		panic("not implemented")
	default:
		panic(fmt.Errorf("invalid expression as func param/return type: %T (%v)", gnoe, gnoe))
	}
}

// returns full import path from package ident
func resolveImport(imports []*ast.ImportSpec, ident string) string {
	for _, i := range imports {
		s, err := strconv.Unquote(i.Path.Value)
		if err != nil {
			panic(fmt.Errorf("could not unquote import path literal: %s", i.Path.Value))
		}

		// TODO: for simplicity, if i.Name is nil we assume the name to be ==
		// to the last part of the import path.
		// ideally, use importer to resolve package directory on user's FS and
		// resolve by parsing and reading package clause
		var name string
		if i.Name != nil {
			name = i.Name.Name
		} else {
			name = path.Base(s)
		}

		if name == ident {
			return s
		}
	}
	return ""
}

func resolveSelectorImport(imports []*ast.ImportSpec, sx *ast.SelectorExpr) string {
	pkgIdent, ok := sx.X.(*ast.Ident)
	if !ok {
		panic(fmt.Errorf("encountered unhandled SelectorExpr.X type: %T (%v)", sx.X, sx))
	}
	impPath := resolveImport(imports, pkgIdent.Name)
	if impPath == "" {
		panic(fmt.Errorf(
			"unknown identifier %q (for resolving type %q)",
			pkgIdent.Name, pkgIdent.Name+"."+sx.Sel.Name,
		))
	}
	return impPath
}

// simple equivalence between two BasicLits.
// Note that this returns true only if the expressions are exactly the same;
// ie. 16 != 0x10, only 16 == 16.
func basicLitsEqual(x1, x2 ast.Expr) bool {
	if x1 == nil || x2 == nil {
		return x1 == nil && x2 == nil
	}
	l1, ok1 := x1.(*ast.BasicLit)
	l2, ok2 := x2.(*ast.BasicLit)
	if !ok1 || !ok2 {
		return false
	}
	return l1.Value == l2.Value
}

// Signatures match when they accept the same, unnamed types.
//
// If the first parameter to the Go function is *[gnolang.Machine], it is
// ignored when matching to the Gno function.
//
// If a Go parameter is [gnolang.TypedValue], it always matches any
// corresponding parameter in Gno.
func (m *mapping) signaturesMatch(gnof, gof funcDecl) bool {
	if gnof.Type.TypeParams != nil || gof.Type.TypeParams != nil {
		panic("type parameters not supported")
	}

	// if first param of go function is *gno.Machine, remove it
	gofp := gof.Type.Params
	if gofp != nil && len(gofp.List) > 0 && m.isGnoMachine(gofp.List[0]) {
		// avoid touching original struct
		n := *gofp
		n.List = n.List[1:]
		gofp = &n

		m.MachineParam = true
	}

	return m.fieldListsMatch(gnof.Type.Params, gofp) &&
		m.fieldListsMatch(gnof.Type.Results, gof.Type.Results)
}

func (m *mapping) fieldListsMatch(gnofl, gofl *ast.FieldList) bool {
	if gnofl == nil || gofl == nil {
		return gnofl == nil && gofl == nil
	}
	if gnofl.NumFields() != gofl.NumFields() {
		return false
	}
	err := iterFields(gnofl.List, gofl.List, func(gnoe, goe ast.Expr) error {
		// if the go type is gno.TypedValue, we just don't perform reflect-based conversion.
		if m.isTypedValue(goe) {
			return nil
		}
		return m.typesEqual(gnoe, goe)
	})
	return err == nil
}

// TODO: this is created based on the uverse definitions. This should be
// centralized, or at least have a CI/make check to make sure this stays the
// same
var builtinTypes = [...]string{
	"bool",
	"string",
	"int",
	"int8",
	"int16",
	"rune",
	"int32",
	"int64",
	"uint",
	"byte",
	"uint8",
	"uint16",
	"uint32",
	"uint64",
	"bigint",
	"float32",
	"float64",
	"error",
}

func isBuiltin(name string) bool {
	for _, x := range builtinTypes {
		if x == name {
			return true
		}
	}
	return false
}
