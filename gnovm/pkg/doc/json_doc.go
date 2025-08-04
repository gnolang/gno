package doc

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/doc"
	"go/doc/comment"
	"go/format"
	"go/token"
	"io"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// JSONDocumentation holds package documentation suitable for transmitting
// as JSON with printable string fields
type JSONDocumentation struct {
	PackagePath string `json:"package_path"`
	PackageLine string `json:"package_line"` // package io // import "io"
	PackageDoc  string `json:"package_doc"`  // markdown of top-level package documentation
	// https://pkg.go.dev/go/doc#Package.Markdown to render markdown

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
}

type JSONFunc struct {
	Type      string       `json:"type"` // if this is a method
	Name      string       `json:"name"`
	Signature string       `json:"signature"`
	Doc       string       `json:"doc"` // markdown
	Params    []*JSONField `json:"params"`
	Results   []*JSONField `json:"results"`
}

type JSONType struct {
	Name      string `json:"name"`
	Signature string `json:"signature"`
	Doc       string `json:"doc"` // markdown
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
func (d *Documentable) WriteJSONDocumentation() (*JSONDocumentation, error) {
	opt := &WriteDocumentationOptions{}
	_, pkg, err := d.pkgData.docPackage(opt)
	if err != nil {
		return nil, err
	}

	// Create custom printer that doesn't generate heading IDs
	printer := createCustomPrinter(pkg)

	// Parse package documentation
	var pkgDoc string
	if pkg.Doc != "" {
		var p comment.Parser
		doc := p.Parse(pkg.Doc)
		pkgDoc = normalizedMarkdownPrinter(printer, doc)
	}

	jsonDoc := &JSONDocumentation{
		PackagePath: d.pkgData.dir.dir,
		PackageLine: fmt.Sprintf("package %s // import %q", pkg.Name, pkg.ImportPath),
		PackageDoc:  pkgDoc,
		Values:      []*JSONValueDecl{},
		Funcs:       []*JSONFunc{},
		Types:       []*JSONType{},
	}

	for _, value := range pkg.Consts {
		var p comment.Parser
		doc := p.Parse(value.Doc)
		jsonDoc.Values = append(jsonDoc.Values, &JSONValueDecl{
			Signature: mustFormatNode(d.pkgData.fset, value.Decl),
			Const:     true,
			Values:    d.extractValueSpecs(value.Decl.Specs, printer),
			Doc:       normalizedMarkdownPrinter(printer, doc),
		})
	}

	for _, value := range pkg.Vars {
		var p comment.Parser
		doc := p.Parse(value.Doc)
		jsonDoc.Values = append(jsonDoc.Values, &JSONValueDecl{
			Signature: mustFormatNode(d.pkgData.fset, value.Decl),
			Const:     false,
			Values:    d.extractValueSpecs(value.Decl.Specs, printer),
			Doc:       normalizedMarkdownPrinter(printer, doc),
		})
	}

	for _, fun := range pkg.Funcs {
		var p comment.Parser
		doc := p.Parse(fun.Doc)
		jsonDoc.Funcs = append(jsonDoc.Funcs, &JSONFunc{
			Name:      fun.Name,
			Signature: mustFormatNode(d.pkgData.fset, fun.Decl),
			Doc:       normalizedMarkdownPrinter(printer, doc),
			Params:    d.extractJSONFields(fun.Decl.Type.Params),
			Results:   d.extractJSONFields(fun.Decl.Type.Results),
		})
	}

	for _, typ := range pkg.Types {
		var p comment.Parser
		doc := p.Parse(typ.Doc)
		jsonDoc.Types = append(jsonDoc.Types, &JSONType{
			Name:      typ.Name,
			Signature: mustFormatNode(d.pkgData.fset, typ.Decl),
			Doc:       normalizedMarkdownPrinter(printer, doc),
		})

		// values of this type
		for _, c := range typ.Consts {
			var p comment.Parser
			doc := p.Parse(c.Doc)
			jsonDoc.Values = append(jsonDoc.Values, &JSONValueDecl{
				Signature: mustFormatNode(d.pkgData.fset, c.Decl),
				Const:     true,
				Values:    d.extractValueSpecs(c.Decl.Specs, printer),
				Doc:       normalizedMarkdownPrinter(printer, doc),
			})
		}
		for _, v := range typ.Vars {
			var p comment.Parser
			doc := p.Parse(v.Doc)
			jsonDoc.Values = append(jsonDoc.Values, &JSONValueDecl{
				Signature: mustFormatNode(d.pkgData.fset, v.Decl),
				Const:     false,
				Values:    d.extractValueSpecs(v.Decl.Specs, printer),
				Doc:       normalizedMarkdownPrinter(printer, doc),
			})
		}

		// constructors for this type
		for _, fun := range typ.Funcs {
			var p comment.Parser
			doc := p.Parse(fun.Doc)
			jsonDoc.Funcs = append(jsonDoc.Funcs, &JSONFunc{
				Name:      fun.Name,
				Signature: mustFormatNode(d.pkgData.fset, fun.Decl),
				Doc:       normalizedMarkdownPrinter(printer, doc),
				Params:    d.extractJSONFields(fun.Decl.Type.Params),
				Results:   d.extractJSONFields(fun.Decl.Type.Results),
			})
		}

		for _, meth := range typ.Methods {
			var p comment.Parser
			doc := p.Parse(meth.Doc)
			jsonDoc.Funcs = append(jsonDoc.Funcs, &JSONFunc{
				Type:      typ.Name,
				Name:      meth.Name,
				Signature: mustFormatNode(d.pkgData.fset, meth.Decl),
				Doc:       normalizedMarkdownPrinter(printer, doc),
				Params:    d.extractJSONFields(meth.Decl.Type.Params),
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
func normalizedMarkdownPrinter(printer *comment.Printer, doc *comment.Doc) string {
	md := string(printer.Markdown(doc))
	md = convertIndentedCodeBlocksToFenced(md)
	md = strings.ReplaceAll(md, `\\`, `\`)

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
			if len(field.Names) == 0 {
				// if there are no names, then the field is unnamed, but still has a type
				f := &JSONField{
					Name: "",
					Type: mustFormatNode(d.pkgData.fset, field.Type),
				}
				results = append(results, f)
			} else {
				// fields can be of the format: (a, b int, c string)
				// so we need to iterate over the names
				for _, name := range field.Names {
					f := &JSONField{
						Name: name.Name,
						Type: mustFormatNode(d.pkgData.fset, field.Type),
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
			typeString = mustFormatNode(d.pkgData.fset, constSpec.Type)
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
			// Parse the comment and use our custom printer
			var p comment.Parser
			doc := p.Parse(commentBuf.String())
			jsonValue := &JSONValue{
				Name: name.Name,
				Type: typeString,
				Doc:  normalizedMarkdownPrinter(printer, doc),
			}
			values = append(values, jsonValue)
		}
	}
	return values
}

// mustFormatNode calls format.Node and returns the result as a string.
// Panic on error, which shouldn't happen since the node is a valid AST from pkgData.parseFile.
func mustFormatNode(fset *token.FileSet, node any) string {
	buf := new(strings.Builder)
	if err := format.Node(buf, fset, node); err != nil {
		panic("Error in format.Node: " + err.Error())
	}
	return buf.String()
}

func (jsonDoc *JSONDocumentation) JSON() string {
	bz := amino.MustMarshalJSON(jsonDoc)
	return string(bz)
}
