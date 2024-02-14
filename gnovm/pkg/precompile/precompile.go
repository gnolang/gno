package precompile

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"go.uber.org/multierr"
	"golang.org/x/tools/go/ast/astutil"

	"github.com/gnolang/gno/tm2/pkg/std"
)

const (
	GnoRealmPkgsPrefixBefore = "gno.land/r/"
	GnoRealmPkgsPrefixAfter  = "github.com/gnolang/gno/examples/gno.land/r/"
	GnoPackagePrefixBefore   = "gno.land/p/demo/"
	GnoPackagePrefixAfter    = "github.com/gnolang/gno/examples/gno.land/p/demo/"
	GnoStdPkgBefore          = "std"
	GnoStdPkgAfter           = "github.com/gnolang/gno/gnovm/stdlibs/stdshim"
	GnoTestExternBefore      = "github.com/gnolang/gno/_test/"
	GnoTestExternAfter       = "github.com/gnolang/gno/gnovm/tests/files/extern/"
)

var stdlibWhitelist = []string{
	// go
	"bufio",
	"bytes",
	"compress/gzip",
	"context",
	"crypto/md5",
	"crypto/sha1",
	"crypto/sha256",
	"encoding/base64",
	"encoding/binary",
	"encoding/hex",
	"encoding/json",
	"encoding/xml",
	"errors",
	"flag",
	"fmt",
	"io",
	"io/util",
	"math",
	"math/big",
	"math/rand",
	"regexp",
	"sort",
	"strconv",
	"strings",
	"text/template",
	"time",
	"unicode/utf8",

	// gno
	"std",
}

var importPrefixWhitelist = []string{
	"github.com/gnolang/gno/_test",
}

const ImportPrefix = "github.com/gnolang/gno"
const testPath = "github.com/gnolang/gno/_test/"

type precompileResult struct {
	Imports    []*ast.ImportSpec
	Translated string
}

// ==================================================================
type ImportPath string

type PrecompileCfg struct {
	Verbose     bool
	SkipFmt     bool
	SkipImports bool
	Gobuild     bool
	GoBinary    string
	GofmtBinary string
	Output      string
}

type PrecompileOptions struct {
	Cfg *PrecompileCfg
	// precompiled is the set of packages already
	// precompiled from .gno to .go.
	Precompiled map[ImportPath]struct{}
}

var DefaultPrecompileCfg = &PrecompileCfg{
	Verbose:  false,
	GoBinary: "go",
	Gobuild:  true,
}

func NewPrecompileOptions(cfg *PrecompileCfg) *PrecompileOptions {
	return &PrecompileOptions{cfg, map[ImportPath]struct{}{}}
}

func (p *PrecompileOptions) GetFlags() *PrecompileCfg {
	return p.Cfg
}

func (p *PrecompileOptions) IsPrecompiled(pkg ImportPath) bool {
	_, precompiled := p.Precompiled[pkg]
	return precompiled
}

func (p *PrecompileOptions) MarkAsPrecompiled(pkg ImportPath) {
	p.Precompiled[pkg] = struct{}{}
}

// TODO: add clean
func (c *PrecompileCfg) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(
		&c.Verbose,
		"verbose",
		false,
		"verbose output when running",
	)

	fs.BoolVar(
		&c.SkipFmt,
		"skip-fmt",
		false,
		"do not check syntax of generated .go files",
	)

	fs.BoolVar(
		&c.SkipImports,
		"skip-imports",
		false,
		"do not precompile imports recursively",
	)

	fs.BoolVar(
		&c.Gobuild,
		"gobuild",
		false,
		"run go build on generated go files, ignoring test files",
	)

	fs.StringVar(
		&c.GoBinary,
		"go-binary",
		"go",
		"go binary to use for building",
	)

	fs.StringVar(
		&c.GofmtBinary,
		"go-fmt-binary",
		"gofmt",
		"gofmt binary to use for syntax checking",
	)

	fs.StringVar(
		&c.Output,
		"output",
		".",
		"output directory",
	)
}

// ==================================================================
func PrecompilePkg(pkgPath ImportPath, opts *PrecompileOptions) error {
	fmt.Println("---precompilePkg, pkgPath: ", pkgPath)
	if opts.IsPrecompiled(pkgPath) {
		fmt.Printf("path: %s isCompiled \n", pkgPath)
		return nil
	}
	opts.MarkAsPrecompiled(pkgPath)

	files, err := filepath.Glob(filepath.Join(string(pkgPath), "*.gno"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("---files: ", files)
	for _, file := range files {
		fmt.Println("---file: ", file)
		if err = PrecompileFile(file, opts); err != nil {
			return fmt.Errorf("%s: %w", file, err)
		}
	}

	return nil
}

// precompile file and imports, xxx.gno -> xxx.gen.go
func PrecompileFile(srcPath string, opts *PrecompileOptions) error {
	fmt.Println("---PrecompileFile at srcPath:", srcPath)
	var importPaths []ImportPath

	flags := opts.GetFlags()
	gofmt := flags.GofmtBinary
	if gofmt == "" {
		gofmt = "gofmt"
	}

	if flags.Verbose {
		fmt.Fprintf(os.Stderr, "%s\n", srcPath)
	}

	// parse .gno.
	source, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}

	// compute attributes based on filename.
	targetFilename, tags := GetPrecompileFilenameAndTags(srcPath)
	fmt.Println("tags: ", tags)
	if !opts.GetFlags().Gobuild {
		tags = "noop"
	}
	precompileRes, err := Precompile(string(source), tags, srcPath)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	//for _, importPath := range precompileRes.Imports {
	//	fmt.Printf("imported path.value: %s \n", importPath.Path.Value)
	//	fmt.Printf("imported path.name: %s \n", importPath.Name)
	//}

	// resolve target path
	var targetPath string
	if flags.Output != "." {
		//path, err := resolvePath(flags.Output, ImportPath(filepath.Dir(srcPath)))
		path, err := resolvePath(flags.Output, ImportPath(filepath.Dir(srcPath)))
		if err != nil {
			return fmt.Errorf("resolve output path: %w", err)
		}
		fmt.Println("---resolved path: ", path)
		fmt.Println("---targetFileName: ", targetFilename)
		targetPath = filepath.Join(path, targetFilename)
	} else {
		dir := filepath.Dir(srcPath)
		fmt.Println("---dir: ", dir)
		fmt.Println("---targetFileName: ", targetFilename)
		targetPath = filepath.Join(dir, targetFilename)
	}

	fmt.Println("---targetPath: ", targetPath)
	// write .go file.
	err = writeDirFile(targetPath, []byte(precompileRes.Translated))
	if err != nil {
		return fmt.Errorf("write .go file: %w", err)
	}

	fmt.Println("---check result at:", filepath.Dir(srcPath))

	// Read the directory contents
	dirEntries, err := os.ReadDir(filepath.Dir(srcPath))
	if err != nil {
		fmt.Println("Error:", err)
	}

	// Iterate over each directory entry
	for _, entry := range dirEntries {
		// Get information about the file
		fileInfo, err := entry.Info()
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		// Print file information
		fmt.Printf("---file name: %s, Size: %d bytes, IsDir: %t\n", fileInfo.Name(), fileInfo.Size(), fileInfo.IsDir())
		//If the entry is a file, print its content
		//if !fileInfo.IsDir() {
		//	fileContent, err := ioutil.ReadFile(filepath.Dir(srcPath) + "/" + entry.Name())
		//	if err != nil {
		//		fmt.Println("Error reading file:", err)
		//		continue
		//	}
		//	fmt.Println("Content:")
		//	fmt.Println(string(fileContent))
		//}
	}

	// check .go fmt, if `SkipFmt` sets to false.
	if !flags.SkipFmt {
		err = PrecompileVerifyFile(targetPath, gofmt)
		if err != nil {
			return fmt.Errorf("check .go file: %w", err)
		}
	}

	// precompile imported packages, if `SkipImports` sets to false
	if !flags.SkipImports {
		importPaths = getPathsFromImportSpec(precompileRes.Imports)
		fmt.Println("---importPaths: ", importPaths)
		for _, path := range importPaths {
			fmt.Println("---precompile imported path: ", path)
			PrecompilePkg(path, opts)
		}
	}
	return nil
}

func GoBuildFileOrPkg(fileOrPkg string, cfg *PrecompileCfg) error {
	verbose := cfg.Verbose
	goBinary := cfg.GoBinary

	if verbose {
		fmt.Fprintf(os.Stderr, "%s\n", fileOrPkg)
	}

	return PrecompileBuildPackage(fileOrPkg, goBinary)
}

// TODO: func PrecompileFile: supports caching.
// TODO: func PrecompilePkg: supports directories.
func guessRootDir(fileOrPkg string, goBinary string) (string, error) {
	abs, err := filepath.Abs(fileOrPkg)
	if err != nil {
		return "", err
	}
	args := []string{"list", "-m", "-mod=mod", "-f", "{{.Dir}}", ImportPrefix}
	cmd := exec.Command(goBinary, args...)
	cmd.Dir = abs
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("can't guess --root-dir")
	}
	rootDir := strings.TrimSpace(string(out))
	return rootDir, nil
}

// PrecompileAndCheckPkg conducts precompile and try build against precompiled files
// used by `gno precompile`, `gnokey run`, `gnokey addpkg`, `client_txs Run`
func PrecompileAndCheckPkg(isMem bool, mempkg *std.MemPackage, paths []string, cfg *PrecompileCfg) (error, string) {
	fmt.Println("---gnolang, PrecompileAndCheckMemPkg")
	var targetPaths []string
	var srcPaths []string
	var tmpDir string
	// precompile with fmt and import files,  and try go build
	precompileCfg := cfg
	if precompileCfg == nil {
		precompileCfg = DefaultPrecompileCfg
	}

	opts := NewPrecompileOptions(precompileCfg)

	defer func() {
		// clean main generated files
		for _, srcPath := range srcPaths {
			fmt.Println("---clean dir:", srcPath)
			err := CleanGeneratedFiles(srcPath)
			if err != nil {
				panic(err)
			}
		}
		// clean imported
		for pkgPath := range opts.Precompiled {
			fmt.Println("precompiled import pkg:", pkgPath)
			fmt.Println("---clean dir:", pkgPath)
			err := CleanGeneratedFiles(string(pkgPath))
			if err != nil {
				panic(err)
			}
		}
		//_, err := os.Stat(tmpDir)
		//if err != nil {
		//	if os.IsNotExist(err) {
		//		fmt.Println("Directory does not exist.")
		//	} else {
		//		fmt.Println("Error:", err)
		//	}
		//	return
		//}
		defer os.RemoveAll(tmpDir) //nolint: errcheck

	}()

	if isMem {
		absPath, err := filepath.Abs("")
		if err != nil {
			panic(err)
		}
		tmpDir, err = os.MkdirTemp(absPath, "*"+mempkg.Name)
		if err != nil {
			panic(err)
		}
		fmt.Println("---tmpDir: ", tmpDir)
		//defer os.RemoveAll(tmpDir) //nolint: errcheck

		// write mem file to tmp dir
		for _, mfile := range mempkg.Files {
			if !strings.HasSuffix(mfile.Name, ".gno") {
				continue // skip spurious file.
			}
			tmpFile := filepath.Join(tmpDir, mfile.Name)
			err = os.WriteFile(tmpFile, []byte(mfile.Body), 0o644)
			if err != nil {
				panic(err)
			}
		}
		targetPaths = append(targetPaths, tmpDir)
	} else {
		targetPaths = paths // existing files
	}

	fmt.Println("---targetPaths: ", targetPaths)

	srcPaths, err := GnoFilesFromArgs(targetPaths)
	if err != nil {
		panic(fmt.Sprintf("list paths: %w", err))
	}

	errCount := 0
	for _, srcPath := range srcPaths {
		fmt.Println("---precompile file at filepath: ", srcPath)
		err = PrecompileFile(srcPath, opts)
		if err != nil {
			err = fmt.Errorf("%s: precompile: %w", srcPath, err)
			errCount++
		}
	}
	if errCount > 0 {
		return fmt.Errorf("%d precompile errors from addpkg", errCount), ""
	}

	// try build
	pkgPaths, err := GnoPackagesFromArgs(targetPaths)
	if err != nil {
		return fmt.Errorf("list packages: %w", err), ""
	}
	errCount = 0
	for _, pkgPath := range pkgPaths {
		fmt.Println("---pkg path: ", pkgPath)
		_ = pkgPath
		if opts.GetFlags().Gobuild == false { // go run
			fmt.Println("---run file")

			// Read the directory contents
			dirEntries, err := os.ReadDir(pkgPath)
			if err != nil {
				fmt.Println("Error:", err)
			}

			// Iterate over each directory entry
			for _, entry := range dirEntries {
				// Get information about the file
				fileInfo, err := entry.Info()
				if err != nil {
					fmt.Println("Error:", err)
					continue
				}

				// Print file information
				fmt.Printf("---file Name: %s, Size: %d bytes, IsDir: %t\n", fileInfo.Name(), fileInfo.Size(), fileInfo.IsDir())
				if !fileInfo.IsDir() {
					//fileContent, err := ioutil.ReadFile(filepath.Dir(pkgPath) + "/" + entry.Name())
					fileContent, err := ioutil.ReadFile(pkgPath + "/" + entry.Name())
					if err != nil {
						fmt.Println("Error reading file:", err)
						continue
					}
					fmt.Println("Content:")
					fmt.Println(string(fileContent))
				}
			}

			err, output := PrecompileRun("main.gno.gen.go", pkgPath, "go run", paths[0]) // always one file per run
			if err != nil {
				//err = fmt.Errorf("%s: run pkg: %w", pkgPath, err)
				errCount++
				return err, ""
			}
			return nil, output

		} else {
			err = GoBuildFileOrPkg(pkgPath, precompileCfg)
			if err != nil {
				err = fmt.Errorf("%s: build pkg: %w", pkgPath, err)
				errCount++
			}
		}
	}

	if errCount > 0 {
		return fmt.Errorf("%d build errors", errCount), ""
	}

	return nil, ""
}

// For single file that with no native injection, most about test files in gnovm/test/files, challenge, etc
// path is uses to make log
func PrecompileAndRunMempkg(mempkg *std.MemPackage, path string) (error, string) {
	fmt.Println("---PrecompileAndRunMempkg, path: ", path)
	goRun := "go run"

	tmpDir, err := os.MkdirTemp("", mempkg.Name)
	if err != nil {
		return err, ""
	}
	defer os.RemoveAll(tmpDir) //nolint: errcheck

	fmt.Println("---tmpDir: ", tmpDir)

	var errs error
	var output string
	for _, mfile := range mempkg.Files { // for gnovm/test/files, only one file contained
		if !strings.HasSuffix(mfile.Name, ".gno") {
			continue // skip spurious file.
		}
		targetFileName, tags := GetPrecompileFilenameAndTags(mfile.Name)
		fmt.Println("---targetFileName:", targetFileName)
		res, err := Precompile(mfile.Body, tags, mfile.Name)
		if err != nil {
			errs = multierr.Append(errs, err)
			continue
		}
		//tmpFile := filepath.Join(tmpDir, mfile.Name)
		err = os.WriteFile(filepath.Join(tmpDir, targetFileName), []byte(res.Translated), 0o644)
		if err != nil {
			errs = multierr.Append(errs, err)
			continue
		}
		// check precompiled file
		err, output = PrecompileRun(targetFileName, tmpDir, goRun, path)
		if err != nil {
			errs = multierr.Append(errs, err)
			continue
		}
	}
	if errs != nil {
		//return fmt.Errorf("precompile package: %w", errs), ""
		return errs, ""
	}
	fmt.Println("---output before return is:", output)
	return nil, output
}

// core translate logic from gno to go
func Precompile(source string, tags string, filename string) (*precompileResult, error) {
	fmt.Println("---Precompile, filename: ", filename)
	var out bytes.Buffer

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, source, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	isTestFile := strings.HasSuffix(filename, "_test.gno") || strings.HasSuffix(filename, "_filetest.gno")
	shouldCheckWhitelist := !isTestFile

	transformed, err := precompileAST(fset, f, shouldCheckWhitelist)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	if tags != "noop" {
		header := "// Code generated by github.com/gnolang/gno. DO NOT EDIT.\n\n"
		if tags != "" {
			header += "//go:build " + tags + "\n\n"
		}
		_, err = out.WriteString(header)
		if err != nil {
			return nil, fmt.Errorf("write to buffer: %w", err)
		}
	}
	err = format.Node(&out, fset, transformed)

	res := &precompileResult{
		Imports:    f.Imports,
		Translated: out.String(),
	}
	return res, nil
}

// PrecompileVerifyFile tries to run `go fmt` against a precompiled .go file.
// This is fast and won't look the imports.
func PrecompileVerifyFile(path string, gofmtBinary string) error {
	// TODO: use cmd/parser instead of exec?

	args := strings.Split(gofmtBinary, " ")
	args = append(args, []string{"-l", "-e", path}...)
	cmd := exec.Command(args[0], args[1:]...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintln(os.Stderr, string(out))
		return fmt.Errorf("%s: %w", gofmtBinary, err)
	}
	return nil
}

// run precompiled go code and handle the result
func PrecompileRun(targetFileName string, targetPath string, goRunBinary string, path string) (error, string) {
	fmt.Printf("---PrecompileRun, dir: %s, gorun: %s \n", targetPath, goRunBinary)
	// TODO: use cmd/parser instead of exec?

	args := strings.Split(goRunBinary, " ")
	args = append(args, targetFileName)
	fmt.Println("---args: ", args)

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = targetPath
	out, err := cmd.CombinedOutput()
	if err != nil { // exit status 1
		fmt.Println("err.Error:", err.Error())
	}

	fmt.Println("combined out: ", string(out))
	res, isErr := parseCmdResult(string(out), path)
	if isErr && res != "" {
		fmt.Println("---return stderr, res: ", res)
		return errors.New(res), ""
	} else if !isErr && res != "" {
		return nil, res
	} else if len(out) != 0 {
		fmt.Println("---return stdout")
		return nil, string(out)
	}
	return nil, ""
}

func parseCmdResult(input string, path string) (string, bool) {
	// List of substrings to be trimmed
	//substrings := []string{"command-line-arguments", "# command-line-arguments"}
	tag := "command-line-arguments"
	var isStdErr bool
	input = strings.TrimSpace(input)
	if strings.Contains(input, tag) {
		fmt.Println("--- contain, input:", input)
		isStdErr = true
	}

	// reassemble log
	parts := strings.Split(input, "main.gno.gen.go")
	// Check if the split resulted in at least two parts
	if len(parts) > 1 {
		// The second part is the string after "main.gno.gen.go"
		input = path + parts[1]
		fmt.Println("Trimmed string:", input)
	} else {
		fmt.Println("String does not contain 'main.go'")
	}
	return input, isStdErr
}

// PrecompileBuildPackage tries to run `go build` against the precompiled .go files.
//
// This method is the most efficient to detect errors but requires that
// all the import are valid and available.
func PrecompileBuildPackage(fileOrPkg string, goBinary string) error {
	fmt.Println("---PrecompileBuildPackage, fileOrPkg: ", fileOrPkg)
	// TODO: use cmd/compile instead of exec?
	// TODO: find the nearest go.mod file, chdir in the same folder, rim prefix?
	// TODO: temporarily create an in-memory go.mod or disable go modules for gno?
	// TODO: ignore .go files that were not generated from gno?
	// TODO: automatically precompile if not yet done.

	//  for test
	files := []string{}

	info, err := os.Stat(fileOrPkg)
	if err != nil {
		return fmt.Errorf("invalid file or package path: %w", err)
	}
	if !info.IsDir() {
		file := fileOrPkg
		files = append(files, file)
	} else {
		pkgDir := fileOrPkg
		goGlob := filepath.Join(pkgDir, "*.go")
		goMatches, err := filepath.Glob(goGlob)
		if err != nil {
			return fmt.Errorf("glob: %w", err)
		}
		for _, goMatch := range goMatches {
			fmt.Println("---goMatch: ", goMatch)
			switch {
			case strings.HasPrefix(goMatch, "."): // skip
			case strings.HasSuffix(goMatch, "_filetest.go"): // skip
			case strings.HasSuffix(goMatch, "_filetest.gno.gen.go"): // skip
			case strings.HasSuffix(goMatch, "_test.go"): // skip
			case strings.HasSuffix(goMatch, "_test.gno.gen.go"): // skip
			default:
				println("append ")
				files = append(files, goMatch)
			}
		}
	}

	for _, f := range files {
		fmt.Println("file: ", f)
	}
	sort.Strings(files)
	args := append([]string{"build", "-v", "-tags=gno"}, files...)
	cmd := exec.Command(goBinary, args...)
	rootDir, err := guessRootDir(fileOrPkg, goBinary)
	if err == nil {
		cmd.Dir = rootDir
	}
	fmt.Println("rootDir: ", rootDir)
	out, err := cmd.CombinedOutput()
	fmt.Println("---out:", string(out))
	if err != nil {
		fmt.Fprintln(os.Stderr, string(out))
		fmt.Printf("---build fail, out: %s \n", string(out))
		return fmt.Errorf("std go compiler: %w", err)
	}

	return nil
}

func precompileAST(fset *token.FileSet, f *ast.File, checkWhitelist bool) (ast.Node, error) {
	fmt.Println("---precompileAst")
	var errs error

	imports := astutil.Imports(fset, f)

	// import whitelist
	if checkWhitelist {
		for _, paragraph := range imports {
			for _, importSpec := range paragraph {
				importPath := strings.TrimPrefix(strings.TrimSuffix(importSpec.Path.Value, `"`), `"`)

				if strings.HasPrefix(importPath, GnoRealmPkgsPrefixBefore) {
					continue
				}

				if strings.HasPrefix(importPath, GnoPackagePrefixBefore) {
					continue
				}

				valid := false
				for _, whitelisted := range stdlibWhitelist {
					if importPath == whitelisted {
						valid = true
						break
					}
				}
				if valid {
					continue
				}

				for _, whitelisted := range importPrefixWhitelist {
					if strings.HasPrefix(importPath, whitelisted) {
						valid = true
						break
					}
				}
				if valid {
					continue
				}

				errs = multierr.Append(errs, fmt.Errorf("import %q is not in the whitelist", importPath))
			}
		}
	}

	// rewrite imports
	for _, paragraph := range imports {
		for _, importSpec := range paragraph {
			importPath := strings.TrimPrefix(strings.TrimSuffix(importSpec.Path.Value, `"`), `"`)
			fmt.Println("---precompileAst, importPath: ", importPath)

			// std package
			if importPath == GnoStdPkgBefore {
				if !astutil.RewriteImport(fset, f, GnoStdPkgBefore, GnoStdPkgAfter) {
					errs = multierr.Append(errs, fmt.Errorf("failed to replace the %q package with %q", GnoStdPkgBefore, GnoStdPkgAfter))
				}
			}

			// p/pkg packages
			if strings.HasPrefix(importPath, GnoPackagePrefixBefore) {
				target := GnoPackagePrefixAfter + strings.TrimPrefix(importPath, GnoPackagePrefixBefore)

				if !astutil.RewriteImport(fset, f, importPath, target) {
					errs = multierr.Append(errs, fmt.Errorf("failed to replace the %q package with %q", importPath, target))
				}
			}

			// r/realm packages
			if strings.HasPrefix(importPath, GnoRealmPkgsPrefixBefore) {
				target := GnoRealmPkgsPrefixAfter + strings.TrimPrefix(importPath, GnoRealmPkgsPrefixBefore)

				if !astutil.RewriteImport(fset, f, importPath, target) {
					errs = multierr.Append(errs, fmt.Errorf("failed to replace the %q package with %q", importPath, target))
				}
			}

			// gnovm/test/files/extern
			if strings.Contains(importPath, GnoTestExternBefore) {
				fmt.Println("---match test extern")
				importAfter := strings.TrimPrefix(importPath, GnoTestExternBefore)
				if !astutil.RewriteImport(fset, f, importPath, importAfter) {
					errs = multierr.Append(errs, fmt.Errorf("failed to replace the %q package with %q", GnoTestExternBefore, GnoTestExternAfter))
				}
			}
		}
	}

	// custom handler
	node := astutil.Apply(f,
		// pre
		func(c *astutil.Cursor) bool {
			// do things here
			return true
		},
		// post
		func(c *astutil.Cursor) bool {
			// and here
			return true
		},
	)

	return node, errs
}
