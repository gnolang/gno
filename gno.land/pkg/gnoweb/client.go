package gnoweb

import (
	"context"
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
	ErrClientTimeout           = errors.New("RPC node request timeout")
	ErrClientResponse          = errors.New("RPC node response error")
)

type FileMeta struct {
	Lines  int
	SizeKB float64
}

type ClientAdapter interface {
	// Realm fetch the content of a realm from a given path and
	// return the data.
	Realm(ctx context.Context, path, args string) ([]byte, error) // raw Render() bytes

	// File fetche the source file from a given
	// package path and filename.
	File(ctx context.Context, path, filename string) ([]byte, FileMeta, error)

	// ListFiles lists all source files available in a specified
	// package path.
	ListFiles(ctx context.Context, path string) ([]string, error)

	// QueryPath list any path given the specified prefix
	ListPaths(ctx context.Context, prefix string, limit int) ([]string, error)

	// Doc retrieves the JSON doc suitable for printing from a
	// specified package path.
	Doc(ctx context.Context, path string) (*doc.JSONDocumentation, error)
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
func (c *rpcClient) Realm(ctx context.Context, path, args string) ([]byte, error) {
	const qpath = "vm/qrender"

	path = strings.Trim(path, "/")
	data := fmt.Sprintf("%s/%s:%s", c.domain, path, args)

	return c.query(ctx, qpath, []byte(data))
}

// SourceFile fetches and writes the source file from a given
// package path and file name to the provided writer. It uses
// Chroma for syntax highlighting or Raw style source.
func (c *rpcClient) File(ctx context.Context, path, fileName string) (out []byte, meta FileMeta, err error) {
	const qpath = "vm/qfile"

	fileName = strings.TrimSpace(fileName)
	if fileName == "" {
		return nil, meta, errors.New("empty filename given") // XXX: Consider creating a specific error variable
	}

	// XXX: Consider moving this into gnoclient
	fullPath := gopath.Join(c.domain, strings.Trim(path, "/"), fileName)

	source, err := c.query(ctx, qpath, []byte(fullPath))
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
func (c *rpcClient) ListFiles(ctx context.Context, path string) ([]string, error) {
	const qpath = "vm/qfile"

	// XXX: Consider moving this into gnoclient
	pkgPath := strings.Trim(path, "/")
	fullPath := fmt.Sprintf("%s/%s", c.domain, pkgPath)
	res, err := c.query(ctx, qpath, []byte(fullPath))
	if err != nil {
		return nil, err
	}

	files := strings.Split(strings.TrimSpace(string(res)), "\n")
	return files, nil
}

// Sources lists all source files available in a specified
// package path by querying the RPC client.
func (c *rpcClient) ListPaths(ctx context.Context, prefix string, limit int) ([]string, error) {
	const qpath = "vm/qpaths"

	// XXX: Consider moving this into gnoclient
	res, err := c.query(ctx, qpath, []byte(prefix))
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
func (c *rpcClient) Doc(ctx context.Context, pkgPath string) (*doc.JSONDocumentation, error) {
	const qpath = "vm/qdoc"

	args := fmt.Sprintf("%s/%s", c.domain, strings.Trim(pkgPath, "/"))
	res, err := c.query(ctx, qpath, []byte(args))
	if err != nil {
		return nil, fmt.Errorf("unable to query qdoc for %s: %w", pkgPath, err)
	}

	jdoc := &doc.JSONDocumentation{}
	if err := amino.UnmarshalJSON(res, jdoc); err != nil {
		c.logger.Warn("unable to unmarshal qdoc, client is probably outdated")
		return nil, fmt.Errorf("unable to unmarshal qdoc for %s: %w", pkgPath, err)
	}

	return jdoc, nil
}

// query sends a query to the RPC client and returns the response
// data.
func (c *rpcClient) query(ctx context.Context, qpath string, data []byte) ([]byte, error) {
	c.logger.Info("querying node", "path", qpath, "data", string(data))

	start := time.Now()
	qres, err := c.client.ABCIQuery(ctx, qpath, data)
	took := time.Since(start)
	if err != nil {
		// Unexpected error from the RPC client itself
		c.logger.Error("query request failed",
			"path", qpath,
			"data", string(data),
			"error", err,
			"took", took,
		)

		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("%w: %s", ErrClientTimeout, err.Error())
		}

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
