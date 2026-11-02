package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// Go has no 256-bit integer, and no Go type can stand in for int256 or
// uint256 (gc's uintptr is 64 bits; a struct has no operators), so Gno code
// that uses them has no faithful Go translation. Rather than translate part of
// such code, the transpiler works at the granularity of whole files and
// packages, using go/types information (the same shims and sizes as Gno's
// type checker), never identifier spelling:
//
//   - in a standard library, the files that use 256-bit types (the stdlib
//     keeps its 256-bit APIs in dedicated files) are left out of the Go
//     output, and the transpile fails if another file of the package refers
//     to anything they declare;
//   - any other package that uses 256-bit types in one of its files is not
//     built as Go (--gobuild), nor is any package that imports it, directly
//     or not. Its files are still transpiled.
//
// Packages it cannot type-check are left to the Go build, which then reports
// their errors.

type go256State struct {
	// pkgPaths maps a package directory (absolute) to its import path.
	pkgPaths map[string]string
	// skip maps a package directory to the reason it is not built as Go.
	skip map[string]string
	// leftOut maps a stdlib package directory to its files left out.
	leftOut map[string]map[string]bool
	// imports maps a package directory to the directories it imports.
	imports map[string][]string
	checked map[string]checked
	cache   gno.TypeCheckCache
}

func newGo256State() *go256State {
	return &go256State{
		pkgPaths: map[string]string{},
		skip:     map[string]string{},
		leftOut:  map[string]map[string]bool{},
		imports:  map[string][]string{},
		checked:  map[string]checked{},
		cache:    gno.TypeCheckCache{},
	}
}

func absDir(dir string) string {
	if abs, err := filepath.Abs(dir); err == nil {
		return filepath.Clean(abs)
	}
	return filepath.Clean(dir)
}

// skipBuild reports whether the package at dir is not to be built as Go, and
// why.
func (st *go256State) skipBuild(dir string) (string, bool) {
	reason, ok := st.skip[absDir(dir)]
	return reason, ok
}

// SkippedPackages returns the packages not built as Go, with the reason for
// each, sorted by directory.
func (st *go256State) SkippedPackages() []string {
	out := make([]string, 0, len(st.skip))
	for dir, reason := range st.skip {
		out = append(out, dir+": "+reason)
	}
	sort.Strings(out)
	return out
}

// transpileImportsOf transpiles the packages imported by files.
func transpileImportsOf(files []string, opts *transpileOptions) error {
	seen := map[string]bool{}
	fset := token.NewFileSet()
	for _, file := range files {
		f, err := parser.ParseFile(fset, file, nil, parser.ImportsOnly)
		if err != nil {
			continue // reported when the file itself is transpiled
		}
		for _, is := range f.Imports {
			path, err := strconv.Unquote(is.Path.Value)
			if err != nil || seen[path] {
				continue
			}
			seen[path] = true
			rel, ok := opts.resolver(path)
			if !ok {
				continue
			}
			if err := transpilePkg(filepath.Join(opts.cfg.rootDir, rel), opts); err != nil {
				return err
			}
		}
	}
	return nil
}

// importPathOf returns the import path of the package at dir, if known.
func (st *go256State) importPathOf(dir, rootDir string) (string, bool) {
	if p, ok := st.pkgPaths[dir]; ok {
		return p, true
	}
	stdlibs := absDir(filepath.Join(rootDir, "gnovm", "stdlibs")) + string(filepath.Separator)
	if rel, ok := strings.CutPrefix(dir, stdlibs); ok {
		return filepath.ToSlash(rel), true
	}
	if mod, err := gnomod.ParseDir(dir); err == nil && mod.Module != "" {
		return mod.Module, true
	}
	return "", false
}

// fsGetter reads the packages imported during type-checking from disk.
type fsGetter struct{ opts *transpileOptions }

func (g fsGetter) GetMemPackage(path string) (mpkg *std.MemPackage) {
	rel, ok := g.opts.resolver(path)
	if !ok {
		return nil
	}
	defer func() {
		if recover() != nil {
			mpkg = nil
		}
	}()
	mpkg, err := gno.ReadMemPackage(filepath.Join(g.opts.cfg.rootDir, rel), path, gno.MPAnyProd)
	if err != nil {
		return nil
	}
	return mpkg
}

// analyze decides how the package at dirPath is handled, once its imports
// have been analyzed: it returns the files to leave out of the Go output (for
// a standard library), and records whether the package is built as Go.
func (st *go256State) analyze(dirPath string, opts *transpileOptions) (leftOut map[string]bool, err error) {
	dir := absDir(dirPath)
	pkgPath, known := st.importPathOf(dir, opts.cfg.rootDir)
	if !known {
		return nil, nil
	}
	files256, ok := st.files256(dir, pkgPath, opts)

	// Imports first: a package importing one that is not built is not
	// built either.
	for _, imp := range st.imports[dir] {
		if _, skipped := st.skip[imp]; skipped {
			impPath, _ := st.importPathOf(imp, opts.cfg.rootDir)
			st.skip[dir] = fmt.Sprintf("imports %s, which is not built as Go", impPath)
			break
		}
	}
	if !ok || len(files256) == 0 {
		return nil, nil
	}

	if gno.IsStdlib(pkgPath) {
		leftOut = map[string]bool{}
		for _, f := range files256 {
			leftOut[f] = true
		}
		// The rest of the package must not depend on what is left out.
		info, fset, ok := st.typeCheck(dir, pkgPath, opts)
		if !ok {
			return nil, fmt.Errorf("standard library %s uses int256/uint256 but could not be type-checked", pkgPath)
		}
		{
			if refs := gno.RefsInto256Files(info, fset, leftOut); len(refs) > 0 {
				return nil, fmt.Errorf("standard library %s: files %s use int256/uint256 and are left out of the Go output, but other files refer to their declarations: %s",
					pkgPath, strings.Join(files256, ", "), strings.Join(refs, "; "))
			}
			for _, f := range files256 {
				src, err := os.ReadFile(filepath.Join(dir, f))
				if err != nil {
					return nil, err
				}
				effects, err := gno.InitEffects256(info, fset, f, string(src))
				if err != nil {
					return nil, err
				}
				if len(effects) > 0 {
					return nil, fmt.Errorf("standard library %s: %s uses int256/uint256 and is left out of the Go output, but has initialization effects: %s",
						pkgPath, f, strings.Join(effects, "; "))
				}
			}
		}
		st.leftOut[dir] = leftOut
		return leftOut, nil
	}
	if _, skipped := st.skip[dir]; !skipped {
		st.skip[dir] = fmt.Sprintf("uses int256/uint256, which Go lacks (%s)", strings.Join(files256, ", "))
	}
	return nil, nil
}

// files256 type-checks the package at dir and returns its files that use
// 256-bit types. ok is false if it could not be type-checked.
func (st *go256State) files256(dir, pkgPath string, opts *transpileOptions) ([]string, bool) {
	info, fset, ok := st.typeCheck(dir, pkgPath, opts)
	if !ok {
		return nil, false
	}
	return gno.Files256(info, fset), true
}

func (st *go256State) typeCheck(dir, pkgPath string, opts *transpileOptions) (info *types.Info, fset *token.FileSet, ok bool) {
	if c, done := st.checked[dir]; done {
		return c.info, c.fset, true
	}
	defer func() {
		if r := recover(); r != nil {
			ok = false
		}
	}()
	mpkg, err := gno.ReadMemPackage(dir, pkgPath, gno.MPAnyProd)
	if err != nil {
		return nil, nil, false
	}
	// Record the imported package directories, for the transitive skip.
	if _, done := st.imports[dir]; !done {
		var imps []string
		ifset := token.NewFileSet()
		for _, mf := range mpkg.Files {
			if !strings.HasSuffix(mf.Name, ".gno") || strings.HasSuffix(mf.Name, "_test.gno") || strings.HasSuffix(mf.Name, "_filetest.gno") {
				continue
			}
			f, err := parser.ParseFile(ifset, mf.Name, mf.Body, parser.ImportsOnly)
			if err != nil {
				continue
			}
			for _, is := range f.Imports {
				path, err := strconv.Unquote(is.Path.Value)
				if err != nil {
					continue
				}
				if rel, ok := opts.resolver(path); ok {
					imps = append(imps, absDir(filepath.Join(opts.cfg.rootDir, rel)))
				}
			}
		}
		st.imports[dir] = imps
	}
	info = &types.Info{
		Types: map[ast.Expr]types.TypeAndValue{},
		Defs:  map[*ast.Ident]types.Object{},
		Uses:  map[*ast.Ident]types.Object{},
	}
	fset = token.NewFileSet()
	getter := fsGetter{opts}
	_, err = gno.TypeCheckMemPackage(mpkg, gno.TypeCheckOptions{
		Getter:     getter,
		TestGetter: getter,
		Mode:       gno.TCLatestRelaxed,
		Cache:      st.cache,
		Fset:       fset,
		ProdOnly:   true,
		Info:       info,
	})
	if err != nil {
		return nil, nil, false
	}
	st.checked[dir] = checked{info, fset}
	return info, fset, true
}

type checked struct {
	info *types.Info
	fset *token.FileSet
}
