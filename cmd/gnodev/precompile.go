package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/pkgs/command"
	"github.com/gnolang/gno/pkgs/errors"
	gno "github.com/gnolang/gno/pkgs/gnolang"
)

type precompileOptions struct {
	Verbose     bool   `flag:"verbose" help:"verbose"`
	SkipFmt     bool   `flag:"skip-fmt" help:"do not check syntax of generated .go files"`
	GoBinary    string `flag:"go-binary" help:"go binary to use for building"`
	GofmtBinary string `flag:"go-binary" help:"gofmt binary to use for syntax checking"`
	Output      string `flag:"output" help:"output directory"`
}

var defaultPrecompileOptions = precompileOptions{
	Verbose:     false,
	SkipFmt:     false,
	GoBinary:    "go",
	GofmtBinary: "gofmt",
	Output:      ".",
}

func precompileApp(cmd *command.Command, args []string, iopts interface{}) error {
	opts := iopts.(precompileOptions)
	if len(args) < 1 {
		cmd.ErrPrintfln("Usage: precompile [precompile flags] [packages]")
		return errors.New("invalid args")
	}

	// precompile .gno files.
	paths, err := gnoFilesFromArgs(args)
	if err != nil {
		return fmt.Errorf("list paths: %w", err)
	}

	errCount := 0
	for _, filepath := range paths {
		err = precompileFile(filepath, opts)
		if err != nil {
			err = fmt.Errorf("%s: precompile: %w", filepath, err)
			cmd.ErrPrintfln("%s", err.Error())
			errCount++
		}
	}

	if errCount > 0 {
		return fmt.Errorf("%d precompile errors", errCount)
	}

	return nil
}

func precompilePkg(pkgPath string, opts precompileOptions) error {
	files, err := filepath.Glob(filepath.Join(pkgPath, "*.gno"))
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if err = precompileFile(file, opts); err != nil {
			return fmt.Errorf("%s: %v", file, err)
		}
	}

	return nil
}

func precompileFile(srcPath string, opts precompileOptions) error {
	shouldCheckFmt := !opts.SkipFmt
	verbose := opts.Verbose
	gofmt := opts.GofmtBinary
	if gofmt == "" {
		gofmt = defaultPrecompileOptions.GofmtBinary
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "%s\n", srcPath)
	}

	// parse .gno.
	source, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}

	// compute attributes based on filename.
	var targetFilename string
	var tags string
	nameNoExtension := strings.TrimSuffix(filepath.Base(srcPath), ".gno")
	switch {
	case strings.HasSuffix(srcPath, "_filetest.gno"):
		tags = "gno,filetest"
		targetFilename = "." + nameNoExtension + ".gno.gen.go"
	case strings.HasSuffix(srcPath, "_test.gno"):
		tags = "gno,test"
		targetFilename = "." + nameNoExtension + ".gno.gen_test.go"
	default:
		tags = "gno"
		targetFilename = nameNoExtension + ".gno.gen.go"
	}

	// preprocess.
	transformed, err := gno.Precompile(string(source), tags, srcPath)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	// write .go file.
	dir := filepath.Dir(srcPath)
	targetPath := filepath.Join(dir, targetFilename)
	err = os.WriteFile(targetPath, []byte(transformed), 0o644)
	if err != nil {
		return fmt.Errorf("write .go file: %w", err)
	}

	// check .go fmt.
	if shouldCheckFmt {
		err = gno.PrecompileVerifyFile(targetPath, gofmt)
		if err != nil {
			return fmt.Errorf("check .go file: %w", err)
		}
	}

	return nil
}
