package doc

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/format"
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
		buf := new(strings.Builder)
		if err := format.Node(buf, d.pkgData.fset, value.Decl); err != nil {
			return nil, err
		}

		values, err := d.extractValueSpecs(pkg, value.Decl.Specs)
		if err != nil {
			return nil, err
		}

		jsonDoc.Values = append(jsonDoc.Values, &JSONValueDecl{
			Signature: buf.String(),
			Const:     true,
			Values:    values,
			Doc:       string(pkg.Markdown(value.Doc)),
		})
	}

	for _, value := range pkg.Vars {
		buf := new(strings.Builder)
		if err := format.Node(buf, d.pkgData.fset, value.Decl); err != nil {
			return nil, err
		}

		values, err := d.extractValueSpecs(pkg, value.Decl.Specs)
		if err != nil {
			return nil, err
		}

		jsonDoc.Values = append(jsonDoc.Values, &JSONValueDecl{
			Signature: buf.String(),
			Const:     false,
			Values:    values,
			Doc:       string(pkg.Markdown(value.Doc)),
		})
	}

	for _, fun := range pkg.Funcs {
		buf := new(strings.Builder)
		if err := format.Node(buf, d.pkgData.fset, fun.Decl); err != nil {
			return nil, err
		}

		params, err := d.extractFuncParams(fun)
		if err != nil {
			return nil, err
		}

		results, err := d.extractFuncResults(fun)
		if err != nil {
			return nil, err
		}

		jsonDoc.Funcs = append(jsonDoc.Funcs, &JSONFunc{
			Name:      fun.Name,
			Signature: buf.String(),
			Doc:       string(pkg.Markdown(fun.Doc)),
			Params:    params,
			Results:   results,
		})
	}

	for _, typ := range pkg.Types {
		buf := new(strings.Builder)
		if err := format.Node(buf, d.pkgData.fset, typ.Decl); err != nil {
			return nil, err
		}
		jsonDoc.Types = append(jsonDoc.Types, &JSONType{
			Name:      typ.Name,
			Signature: buf.String(),
			Doc:       string(pkg.Markdown(typ.Doc)),
		})

		// constructors for this type
		for _, fun := range typ.Funcs {
			buf := new(strings.Builder)
			if err := format.Node(buf, d.pkgData.fset, fun.Decl); err != nil {
				return nil, err
			}

			params, err := d.extractFuncParams(fun)
			if err != nil {
				return nil, err
			}

			results, err := d.extractFuncResults(fun)
			if err != nil {
				return nil, err
			}

			jsonDoc.Funcs = append(jsonDoc.Funcs, &JSONFunc{
				Name:      fun.Name,
				Signature: buf.String(),
				Doc:       string(pkg.Markdown(fun.Doc)),
				Params:    params,
				Results:   results,
			})
		}

		for _, meth := range typ.Methods {
			buf := new(strings.Builder)
			if err := format.Node(buf, d.pkgData.fset, meth.Decl); err != nil {
				return nil, err
			}

			params, err := d.extractFuncParams(meth)
			if err != nil {
				return nil, err
			}

			results, err := d.extractFuncResults(meth)
			if err != nil {
				return nil, err
			}

			jsonDoc.Funcs = append(jsonDoc.Funcs, &JSONFunc{
				Type:      typ.Name,
				Name:      meth.Name,
				Signature: buf.String(),
				Doc:       string(pkg.Markdown(meth.Doc)),
				Params:    params,
				Results:   results,
			})
		}
	}

	return jsonDoc, nil
}

func (d *Documentable) extractFuncParams(fun *doc.Func) ([]*JSONField, error) {
	params := []*JSONField{}
	for _, param := range fun.Decl.Type.Params.List {
		buf := new(strings.Builder)
		if err := format.Node(buf, d.pkgData.fset, param.Type); err != nil {
			return nil, err
		}

		// parameters can be of the format: (a, b int, c string)
		// so we need to iterate over the names
		for _, name := range param.Names {
			field := &JSONField{
				Name: name.Name,
				Type: buf.String(),
			}

			params = append(params, field)
		}
	}

	return params, nil
}

func (d *Documentable) extractFuncResults(fun *doc.Func) ([]*JSONField, error) {
	results := []*JSONField{}
	if fun.Decl.Type.Results != nil {
		for _, result := range fun.Decl.Type.Results.List {
			buf := new(strings.Builder)
			if err := format.Node(buf, d.pkgData.fset, result.Type); err != nil {
				return nil, err
			}

			// results can be of the format: (a, b int, c string)
			// so we need to iterate over the names
			for _, name := range result.Names {
				result := &JSONField{
					Name: name.Name,
					Type: buf.String(),
				}

				results = append(results, result)
			}

			// if there are no names, then the result is an unnamed return
			if len(result.Names) == 0 {
				result := &JSONField{
					Name: "",
					Type: buf.String(),
				}
				results = append(results, result)
			}
		}
	}
	return results, nil
}

func (d *Documentable) extractValueSpecs(pkg *doc.Package, specs []ast.Spec) ([]*JSONValue, error) {
	values := []*JSONValue{}

	for _, value := range specs {
		constSpec := value.(*ast.ValueSpec)
		buf := new(strings.Builder)

		if constSpec.Type != nil {
			if err := format.Node(buf, d.pkgData.fset, constSpec.Type); err != nil {
				return nil, err
			}
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
				Type: buf.String(),
				Doc:  string(pkg.Markdown(commentBuf.String())),
			}
			values = append(values, jsonValue)
		}
	}
	return values, nil
}

func (jsonDoc *JSONDocumentation) JSON() string {
	bz := amino.MustMarshalJSON(jsonDoc)
	return string(bz)
}
