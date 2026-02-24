package packages

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

type overlayFS struct {
	dirs  map[string][]string
	files map[string][]byte
	root  string
}

func newOverlayFS(files map[string][]byte, root string) *overlayFS {
	ofs := &overlayFS{files: files, root: root}
	ofs.fillDirs()
	return ofs
}

// ReadDir implements fs.ReadDirFS.
func (o *overlayFS) ReadDir(name string) ([]fs.DirEntry, error) {
	overlayName, err := o.overlayName(name)
	if err != nil {
		return nil, err
	}
	overlayEntries := o.dirs[overlayName]
	osEntries, err := os.ReadDir(name)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	fsEntries := append([]fs.DirEntry{}, osEntries...)
	for _, entryName := range overlayEntries {
		osIdx := slices.IndexFunc(osEntries, func(entry os.DirEntry) bool {
			return entry.Name() == entryName
		})
		_, entryIsDir := o.dirs[filepath.Join(name, entryName)]
		entry := &overlayDirEntry{name: entryName, isDir: entryIsDir}
		if osIdx == -1 {
			fsEntries = append(fsEntries, entry)
			continue
		}
		fsEntries[osIdx] = entry
	}
	slices.SortFunc(fsEntries, func(a, b fs.DirEntry) int {
		return strings.Compare(a.Name(), b.Name())
	})
	return fsEntries, nil
}

// Stat implements fs.StatFS.
func (o *overlayFS) Stat(name string) (fs.FileInfo, error) {
	f, err := o.Open(name)
	if err != nil {
		return nil, err
	}
	return f.Stat()
}

// Open implements fs.FS.
func (o *overlayFS) Open(name string) (fs.File, error) {
	overlayName, err := o.overlayName(name)
	if err != nil {
		return nil, err
	}
	body, ok := o.files[overlayName]
	if ok {
		return &overlayFile{
			buf:  bytes.NewBuffer(body),
			size: int64(len(body)),
		}, nil
	}
	return os.Open(name)
}

func (ofs *overlayFS) fillDirs() {
	ofs.dirs = map[string][]string{}

	for file := range ofs.files {
		var (
			prefix = file
			name   string
		)
		for prefix != "." {
			prefix, name = filepath.Split(prefix) // QUESTION: should use path lib or filepath lib?
			prefix = filepath.Clean(prefix)
			if _, hasDir := ofs.dirs[prefix]; !hasDir {
				ofs.dirs[prefix] = []string{}
			}
			if !slices.Contains(ofs.dirs[prefix], name) {
				ofs.dirs[prefix] = append(ofs.dirs[prefix], name)
			}
		}
	}
}

func (ofs *overlayFS) overlayName(name string) (string, error) {
	overlayName := name
	if filepath.IsAbs(overlayName) {
		rel, err := filepath.Rel(ofs.root, name) // QUESTION: should use path lib or filepath lib?
		if err != nil {
			return "", err
		}
		overlayName = rel
	}
	return filepath.Clean(overlayName), nil
}

var _ fs.StatFS = (*overlayFS)(nil)
var _ fs.ReadDirFS = (*overlayFS)(nil)

type overlayFile struct {
	buf  *bytes.Buffer
	size int64
}

// Close implements fs.File.
func (o *overlayFile) Close() error {
	return nil
}

// Read implements fs.File.
func (o *overlayFile) Read(bz []byte) (int, error) {
	return o.buf.Read(bz)
}

// Stat implements fs.File.
func (o *overlayFile) Stat() (fs.FileInfo, error) {
	return &overlayFileInfo{size: o.size}, nil
}

var _ fs.File = (*overlayFile)(nil)

type overlayFileInfo struct {
	isDir bool
	size  int64
}

// IsDir implements fs.FileInfo.
func (o *overlayFileInfo) IsDir() bool {
	return o.isDir
}

// ModTime implements fs.FileInfo.
func (o *overlayFileInfo) ModTime() time.Time {
	panic("ModTime unimplemented")
}

// Mode implements fs.FileInfo.
func (o *overlayFileInfo) Mode() fs.FileMode {
	panic("Mode unimplemented")
}

// Name implements fs.FileInfo.
func (o *overlayFileInfo) Name() string {
	panic("Name unimplemented")
}

// Size implements fs.FileInfo.
func (o *overlayFileInfo) Size() int64 {
	return o.size
}

// Sys implements fs.FileInfo.
func (o *overlayFileInfo) Sys() any {
	panic("Sys unimplemented")
}

var _ fs.FileInfo = (*overlayFileInfo)(nil)

type overlayDirEntry struct {
	isDir bool
	name  string
}

// Info implements fs.DirEntry.
func (o *overlayDirEntry) Info() (fs.FileInfo, error) {
	return &overlayFileInfo{isDir: true}, nil
}

// IsDir implements fs.DirEntry.
func (o *overlayDirEntry) IsDir() bool {
	return o.isDir
}

// Name implements fs.DirEntry.
func (o *overlayDirEntry) Name() string {
	return o.name
}

// Type implements fs.DirEntry.
func (o *overlayDirEntry) Type() fs.FileMode {
	panic("unimplemented")
}

var _ fs.DirEntry = (*overlayDirEntry)(nil)
