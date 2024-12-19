package gnoweb

import (
	"errors"
	"io"

	md "github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm" // for error types
)

var (
	ErrClientPathNotFound = errors.New("package path not found")
	ErrClientBadRequest   = errors.New("bad request")
	ErrClientResponse     = errors.New("node response error") // Corrected typo in "response"
)

type FileMeta struct {
	Lines  int
	SizeKb float64
}

type RealmMeta struct {
	*md.Toc
}

// WebClient is an interface for interacting with package ressources.
type WebClient interface {
	// RenderRealm renders the content of a realm from a given path and
	// arguments into the giver `writer`. The method should ensures the rendered
	// content is safely handled and formatted.
	RenderRealm(w io.Writer, path string, args string) (*RealmMeta, error)

	// SourceFile fetches and writes the source file from a given
	// package path and file name. The method should ensures the source
	// file's content is safely handled and formatted.
	SourceFile(w io.Writer, pkgPath, fileName string) (*FileMeta, error)

	// Functions retrieves a list of function signatures from a
	// specified package path.
	Functions(path string) ([]vm.FunctionSignature, error)

	// Sources lists all source files available in a specified
	// package path.
	Sources(path string) ([]string, error)
}
