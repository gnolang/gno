package doc

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/doc"
	"go/doc/comment"
	"go/format"
	"go/printer"
	"go/token"
	"io"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// JSONDocumentation holds package documentation suitable for transmitting
// as JSON with printable string fields
type JSONDocumentation struct {
	PackagePath string   `json:"package_path"`
	PackageLine string   `json:"package_line"` // package io // import "io"
	PackageDoc  string   `json:"package_doc"`  // markdown of top-level package documentation
	Bugs        []string `json:"bugs"`         // From comments with "BUG(who): Details"

	// These match each of the sections in a pkg.go.dev package documentation
	Values []*JSONValueDecl `json:"values"` // constants and variables declared
	Funcs  []*JSONFunc      `json:"funcs"`  // Funcs and methods
	Types  []*JSONType      `json:"types"`
}

type JSONValueDecl struct {
	Signature string       `json:"signature"`
	Const     bool         `json:"const"`
	Values    []*JSONValue `json:"values"`
	Doc       string       `json:"doc"` // markdown
}

type JSONValue struct {
	Name string `json:"name"`
	Doc  string `json:"doc"`
	Type string `json:"type"` // often empty
}

type JSONField struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Doc  string `json:"doc"` // markdown
}

type JSONFunc struct {
	Type      string       `json:"type"` // if this is a method
	Name      string       `json:"name"`
	Crossing  bool         `json:"crossing"` // true if the first param is "cur realm"
	Signature string       `json:"signature"`
	Doc       string       `json:"doc"` // markdown
	Params    []*JSONField `json:"params"`
	Results   []*JSONField `json:"results"`
}

const (
	structKind    = "struct"
	interfaceKind = "interface"
	arrayKind     = "array"
	sliceKind     = "slice"
	mapKind       = "map"
	chanKind      = "chan"
	funcKind      = "func"
	pointerKind   = "pointer"
	identKind     = "ident"
)

type JSONInterfaceElement struct {
	Method *JSONFunc `json:"method,omitempty"` // Normal interface method
	Type   string    `json:"type,omitempty"`   // Embedded type
}

type JSONType struct {
	Name  string `json:"name"`  // "MyType"
	Type  string `json:"type"`  // "struct { ... }"
	Doc   string `json:"doc"`   // godoc documentation...
	Alias bool   `json:"alias"` // if an alias like `type A = B`
	Kind  string `json:"kind"`  // struct | interface | array | slice | map | channel | func | pointer | ident
	// TODO: Use omitzero when upgraded to Go 1.24
	InterElems []*JSONInterfaceElement `json:"inter_elems,omitempty"` // interface methods or embedded types (Kind == "interface") (struct methods are in JSONDocumentation.Funcs)
	Fields     []*JSONField            `json:"fields,omitempty"`      // struct fields (Kind == "struct")
}

// NewDocumentableFromMemPkg gets the pkgData from mpkg and returns a Documentable
func NewDocumentableFromMemPkg(mpkg *std.MemPackage, unexported bool, symbol, accessible string) (*Documentable, error) {
	pd, err := newPkgDataFromMemPkg(mpkg, unexported)
	if err != nil {
		return nil, err
	}

	doc := &Documentable{
		bfsDir:     pd.dir,
		pkgData:    pd,
		symbol:     symbol,
		accessible: accessible,
	}
	return doc, nil
}

// WriteJSONDocumentation returns a JSONDocumentation for the package
// A useful opt is Source=true. opt may be nil
func (d *Documentable) WriteJSONDocumentation(opt *WriteDocumentationOptions) (*JSONDocumentation, error) {
	if opt == nil {
		opt = &WriteDocumentationOptions{}
	}
	astpkg, pkg, err := d.pkgData.docPackage()
	if err != nil {
		return nil, err
	}
	file := ast.MergePackageFiles(astpkg, 0)

	// Create custom printer that doesn't generate heading IDs
	printer := createCustomPrinter(pkg)

	// Parse package documentation
	var pkgDoc string
	if pkg.Doc != "" {
		pkgDoc = normalizedMarkdownPrinter(printer, pkg.Doc)
	}

	jsonDoc := &JSONDocumentation{
		PackagePath: d.pkgData.dir.dir,
		PackageLine: fmt.Sprintf("package %s // import %q", pkg.Name, pkg.ImportPath),
		PackageDoc:  pkgDoc,
		Values:      []*JSONValueDecl{},
		Funcs:       []*JSONFunc{},
		Types:       []*JSONType{},
	}

	if pkg.Notes["BUG"] != nil {
		for _, note := range pkg.Notes["BUG"] {
			jsonDoc.Bugs = append(jsonDoc.Bugs, note.Body)
		}
	}

	for _, value := range pkg.Consts {
		var p comment.Parser
		doc := p.Parse(value.Doc)
		jsonDoc.Values = append(jsonDoc.Values, &JSONValueDecl{
			Signature: mustFormatNode(d.pkgData.fset, value.Decl, opt.Source, file),
			Const:     true,
			Values:    d.extractValueSpecs(value.Decl.Specs, printer),
			Doc:       normalizedMarkdownPrinter(printer, doc),
		})
	}

	for _, value := range pkg.Vars {
		var p comment.Parser
		doc := p.Parse(value.Doc)
		jsonDoc.Values = append(jsonDoc.Values, &JSONValueDecl{
			Signature: mustFormatNode(d.pkgData.fset, value.Decl, opt.Source, file),
			Const:     false,
			Values:    d.extractValueSpecs(value.Decl.Specs, printer),
			Doc:       normalizedMarkdownPrinter(printer, doc),
		})
	}

	for _, fun := range pkg.Funcs {
		params := d.extractJSONFields(fun.Decl.Type.Params)
		jsonDoc.Funcs = append(jsonDoc.Funcs, &JSONFunc{
			Name:      fun.Name,
			Crossing:  isCrossing(params),
			Signature: mustFormatNode(d.pkgData.fset, fun.Decl, opt.Source, file),
			Doc:       normalizedMarkdownPrinter(printer, fun.Doc),
			Params:    params,
			Results:   d.extractJSONFields(fun.Decl.Type.Results),
		})
	}

	for _, typ := range pkg.Types {
		typeSpec := getTypeSpec(typ)
		if typeSpec == nil || typeSpec.Type == nil {
			// We don't expect this
			continue
		}
		typeExpr := deparenthesize(typeSpec.Type)

		kind := ""
		var interElems []*JSONInterfaceElement
		var fields []*JSONField

		switch t := typeExpr.(type) {
		case *ast.StructType:
			kind = structKind
			// TODO: Anonymous fields.
			fields = d.extractJSONFields(t.Fields)
		case *ast.InterfaceType:
			kind = interfaceKind
			for _, iMethod := range t.Methods.List {
				if len(iMethod.Names) == 0 {
					// Embedded type
					interElems = append(interElems, &JSONInterfaceElement{
						Type: mustFormatNode(d.pkgData.fset, iMethod.Type, false, nil),
					})
					continue
				}

				// Method
				fun, ok := iMethod.Type.(*ast.FuncType)
				if !ok {
					// We don't expect this
					continue
				}
				// This is an interface, so we should expect only one name
				if len(iMethod.Names) != 1 {
					continue
				}
				name := iMethod.Names[0].Name

				docBuf := new(strings.Builder)
				if iMethod.Doc != nil {
					for _, comment := range iMethod.Doc.List {
						docBuf.WriteString(comment.Text)
						docBuf.WriteString("\n")
					}
				}
				if iMethod.Comment != nil {
					for _, comment := range iMethod.Comment.List {
						docBuf.WriteString(comment.Text)
						docBuf.WriteString("\n")
					}
				}

				interElems = append(interElems, &JSONInterfaceElement{
					Method: &JSONFunc{
						Type:      typ.Name,
						Name:      name,
						Signature: name + strings.TrimPrefix(mustFormatNode(d.pkgData.fset, fun, false, file), "func"),
						Doc:       string(pkg.Markdown(docBuf.String())),
						Params:    d.extractJSONFields(fun.Params),
						Results:   d.extractJSONFields(fun.Results),
					},
				})
			}
		case *ast.ArrayType:
			if t.Len == nil {
				kind = sliceKind
			} else {
				kind = arrayKind
			}
		case *ast.MapType:
			kind = mapKind
		case *ast.ChanType:
			kind = chanKind
		case *ast.FuncType:
			kind = funcKind
		case *ast.StarExpr:
			kind = pointerKind
		default:
			// Default to ident
			kind = identKind
		}

		jsonDoc.Types = append(jsonDoc.Types, &JSONType{
			Name:       typ.Name,
			Type:       mustFormatNode(d.pkgData.fset, typeExpr, false, file),
			Doc:        normalizedMarkdownPrinter(printer, typ.Doc),
			Alias:      typeSpec.Assign != 0,
			Kind:       kind,
			InterElems: interElems,
			Fields:     fields,
		})

		// values of this type
		for _, c := range typ.Consts {
			var p comment.Parser
			doc := p.Parse(c.Doc)
			jsonDoc.Values = append(jsonDoc.Values, &JSONValueDecl{
				Signature: mustFormatNode(d.pkgData.fset, c.Decl, opt.Source, file),
				Const:     true,
				Values:    d.extractValueSpecs(c.Decl.Specs, printer),
				Doc:       normalizedMarkdownPrinter(printer, doc),
			})
		}
		for _, v := range typ.Vars {
			var p comment.Parser
			doc := p.Parse(v.Doc)
			jsonDoc.Values = append(jsonDoc.Values, &JSONValueDecl{
				Signature: mustFormatNode(d.pkgData.fset, v.Decl, opt.Source, file),
				Const:     false,
				Values:    d.extractValueSpecs(v.Decl.Specs, printer),
				Doc:       normalizedMarkdownPrinter(printer, doc),
			})
		}

		// constructors for this type
		for _, fun := range typ.Funcs {
			params := d.extractJSONFields(fun.Decl.Type.Params)
			jsonDoc.Funcs = append(jsonDoc.Funcs, &JSONFunc{
				Name:      fun.Name,
				Crossing:  isCrossing(params),
				Signature: mustFormatNode(d.pkgData.fset, fun.Decl, opt.Source, file),
				Doc:       normalizedMarkdownPrinter(printer, fun.Doc),
				Params:    params,
				Results:   d.extractJSONFields(fun.Decl.Type.Results),
			})
		}

		for _, meth := range typ.Methods {
			params := d.extractJSONFields(meth.Decl.Type.Params)
			jsonDoc.Funcs = append(jsonDoc.Funcs, &JSONFunc{
				Type:      typ.Name,
				Name:      meth.Name,
				Crossing:  isCrossing(params),
				Signature: mustFormatNode(d.pkgData.fset, meth.Decl, opt.Source, file),
				Doc:       normalizedMarkdownPrinter(printer, meth.Doc),
				Params:    params,
				Results:   d.extractJSONFields(meth.Decl.Type.Results),
			})
		}
	}

	return jsonDoc, nil
}

// createCustomPrinter creates a printer that doesn't generate heading IDs
// and handles backslash escaping properly
func createCustomPrinter(pkg *doc.Package) *comment.Printer {
	printer := pkg.Printer()
	printer.HeadingID = func(h *comment.Heading) string {
		return "" // Return empty string to omit heading IDs
	}
	return printer
}

// normalizedMarkdownPrinter converts a doc comment to markdown without double backslashes
// and converts indented code blocks to fenced code blocks for Chroma syntax highlighting
func normalizedMarkdownPrinter(printer *comment.Printer, doc interface{}) string {
	if doc == "" {
		return ""
	}

	var md string
	switch d := doc.(type) {
	case string:
		if d == "" {
			return ""
		}
		var p comment.Parser
		parsedDoc := p.Parse(d)
		md = string(printer.Markdown(parsedDoc))
	case *comment.Doc:
		md = string(printer.Markdown(d))
	default:
		return ""
	}

	md = convertIndentedCodeBlocksToFenced(md)

	return md
}

// convertIndentedCodeBlocksToFenced converts 4-space indented code blocks to fenced code blocks
// This is needed because Chroma only works with fenced code blocks, not indented ones
func convertIndentedCodeBlocksToFenced(markdown string) string {
	var buf bytes.Buffer
	reader := strings.NewReader(markdown)

	if err := normalizeCodeBlockStream(reader, &buf); err != nil {
		// If conversion fails, return original markdown
		return markdown
	}

	return buf.String()
}

// normalizeCodeBlockStream converts indented code blocks to fenced code blocks using streams
func normalizeCodeBlockStream(r io.Reader, w io.Writer) error {
	scanner := bufio.NewScanner(r)
	writer := bufio.NewWriter(w)
	defer writer.Flush()

	inCode := false
	write := func(s string) error { _, err := writer.WriteString(s + "\n"); return err }

	for scanner.Scan() {
		line := scanner.Text()
		isCode := strings.HasPrefix(line, "\t") || (len(line) >= 4 && line[:4] == "    ")

		if isCode && !inCode {
			if err := write("```go"); err != nil {
				return err
			}
			inCode = true
		}
		if !isCode && inCode {
			if err := write("```"); err != nil {
				return err
			}
			inCode = false
		}

		if isCode {
			if strings.HasPrefix(line, "\t") {
				line = line[1:]
			} else {
				line = line[4:]
			}
		}
		if err := write(line); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("lecture failed: %w", err)
	}
	if inCode {
		if err := write("```"); err != nil {
			return err
		}
	}
	return nil
}

func (d *Documentable) extractJSONFields(fieldList *ast.FieldList) []*JSONField {
	results := []*JSONField{}
	if fieldList != nil {
		for _, field := range fieldList.List {
			commentBuf := new(strings.Builder)
			if field.Doc != nil {
				for _, comment := range field.Doc.List {
					commentBuf.WriteString(comment.Text)
					commentBuf.WriteString("\n")
				}
			}
			if field.Comment != nil {
				for _, comment := range field.Comment.List {
					commentBuf.WriteString(comment.Text)
					commentBuf.WriteString("\n")
				}
			}

			if len(field.Names) == 0 {
				// if there are no names, then the field is unnamed, but still has a type
				f := &JSONField{
					Name: "",
					Type: mustFormatNode(d.pkgData.fset, field.Type, false, nil),
					Doc:  commentBuf.String(),
				}
				results = append(results, f)
			} else {
				// fields can be of the format: (a, b int, c string)
				// so we need to iterate over the names
				for _, name := range field.Names {
					f := &JSONField{
						Name: name.Name,
						Type: mustFormatNode(d.pkgData.fset, field.Type, false, nil),
						Doc:  commentBuf.String(),
					}
					results = append(results, f)
				}
			}
		}
	}
	return results
}

func (d *Documentable) extractValueSpecs(specs []ast.Spec, printer *comment.Printer) []*JSONValue {
	values := []*JSONValue{}

	for _, value := range specs {
		constSpec := value.(*ast.ValueSpec)

		typeString := ""
		if constSpec.Type != nil {
			typeString = mustFormatNode(d.pkgData.fset, constSpec.Type, false, nil)
		}

		commentBuf := new(strings.Builder)
		if constSpec.Comment != nil {
			for _, comment := range constSpec.Comment.List {
				commentBuf.WriteString(comment.Text)
			}
		}

		// Const declaration can be of the form: const a, b, c = 1, 2, 3
		// so we need to iterate over the names
		for _, name := range constSpec.Names {
			jsonValue := &JSONValue{
				Name: name.Name,
				Type: typeString,
				Doc:  normalizedMarkdownPrinter(printer, commentBuf.String()),
			}
			values = append(values, jsonValue)
		}
	}
	return values
}

// mustFormatNode calls format.Node and returns the result as a string.
// Panic on error, which shouldn't happen since the node is a valid AST from pkgData.parseFile.
// If source is true and the optional ast.File is given, then use it to get internal comments.
func mustFormatNode(fset *token.FileSet, node any, source bool, file *ast.File) string {
	if !source {
		// Omit the Doc and Body so that it's not in the signature
		switch n := node.(type) {
		case *ast.FuncDecl:
			node = &ast.FuncDecl{
				Recv: n.Recv,
				Name: n.Name,
				Type: n.Type,
			}
		case *ast.GenDecl:
			node = &ast.GenDecl{
				TokPos: n.TokPos,
				Tok:    n.Tok,
				Lparen: n.Lparen,
				Specs:  n.Specs,
				Rparen: n.Rparen,
			}
		}
	}

	if file != nil && source {
		// Need an extra little dance to get internal comments to appear.
		node = &printer.CommentedNode{
			Node:     node,
			Comments: file.Comments,
		}
	}

	buf := new(strings.Builder)
	if err := format.Node(buf, fset, node); err != nil {
		panic("Error in format.Node: " + err.Error())
	}
	return buf.String()
}

// isCrossing returns true if the first param has type "realm"
func isCrossing(params []*JSONField) bool {
	if len(params) < 1 {
		return false
	}

	return params[0].Type == "realm"
}

// Search typ for the ast.TypeSpec with the same name and return it, or nil if not found
func getTypeSpec(typ *doc.Type) *ast.TypeSpec {
	for _, spec := range typ.Decl.Specs {
		tSpec := spec.(*ast.TypeSpec) // Must succeed
		if typ.Name == tSpec.Name.Name {
			return tSpec
		}
	}

	return nil
}

// Return the expression inside the parentheses, if any
func deparenthesize(expr ast.Expr) ast.Expr {
	x := expr
	for {
		if t, ok := x.(*ast.ParenExpr); ok {
			x = t.X
		} else {
			break
		}
	}
	return x
}

func (jsonDoc *JSONDocumentation) JSON() string {
	bz := amino.MustMarshalJSON(jsonDoc)
	return string(bz)
}
