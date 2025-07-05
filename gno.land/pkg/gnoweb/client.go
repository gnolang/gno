package gnoweb

import (
	"errors"
	"fmt"
	"log/slog"
	gopath "path"
	"strings"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/doc"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
)

var (
	ErrClientPackageNotFound   = errors.New("package not found")
	ErrClientFileNotFound      = errors.New("file not found")
	ErrClientRenderNotDeclared = errors.New("render function not declared")
	ErrClientBadRequest        = errors.New("bad request")
	ErrClientResponse          = errors.New("node response error")
	ErrClientFile              = errors.New("unknown or invalid filename for path")
)

type FileMeta struct {
	Lines  int
	SizeKB float64
}

type ClientAdapter interface {
	// Realm fetch the content of a realm from a given path and
	// return the data.
	Realm(path, args string) ([]byte, error) // raw Render() bytes

	// File fetche the source file from a given
	// package path and filename.
	File(path, filename string) ([]byte, FileMeta, error)

	// Sources lists all source files available in a specified
	// package path.
	ListFiles(path string) ([]string, error)

	// QueryPath list any path given the specified prefix
	ListPaths(prefix string, limit int) ([]string, error)

	// Doc retrieves the JSON doc suitable for printing from a
	// specified package path.
	Doc(path string) (*doc.JSONDocumentation, error)
}

type rpcClient struct {
	domain string
	logger *slog.Logger
	client *client.RPCClient
}

var _ ClientAdapter = (*rpcClient)(nil)

// NewHTMLClient creates a new instance of WebClient.
// It requires a configured logger and WebClientConfig.
func NewRPCClientAdapter(logger *slog.Logger, cli *client.RPCClient, domain string) ClientAdapter {
	return &rpcClient{
		logger: logger,
		domain: domain,
		client: cli,
	}
}

// RenderRealm renders the content of a realm from a given path
// and arguments into the provided writer. It uses Goldmark for
// Markdown processing to generate HTML content.
func (c *rpcClient) Realm(path, args string) ([]byte, error) {
	const qpath = "vm/qrender"

	path = strings.Trim(path, "/")
	data := fmt.Sprintf("%s/%s:%s", c.domain, path, args)

	return c.query(qpath, []byte(data))
}

// SourceFile fetches and writes the source file from a given
// package path and file name to the provided writer. It uses
// Chroma for syntax highlighting or Raw style source.
func (c *rpcClient) File(path, fileName string) (out []byte, meta FileMeta, err error) {
	const qpath = "vm/qfile"

	fileName = strings.TrimSpace(fileName)
	if fileName == "" {
		return nil, meta, errors.New("empty filename given") // XXX: Consider creating a specific error variable
	}

	// XXX: Consider moving this into gnoclient
	fullPath := gopath.Join(c.domain, strings.Trim(path, "/"), fileName)

	source, err := c.query(qpath, []byte(fullPath))
	if err != nil {
		// XXX: this is a bit ugly, we should make the keeper return an
		// assertable error.
		if strings.Contains(err.Error(), "not available") {
			return nil, meta, ErrClientPackageNotFound
		}

		return nil, meta, err
	}

	meta = FileMeta{
		Lines:  strings.Count(string(source), "\n"),
		SizeKB: float64(len(source)) / 1024.0,
	}

	// Use raw syntax for source
	return source, meta, nil
}

// ListFiles lists all source files available in a specified
// package path by querying the RPC client.
func (c *rpcClient) ListFiles(path string) ([]string, error) {
	const qpath = "vm/qfile"

	// XXX: Consider moving this into gnoclient
	pkgPath := strings.Trim(path, "/")
	fullPath := fmt.Sprintf("%s/%s", c.domain, pkgPath)
	res, err := c.query(qpath, []byte(fullPath))
	if err != nil {
		return nil, err
	}

	files := strings.Split(strings.TrimSpace(string(res)), "\n")
	return files, nil
}

// Sources lists all source files available in a specified
// package path by querying the RPC client.
func (c *rpcClient) ListPaths(prefix string, limit int) ([]string, error) {
	const qpath = "vm/qpaths"

	// XXX: Consider moving this into gnoclient
	res, err := c.query(qpath, []byte(prefix))
	if err != nil {
		return nil, err
	}

	// update the paths to be relative to the root instead of the domain
	paths := strings.Split(strings.TrimSpace(string(res)), "\n")
	for i, path := range paths {
		paths[i] = strings.TrimPrefix(path, c.domain)
	}

	return paths, nil
}

// Doc retrieves the JSON doc suitable for printing from a
// specified package path.
func (c *rpcClient) Doc(pkgPath string) (*doc.JSONDocumentation, error) {
	const qpath = "vm/qdoc"

	args := fmt.Sprintf("%s/%s", c.domain, strings.Trim(pkgPath, "/"))
	res, err := c.query(qpath, []byte(args))
	if err != nil {
		return nil, fmt.Errorf("unable to query qdoc: %w", err)
	}

	jdoc := &doc.JSONDocumentation{}
	if err := amino.UnmarshalJSON(res, jdoc); err != nil {
		c.logger.Warn("unable to unmarshal qdoc, client is probably outdated")
		return nil, fmt.Errorf("unable to unmarshal qdoc: %w", err)
	}

	return jdoc, nil
}

// query sends a query to the RPC client and returns the response
// data.
func (c *rpcClient) query(qpath string, data []byte) ([]byte, error) {
	c.logger.Info("querying node", "path", qpath, "data", string(data))

	start := time.Now()
	qres, err := c.client.ABCIQuery(qpath, data)
	took := time.Since(start)
	if err != nil {
		// Unexpected error from the RPC client itself
		c.logger.Error("query request failed",
			"path", qpath,
			"data", string(data),
			"error", err,
			"took", took,
		)
		return nil, fmt.Errorf("%w: %s", ErrClientBadRequest, err.Error())
	}

	// Log the response at debug level for detailed tracing
	c.logger.Debug("query response received",
		"path", qpath,
		"data", string(data),
		"response_error", qres.Response.Error,
		"took", took,
	)

	qerr := qres.Response.Error
	if qerr == nil {
		return qres.Response.Data, nil
	}

	// Handle and log known error types
	switch {
	case errors.Is(qerr, vm.InvalidPkgPathError{}), errors.Is(qerr, vm.InvalidPackageError{}):
		c.logger.Warn("package not found",
			"path", qpath,
			"data", string(data),
			"error", qres.Response.Error,
		)
		return nil, ErrClientPackageNotFound
	case errors.Is(qerr, vm.InvalidFileError{}):
		c.logger.Warn("file not found",
			"path", qpath,
			"data", string(data),
			"error", qres.Response.Error,
		)
		return nil, ErrClientFileNotFound
	case errors.Is(qerr, vm.NoRenderDeclError{}):
		c.logger.Warn("render function not declared",
			"path", qpath,
			"data", string(data),
			"error", qres.Response.Error,
		)
		return nil, ErrClientRenderNotDeclared
	default:
	}

	// fallback on general error
	c.logger.Error("node response error",
		"path", qpath,
		"data", string(data),
		"error", qres.Response.Error,
	)
	return nil, fmt.Errorf("%w: %w", ErrClientResponse, qres.Response.Error)
}
