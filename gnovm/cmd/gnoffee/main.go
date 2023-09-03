package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/gnolang/gno/gnovm/pkg/gnoffee"
)

var writeFlag bool

func init() {
	flag.BoolVar(&writeFlag, "w", false, "write result to gnoffee.gen.go file instead of stdout")
}

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Usage: gnoffee [-w] <package-path or file.gnoffee or '-'>")
		return
	}

	err := doMain(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func doMain(arg string) error {
	fset, pkg, err := processPackageOrFileOrStdin(arg)
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	newFile, err := gnoffee.Stage2(pkg)
	if err != nil {
		return fmt.Errorf("processing the AST: %w", err)
	}

	// combine existing files into newFile to generate a unique file for the whole package.
	for _, file := range pkg {
		newFile.Decls = append(newFile.Decls, file.Decls...)
	}

	if writeFlag {
		filename := "gnoffee.gen.go"
		f, err := os.Create(filename)
		if err != nil {
			return fmt.Errorf("creating file %q: %w", filename, err)
		}
		defer f.Close()

		err = printer.Fprint(f, fset, newFile)
		if err != nil {
			return fmt.Errorf("writing to file %q: %w", filename, err)
		}
	} else {
		_ = printer.Fprint(os.Stdout, fset, newFile)
	}
	return nil
}

func processPackageOrFileOrStdin(arg string) (*token.FileSet, map[string]*ast.File, error) {
	fset := token.NewFileSet()
	pkg := map[string]*ast.File{}

	processFile := func(data []byte, filename string) error {
		source := string(data)
		source = gnoffee.Stage1(source)

		parsedFile, err := parser.ParseFile(fset, filename, source, parser.ParseComments)
		if err != nil {
			return fmt.Errorf("parsing file %q: %v", filename, err)
		}
		pkg[filename] = parsedFile
		return nil
	}

	// process arg
	if arg == "-" {
		// Read from stdin and process
		data, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return nil, nil, fmt.Errorf("reading from stdin: %w", err)
		}
		if err := processFile(data, "stdin.gnoffee"); err != nil {
			return nil, nil, err
		}
	} else {
		// If it's a directory, gather all .go and .gnoffee files and process accordingly
		if info, err := os.Stat(arg); err == nil && info.IsDir() {
			err := filepath.Walk(arg, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				ext := filepath.Ext(path)
				if ext == ".gnoffee" {
					data, err := ioutil.ReadFile(path)
					if err != nil {
						return fmt.Errorf("reading file %q: %v", path, err)
					}
					if err := processFile(data, path); err != nil {
						return err
					}
				}
				return nil
			})
			if err != nil {
				return nil, nil, err
			}
		} else {
			data, err := ioutil.ReadFile(arg)
			if err != nil {
				return nil, nil, fmt.Errorf("reading file %q: %w", arg, err)
			}
			if err := processFile(data, arg); err != nil {
				return nil, nil, err
			}
		}
	}
	return fset, pkg, nil
}
