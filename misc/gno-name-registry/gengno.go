// This file generates a Gno file that can be used to populate a list of reserved names.
//
// It uses Handshake's protocol registered names "lockup.json" file to initialize the names.
// Registered names has been curated and extracted from TLDs and Alexa's top 100k domain names.
//
// Lockup file can be found here:
// https://github.com/handshake-org/hs-names-2023/blob/3482d12e9c680030f1cec729f5e3a7aa454d0f15/build/updated/lockup.json
package main

import (
	_ "embed"
	"encoding/json"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
)

//go:embed source.gno
var source string

// Creates a new AST node for a name entry.
func newEntryNode(name string, typ int) *ast.CompositeLit {
	return &ast.CompositeLit{Elts: []ast.Expr{
		&ast.KeyValueExpr{
			Key: ast.NewIdent("Name"),
			Value: &ast.BasicLit{
				Kind:  token.STRING,
				Value: strconv.Quote(name),
			},
		},
		&ast.KeyValueExpr{
			Key: ast.NewIdent("Type"),
			Value: &ast.BasicLit{
				Kind:  token.INT,
				Value: strconv.Itoa(typ),
			},
		},
	}}
}

func main() {
	// Open Handshake's protocol list of reserved names
	data, err := os.Open("lockup.json")
	if err != nil {
		fatal(err)
	}

	defer data.Close()

	// Read list of reserved names
	var entries map[string][]any
	err = json.NewDecoder(data).Decode(&entries)
	if err != nil {
		fatal(err)
	}

	// Parse Gno source into a tree
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", source, parser.ParseComments)
	if err != nil {
		fatal(err)
	}

	decl := node.Decls[1].(*ast.GenDecl)
	spec := decl.Specs[0].(*ast.ValueSpec)
	array := spec.Values[0].(*ast.CompositeLit)

	// For each reserved name create an entry item
	for _, v := range entries {
		if len(v) != 2 {
			// Skip invalid rows (usually the first one)
			continue
		}

		name := v[0].(string)
		nameType := int(v[1].(float64))
		array.Elts = append(array.Elts, newEntryNode(name, nameType))
	}

	// Generate a Gno file with all the entries initialized
	file, err := os.Create("main.gen.gno")
	if err != nil {
		fatal(err)
	}

	defer file.Close()

	var buf strings.Builder
	err = format.Node(&buf, fset, node)
	if err != nil {
		fatal(err)
	}

	// Write the Gno file making sure sure each entry is defined in a single line
	_, err = strings.NewReplacer("{Name", "\n\t{Name").WriteString(file, buf.String())
	if err != nil {
		fatal(err)
	}

	println("generated file " + file.Name())
}

func fatal(err error) {
	println(err.Error())
	os.Exit(1)
}
