package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"

	"github.com/davecgh/go-spew/spew"
)

func main() {
	fset := token.NewFileSet() // positions are relative to fset

	filename := os.Args[1]
	bz, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	// Parse src but stop after processing the imports.
	f, err := parser.ParseFile(fset, "", string(bz), parser.ParseComments|parser.DeclarationErrors)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Print the imports from the file's AST.
	spew.Dump(f)
}
