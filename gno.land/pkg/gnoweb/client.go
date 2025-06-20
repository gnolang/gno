package gnoweb

import (
	"errors"
)

var (
	ErrNotFound     = errors.New("path not found")
	ErrNoRenderDecl = errors.New("render func not declared")
)

type FileMeta struct {
	Lines  int
	SizeKB float64
}

type Client interface {
	Realm(path, args string) ([]byte, error)          // raw Render() bytes
	File(path, file string) ([]byte, FileMeta, error) // raw source
	ListFiles(path string) ([]string, error)
	ListPaths(prefix string, limit int) ([]string, error)
	Doc(path string) (*doc.JSONDocumentation, error)
}
