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
	"golang.org/x/mod/modfile"
)

const recursiveSuffix = string(os.PathSeparator) + "..."
const modFileBaseName = "gno.mod"

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

	potentialMod := filepath.Join(dir, modFileBaseName)

	if _, err := os.Stat(potentialMod); os.IsNotExist(err) {
		parent, file := filepath.Split(dir)
		if file == "" {
			return "", os.ErrNotExist
		}
		return findModDir(parent)
	} else if err != nil {
		return "", err
	}

	return filepath.Clean(dir), nil
}

func ListPackages(paths ...string) ([]string, error) {
	details, _, err := ListPackagesDetails(paths...)
	if err != nil {
		return nil, err
	}
	res := make([]string, len(details))
	for i, p := range details {
		res[i] = p.PkgPath
	}
	return res, nil
}

func ListPackagesDetails(paths ...string) ([]PackageDetails, *modfile.File, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get workdir: %w", err)
	}

	modDir, err := findModDir(wd)
	if os.IsNotExist(err) {
		return nil, nil, errors.New("gno.mod file not found in current directory or any parent directory")
	} else if err != nil {
		return nil, nil, fmt.Errorf("failed to find parent module: %w", err)
	}
	modFilePath := filepath.Join(modDir, modFileBaseName)
	modFileBytes, err := os.ReadFile(modFilePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read modfile: %w", err)
	}
	modFile, err := modfile.ParseLax(modFilePath, modFileBytes, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse modfile: %w", err)
	}

	pkgPath := modFile.Module.Mod.Path

	toVisit := []string{}
	for _, p := range paths {
		toVisit = append(toVisit, filepath.Clean(p))
	}
	visited := map[string]struct{}{}
	packages := []PackageDetails{}
	errs := []error{}

	for len(toVisit) > 0 {
		p := toVisit[0]
		visited[p] = struct{}{}
		toVisit = toVisit[1:]

		root := convertRecursivePathToDir(p)

		if !isRecursivePath(p) {
			if p != "." && strings.ContainsRune(p, '.') {
				packages = append(packages, PackageDetails{
					PkgPath: p,
					Remote:  true,
				})
				continue
			}
			p = path.Join(pkgPath, p)
			packages = append(packages, PackageDetails{
				PkgPath: p,
				Root:    root,
			})
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
				if _, ok := visited[dirPath]; !ok {
					toVisit = append(toVisit, dirPath)
				}
			}
			if !hasGnoFiles && IsGnoFile(fileName) {
				hasGnoFiles = true
			}
		}

		if hasGnoFiles {
			if _, ok := visited[root]; !ok {
				toVisit = append(toVisit, root)
			}
		}
	}

	return packages, modFile, errors.Join(errs...)
}

type PackageDetails struct {
	PkgPath string
	Root    string
	Remote  bool
}

func ResolvePackages(paths ...string) ([]*Package, error) {
	pkgs, modFile, err := ListPackagesDetails(paths...)
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}

	res := make([]*Package, len(pkgs))
	errs := []error{}
	for i, meta := range pkgs {
		if meta.Remote {
			res[i], err = ResolveRemote(meta.PkgPath)
			if err != nil {
				errs = append(errs, err)
			}
			continue
		}

		absRoot, err := filepath.Abs(meta.Root)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to find absolute root %q: %w", meta.Root, err))
		}

		pkgPath := meta.PkgPath
		pkg, err := fillPackage(pkgPath, absRoot, modFile)
		if err != nil {
			return nil, fmt.Errorf("failed to fill Package %q: %w", pkgPath, err)
		}

		res[i] = pkg
	}
	return res, errors.Join(errs...)
}

// Does not fill deps
func ResolveRemote(pkgPath string) (*Package, error) {
	modCache := filepath.Join(gnoenv.HomeDir(), "pkg", "mod")
	pkgDir := filepath.Join(modCache, pkgPath)
	if err := DownloadModule(pkgPath, pkgDir); err != nil {
		return nil, fmt.Errorf("failed to download module %q: %w", pkgPath, err)
	}

	pkg, err := fillPackage(pkgPath, pkgDir, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fill Package %q: %w", pkgPath, err)
	}

	return pkg, nil
}

func fillPackage(pkgPath, pkgDir string, modFile *modfile.File) (*Package, error) {
	fsFiles := []string{}
	modFiles := []string{}
	fsTestFiles := []string{}
	testFiles := []string{}
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
		} else if IsGnoFile(fileName) {
			fsFiles = append(fsFiles, filepath.Join(pkgDir, fileName))
			modFiles = append(modFiles, fileName)
		}
	}
	name, imports, err := resolveNameAndImports(fsFiles)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve name and imports for %q: %w", pkgPath, err)
	}
	_, testImports, err := resolveNameAndImports(fsTestFiles)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve name test imports for %q: %w", pkgPath, err)
	}

	module := Module{}
	if modFile != nil {
		module = Module{
			Path: modFile.Module.Mod.Path,
		}
	}

	return &Package{
		ImportPath:   pkgPath,
		Dir:          pkgDir,
		Name:         name,
		Module:       module,
		GnoFiles:     modFiles,
		Imports:      imports,
		TestGnoFiles: testFiles,
		TestImports:  testImports,
	}, nil
}

func DownloadModule(pkgPath string, dst string) error {
	modFilePath := filepath.Join(dst, modFileBaseName)
	if _, err := os.Stat(modFilePath); os.IsNotExist(err) {
		fmt.Println("gno: downloading", pkgPath)

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
		if err := os.WriteFile(modFilePath, []byte("package "+pkgPath+"\n"), 0644); err != nil {
			return fmt.Errorf("failed to write modfile at %q: %w", modFilePath, err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to stat downloaded module %q at %q: %w", pkgPath, dst, err)
	}

	return nil
}

func IsGnoTestFile(p string) bool {
	if !IsGnoFile(p) {
		return false
	}
	return strings.HasSuffix(p, "_test.gno") || strings.HasSuffix(p, "_filetest.gno")
}

type Package struct {
	Dir          string   `json:",omitempty"`
	ImportPath   string   `json:",omitempty"`
	Name         string   `json:",omitempty"`
	Root         string   `json:",omitempty"`
	Module       Module   `json:",omitempty"`
	Match        []string `json:",omitempty"`
	GnoFiles     []string `json:",omitempty"`
	Imports      []string `json:",omitempty"`
	Deps         []string `json:",omitempty"`
	TestGnoFiles []string `json:",omitempty"`
	TestImports  []string `json:",omitempty"`
}

type Module struct {
	Path      string `json:",omitempty"`
	Main      bool   `json:",omitempty"`
	Dir       string `json:",omitempty"`
	GoMod     string `json:",omitempty"`
	GoVersion string `json:",omitempty"`
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
