package doctest

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"sort"
	"strconv"
	"strings"
)

// supported stdlib packages in gno.
// ref: go-gno-compatibility.md
var stdLibPackages = map[string]bool{
	"bufio":           true,
	"builtin":         true,
	"bytes":           true,
	"encoding":        true,
	"encoding/base64": true,
	"encoding/hex":    true,
	"hash":            true,
	"hash/adler32":    true,
	"io":              true,
	"math":            true,
	"math/bits":       true,
	"net/url":         true,
	"path":            true,
	"regexp":          true,
	"regexp/syntax":   true,
	"std":             true,
	"strings":         true,
	"time":            true,
	"unicode":         true,
	"unicode/utf16":   true,
	"unicode/utf8":    true,

	// partially supported packages
	"crypto/cipher":   true,
	"crypto/ed25519":  true,
	"crypto/sha256":   true,
	"encoding/binary": true,
	"errors":          true,
	"sort":            true,
	"strconv":         true,
	"testing":         true,
}

// analyzeAndModifyCode analyzes the given code block, adds package declaration if missing,
// ensures a main function exists, and updates imports. It returns the modified code as a string.
func analyzeAndModifyCode(code string) (string, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", code, parser.AllErrors)
	if err != nil {
		// Prepend package main to the code and try to parse again.
		node, err = parser.ParseFile(fset, "", "package main\n"+code, parser.ParseComments)
		if err != nil {
			return "", fmt.Errorf("failed to parse code: %w", err)
		}
	}

	if node.Name == nil {
		node.Name = ast.NewIdent("main")
	}

	if !hasMainFunction(node) {
		return "", fmt.Errorf("main function is missing")
	}

	updateImports(node)

	src, err := codePrettier(fset, node)
	if err != nil {
		return "", err
	}

	return src, nil
}

// hasMainFunction checks if a main function exists in the AST.
// It returns an error if the main function is missing.
func hasMainFunction(node *ast.File) bool {
	for _, decl := range node.Decls {
		if fn, isFn := decl.(*ast.FuncDecl); isFn && fn.Name.Name == "main" {
			return true
		}
	}
	return false
}

// detectUsedPackages inspects the AST and returns a map of used stdlib packages.
func detectUsedPackages(node *ast.File) map[string]bool {
	usedPackages := make(map[string]bool)
	remainingPackages := make(map[string]bool)
	for pkg := range stdLibPackages {
		remainingPackages[pkg] = true
	}

	ast.Inspect(node, func(n ast.Node) bool {
		if len(remainingPackages) == 0 {
			return false
		}

		selectorExpr, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		ident, ok := selectorExpr.X.(*ast.Ident)
		if !ok {
			return true
		}

		if remainingPackages[ident.Name] {
			usedPackages[ident.Name] = true
			delete(remainingPackages, ident.Name)
			return false
		}

		for fullPkg := range stdLibPackages {
			if isMatchingSubpackage(fullPkg, ident.Name, selectorExpr.Sel.Name) {
				usedPackages[fullPkg] = true
				delete(remainingPackages, fullPkg)
				return false
			}
		}
		return true
	})
	return usedPackages
}

func isMatchingSubpackage(fullPkg, prefix, suffix string) bool {
	if !strings.HasPrefix(fullPkg, prefix+"/") {
		return false
	}
	parts := strings.SplitN(fullPkg, "/", 2)
	return len(parts) == 2 && parts[1] == suffix
}

// updateImports modifies the AST to include all necessary import statements.
// based on the packages used in the code and existing imports.
func updateImports(node *ast.File) {
	usedPackages := detectUsedPackages(node)

	// Remove existing imports
	node.Decls = removeImportDecls(node.Decls)

	// Add new imports only for used packages
	if len(usedPackages) > 0 {
		importSpecs := createImportSpecs(usedPackages)
		importDecl := &ast.GenDecl{
			Tok:    token.IMPORT,
			Lparen: token.Pos(1),
			Specs:  importSpecs,
		}
		node.Decls = append([]ast.Decl{importDecl}, node.Decls...)
	}
}

// createImportSpecs generates a slice of import specifications from a map of importable package paths.
// It sorts the paths alphabetically before creating the import specs.
func createImportSpecs(imports map[string]bool) []ast.Spec {
	paths := make([]string, 0, len(imports))
	for path := range imports {
		paths = append(paths, path)
	}

	sort.Strings(paths)

	specs := make([]ast.Spec, 0, len(imports))
	for _, path := range paths {
		specs = append(specs, &ast.ImportSpec{
			Path: &ast.BasicLit{
				Kind:  token.STRING,
				Value: strconv.Quote(path),
			},
		})
	}
	return specs
}

// removeImportDecls filters out import declarations from a slice of declarations.
func removeImportDecls(decls []ast.Decl) []ast.Decl {
	result := make([]ast.Decl, 0, len(decls))
	for _, decl := range decls {
		if genDecl, ok := decl.(*ast.GenDecl); !ok || genDecl.Tok != token.IMPORT {
			result = append(result, decl)
		}
	}
	return result
}

func codePrettier(fset *token.FileSet, node *ast.File) (string, error) {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, node); err != nil {
		return "", fmt.Errorf("failed to print code: %w", err)
	}
	return buf.String(), nil
}
