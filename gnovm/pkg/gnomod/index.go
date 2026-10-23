package gnomod

import (
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

// ModuleIndex maps a module path to the directory declaring it.
type ModuleIndex map[string]string

// IndexDir walks root and indexes every module under it by the module path its
// gnomod.toml declares, not by its directory name. A package whose directory
// does not repeat the version its module path carries is therefore still found
// at its module path. On a duplicate module path the first directory in lexical
// order wins, so a module in the tree shadows a copy of it placed deeper.
func IndexDir(root string) (ModuleIndex, error) {
	index := make(ModuleIndex, 512)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || (entry.Name() != "gnomod.toml" && entry.Name() != "gno.mod") {
			return nil
		}
		file, err := ParseFilepath(path)
		if err != nil || file.Module == "" {
			// Not resolvable. Whoever ends up loading this directory reports
			// the error with the context to make sense of it.
			return nil
		}
		if _, dup := index[file.Module]; !dup {
			index[file.Module] = filepath.Dir(path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return index, nil
}

type cachedIndex struct {
	once  sync.Once
	index ModuleIndex
	err   error
}

var indexCache sync.Map // cleaned root -> *cachedIndex

// IndexDirCached returns the module index of root, walked on first use and kept
// for the lifetime of the process. A module written under root after the first
// call is not seen; callers that need it fresh use [IndexDir].
func IndexDirCached(root string) (ModuleIndex, error) {
	value, _ := indexCache.LoadOrStore(filepath.Clean(root), &cachedIndex{})
	cached := value.(*cachedIndex)
	cached.once.Do(func() { cached.index, cached.err = IndexDir(root) })
	return cached.index, cached.err
}

// FindModuleDir returns the directory under root that provides modPath, or ""
// when nothing under root does.
//
// root/modPath answers when it declares modPath, or when it declares no module
// at all and so can only be identified by where it sits. Otherwise the module
// index of root answers, which is what resolves a package whose directory does
// not repeat the version its module path carries.
func FindModuleDir(root, modPath string) string {
	dir := filepath.Join(root, filepath.FromSlash(modPath))
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		file, err := ParseDir(dir)
		if err != nil || file.Module == modPath {
			return dir
		}
	}
	index, err := IndexDirCached(root)
	if err != nil {
		return ""
	}
	return index[modPath]
}
