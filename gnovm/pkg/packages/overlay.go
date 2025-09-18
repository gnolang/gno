package packages

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

type overlayFS struct {
	files map[string][]byte
	root  string
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
	overlayName := name
	if filepath.IsAbs(overlayName) {
		rel, err := filepath.Rel(o.root, name) // QUESTION: should use path lib or filepath lib?
		if err != nil {
			return nil, err
		}
		overlayName = rel
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

var _ fs.StatFS = (*overlayFS)(nil)

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
