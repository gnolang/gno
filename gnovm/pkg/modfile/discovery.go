package modfile

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/gnolang/gno/gnovm"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/packages"
)

// Module represents a Gno module, typically defined by a gno.toml file in its root directory.
type Module struct {
	Dir     string   // Absolute path to the module directory (where gno.toml is found).
	Name    string   // Module name (from Modfile.PkgPath).
	Imports []string // Direct imports of this module (from source analysis).
	// TODO: NontestImports []string
	Modfile Modfile  // The parsed gno.toml file content.
	Folders []Folder // File listings for the module's root and all its sub-directories.
}

// Folder represents the file contents of a specific directory.
type Folder struct {
	Dir              string   // Absolute path to the folder.
	GnoFiles         []string // .gno source files (basename only).
	TestGnoFiles     []string // _test.gno source files (basename only).
	FiletestGnoFiles []string // _filetest.gno source files (basename only).
	OtherFiles       []string // Other files in the directory (basename only).
	Subdirs          []string // Names of immediate sub-directories.
}

type (
	ModuleList       []*Module
	SortedModuleList []*Module
)

// Sort sorts the given modules by their dependencies.
func (ml ModuleList) Sort() (SortedModuleList, error) {
	visited := make(map[string]bool)
	onStack := make(map[string]bool)
	sortedModules := make([]*Module, 0, len(ml))

	for _, m := range ml {
		if err := visitModule(m, ml, visited, onStack, &sortedModules); err != nil {
			return nil, err
		}
	}
	return sortedModules, nil
}

// ListModules lists all Gno modules found by scanning the root directory for gno.toml files.
// It also populates the Folders field for each module with the content of its root and all sub-directories.
func ListModules(root string) (ModuleList, error) {
	var modules ModuleList

	// First, find all module roots (directories with gno.toml)
	type moduleRootInfo struct {
		path string
		mf   *Modfile
	}
	var rootsToProcess []moduleRootInfo

	// Phase 1: Identify module roots
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, errIn error) error {
		if errIn != nil {
			return errIn // Propagate errors from WalkDir itself
		}
		if !d.IsDir() {
			return nil // Only interested in directories for finding module roots
		}

		// Check for gno.toml in this directory.
		mf, readErr := ReadModfile(path, false) // findInParents is false
		if readErr == nil {
			// Found a directory that contains a gno.toml
			rootsToProcess = append(rootsToProcess, moduleRootInfo{path: path, mf: mf})
			// CRITICAL: There should be NO `return filepath.SkipDir` here.
			// We need to continue walking to find all potential module roots, even if nested.
		} else if !errors.Is(readErr, ErrModfileNotFound) {
			// An error other than "not found" occurred trying to read/parse a gno.toml
			return fmt.Errorf("error processing potential modfile in %s: %w", path, readErr)
		}
		return nil // Continue walking in all cases (unless a real error above was returned)
	})

	if err != nil {
		return nil, fmt.Errorf("error during module root identification: %w", err)
	}

	// Phase 2: For each identified module root, build the Module struct, including all its folders.
	for _, rootInfo := range rootsToProcess {
		modulePath := rootInfo.path
		mf := rootInfo.mf

		// Determine imports (this part remains the same)
		pkgInfo, err := gnolang.ReadMemPackage(modulePath, mf.PkgPath)
		if err != nil {
			pkgInfo = &gnovm.MemPackage{Name: filepath.Base(mf.PkgPath)}
		}
		importsMap, err := packages.Imports(pkgInfo, nil)
		if err != nil {
			importsMap = nil
		}
		importsRaw := importsMap.Merge(packages.FileKindPackageSource, packages.FileKindTest, packages.FileKindXTest)
		imports := make([]string, 0, len(importsRaw))
		for _, imp := range importsRaw {
			if imp.PkgPath != mf.PkgPath && !gnolang.IsStdlib(imp.PkgPath) {
				imports = append(imports, imp.PkgPath)
			}
		}

		var moduleFolders []Folder
		// Scan the module root and all its subdirectories to populate Folders
		walkErr := filepath.WalkDir(modulePath, func(currentPath string, d fs.DirEntry, walkErrIn error) error {
			if walkErrIn != nil {
				return walkErrIn
			}
			if !d.IsDir() {
				return nil
			}

			// Ensure we are within the identified modulePath.
			// This check is mostly to be robust if modulePath was relative or complex,
			// though modulePath from Phase 1 should be absolute and clean.
			if !strings.HasPrefix(currentPath, modulePath) && currentPath != modulePath {
				return filepath.SkipDir // Should not happen if modulePath is a canonical root
			}

			// If currentPath is a sub-directory of modulePath that *also* contains a gno.toml,
			// it's an inner module. We skip scanning it as a folder of the outer module,
			// as it will be (or has been) processed as its own separate module by rootsToProcess.
			// This check only applies if currentPath is not the modulePath itself.
			if currentPath != modulePath {
				if _, innerMfErr := ReadModfile(currentPath, false); innerMfErr == nil {
					return filepath.SkipDir
				} else if !errors.Is(innerMfErr, ErrModfileNotFound) {
					return fmt.Errorf("error checking for inner modfile at %s: %w", currentPath, innerMfErr)
				}
			}

			folder, scanErr := scanDirectoryContents(currentPath)
			if scanErr != nil {
				return fmt.Errorf("error scanning directory %s for module %s: %w", currentPath, mf.PkgPath, scanErr)
			}
			moduleFolders = append(moduleFolders, folder)
			return nil
		})

		if walkErr != nil {
			// It might be better to log this and continue, or collect errors,
			// rather than failing the entire ListModules operation. For now, fail fast.
			return nil, fmt.Errorf("error scanning folders for module %s at %s: %w", mf.PkgPath, modulePath, walkErr)
		}

		modules = append(modules, &Module{
			Dir:     modulePath,
			Name:    mf.PkgPath,
			Modfile: *mf,
			Imports: imports,
			Folders: moduleFolders,
		})
	}

	return modules, nil
}

// HasCode returns true if the folder contains any Gno source files.
func (f *Folder) HasCode() bool {
	if f == nil {
		return false
	}
	return len(f.GnoFiles) > 0
}

// HasTests returns true if the folder contains any Gno test files (_test.gno or _filetest.gno).
func (f *Folder) HasTests() bool {
	if f == nil {
		return false
	}
	return len(f.TestGnoFiles) > 0 || len(f.FiletestGnoFiles) > 0
}

// HasCode returns true if any folder within the module contains Gno source files.
func (m *Module) HasCode() bool {
	if m == nil {
		return false
	}
	for _, f := range m.Folders {
		if f.HasCode() {
			return true
		}
	}
	return false
}

// HasTests returns true if any folder within the module contains Gno test files.
func (m *Module) HasTests() bool {
	if m == nil {
		return false
	}
	for _, f := range m.Folders {
		if f.HasTests() {
			return true
		}
	}
	return false
}

// scanDirectoryContents creates a Folder struct for the given directory path by reading its direct entries.
func scanDirectoryContents(dirPath string) (Folder, error) {
	folder := Folder{Dir: dirPath}
	dentries, err := os.ReadDir(dirPath)
	if err != nil {
		return folder, fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	for _, de := range dentries {
		name := de.Name()
		if de.IsDir() {
			folder.Subdirs = append(folder.Subdirs, name)
			// This function only processes files and lists immediate subdirs in the current directory.
			// Recursive scanning for sub-directory *contents* (i.e., creating Folder structs for subdirs)
			// is handled by the caller (e.g., ListModules using filepath.WalkDir).
			continue
		}

		// It's a file, proceed with file type categorization
		if strings.HasSuffix(name, ".gno") {
			if strings.HasSuffix(name, "_filetest.gno") {
				folder.FiletestGnoFiles = append(folder.FiletestGnoFiles, name)
			} else if strings.HasSuffix(name, "_test.gno") {
				folder.TestGnoFiles = append(folder.TestGnoFiles, name)
			} else {
				folder.GnoFiles = append(folder.GnoFiles, name)
			}
		} else if name != "gno.toml" { // Exclude gno.toml itself from OtherFiles
			folder.OtherFiles = append(folder.OtherFiles, name)
		}
	}
	return folder, nil
}

// visitModule is a helper for the Sort method.
func visitModule(module *Module, modules []*Module, visited, onStack map[string]bool, sortedModules *[]*Module) error {
	if onStack[module.Name] {
		return fmt.Errorf("cycle detected: %s", module.Name)
	}
	if visited[module.Name] {
		return nil
	}

	visited[module.Name] = true
	onStack[module.Name] = true

	for _, imp := range module.Imports {
		found := false
		for _, m := range modules {
			if m.Name != imp {
				continue
			}
			if err := visitModule(m, modules, visited, onStack, sortedModules); err != nil {
				return err
			}
			found = true
			break
		}
		if !found {
			return fmt.Errorf("missing dependency '%s' for module '%s'", imp, module.Name)
		}
	}

	onStack[module.Name] = false
	*sortedModules = append(*sortedModules, module)
	return nil
}
