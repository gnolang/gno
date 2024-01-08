package main

import (
	"context"
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/gnolang/gno/tm2/pkg/commands"
)

func main() {
	cmd := commands.NewCommand(
		commands.Metadata{
			ShortUsage: "<file-path>",
			LongHelp:   "Prints out the imports for a given file's AST",
		},
		commands.NewEmptyConfig(),
		execScan,
	)

	cmd.Execute(context.Background(), os.Args[1:])
}

func execScan(_ context.Context, args []string) error {
	if len(args) < 1 {
		return flag.ErrHelp
	}

	fset := token.NewFileSet() // positions are relative to fset

	filename := args[0]
	bz, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("unable to read file, %w", err)
	}

	// Parse src but stop after processing the imports.
	f, err := parser.ParseFile(fset, "", string(bz), parser.ParseComments|parser.DeclarationErrors)
	if err != nil {
		return fmt.Errorf("unable to parse file, %w", err)
	}

	// Print the imports from the file's AST.
	spew.Dump(f)

	return nil
}
