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
	ErrClientObjectNotFound    = errors.New("object not found")
	ErrClientRenderNotDeclared = errors.New("render function not declared")
	ErrClientBadRequest        = errors.New("bad request")
	ErrClientTimeout           = errors.New("RPC node request timeout")
	ErrClientResponse          = errors.New("RPC node response error")
	ErrClientResponseTooLarge  = errors.New("RPC node response too large")
)

// maxRPCResponseSize caps every per-query response from the RPC node.
// A legit realm package + qobject_json response is well under 1 MB; 8 MB
// is a generous headroom that still bounds memory amplification when a
// misbehaving/compromised node ships a multi-MB amino blob.
const maxRPCResponseSize = 8 << 20 // 8 MiB

// defaultMaxConcurrentRPC caps in-flight outbound RPCs per rpcClient
// when AppConfig.MaxConcurrentRPC is left at zero. Under the lazy-
// preview model, one viewport can fire dozens of fragment requests in
// parallel; without this gate they would saturate the chain node. 32
// keeps chain-side load tight while still absorbing a typical viewport
// burst — queueing past that is preferable to fanning out unbounded
// work to the node. Operators can tune via AppConfig when the chain
// node has more headroom (or less).
const defaultMaxConcurrentRPC = 32

// acquireRPCSlot blocks until a slot is free or ctx is cancelled, then
// returns a release function that frees the slot. Always returns a
// non-nil release fn so callers can defer unconditionally.
func acquireRPCSlot(ctx context.Context, slots chan struct{}) (func(), error) {
	select {
	case slots <- struct{}{}:
		return func() { <-slots }, nil
	case <-ctx.Done():
		return func() {}, ctx.Err()
	}
}

// checkResponseSize enforces maxRPCResponseSize on a raw qres.Data slice.
// The wrapped error preserves errors.Is matching against the sentinel
// while exposing the size in logs for forensics.
func checkResponseSize(data []byte) error {
	if len(data) > maxRPCResponseSize {
		return fmt.Errorf("%w: %d bytes (max %d)", ErrClientResponseTooLarge, len(data), maxRPCResponseSize)
	}
	return nil
}

type FileMeta struct {
	Lines  int
	SizeKB float64
}

type ClientAdapter interface {
	// Realm fetch the content of a realm from a given path and
	// return the data.
	Realm(ctx context.Context, path, args string) ([]byte, error) // raw Render() bytes

	// File fetches the source file from a given package path and
	// filename. `height = 0` queries the latest block; positive values
	// pin the query to that historical height — required for the
	// state explorer's time-travel mode to render source consistent
	// with the value snapshot.
	File(ctx context.Context, path, filename string, height int64) ([]byte, FileMeta, error)

	// ListFiles lists all source files available in a specified
	// package path. `height = 0` queries the latest block; non-zero
	// pins the listing to a historical block so time-travel views
	// stay consistent between file names and contents.
	ListFiles(ctx context.Context, path string, height int64) ([]string, error)

	// QueryPath list any path given the specified prefix
	ListPaths(ctx context.Context, prefix string, limit int) ([]string, error)

	// Doc retrieves the JSON doc suitable for printing from a
	// specified package path. `height = 0` queries the latest block;
	// any positive value pins the query to that historical height.
	Doc(ctx context.Context, path string, height int64) (*doc.JSONDocumentation, error)

	// StatePkg retrieves the root state tree for a package. `height
	// = 0` queries the latest block; any positive value pins the
	// query to that historical height (time-travel).
	StatePkg(ctx context.Context, path string, height int64) ([]byte, error)

	// StateObject retrieves the children of an object by ObjectID at
	// the given block height (0 for latest).
	StateObject(ctx context.Context, oid string, height int64) ([]byte, error)

	// StateType retrieves a type definition by TypeID at the given
	// block height (0 for latest).
	StateType(ctx context.Context, typeId string, height int64) ([]byte, error)
}

type rpcClient struct {
	domain   string
	logger   *slog.Logger
	client   *client.RPCClient
	rpcSlots chan struct{}
}

var _ ClientAdapter = (*rpcClient)(nil)

// NewRPCClientAdapter creates a new instance of rpcClient.
// maxConcurrentRPC ≤ 0 ⇒ defaultMaxConcurrentRPC; positive values cap
// in-flight outbound RPCs at that count.
func NewRPCClientAdapter(logger *slog.Logger, cli *client.RPCClient, domain string, maxConcurrentRPC int) ClientAdapter {
	if maxConcurrentRPC <= 0 {
		maxConcurrentRPC = defaultMaxConcurrentRPC
	}
	return &rpcClient{
		logger:   logger,
		domain:   domain,
		client:   cli,
		rpcSlots: make(chan struct{}, maxConcurrentRPC),
	}
}

// RenderRealm renders the content of a realm from a given path
// and arguments into the provided writer. It uses Goldmark for
// Markdown processing to generate HTML content.
func (c *rpcClient) Realm(ctx context.Context, path, args string) ([]byte, error) {
	const qpath = "vm/qrender"

	path = strings.Trim(path, "/")
	data := fmt.Sprintf("%s/%s:%s", c.domain, path, args)

	return c.query(ctx, qpath, []byte(data), 0)
}

// SourceFile fetches and writes the source file from a given
// package path and file name to the provided writer. It uses
// Chroma for syntax highlighting or Raw style source.
func (c *rpcClient) File(ctx context.Context, path, fileName string, height int64) (out []byte, meta FileMeta, err error) {
	const qpath = "vm/qfile"

	fileName = strings.TrimSpace(fileName)
	if fileName == "" {
		return nil, meta, errors.New("empty filename given") // XXX: Consider creating a specific error variable
	}

	// XXX: Consider moving this into gnoclient
	fullPath := gopath.Join(c.domain, strings.Trim(path, "/"), fileName)

	source, err := c.query(ctx, qpath, []byte(fullPath), height)
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
func (c *rpcClient) ListFiles(ctx context.Context, path string, height int64) ([]string, error) {
	const qpath = "vm/qfile"

	// XXX: Consider moving this into gnoclient
	pkgPath := strings.Trim(path, "/")
	fullPath := fmt.Sprintf("%s/%s", c.domain, pkgPath)
	res, err := c.query(ctx, qpath, []byte(fullPath), height)
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
	res, err := c.query(ctx, qpath, []byte(prefix), 0)
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
// specified package path. `height = 0` queries the latest block.
func (c *rpcClient) Doc(ctx context.Context, pkgPath string, height int64) (*doc.JSONDocumentation, error) {
	const qpath = "vm/qdoc"

	args := fmt.Sprintf("%s/%s", c.domain, strings.Trim(pkgPath, "/"))
	res, err := c.query(ctx, qpath, []byte(args), height)
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

// StatePkg retrieves root state tree for a package via vm/qpkg_json.
// `height = 0` queries the latest block; positive value pins to that
// historical height (time-travel).
func (c *rpcClient) StatePkg(ctx context.Context, path string, height int64) ([]byte, error) {
	const qpath = "vm/qpkg_json"

	path = strings.Trim(path, "/")
	data := fmt.Sprintf("%s/%s", c.domain, path)
	return c.query(ctx, qpath, []byte(data), height)
}

// StateObject retrieves an object by ObjectID via vm/qobject_json
// at the given block height (0 for latest).
func (c *rpcClient) StateObject(ctx context.Context, oid string, height int64) ([]byte, error) {
	const qpath = "vm/qobject_json"

	return c.query(ctx, qpath, []byte(oid), height)
}

// StateType retrieves a type definition by TypeID via vm/qtype_json
// at the given block height (0 for latest).
func (c *rpcClient) StateType(ctx context.Context, typeId string, height int64) ([]byte, error) {
	const qpath = "vm/qtype_json"

	return c.query(ctx, qpath, []byte(typeId), height)
}

// query sends a query to the RPC client and returns the response
// data. `height = 0` uses the latest block; any positive value pins
// the query to that historical height via ABCIQueryWithOptions.
func (c *rpcClient) query(ctx context.Context, qpath string, data []byte, height int64) ([]byte, error) {
	// Debug, not Info: `data` carries attacker-supplied OID/TID/file —
	// at hot-path rates this would amplify log volume by an order of
	// magnitude. Failures still log at Warn/Error below.
	c.logger.Debug("querying node", "path", qpath, "data", string(data), "height", height)

	// Bound concurrent outbound RPCs process-wide so gnoweb cannot
	// hammer the chain node under HTTP burst — orthogonal to any
	// HTTP-level rate limit (gnoweb's or nginx's), which sees neither
	// the fan-out nor the per-request RPC count.
	release, err := acquireRPCSlot(ctx, c.rpcSlots)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, ErrClientTimeout
		}
		return nil, err
	}
	defer release()

	start := time.Now()
	opts := client.DefaultABCIQueryOptions
	if height > 0 {
		opts.Height = height
	}
	qres, err := c.client.ABCIQueryWithOptions(ctx, qpath, data, opts)
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
			// Don't leak transport URL / endpoint detail into the wrapped
			// error — the typed sentinel + qpath via logger is enough.
			return nil, ErrClientTimeout
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
		if sizeErr := checkResponseSize(qres.Response.Data); sizeErr != nil {
			c.logger.Error("RPC response exceeded size cap",
				"path", qpath, "size", len(qres.Response.Data))
			return nil, sizeErr
		}
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
	case errors.Is(qerr, vm.ObjectNotFoundError{}):
		c.logger.Warn("object not found",
			"path", qpath,
			"data", string(data),
			"error", qres.Response.Error,
		)
		return nil, ErrClientObjectNotFound
	case errors.Is(qerr, vm.InvalidExprError{}):
		// vm/qtype_json returns InvalidExprError("type not found: …") when
		// the TypeID is unknown. Map to the same not-found sentinel so the
		// state explorer renders a 404 (the type page is gone) instead of
		// the generic 500.
		c.logger.Warn("type not found",
			"path", qpath,
			"data", string(data),
			"error", qres.Response.Error,
		)
		return nil, ErrClientObjectNotFound
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
