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
	"go/doc/comment"
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

// oneLineNode returns a one-line summary of the given input node.
func (pkg *pkgPrinter) oneLineNode(node any) string {
	const maxDepth = 10
	return pkg.oneLineNodeDepth(node, maxDepth)
}

// oneLineNodeDepth returns a one-line summary of the given input node.
// The depth specifies the maximum depth when traversing.
func (pkg *pkgPrinter) oneLineNodeDepth(node any, depth int) string {
	const dotDotDot = "..."
	if depth == 0 {
		return dotDotDot
	}
	depth--

	switch n := node.(type) {
	case nil:
		return ""

	case *JSONValueDecl:
		// Formats const and var declarations.
		trailer := ""
		if len(n.Values) > 1 {
			trailer = " " + dotDotDot
		}

		// Find the first relevant spec.
		typ := n.Values[0].Type
		for _, spec := range n.Values {
			// The type name may carry over from a previous specification in the
			// case of constants and iota.
			if spec.Type != "" {
				typ = spec.Type
			}

			if !pkg.isExported(spec.Name) {
				continue
			}
			var token string
			if n.Const {
				token = "const"
			} else {
				token = "var"
			}
			return fmt.Sprintf("%s %s %s%s", token, spec.Name, typ, trailer)
		}
		return ""

	default:
		return ""
	}
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
				if typeName != "" && strings.Replace(v.Type, "*", "", -1) != typeName {
					break
				}
				// Make a singleton for oneLineNode
				oneValue := &JSONValueDecl{
					Const:  value.Const,
					Values: []*JSONValue{v},
				}
				if decl := pkg.oneLineNode(oneValue); decl != "" {
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
		// pkg.Printf("%s\n", pkg.oneLineNode(typ))
		pkg.Printf("type %s struct{ ... }\n", typ.Name)
		// Now print the consts, vars, and constructors.
		for _, value := range pkg.doc.Values {
			for _, v := range value.Values {
				if pkg.isExported(v.Name) && pkg.typedValue[v.Name] == typ.Name {
					// Make a singleton for oneLineNode
					oneValue := &JSONValueDecl{
						Const:  value.Const,
						Values: []*JSONValue{v},
					}
					if decl := pkg.oneLineNode(oneValue); decl != "" {
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
	/*
		// Print each spec only if there is at least one exported symbol in it.
		// (See issue 11008.)
		// TODO: Should we elide unexported symbols from a single spec?
		// It's an unlikely scenario, probably not worth the trouble.
		// TODO: Would be nice if go/doc did this for us.
		specs := make([]ast.Spec, 0, len(value.Decl.Specs))
		var typ ast.Expr
		for _, spec := range value.Decl.Specs {
			vspec := spec.(*ast.ValueSpec)

			// The type name may carry over from a previous specification in the
			// case of constants and iota.
			if vspec.Type != nil {
				typ = vspec.Type
			}

			for _, ident := range vspec.Names {
				if pkg.opt.Source || pkg.isExported(ident.Name) {
					if vspec.Type == nil && vspec.Values == nil && typ != nil {
						// This a standalone identifier, as in the case of iota usage.
						// Thus, assume the type comes from the previous type.
						vspec.Type = &ast.Ident{
							Name:    pkg.oneLineNode(typ),
							NamePos: vspec.End() - 1,
						}
					}

					specs = append(specs, vspec)
					typ = nil // Only inject type on first exported identifier
					break
				}
			}
		}
		if len(specs) == 0 {
			return
		}
		value.Decl.Specs = specs
		pkg.emit(value.Doc, value.Signature)
	*/
	pkg.emit(value.Doc, value.Signature)
	printed[value] = true
}

// typeDoc prints the docs for a type, including constructors and other items
// related to it.
func (pkg *pkgPrinter) typeDoc(typ *JSONType) {
	if typ.Kind == interfaceKind {
		saveMethodCount := len(typ.Methods)
		pkg.trimUnexportedElems(typ)
		if len(typ.Methods) == saveMethodCount {
			pkg.emit(typ.Doc, "type "+typ.Name+" "+typ.Type)
		} else {
			pkg.Printf("type %s interface {\n", typ.Name)
			for _, meth := range typ.Methods {
				lineComment := ""
				if meth.Doc != "" {
					lineComment = fmt.Sprintf("  %s", meth.Doc)
				}
				pkg.Printf("%s %s%s\n", indent, meth.Signature, lineComment)
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
		typ.Methods = pkg.trimUnexportedMethods(typ.Methods)
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
			name = strings.Replace(field.Type, "*", "", -1)
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

// trimUnexportedMethods returns the method list trimmed of unexported methods.
func (pkg *pkgPrinter) trimUnexportedMethods(methods []*JSONFunc) []*JSONFunc {
	trimmed := false
	list := make([]*JSONFunc, 0, len(methods))
	for _, meth := range methods {
		name := meth.Name
		if name == "" {
			// Embedded type. Use the name of the type.
			name = strings.Replace(meth.Type, "*", "", -1)
		}
		// Trims if any is unexported. Good enough in practice.
		ok := true
		if !pkg.isExported(name) {
			trimmed = true
			ok = false
		}
		if ok {
			list = append(list, meth)
		}
	}
	if !trimmed {
		return methods
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
		/*
			// Type may be an interface. The go/doc package does not attach
			// an interface's methods to the doc.Type. We need to dig around.
			spec := pkg.findTypeSpec(typ.Decl, typ.Name)
			inter, ok := spec.Type.(*ast.InterfaceType)
			if !ok {
				// Not an interface type.
				continue
			}

			// Collect and print only the methods that match.
			var methods []*ast.Field
			for _, iMethod := range inter.Methods.List {
				// This is an interface, so there can be only one name.
				// TODO: Anonymous methods (embedding)
				if len(iMethod.Names) == 0 {
					continue
				}
				name := iMethod.Names[0].Name
				if pkg.match(method, name) {
					methods = append(methods, iMethod)
					found = true
				}
			}
			if found {
				pkg.Printf("type %s ", spec.Name)
				inter.Methods.List, methods = methods, inter.Methods.List
				err := format.Node(&pkg.buf, pkg.fs, inter)
				if err != nil {
					pkg.Fatalf("%v", err)
				}
				pkg.newlines(1)
				// Restore the original methods.
				inter.Methods.List = methods
			}
		*/
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
			lineComment := ""
			if field.Doc != "" {
				lineComment = fmt.Sprintf("  %s", field.Doc)
			}
			pkg.Printf("%s%s %s%s\n", indent, field.Name, field.Type, lineComment)
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
