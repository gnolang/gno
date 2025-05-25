package gnoweb

import (
	"errors"
	"io"

	md "github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/gnolang/gno/gnovm/pkg/doc"
)

var (
	ErrClientPathNotFound = errors.New("package not found")
	ErrRenderNotDeclared  = errors.New("render function not declared")
	ErrClientBadRequest   = errors.New("bad request")
	ErrClientResponse     = errors.New("node response error")
)

type FileMeta struct {
	Lines  int
	SizeKb float64
}

type RealmMeta struct {
	Toc md.Toc
}

// Renderer is an interface for rendering content from source.
type ContentRenderer interface {
	// Render renders the content of a source file and write it on the given writer.
	// It returns a Table of Contents (Toc) and an error if any occurs.
	Render(w io.Writer, u *weburl.GnoURL, src []byte) (md.Toc, error)
}

// WebClient is an interface for interacting with package and node resources.
type WebClient interface {
	// RenderRealm renders the content of a realm from a given path and
	// arguments into the giver `writer`. The method should ensures the rendered
	// content is safely handled and formatted.
	RenderRealm(w io.Writer, u *weburl.GnoURL, cr ContentRenderer) (*RealmMeta, error)

	// SourceFile fetches and writes the source file from a given
	// package path, file name and if raw. The method should ensures the source
	// file's content is safely handled and formatted.
	SourceFile(w io.Writer, pkgPath, fileName string, isRaw bool) (*FileMeta, error)

	// QueryPath list any path given the specified prefix
	QueryPaths(prefix string, limit int) ([]string, error)

	// Doc retrieves the JSON doc suitable for printing from a
	// specified package path.
	Doc(path string) (*doc.JSONDocumentation, error)

	// Sources lists all source files available in a specified
	// package path.
	Sources(path string) ([]string, error)
}
