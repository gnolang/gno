package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
)

// WalkDir is a forked version of [filepath.WalkDir].
// Special case errors are removed (SkipDir/SkipAll), and the walk does
// a first pass through the files and then a second pass through the directories.
func WalkDir(root string, fn fs.WalkDirFunc) error {
	info, err := os.Lstat(root)
	if err != nil {
		err = fn(root, nil, err)
	} else {
		err = walkDir(root, fs.FileInfoToDirEntry(info), fn)
	}
	return err
}

// walkDir recursively descends path, calling walkDirFn.
func walkDir(path string, d fs.DirEntry, walkDirFn fs.WalkDirFunc) error {
	if err := walkDirFn(path, d, nil); err != nil {
		return err
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		// Second call, to report ReadDir error.
		err = walkDirFn(path, d, err)
		if err != nil {
			return err
		}
	}

	// First pass, for files
	for _, f := range entries {
		if !f.IsDir() {
			fullPath := filepath.Join(path, f.Name())
			if err := walkDirFn(fullPath, f, nil); err != nil {
				return err
			}
		}
	}
	// Second pass, for directories
	for _, d := range entries {
		if d.IsDir() {
			fullPath := filepath.Join(path, d.Name())
			if err := walkDir(fullPath, d, walkDirFn); err != nil {
				return err
			}
		}
	}
	return nil
}

func mustUnquote(v string) string {
	s, err := strconv.Unquote(v)
	if err != nil {
		panic(fmt.Errorf("could not unquote import path literal: %s", v))
	}
	return s
}
