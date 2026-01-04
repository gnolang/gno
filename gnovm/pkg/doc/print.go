// Copyright 2015 The Go Authors. All rights reserved.
// Copied and modified from Go source: cmd/doc/pkg.go
// Modifications done include:
// - Removing code for supporting documenting commands
// - Removing code for supporting import commands

package doc

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/doc/comment"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	punchedCardWidth = 80
	indent           = "    "
)

type pkgPrinter struct {
	name        string             // Package name, json for encoding/json.
	doc         *JSONDocumentation // From WriteJSONDocumentation
	typedValue  map[string]string  // Consts and vars related to types. val_name -> type_name
	constructor map[string]string  // Constructors. func_name -> type_name
	buf         pkgBuffer
	opt         *WriteDocumentationOptions
	importPath  string

	// this is set when an error should be returned up the call chain.
	// it is set together with a panic(errFatal), so it can be checked easily
	// when calling recover.
	err error
}

func (pkg *pkgPrinter) isExported(name string) bool {
	// cmd/doc uses a global here, so we change this to be a method.
	return pkg.opt.Unexported || token.IsExported(name)
}

func ToText(w io.Writer, text, prefix, codePrefix string) {
	// We don't have the package AST, so use a default Parser and Printer
	p := &comment.Parser{}
	d := p.Parse(text)
	pr := &comment.Printer{}
	pr.TextPrefix = prefix
	pr.TextCodePrefix = codePrefix
	w.Write(pr.Text(d))
}

// pkgBuffer is a wrapper for bytes.Buffer that prints a package clause the
// first time Write is called.
type pkgBuffer struct {
	pkg     *pkgPrinter
	printed bool // Prevent repeated package clauses.
	bytes.Buffer
}

func (pb *pkgBuffer) Write(p []byte) (int, error) {
	pb.packageClause()
	return pb.Buffer.Write(p)
}

func (pb *pkgBuffer) packageClause() {
	if !pb.printed {
		pb.printed = true
		pb.pkg.packageClause()
	}
}

var errFatal = errors.New("pkg/doc: pkgPrinter.Fatalf called")

// in cmd/go, pkg.Fatalf is like log.Fatalf, but panics so it can be recovered in the
// main do function, so it doesn't cause an exit. Allows testing to work
// without running a subprocess.
// For our purposes, we store the error in .err - the caller knows about this and will check it.
func (pkg *pkgPrinter) Fatalf(format string, args ...any) {
	pkg.err = fmt.Errorf(format, args...)
	panic(errFatal)
}

func (pkg *pkgPrinter) Printf(format string, args ...any) {
	fmt.Fprintf(&pkg.buf, format, args...)
}

func (pkg *pkgPrinter) flush() error {
	_, err := pkg.opt.w.Write(pkg.buf.Bytes())
	if err != nil {
		return err
	}
	pkg.buf.Reset() // Not needed, but it's a flush.
	return nil
}

var newlineBytes = []byte("\n\n") // We never ask for more than 2.

// newlines guarantees there are n newlines at the end of the buffer.
func (pkg *pkgPrinter) newlines(n int) {
	for !bytes.HasSuffix(pkg.buf.Bytes(), newlineBytes[:n]) {
		pkg.buf.WriteRune('\n')
	}
}

// emit prints the node signature. If pkg.opt.Source is true, it ignores the provided comment,
// assuming the comment is in the node itself. Otherwise, the go/doc package
// clears the stuff we don't want to print anyway. It's a bit of a magic trick.
func (pkg *pkgPrinter) emit(comment string, node string) {
	if node != "" {
		pkg.Printf("%s\n", node)
		if comment != "" && !pkg.opt.Source {
			pkg.newlines(1)
			ToText(&pkg.buf, comment, indent, indent+indent)
			pkg.newlines(2) // Blank line after comment to separate from next item.
		} else {
			pkg.newlines(1)
		}
	}
}

// oneLineType parses the expression and returns a one-line summary of the node.
func (pkg *pkgPrinter) oneLineType(expression string) string {
	node, err := parser.ParseExpr(expression)
	if err != nil {
		// Panic on error, which shouldn't happen since it should be a valid expression from JSONDocumentation
		pkg.Fatalf("Error %s on parsing type expression: %s", err.Error(), expression)
	}
	const maxDepth = 10
	return pkg.oneLineNodeDepth(node, maxDepth)
}

// oneLineNodeDepth returns a one-line summary of the given input node.
// The depth specifies the maximum depth when traversing the AST.
func (pkg *pkgPrinter) oneLineNodeDepth(node ast.Node, depth int) string {
	const dotDotDot = "..."
	if depth == 0 {
		return dotDotDot
	}
	depth--

	switch n := node.(type) {
	case nil:
		return ""

	case *ast.FuncDecl:
		// Formats func declarations.
		name := n.Name.Name
		recv := pkg.oneLineNodeDepth(n.Recv, depth)
		if len(recv) > 0 {
			recv = "(" + recv + ") "
		}
		fnc := pkg.oneLineNodeDepth(n.Type, depth)
		fnc = strings.TrimPrefix(fnc, "func")
		return fmt.Sprintf("func %s%s%s", recv, name, fnc)

	case *ast.TypeSpec:
		sep := " "
		if n.Assign.IsValid() {
			sep = " = "
		}
		tparams := pkg.formatTypeParams(n.TypeParams, depth)
		return fmt.Sprintf("type %s%s%s%s", n.Name.Name, tparams, sep, pkg.oneLineNodeDepth(n.Type, depth))

	case *ast.FuncType:
		var params []string
		if n.Params != nil {
			for _, field := range n.Params.List {
				params = append(params, pkg.oneLineField(field, depth))
			}
		}
		needParens := false
		var results []string
		if n.Results != nil {
			needParens = needParens || len(n.Results.List) > 1
			for _, field := range n.Results.List {
				needParens = needParens || len(field.Names) > 0
				results = append(results, pkg.oneLineField(field, depth))
			}
		}

		tparam := pkg.formatTypeParams(n.TypeParams, depth)
		param := joinStrings(params)
		if len(results) == 0 {
			return fmt.Sprintf("func%s(%s)", tparam, param)
		}
		result := joinStrings(results)
		if !needParens {
			return fmt.Sprintf("func%s(%s) %s", tparam, param, result)
		}
		return fmt.Sprintf("func%s(%s) (%s)", tparam, param, result)

	case *ast.StructType:
		if n.Fields == nil || len(n.Fields.List) == 0 {
			return "struct{}"
		}
		return "struct{ ... }"

	case *ast.InterfaceType:
		if n.Methods == nil || len(n.Methods.List) == 0 {
			return "interface{}"
		}
		return "interface{ ... }"

	case *ast.FieldList:
		if n == nil || len(n.List) == 0 {
			return ""
		}
		if len(n.List) == 1 {
			return pkg.oneLineField(n.List[0], depth)
		}
		return dotDotDot

	case *ast.FuncLit:
		return pkg.oneLineNodeDepth(n.Type, depth) + " { ... }"

	case *ast.CompositeLit:
		typ := pkg.oneLineNodeDepth(n.Type, depth)
		if len(n.Elts) == 0 {
			return fmt.Sprintf("%s{}", typ)
		}
		return fmt.Sprintf("%s{ %s }", typ, dotDotDot)

	case *ast.ArrayType:
		length := pkg.oneLineNodeDepth(n.Len, depth)
		element := pkg.oneLineNodeDepth(n.Elt, depth)
		return fmt.Sprintf("[%s]%s", length, element)

	case *ast.MapType:
		key := pkg.oneLineNodeDepth(n.Key, depth)
		value := pkg.oneLineNodeDepth(n.Value, depth)
		return fmt.Sprintf("map[%s]%s", key, value)

	case *ast.CallExpr:
		fnc := pkg.oneLineNodeDepth(n.Fun, depth)
		var args []string
		for _, arg := range n.Args {
			args = append(args, pkg.oneLineNodeDepth(arg, depth))
		}
		return fmt.Sprintf("%s(%s)", fnc, joinStrings(args))

	case *ast.UnaryExpr:
		return fmt.Sprintf("%s%s", n.Op, pkg.oneLineNodeDepth(n.X, depth))

	case *ast.Ident:
		return n.Name

	default:
		// As a fallback, use default formatter for all unknown node types.
		buf := new(strings.Builder)
		format.Node(buf, token.NewFileSet(), node)
		s := buf.String()
		if strings.Contains(s, "\n") {
			return dotDotDot
		}
		return s
	}
}

func (pkg *pkgPrinter) formatTypeParams(list *ast.FieldList, depth int) string {
	if list.NumFields() == 0 {
		return ""
	}
	tparams := make([]string, 0, len(list.List))
	for _, field := range list.List {
		tparams = append(tparams, pkg.oneLineField(field, depth))
	}
	return "[" + joinStrings(tparams) + "]"
}

// oneLineField returns a one-line summary of the field.
func (pkg *pkgPrinter) oneLineField(field *ast.Field, depth int) string {
	names := make([]string, 0, len(field.Names))
	for _, name := range field.Names {
		names = append(names, name.Name)
	}
	if len(names) == 0 {
		return pkg.oneLineNodeDepth(field.Type, depth)
	}
	return joinStrings(names) + " " + pkg.oneLineNodeDepth(field.Type, depth)
}

// joinStrings formats the input as a comma-separated list,
// but truncates the list at some reasonable length if necessary.
func joinStrings(ss []string) string {
	var n int
	for i, s := range ss {
		n += len(s) + len(", ")
		if n > punchedCardWidth {
			ss = append(ss[:i:i], "...")
			break
		}
	}
	return strings.Join(ss, ", ")
}

// oneLineDecl returns a one-line summary of the given input value.
func (pkg *pkgPrinter) oneLineDecl(value *JSONValueDecl) string {
	const dotDotDot = "..."

	// Formats const and var declarations.
	trailer := ""
	if len(value.Values) > 1 {
		trailer = " " + dotDotDot
	}

	// Find the first relevant spec.
	typ := value.Values[0].Type
	for _, spec := range value.Values {
		// The type name may carry over from a previous specification in the
		// case of constants and iota.
		if spec.Type != "" {
			typ = spec.Type
		}

		if !pkg.isExported(spec.Name) {
			continue
		}
		var token string
		if value.Const {
			token = "const"
		} else {
			token = "var"
		}

		typeString := typ
		if typ != "" {
			typeString = pkg.oneLineType(typ)
		}
		return fmt.Sprintf("%s %s %s%s", token, spec.Name, typeString, trailer)
	}
	return ""
}

// allDoc prints all the docs for the package.
func (pkg *pkgPrinter) allDoc() {
	pkg.Printf("") // Trigger the package clause; we know the package exists.
	ToText(&pkg.buf, pkg.doc.PackageDoc, "", indent)
	pkg.newlines(1)

	printed := make(map[*JSONValueDecl]bool)

	hdr := ""
	printHdr := func(s string) {
		if hdr != s {
			pkg.Printf("\n%s\n\n", s)
			hdr = s
		}
	}

	// Constants.
	for _, value := range pkg.doc.Values {
		if !value.Const {
			continue
		}
		// Constants and variables come in groups, and valueDoc prints
		// all the items in the group. We only need to find one exported symbol.
		for _, v := range value.Values {
			if pkg.isExported(v.Name) && pkg.typedValue[v.Name] == "" {
				printHdr("CONSTANTS")
				pkg.valueDoc(value, printed)
				break
			}
		}
	}

	// Variables.
	for _, value := range pkg.doc.Values {
		if value.Const {
			continue
		}
		// Constants and variables come in groups, and valueDoc prints
		// all the items in the group. We only need to find one exported symbol.
		for _, v := range value.Values {
			if pkg.isExported(v.Name) && pkg.typedValue[v.Name] == "" {
				printHdr("VARIABLES")
				pkg.valueDoc(value, printed)
				break
			}
		}
	}

	// Functions.
	for _, fun := range pkg.doc.Funcs {
		if fun.Type == "" && pkg.isExported(fun.Name) && pkg.constructor[fun.Name] == "" {
			printHdr("FUNCTIONS")
			pkg.emit(fun.Doc, fun.Signature)
		}
	}

	// Types.
	for _, typ := range pkg.doc.Types {
		if pkg.isExported(typ.Name) {
			printHdr("TYPES")
			pkg.typeDoc(typ)
		}
	}
}

// packageDoc prints the docs for the package (package doc plus one-liners of the rest).
func (pkg *pkgPrinter) packageDoc() {
	pkg.Printf("") // Trigger the package clause; we know the package exists.
	if !pkg.opt.Short {
		ToText(&pkg.buf, pkg.doc.PackageDoc, "", indent)
		pkg.newlines(1)
	}

	if !pkg.opt.Short {
		pkg.newlines(2) // Guarantee blank line before the components.
	}

	pkg.valueSummary(pkg.doc.Values, false, "")
	pkg.funcSummary(pkg.doc.Funcs, false, "")
	pkg.typeSummary()
	if !pkg.opt.Short {
		pkg.bugs()
	}
}

// packageClause prints the package clause.
func (pkg *pkgPrinter) packageClause() {
	if pkg.opt.Short {
		return
	}

	// If we're using modules, the import path derived from module code locations wins.
	// If we did a file system scan, we knew the import path when we found the directory.
	// But if we started with a directory name, we never knew the import path.
	// Either way, we don't know it now, and it's cheap to (re)compute it.
	/* TODO: add when supporting gno doc on local directories
	if usingModules {
		for _, root := range codeRoots() {
			if pkg.build.Dir == root.dir {
				importPath = root.importPath
				break
			}
			if strings.HasPrefix(pkg.build.Dir, root.dir+string(filepath.Separator)) {
				suffix := filepath.ToSlash(pkg.build.Dir[len(root.dir)+1:])
				if root.importPath == "" {
					importPath = suffix
				} else {
					importPath = root.importPath + "/" + suffix
				}
				break
			}
		}
	}
	*/

	pkg.Printf("%s\n\n", pkg.doc.PackageLine)
	/* TODO
	if !usingModules && importPath != pkg.build.ImportPath {
		pkg.Printf("WARNING: package source is installed in %q\n", pkg.build.ImportPath)
	} */
}

// valueSummary prints a one-line summary for each set of values and constants.
// If all the types in a constant or variable declaration belong to the same
// type they can be printed by typeSummary, and so can be suppressed here.
// If typeName is not "" then only print values for the type.
func (pkg *pkgPrinter) valueSummary(values []*JSONValueDecl, showGrouped bool, typeName string) {
	var isGrouped map[*JSONValue]bool
	if !showGrouped {
		isGrouped = make(map[*JSONValue]bool)
		for _, typ := range pkg.doc.Types {
			if !pkg.isExported(typ.Name) {
				continue
			}
			for _, value := range pkg.doc.Values {
				// Remember the previous value's Type in case the Type is "" (for example, iota)
				vType := value.Values[0].Type
				for _, v := range value.Values {
					if v.Type != "" {
						vType = v.Type
					}
					if vType == typ.Name {
						isGrouped[v] = true
					}
				}
			}
		}
	}

	for _, value := range values {
		for _, v := range value.Values {
			if !isGrouped[v] {
				if typeName != "" && strings.TrimPrefix(v.Type, "*") != typeName {
					break
				}
				// Make a singleton for oneLineDecl
				oneValue := &JSONValueDecl{
					Const:  value.Const,
					Values: []*JSONValue{v},
				}
				if decl := pkg.oneLineDecl(oneValue); decl != "" {
					pkg.Printf("%s\n", decl)
					break
				}
			}
		}
	}
}

// funcSummary prints a one-line summary for each function. Constructors
// are printed by typeSummary, below, and so can be suppressed here.
// Only show functions whose Type is typeName (including if typeName is "").
func (pkg *pkgPrinter) funcSummary(funcs []*JSONFunc, showConstructors bool, typeName string) {
	for _, fun := range funcs {
		if fun.Type != typeName {
			continue
		}
		if pkg.isExported(fun.Name) {
			if showConstructors || pkg.constructor[fun.Name] == "" {
				pkg.Printf("%s\n", fun.Signature)
			}
		}
	}
}

// typeSummary prints a one-line summary for each type, followed by its constructors.
func (pkg *pkgPrinter) typeSummary() {
	for _, typ := range pkg.doc.Types {
		if pkg.isExported(typ.Name) {
			pkg.Printf("type %s %s\n", typ.Name, pkg.oneLineType(typ.Type))
			// Now print the consts, vars, and constructors.
			for _, value := range pkg.doc.Values {
				for _, v := range value.Values {
					if pkg.isExported(v.Name) && pkg.typedValue[v.Name] == typ.Name {
						// Make a singleton for oneLineDecl
						oneValue := &JSONValueDecl{
							Const:  value.Const,
							Values: []*JSONValue{v},
						}
						if decl := pkg.oneLineDecl(oneValue); decl != "" {
							pkg.Printf(indent+"%s\n", decl)
							break
						}
					}
				}
			}
			for _, constructor := range pkg.doc.Funcs {
				if constructor.Type != "" {
					// Constructors are not methods
					continue
				}
				if pkg.constructor[constructor.Name] != typ.Name {
					continue
				}
				if pkg.isExported(constructor.Name) {
					ToText(&pkg.buf, constructor.Signature, indent, "")
				}
			}
		}
	}
}

// bugs prints the BUGS information for the package.
// TODO: Provide access to TODOs and NOTEs as well (very noisy so off by default)?
func (pkg *pkgPrinter) bugs() {
	if len(pkg.doc.Bugs) == 0 {
		return
	}
	pkg.Printf("\n")
	for _, note := range pkg.doc.Bugs {
		pkg.Printf("%s: %v\n", "BUG", note)
	}
}

// findValues finds the constants and variables in doc.Values that describe the symbol.
func (pkg *pkgPrinter) findValues(symbol string) (values []*JSONValueDecl) {
	for _, value := range pkg.doc.Values {
		for _, v := range value.Values {
			if pkg.match(symbol, v.Name) {
				values = append(values, value)
			}
		}
	}
	return
}

// findFuncs finds the doc.Funcs that describes the symbol.
func (pkg *pkgPrinter) findFuncs(symbol string) (funcs []*JSONFunc) {
	for _, fun := range pkg.doc.Funcs {
		if pkg.match(symbol, fun.Name) {
			funcs = append(funcs, fun)
		}
	}
	return
}

// findTypes finds the JSONType that describes the symbol.
// If symbol is empty, it finds all exported types.
func (pkg *pkgPrinter) findTypes(symbol string) (types []*JSONType) {
	for _, typ := range pkg.doc.Types {
		if symbol == "" && pkg.isExported(typ.Name) || pkg.match(symbol, typ.Name) {
			types = append(types, typ)
		}
	}
	return
}

// symbolDoc prints the docs for symbol. There may be multiple matches.
// If symbol matches a type, output includes its methods factories and associated constants.
// If there is no top-level symbol, symbolDoc looks for methods that match.
func (pkg *pkgPrinter) symbolDoc(symbol string) {
	found := false
	// Functions.
	for _, fun := range pkg.findFuncs(symbol) {
		if fun.Type != "" {
			continue
		}

		// Symbol is a function.
		pkg.emit(fun.Doc, fun.Signature)
		found = true
	}
	// Constants and variables behave the same.
	values := pkg.findValues(symbol)
	// A declaration like
	//	const ( c = 1; C = 2 )
	// could be printed twice if the -u flag is set, as it matches twice.
	// So we remember which declarations we've printed to avoid duplication.
	printed := make(map[*JSONValueDecl]bool)
	for _, value := range values {
		pkg.valueDoc(value, printed)
		found = true
	}
	// Types.
	for _, typ := range pkg.findTypes(symbol) {
		pkg.typeDoc(typ)
		found = true
	}
	if !found {
		// See if there are methods.
		pkg.printMethodDoc("", symbol)
	}
}

// valueDoc prints the docs for a constant or variable.
func (pkg *pkgPrinter) valueDoc(value *JSONValueDecl, printed map[*JSONValueDecl]bool) {
	if printed[value] {
		return
	}
	pkg.emit(value.Doc, value.Signature)
	printed[value] = true
}

// typeDoc prints the docs for a type, including constructors and other items
// related to it.
func (pkg *pkgPrinter) typeDoc(typ *JSONType) {
	if typ.Kind == interfaceKind {
		saveInterElemsCount := len(typ.InterElems)
		pkg.trimUnexportedElems(typ)
		if len(typ.InterElems) == saveInterElemsCount {
			pkg.emit(typ.Doc, "type "+typ.Name+" "+typ.Type)
		} else {
			pkg.Printf("type %s interface {\n", typ.Name)
			for _, interElem := range typ.InterElems {
				if interElem.Type != "" {
					// Embedded type
					pkg.Printf("%s%s\n", indent, interElem.Type)
				} else {
					lineComment := ""
					if interElem.Method.Doc != "" {
						lineComment = fmt.Sprintf("  %s", interElem.Method.Doc)
					}
					pkg.Printf("%s%s%s\n", indent, interElem.Method.Signature, lineComment)
				}
			}
			pkg.Printf("%s// Has unexported methods.\n", indent)
			pkg.Printf("}\n")
		}
	} else {
		saveFieldCount := len(typ.Fields)
		pkg.trimUnexportedElems(typ)
		if len(typ.Fields) == saveFieldCount {
			assign := " "
			if typ.Alias {
				assign = " = "
			}
			pkg.emit(typ.Doc, "type "+typ.Name+assign+typ.Type)
		} else {
			pkg.Printf("type %s struct {\n", typ.Name)
			for _, field := range typ.Fields {
				lineComment := ""
				if field.Doc != "" {
					lineComment = fmt.Sprintf("  %s", field.Doc)
				}
				pkg.Printf("%s%s %s%s\n", indent, field.Name, field.Type, lineComment)
			}
			pkg.Printf("%s// Has unexported fields.\n", indent)
			pkg.Printf("}\n")
		}
	}
	pkg.newlines(2)
	// Show associated methods, constants, etc.
	if pkg.opt.ShowAll {
		printed := make(map[*JSONValueDecl]bool)
		for _, value := range pkg.doc.Values {
			for _, v := range value.Values {
				if pkg.isExported(v.Name) && pkg.typedValue[v.Name] == typ.Name {
					if pkg.opt.ShowAll {
						pkg.valueDoc(value, printed)
					} else {
						pkg.Printf("%s\n", value.Signature)
					}
					break
				}
			}
		}
		for _, constructor := range pkg.doc.Funcs {
			if constructor.Type != "" {
				// Constructors are not methods
				continue
			}
			if pkg.constructor[constructor.Name] != typ.Name {
				continue
			}
			if pkg.isExported(constructor.Name) {
				pkg.emit(constructor.Doc, constructor.Signature)
			}
		}
		for _, meth := range pkg.doc.Funcs {
			if meth.Type != typ.Name {
				continue
			}
			if pkg.isExported(meth.Name) {
				pkg.Printf("%s\n", meth.Signature)
				if pkg.opt.ShowAll && meth.Doc != "" {
					pkg.Printf("    %s\n", meth.Doc)
				}
			}
		}
	} else {
		pkg.valueSummary(pkg.doc.Values, true, typ.Name)
		// constructors
		for _, constructor := range pkg.doc.Funcs {
			if constructor.Type != "" || pkg.constructor[constructor.Name] != typ.Name {
				continue
			}
			pkg.funcSummary([]*JSONFunc{constructor}, true, "")
		}
		pkg.funcSummary(pkg.doc.Funcs, true, typ.Name)
	}
}

// trimUnexportedElems modifies typ in place to elide unexported fields from
// structs and methods from interfaces (unless the unexported flag is set or we
// are asked to show the original source).
func (pkg *pkgPrinter) trimUnexportedElems(typ *JSONType) {
	if pkg.opt.Unexported || pkg.opt.Source {
		return
	}
	if typ.Kind == interfaceKind {
		typ.InterElems = pkg.trimUnexportedMethods(typ.InterElems)
	} else {
		typ.Fields = pkg.trimUnexportedFields(typ.Fields)
	}
}

// trimUnexportedFields returns the field list trimmed of unexported fields.
func (pkg *pkgPrinter) trimUnexportedFields(fields []*JSONField) []*JSONField {
	trimmed := false
	list := make([]*JSONField, 0, len(fields))
	for _, field := range fields {
		name := field.Name
		if name == "" {
			// Embedded type. Use the name of the type.
			name = strings.TrimPrefix(field.Type, "*")
		}
		// Trims if any is unexported. Good enough in practice.
		ok := true
		if !pkg.isExported(name) {
			trimmed = true
			ok = false
		}
		if ok {
			list = append(list, field)
		}
	}
	if !trimmed {
		return fields
	}
	return list
}

// trimUnexportedMethods returns the method list trimmed of unexported methods and embedded types.
func (pkg *pkgPrinter) trimUnexportedMethods(interElems []*JSONInterfaceElement) []*JSONInterfaceElement {
	trimmed := false
	list := make([]*JSONInterfaceElement, 0, len(interElems))
	for _, interElem := range interElems {
		var name string
		constraint := false
		if interElem.Type != "" {
			// Embedded type. Use the name of the type.
			if strings.Contains(interElem.Type, " ") {
				// Assume this is a constraint or an interface literals like "interface { A() }"
				constraint = true
			} else {
				name = strings.TrimPrefix(interElem.Type, "*")
			}
		} else {
			name = interElem.Method.Name
		}
		// Trims if any is unexported. Good enough in practice.
		ok := true
		if !constraint && !pkg.isExported(name) {
			trimmed = true
			ok = false
		}
		if ok {
			list = append(list, interElem)
		}
	}
	if !trimmed {
		return interElems
	}
	return list
}

// printMethodDoc prints the docs for matches of symbol.method.
// If symbol is empty, it prints all methods for any concrete type
// that match the name. It reports whether it found any methods.
func (pkg *pkgPrinter) printMethodDoc(symbol, method string) bool {
	types := pkg.findTypes(symbol)
	if types == nil {
		if symbol == "" {
			return false
		}
		pkg.Fatalf("symbol %s is not a type in package %s installed in %q", symbol, pkg.name, pkg.importPath)
	}
	found := false
	for _, typ := range types {
		hasMethods := false
		for _, meth := range pkg.doc.Funcs {
			if meth.Type != typ.Name {
				continue
			}
			hasMethods = true
			if pkg.match(method, meth.Name) {
				pkg.emit(meth.Doc, meth.Signature)
				found = true
			}
		}
		if hasMethods {
			continue
		}
		if symbol == "" {
			continue
		}
	}
	return found
}

// printFieldDoc prints the docs for matches of symbol.fieldName.
// It reports whether it found any field.
// Both symbol and fieldName must be non-empty or it returns false.
func (pkg *pkgPrinter) printFieldDoc(symbol, fieldName string) bool {
	if symbol == "" || fieldName == "" {
		return false
	}
	types := pkg.findTypes(symbol)
	if types == nil {
		pkg.Fatalf("symbol %s is not a type in package %s installed in %q", symbol, pkg.name, pkg.importPath)
	}
	found := false
	numUnmatched := 0
	for _, typ := range types {
		for _, field := range typ.Fields {
			if !pkg.match(fieldName, field.Name) {
				numUnmatched++
				continue
			}
			if !found {
				pkg.Printf("type %s struct {\n", typ.Name)
			}
			s := pkg.oneLineType(field.Type)
			lineComment := ""
			if field.Doc != "" {
				lineComment = fmt.Sprintf("  %s", strings.TrimSpace(field.Doc))
			}
			pkg.Printf("%s%s %s%s\n", indent, field.Name, s, lineComment)
			found = true
		}
	}
	if found {
		if numUnmatched > 0 {
			pkg.Printf("\n    // ... other fields elided ...\n")
		}
		pkg.Printf("}\n")
	}
	return found
}

// printInterfaceMethodDoc prints the docs for matches of symbol.methodName.
// It reports whether it found any field.
// Both symbol and methodName must be non-empty or it returns false.
func (pkg *pkgPrinter) printInterfaceMethodDoc(symbol, methodName string) bool {
	if symbol == "" || methodName == "" {
		return false
	}
	types := pkg.findTypes(symbol)
	if types == nil {
		pkg.Fatalf("symbol %s is not a type in package %s installed in %q", symbol, pkg.name, pkg.importPath)
	}
	found := false
	numUnmatched := 0
	for _, typ := range types {
		for _, interElem := range typ.InterElems {
			if interElem.Method == nil || !pkg.match(methodName, interElem.Method.Name) {
				numUnmatched++
				continue
			}
			if !found {
				pkg.Printf("type %s interface {\n", typ.Name)
			}
			lineComment := ""
			if interElem.Method.Doc != "" {
				lineComment = fmt.Sprintf("  %s", strings.TrimSpace(interElem.Method.Doc))
			}
			pkg.Printf("%s%s%s\n", indent, interElem.Method.Signature, lineComment)
			found = true
		}
	}
	if found {
		if numUnmatched > 0 {
			pkg.Printf("\n    // ... other methods elided ...\n")
		}
		pkg.Printf("}\n")
	}
	return found
}

// methodDoc prints the docs for matches of symbol.method.
func (pkg *pkgPrinter) methodDoc(symbol, method string) bool {
	return pkg.printMethodDoc(symbol, method)
}

// fieldDoc prints the docs for matches of symbol.field.
func (pkg *pkgPrinter) fieldDoc(symbol, field string) bool {
	return pkg.printFieldDoc(symbol, field)
}

func (pkg *pkgPrinter) match(user, program string) bool {
	if !pkg.isExported(program) {
		return false
	}
	return symbolMatch(user, program)
}

// match reports whether the user's symbol matches the program's.
// A lower-case character in the user's string matches either case in the program's.
func symbolMatch(user, program string) bool {
	/* TODO: might be useful to add for tooling.
	if matchCase {
		return user == program
	} */
	for _, u := range user {
		p, w := utf8.DecodeRuneInString(program)
		program = program[w:]
		if u == p {
			continue
		}
		if unicode.IsLower(u) && simpleFold(u) == simpleFold(p) {
			continue
		}
		return false
	}
	return program == ""
}

// simpleFold returns the minimum rune equivalent to r
// under Unicode-defined simple case folding.
func simpleFold(r rune) rune {
	for {
		r1 := unicode.SimpleFold(r)
		if r1 <= r {
			return r1 // wrapped around, found min
		}
		r = r1
	}
}
