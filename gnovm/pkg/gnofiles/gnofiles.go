package gnofiles

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func DirEntryIsGnoFile(f fs.DirEntry) bool {
	name := f.Name()
	return !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".gno") && !f.IsDir()
}

func GnoFilesFromArgsRecursively(args []string) ([]string, error) {
	var paths []string

	for _, argPath := range args {
		info, err := os.Stat(argPath)
		if err != nil {
			return nil, fmt.Errorf("invalid file or package path: %w", err)
		}

		if !info.IsDir() {
			if DirEntryIsGnoFile(fs.FileInfoToDirEntry(info)) {
				paths = append(paths, ensurePathPrefix(argPath))
			}

			continue
		}

		err = walkDirForGnoDirs(argPath, func(path string) {
			dir := ensurePathPrefix(path)
			files, err := os.ReadDir(dir)
			if err != nil {
				return
			}
			for _, f := range files {
				if DirEntryIsGnoFile(f) {
					path := filepath.Join(dir, f.Name())
					paths = append(paths, ensurePathPrefix(path))
				}
			}
		})
		if err != nil {
			return nil, fmt.Errorf("unable to walk dir: %w", err)
		}
	}

	return paths, nil
}

func GnoDirsFromArgsRecursively(args []string) ([]string, error) {
	var paths []string

	for _, argPath := range args {
		info, err := os.Stat(argPath)
		if err != nil {
			return nil, fmt.Errorf("invalid file or package path: %w", err)
		}

		if !info.IsDir() {
			if DirEntryIsGnoFile(fs.FileInfoToDirEntry(info)) {
				paths = append(paths, ensurePathPrefix(argPath))
			}

			continue
		}

		// Gather package paths from the directory
		err = walkDirForGnoDirs(argPath, func(path string) {
			paths = append(paths, ensurePathPrefix(path))
		})
		if err != nil {
			return nil, fmt.Errorf("unable to walk dir: %w", err)
		}
	}

	return paths, nil
}

func GnoFilesFromArgs(args []string) ([]string, error) {
	var paths []string

	for _, argPath := range args {
		info, err := os.Stat(argPath)
		if err != nil {
			return nil, fmt.Errorf("invalid file or package path: %w", err)
		}

		if !info.IsDir() {
			if DirEntryIsGnoFile(fs.FileInfoToDirEntry(info)) {
				paths = append(paths, ensurePathPrefix(argPath))
			}
			continue
		}

		files, err := os.ReadDir(argPath)
		if err != nil {
			return nil, err
		}
		for _, f := range files {
			if DirEntryIsGnoFile(f) {
				path := filepath.Join(argPath, f.Name())
				paths = append(paths, ensurePathPrefix(path))
			}
		}
	}

	return paths, nil
}

func ensurePathPrefix(path string) string {
	if filepath.IsAbs(path) {
		return path
	}

	// cannot use path.Join or filepath.Join, because we need
	// to ensure that ./ is the prefix to pass to go build.
	// if not absolute.
	return "." + string(filepath.Separator) + path
}

func walkDirForGnoDirs(root string, addPath func(path string)) error {
	visited := make(map[string]struct{})

	walkFn := func(currPath string, f fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("%s: walk dir: %w", root, err)
		}

		if f.IsDir() || !DirEntryIsGnoFile(f) {
			return nil
		}

		parentDir := filepath.Dir(currPath)
		if _, found := visited[parentDir]; found {
			return nil
		}

		visited[parentDir] = struct{}{}

		addPath(parentDir)

		return nil
	}

	return filepath.WalkDir(root, walkFn)
}

func GnoPackagesFromArgsRecursively(args []string) ([]string, error) {
	var paths []string

	for _, argPath := range args {
		info, err := os.Stat(argPath)
		if err != nil {
			return nil, fmt.Errorf("invalid file or package path: %w", err)
		}

		if !info.IsDir() {
			paths = append(paths, ensurePathPrefix(argPath))

			continue
		}

		// Gather package paths from the directory
		err = walkDirForGnoDirs(argPath, func(path string) {
			paths = append(paths, ensurePathPrefix(path))
		})
		if err != nil {
			return nil, fmt.Errorf("unable to walk dir: %w", err)
		}
	}

	return paths, nil
}
