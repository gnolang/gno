package doc

import (
	"fmt"
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
	Values []*JSONValue `json:"values"` // constants and variables declared
	Funcs  []*JSONFunc  `json:"funcs"`  // Funcs and methods
	Types  []*JSONType  `json:"types"`
}

type JSONValue struct {
	Signature string `json:"signature"`
	Doc       string `json:"doc"` // markdown
}

type JSONFunc struct {
	Type      string `json:"type"` // if this is a method
	Name      string `json:"name"`
	Signature string `json:"signature"`
	Doc       string `json:"doc"` // markdown
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
		Values:      []*JSONValue{},
		Funcs:       []*JSONFunc{},
		Types:       []*JSONType{},
	}

	for _, value := range pkg.Consts {
		buf := new(strings.Builder)
		if err := format.Node(buf, d.pkgData.fset, value.Decl); err != nil {
			return nil, err
		}
		jsonDoc.Values = append(jsonDoc.Values, &JSONValue{
			Signature: buf.String(),
			Doc:       string(pkg.Markdown(value.Doc)),
		})
	}

	for _, value := range pkg.Vars {
		buf := new(strings.Builder)
		if err := format.Node(buf, d.pkgData.fset, value.Decl); err != nil {
			return nil, err
		}
		jsonDoc.Values = append(jsonDoc.Values, &JSONValue{
			Signature: buf.String(),
			Doc:       string(pkg.Markdown(value.Doc)),
		})
	}

	for _, fun := range pkg.Funcs {
		buf := new(strings.Builder)
		if err := format.Node(buf, d.pkgData.fset, fun.Decl); err != nil {
			return nil, err
		}
		jsonDoc.Funcs = append(jsonDoc.Funcs, &JSONFunc{
			Name:      fun.Name,
			Signature: buf.String(),
			Doc:       string(pkg.Markdown(fun.Doc)),
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

		for _, meth := range typ.Methods {
			buf := new(strings.Builder)
			if err := format.Node(buf, d.pkgData.fset, meth.Decl); err != nil {
				return nil, err
			}
			jsonDoc.Funcs = append(jsonDoc.Funcs, &JSONFunc{
				Type:      typ.Name,
				Name:      meth.Name,
				Signature: buf.String(),
				Doc:       string(pkg.Markdown(meth.Doc)),
			})
		}
	}

	return jsonDoc, nil
}

func (jsonDoc *JSONDocumentation) JSON() string {
	bz := amino.MustMarshalJSON(jsonDoc)
	return string(bz)
}
