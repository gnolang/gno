package importer

import (
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/std"
	"golang.org/x/mod/modfile"
)

const recursiveSuffix = string(os.PathSeparator) + "..."

type visitTarget struct {
	path  string
	match string
}

func DiscoverPackages(paths ...string) ([]*PackageSummary, error) {
	toVisit := []visitTarget{}
	for _, p := range paths {
		toVisit = append(toVisit, visitTarget{path: p, match: p})
	}
	visited := map[visitTarget]struct{}{}
	cache := map[string]*PackageSummary{}
	packages := []*PackageSummary{}
	errs := []error{}

	for len(toVisit) > 0 {
		tgt := toVisit[0]
		toVisit = toVisit[1:]

		if _, ok := visited[tgt]; ok {
			continue
		}
		visited[tgt] = struct{}{}

		if tgt.path == "" {
			continue
		}

		if tgt.path[0] == '.' {
			absPath, err := filepath.Abs(tgt.path)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to get absolute path for %q: %w", tgt, err))
			}
			toVisit = append(toVisit, visitTarget{path: absPath, match: tgt.match})
			continue
		}

		root := convertRecursivePathToDir(tgt.path)

		if !isRecursivePath(tgt.path) {
			if tgt.path[0] != '/' {
				pkgPath := tgt.path
				if pkg, ok := cache[pkgPath]; ok {
					pkg.Match = append(pkg.Match, tgt.match)
				} else {
					cache[pkgPath] = &PackageSummary{
						PkgPath: pkgPath,
						Match:   []string{tgt.match},
					}
					packages = append(packages, cache[pkgPath])
				}
				continue
			}
			modDir, err := findModDir(root)
			if os.IsNotExist(err) {
				continue
			} else if err != nil {
				return nil, fmt.Errorf("failed to find parent module: %w", err)
			}
			modFilePath := filepath.Join(modDir, ModfileName)
			modFileBytes, err := os.ReadFile(modFilePath)
			if err != nil {
				return nil, fmt.Errorf("failed to read modfile: %w", err)
			}
			modFile, err := modfile.ParseLax(modFilePath, modFileBytes, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to parse modfile: %w", err)
			}
			if modFile == nil || modFile.Module == nil {
				continue
			}
			globalPkgPath := modFile.Module.Mod.Path
			relfpath, err := filepath.Rel(modDir, tgt.path)
			if err != nil {
				return nil, fmt.Errorf("failed to get pkg path relative to mod root: %w", err)
			}
			relpath := strings.Join(filepath.SplitList(relfpath), "/")
			absroot, err := filepath.Abs(root)
			if err != nil {
				return nil, fmt.Errorf("failed to get absolute pkg root")
			}
			pkgPath := path.Join(globalPkgPath, relpath)
			if pkg, ok := cache[pkgPath]; ok {
				pkg.Match = append(pkg.Match, tgt.match)
			} else {
				cache[pkgPath] = &PackageSummary{
					PkgPath: path.Join(globalPkgPath, relpath),
					Root:    absroot,
					Module: &Module{
						Path:   globalPkgPath,
						Dir:    modDir,
						GnoMod: modFilePath,
					},
					Match: []string{tgt.match},
				}
				packages = append(packages, cache[pkgPath])
			}
			continue
		}

		dirEntry, err := os.ReadDir(root)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		hasGnoFiles := false
		for _, entry := range dirEntry {
			fileName := entry.Name()
			if entry.IsDir() {
				dirPath := filepath.Join(root, fileName) + recursiveSuffix
				toVisit = append(toVisit, visitTarget{path: dirPath, match: tgt.match})
			}
			if !hasGnoFiles && IsGnoFile(fileName) {
				hasGnoFiles = true
			}
		}

		if hasGnoFiles {
			toVisit = append(toVisit, visitTarget{path: root, match: tgt.match})
		}
	}

	return packages, errors.Join(errs...)
}

type PackageSummary struct {
	PkgPath string
	Root    string
	Module  *Module
	Match   []string
}

// FIXME: support files
func Load(paths ...string) ([]*Package, error) {
	pkgs, err := DiscoverPackages(paths...)
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}

	visited := map[string]struct{}{}
	cache := make(map[string]*Package)
	list := []*Package{}
	errs := []error{}
	for pile := pkgs; len(pile) > 0; pile = pile[1:] {
		pkgSum := pile[0]
		if _, ok := visited[pkgSum.PkgPath]; ok {
			continue
		}
		visited[pkgSum.PkgPath] = struct{}{}

		pkg, err := resolvePackage(pkgSum)
		if err != nil {
			pkg = &Package{
				ImportPath: pkgSum.PkgPath,
				Dir:        pkgSum.Root,
				Match:      pkgSum.Match,
			}
			pkg.Errors = append(pkg.Errors, fmt.Errorf("failed to resolve package %q: %w", pkgSum.PkgPath, err))
		}

		// TODO: what about TestImports
		for _, imp := range pkg.Imports {
			pile = append(pile, &PackageSummary{
				PkgPath: imp,
			})
		}

		cache[pkg.ImportPath] = pkg
		list = append(list, pkg)
	}

	for _, pkg := range list {
		if len(pkg.Errors) > 0 {
			continue
		}
		// TODO: this could be optimized
		pkg.Deps, err = listDeps(pkg.ImportPath, cache)
		if err != nil {
			pkg.Errors = append(pkg.Errors, err)
		}
	}

	return list, errors.Join(errs...)
}

func listDeps(target string, pkgs map[string]*Package) ([]string, error) {
	deps := []string{}
	err := listDepsRecursive(target, target, pkgs, &deps, make(map[string]struct{}))
	return deps, err
}

func listDepsRecursive(rootTarget string, target string, pkgs map[string]*Package, deps *[]string, visited map[string]struct{}) error {
	if _, ok := visited[target]; ok {
		return nil
	}
	pkg, ok := pkgs[target]
	if !ok {
		return fmt.Errorf("%s not found in cache", target)
	}
	visited[target] = struct{}{}
	var errs []error
	for _, imp := range pkg.Imports {
		err := listDepsRecursive(rootTarget, imp, pkgs, deps, visited)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if target != rootTarget {
		(*deps) = append(*deps, target)
	}
	return errors.Join(errs...)
}

func resolvePackage(meta *PackageSummary) (*Package, error) {
	if meta.Root == "" {
		if !strings.ContainsRune(meta.PkgPath, '.') {
			return resolveStdlib(meta)
		} else {
			return resolveRemote(meta)
		}
	}

	if meta.Module == nil {
		return nil, errors.New("unexpected nil module")
	}

	return fillPackage(meta)
}

func resolveStdlib(ometa *PackageSummary) (*Package, error) {
	meta := *ometa
	gnoRoot, err := gnoenv.GuessRootDir()
	if err != nil {
		return nil, fmt.Errorf("failed to guess gno root dir: %w", err)
	}
	parts := strings.Split(meta.PkgPath, "/")
	meta.Root = filepath.Join(append([]string{gnoRoot, "gnovm", "stdlibs"}, parts...)...)
	return fillPackage(&meta)
}

// Does not fill deps
func resolveRemote(ometa *PackageSummary) (*Package, error) {
	meta := *ometa

	modCache := filepath.Join(gnoenv.HomeDir(), "pkg", "mod")
	meta.Root = filepath.Join(modCache, meta.PkgPath)
	if err := DownloadModule(meta.PkgPath, meta.Root); err != nil {
		return nil, fmt.Errorf("failed to download module %q: %w", meta.PkgPath, err)
	}
	modDir, err := findModDir(meta.Root)
	if os.IsNotExist(err) {
		return nil, errors.New("failed to clone mod")
	} else if err != nil {
		return nil, fmt.Errorf("failed to find parent module: %w", err)
	}
	modFilePath := filepath.Join(modDir, ModfileName)
	modFileBytes, err := os.ReadFile(modFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read modfile: %w", err)
	}
	modFile, err := modfile.ParseLax(modFilePath, modFileBytes, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to parse modfile: %w", err)
	}
	meta.Module = &Module{
		Path:   modFile.Module.Mod.Path,
		Dir:    modDir,
		GnoMod: modFilePath,
	}

	pkg, err := fillPackage(&meta)
	if err != nil {
		return nil, fmt.Errorf("failed to fill Package %q: %w", meta.PkgPath, err)
	}

	return pkg, nil
}

func fillPackage(meta *PackageSummary) (*Package, error) {
	fsFiles := []string{}
	files := []string{}
	fsTestFiles := []string{}
	testFiles := []string{}
	fsFiletestFiles := []string{}
	filetestFiles := []string{}

	pkgDir := meta.Root
	pkgPath := meta.PkgPath

	dir, err := os.ReadDir(pkgDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read module files list: %w", err)
	}
	for _, entry := range dir {
		if entry.IsDir() {
			continue
		}

		fileName := entry.Name()
		if IsGnoTestFile(fileName) {
			fsTestFiles = append(fsTestFiles, filepath.Join(pkgDir, fileName))
			testFiles = append(testFiles, fileName)
		} else if IsGnoFiletestFile(fileName) {
			fsFiletestFiles = append(fsFiletestFiles, filepath.Join(pkgDir, fileName))
			filetestFiles = append(filetestFiles, fileName)
		} else if IsGnoFile(fileName) {
			fsFiles = append(fsFiles, filepath.Join(pkgDir, fileName))
			files = append(files, fileName)
		}
	}
	name, imports, err := resolveNameAndImports(fsFiles)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve name and imports for %q: %w", pkgPath, err)
	}
	_, testImports, err := resolveNameAndImports(fsTestFiles)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve test name and imports for %q: %w", pkgPath, err)
	}
	_, filetestImports, err := resolveNameAndImports(fsFiletestFiles)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve filetest name and imports for %q: %w", pkgPath, err)
	}

	module := Module{}
	if meta.Module != nil {
		module = *meta.Module
	}

	return &Package{
		ImportPath:       pkgPath,
		Dir:              pkgDir,
		Name:             name,
		Module:           module,
		Match:            meta.Match,
		GnoFiles:         files,
		Imports:          imports,
		TestGnoFiles:     testFiles,
		TestImports:      testImports,
		FiletestGnoFiles: filetestFiles,
		FiletestImports:  filetestImports,
	}, nil
}

func DownloadModule(pkgPath string, dst string) error {
	modFilePath := filepath.Join(dst, ModfileName)
	if _, err := os.Stat(modFilePath); os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, "gno: downloading", pkgPath)

		// create client from pkgpath
		parts := strings.Split(pkgPath, "/")
		if len(parts) < 1 {
			return fmt.Errorf("bad pkg path %q", pkgPath)
		}
		rpcURL := (&url.URL{
			Scheme: "https",
			Host:   "rpc." + parts[0] + ":443",
		}).String()
		tmClient, err := client.NewHTTPClient(rpcURL)
		if err != nil {
			return fmt.Errorf("failed to instantiate tm2 client with remote %q: %w", rpcURL, err)
		}
		client := gnoclient.Client{RPCClient: tmClient}

		// fetch files
		data, _, err := client.QFile(pkgPath)
		if err != nil {
			return fmt.Errorf("failed to query files list for pkg %q: %w", pkgPath, err)
		}
		if err := os.MkdirAll(dst, 0744); err != nil {
			return fmt.Errorf("failed to create cache dir for %q at %q: %w", pkgPath, dst, err)
		}
		files := strings.Split(string(data), "\n")
		for _, file := range files {
			filePath := path.Join(pkgPath, file)
			data, _, err := client.QFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to query package file %q: %w", filePath, err)
			}
			dst := filepath.Join(dst, file)
			if err := os.WriteFile(dst, []byte(data), 0644); err != nil {
				return fmt.Errorf("failed to write file at %q: %w", dst, err)
			}
		}

		// write gno.mod
		if err := os.WriteFile(modFilePath, []byte("module "+pkgPath+"\n"), 0644); err != nil {
			return fmt.Errorf("failed to write modfile at %q: %w", modFilePath, err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to stat downloaded module %q at %q: %w", pkgPath, dst, err)
	}

	return nil
}

type Package struct {
	Dir              string   `json:",omitempty"`
	ImportPath       string   `json:",omitempty"`
	Name             string   `json:",omitempty"`
	Root             string   `json:",omitempty"`
	Module           Module   `json:",omitempty"`
	Match            []string `json:",omitempty"`
	GnoFiles         []string `json:",omitempty"`
	Imports          []string `json:",omitempty"`
	Deps             []string `json:",omitempty"`
	TestGnoFiles     []string `json:",omitempty"`
	TestImports      []string `json:",omitempty"`
	FiletestGnoFiles []string `json:",omitempty"`
	FiletestImports  []string `json:",omitempty"`
	Errors           []error  `json:",omitempty"`
}

func (p *Package) MemPkg() (*std.MemPackage, error) {
	allFiles := append(p.GnoFiles, p.TestGnoFiles...)
	allFiles = append(allFiles, p.FiletestGnoFiles...)
	files := make([]*std.MemFile, len(allFiles))
	for i, f := range allFiles {
		body, err := os.ReadFile(filepath.Join(p.Dir, f))
		if err != nil {
			return nil, err
		}
		files[i] = &std.MemFile{
			Name: f,
			Body: string(body),
		}
	}
	return &std.MemPackage{
		Name:  p.Name,
		Path:  p.ImportPath,
		Files: files,
	}, nil
}

type Module struct {
	Path   string `json:",omitempty"`
	Dir    string `json:",omitempty"`
	GnoMod string `json:",omitempty"`
}

func resolveNameAndImports(gnoFiles []string) (string, []string, error) {
	names := map[string]int{}
	imports := map[string]struct{}{}
	bestName := ""
	bestNameCount := 0
	for _, srcPath := range gnoFiles {
		src, err := os.ReadFile(srcPath)
		if err != nil {
			return "", nil, fmt.Errorf("failed to read file %q: %w", srcPath, err)
		}
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, srcPath, src,
			// SkipObjectResolution -- unused here.
			// ParseComments -- so that they show up when re-building the AST.
			parser.SkipObjectResolution|parser.ImportsOnly)
		if err != nil {
			return "", nil, fmt.Errorf("parse: %w", err)
		}
		name := f.Name.String()
		names[name] += 1
		count := names[name]
		if count > bestNameCount {
			bestName = name
			bestNameCount = count
		}
		for _, imp := range f.Imports {
			importPath := imp.Path.Value
			// trim quotes
			if len(importPath) >= 2 {
				importPath = importPath[1 : len(importPath)-1]
			}
			imports[importPath] = struct{}{}
		}
	}
	importsSlice := make([]string, len(imports))
	i := 0
	for imp := range imports {
		importsSlice[i] = imp
		i++
	}
	return bestName, importsSlice, nil
}

func isRecursivePath(p string) bool {
	return strings.HasSuffix(p, recursiveSuffix) || p == "..."
}

func convertRecursivePathToDir(p string) string {
	if p == "..." {
		return "."
	}
	if !strings.HasSuffix(p, recursiveSuffix) {
		return p
	}
	return p[:len(p)-4]
}

func findModDir(dir string) (string, error) {
	dir = filepath.Clean(dir)

	potentialMod := filepath.Join(dir, ModfileName)

	if _, err := os.Stat(potentialMod); os.IsNotExist(err) {
		parent, file := filepath.Split(dir)
		if file == "" || (parent == "" && file == ".") {
			return "", os.ErrNotExist
		}
		return findModDir(parent)
	} else if err != nil {
		return "", err
	}

	return filepath.Clean(dir), nil
}
