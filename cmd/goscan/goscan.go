package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"

	"github.com/davecgh/go-spew/spew"
)

/*
Goscan:


 */

func main() {
	fset := token.NewFileSet() // positions are relative to fset

	filename := os.Args[1] // Take a filename as an argument.
	bz, err := ioutil.ReadFile(filename)
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
	// AST: https://en.wikipedia.org/wiki/Abstract_syntax_tree
	spew.Dump(f)

}
