package doc

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/format"
	"go/token"
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
	Imports []string        `json:"imports"` // list of import paths
}

type JSONValueDecl struct {
	Signature string       `json:"signature"`
	Const     bool         `json:"const"`
	Values    []*JSONValue `json:"values"`
	Doc       string       `json:"doc"` // markdown
	File      string       `json:"file"`
	Line      int          `json:"line"`
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
	File      string       `json:"file"`
	Line      int          `json:"line"`
}

type JSONType struct {
	Name      string `json:"name"`
	Signature string `json:"signature"`
	Doc       string `json:"doc"` // markdown
	File      string `json:"file"`
	Line      int    `json:"line"`
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

// Helper functions to reduce code duplication

// createJSONValueDecl creates a JSONValueDecl from a doc.Value
func (d *Documentable) createJSONValueDecl(pkg *doc.Package, value *doc.Value, isConst bool) *JSONValueDecl {
	file, line := d.pkgData.extractPosition(value.Decl)
	return &JSONValueDecl{
		Signature: mustFormatNode(d.pkgData.fset, value.Decl),
		Const:     isConst,
		Values:    d.extractValueSpecs(pkg, value.Decl.Specs),
		Doc:       string(pkg.Markdown(value.Doc)),
		File:      file,
		Line:      line,
	}
}

// createJSONFunc creates a JSONFunc from a doc.Func
func (d *Documentable) createJSONFunc(pkg *doc.Package, fun *doc.Func, typeName string) *JSONFunc {
	file, line := d.pkgData.extractPosition(fun.Decl)
	return &JSONFunc{
		Type:      typeName,
		Name:      fun.Name,
		Signature: mustFormatNode(d.pkgData.fset, fun.Decl),
		Doc:       string(pkg.Markdown(fun.Doc)),
		Params:    d.extractJSONFields(fun.Decl.Type.Params),
		Results:   d.extractJSONFields(fun.Decl.Type.Results),
		File:      file,
		Line:      line,
	}
}

// createJSONType creates a JSONType from a doc.Type
func (d *Documentable) createJSONType(pkg *doc.Package, typ *doc.Type) *JSONType {
	file, line := d.pkgData.extractPosition(typ.Decl)
	return &JSONType{
		Name:      typ.Name,
		Signature: mustFormatNode(d.pkgData.fset, typ.Decl),
		Doc:       string(pkg.Markdown(typ.Doc)),
		File:      file,
		Line:      line,
	}
}

// WriteJSONDocumentation returns a JSONDocumentation for the package
func (d *Documentable) WriteJSONDocumentation() (*JSONDocumentation, error) {
	opt := &WriteDocumentationOptions{}
	_, pkg, err := d.pkgData.docPackage(opt)
	if err != nil {
		return nil, err
	}

	// Extract imports from the package files
	imports := d.pkgData.extractImports()

	jsonDoc := &JSONDocumentation{
		PackagePath: d.pkgData.dir.dir,
		PackageLine: fmt.Sprintf("package %s // import %q", pkg.Name, pkg.ImportPath),
		PackageDoc:  string(pkg.Markdown(pkg.Doc)),
		Values:      []*JSONValueDecl{},
		Funcs:       []*JSONFunc{},
		Types:       []*JSONType{},
		Imports:     imports,
	}

	for _, value := range pkg.Consts {
		jsonDoc.Values = append(jsonDoc.Values, d.createJSONValueDecl(pkg, value, true))
	}

	for _, value := range pkg.Vars {
		jsonDoc.Values = append(jsonDoc.Values, d.createJSONValueDecl(pkg, value, false))
	}

	for _, fun := range pkg.Funcs {
		jsonDoc.Funcs = append(jsonDoc.Funcs, d.createJSONFunc(pkg, fun, ""))
	}

	for _, typ := range pkg.Types {
		jsonDoc.Types = append(jsonDoc.Types, d.createJSONType(pkg, typ))

		// values of this type
		for _, c := range typ.Consts {
			jsonDoc.Values = append(jsonDoc.Values, d.createJSONValueDecl(pkg, c, true))
		}
		for _, v := range typ.Vars {
			jsonDoc.Values = append(jsonDoc.Values, d.createJSONValueDecl(pkg, v, false))
		}

		// constructors for this type
		for _, fun := range typ.Funcs {
			jsonDoc.Funcs = append(jsonDoc.Funcs, d.createJSONFunc(pkg, fun, typ.Name))
		}

		for _, meth := range typ.Methods {
			jsonDoc.Funcs = append(jsonDoc.Funcs, d.createJSONFunc(pkg, meth, typ.Name))
		}
	}

	return jsonDoc, nil
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

func (d *Documentable) extractValueSpecs(pkg *doc.Package, specs []ast.Spec) []*JSONValue {
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
			jsonValue := &JSONValue{
				Name: name.Name,
				Type: typeString,
				Doc:  string(pkg.Markdown(commentBuf.String())),
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
