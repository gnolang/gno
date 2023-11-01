// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package fs

import (
	"errors"
)

// XXX: To use `Sort` package, we need to implement `internal/reflectile`, which I think we should avoid for now.
// So, I've implemented basic sorting, which we'll need to fix later.
type ByName []DirEntry

func (s ByName) Len() int           { return len(s) }
func (s ByName) Less(i, j int) bool { return s[i].Name() < s[j].Name() }
func (s ByName) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func quickSort(entries []DirEntry, low, high int) {
	if low >= high {
		return
	}

	pivot := entries[low]
	i, j := low, high
	for i < j {
		for i < j && entries[j].Name() >= pivot.Name() {
			j--
		}

		entries[i] = entries[j]
		for i < j && entries[i].Name() <= pivot.Name() {
			i++
		}

		entries[j] = entries[i]
	}

	entries[i] = pivot
	quickSort(entries, low, i-1)
	quickSort(entries, i+1, high)
}

func sortDirEntries(entries []DirEntry) {
	quickSort(entries, 0, len(entries)-1)
}

// ReadDirFS is the interface implemented by a file system
// that provides an optimized implementation of ReadDir.
type ReadDirFS interface {
	FS

	// ReadDir reads the named directory
	// and returns a list of directory entries sorted by filename.
	ReadDir(name string) ([]DirEntry, error)
}

// ReadDir reads the named directory
// and returns a list of directory entries sorted by filename.
//
// If fs implements ReadDirFS, ReadDir calls fs.ReadDir.
// Otherwise ReadDir calls fs.Open and uses ReadDir and Close
// on the returned file.
func ReadDir(fsys FS, name string) ([]DirEntry, error) {
	if fsys, ok := fsys.(ReadDirFS); ok {
		return fsys.ReadDir(name)
	}

	file, err := fsys.Open(name)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	dir, ok := file.(ReadDirFile)
	if !ok {
		return nil, &PathError{Op: "readdir", Path: name, Err: errors.New("not implemented")}
	}

	ls, err := dir.ReadDir(-1)
	if err != nil {
		return nil, err
	}

	sortDirEntries(ls)

	return ls, err
}

// dirInfo is a DirEntry based on a FileInfo.
type dirInfo struct {
	fileInfo FileInfo
}

func (d dirInfo) IsDir() bool {
	return d.fileInfo.IsDir()
}

func (d dirInfo) Typer() Typer {
	return d.fileInfo.Typer()
}

func (d dirInfo) Info() (FileInfo, error) {
	return d.fileInfo, nil
}

func (d dirInfo) Name() string {
	return d.fileInfo.Name()
}

// func (d dirInfo) String() string {
//
// }

func FileInfoToDirEntry(info FileInfo) DirEntry {
	if info == nil {
		return nil
	}

	return dirInfo{fileInfo: info}
}
