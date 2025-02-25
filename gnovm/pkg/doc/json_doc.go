package doc

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/format"
	"go/token"
	"strings"

	"github.com/gnolang/gno/gnovm"
	"github.com/gnolang/gno/tm2/pkg/amino"
)

// JSONDocumentation holds package documentation suitable for transmitting
// as JSON with printable string fields
type JSONDocumentation struct {
	PackagePath string `json:"package_path"`
	PackageLine string `json:"package_line"` // package io // import "io"
	PackageDoc  string `json:"package_doc"`  // markdown of top-level package documentation
	// https://pkg.go.dev/go/doc#Package.Markdown to render markdown

	// These match each of the sections in a pkg.go.dev package documentationj
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
	Name string
	Type string
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

// NewDocumentableFromMemPkg gets the pkgData from memPkg and returns a Documentable
func NewDocumentableFromMemPkg(memPkg *gnovm.MemPackage, unexported bool, symbol, accessible string) (*Documentable, error) {
	pd, err := newPkgDataFromMemPkg(memPkg, unexported)
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

	jsonDoc := &JSONDocumentation{
		PackagePath: d.pkgData.dir.dir,
		PackageLine: fmt.Sprintf("package %s // import %q", pkg.Name, pkg.ImportPath),
		PackageDoc:  string(pkg.Markdown(pkg.Doc)),
		Values:      []*JSONValueDecl{},
		Funcs:       []*JSONFunc{},
		Types:       []*JSONType{},
	}

	for _, value := range pkg.Consts {
		jsonDoc.Values = append(jsonDoc.Values, &JSONValueDecl{
			Signature: mustFormatNode(d.pkgData.fset, value.Decl),
			Const:     true,
			Values:    d.extractValueSpecs(pkg, value.Decl.Specs),
			Doc:       string(pkg.Markdown(value.Doc)),
		})
	}

	for _, value := range pkg.Vars {
		jsonDoc.Values = append(jsonDoc.Values, &JSONValueDecl{
			Signature: mustFormatNode(d.pkgData.fset, value.Decl),
			Const:     false,
			Values:    d.extractValueSpecs(pkg, value.Decl.Specs),
			Doc:       string(pkg.Markdown(value.Doc)),
		})
	}

	for _, fun := range pkg.Funcs {
		jsonDoc.Funcs = append(jsonDoc.Funcs, &JSONFunc{
			Name:      fun.Name,
			Signature: mustFormatNode(d.pkgData.fset, fun.Decl),
			Doc:       string(pkg.Markdown(fun.Doc)),
			Params:    d.extractFuncParams(fun),
			Results:   d.extractFuncResults(fun),
		})
	}

	for _, typ := range pkg.Types {
		jsonDoc.Types = append(jsonDoc.Types, &JSONType{
			Name:      typ.Name,
			Signature: mustFormatNode(d.pkgData.fset, typ.Decl),
			Doc:       string(pkg.Markdown(typ.Doc)),
		})

		// constructors for this type
		for _, fun := range typ.Funcs {
			jsonDoc.Funcs = append(jsonDoc.Funcs, &JSONFunc{
				Name:      fun.Name,
				Signature: mustFormatNode(d.pkgData.fset, fun.Decl),
				Doc:       string(pkg.Markdown(fun.Doc)),
				Params:    d.extractFuncParams(fun),
				Results:   d.extractFuncResults(fun),
			})
		}

		for _, meth := range typ.Methods {
			jsonDoc.Funcs = append(jsonDoc.Funcs, &JSONFunc{
				Type:      typ.Name,
				Name:      meth.Name,
				Signature: mustFormatNode(d.pkgData.fset, meth.Decl),
				Doc:       string(pkg.Markdown(meth.Doc)),
				Params:    d.extractFuncParams(meth),
				Results:   d.extractFuncResults(meth),
			})
		}
	}

	return jsonDoc, nil
}

func (d *Documentable) extractFuncParams(fun *doc.Func) []*JSONField {
	params := []*JSONField{}
	for _, param := range fun.Decl.Type.Params.List {
		// parameters can be of the format: (a, b int, c string)
		// so we need to iterate over the names
		for _, name := range param.Names {
			field := &JSONField{
				Name: name.Name,
				Type: mustFormatNode(d.pkgData.fset, param.Type),
			}

			params = append(params, field)
		}
	}

	return params
}

func (d *Documentable) extractFuncResults(fun *doc.Func) []*JSONField {
	results := []*JSONField{}
	if fun.Decl.Type.Results != nil {
		for _, result := range fun.Decl.Type.Results.List {
			if len(result.Names) == 0 {
				// if there are no names, then the result is an unnamed return
				result := &JSONField{
					Name: "",
					Type: mustFormatNode(d.pkgData.fset, result.Type),
				}
				results = append(results, result)
			} else {
				// results can be of the format: (a, b int, c string)
				// so we need to iterate over the names
				for _, name := range result.Names {
					result := &JSONField{
						Name: name.Name,
						Type: mustFormatNode(d.pkgData.fset, result.Type),
					}
					results = append(results, result)
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
