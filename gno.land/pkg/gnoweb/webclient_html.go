package gnoweb

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"strings"

	md "github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm" // for error types
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

type HTMLWebClientConfig struct {
	Domain      string
	UnsafeHTML  bool
	RPCClient   *client.RPCClient
	Highlighter Highlighter
	Markdown    goldmark.Markdown
}

// NewDefaultHTMLWebClientConfig initializes a WebClientConfig with default settings.
// It sets up goldmark Markdown parsing options and default domain and highlighter.
func NewDefaultHTMLWebClientConfig(client *client.RPCClient) *HTMLWebClientConfig {
	mdopts := []goldmark.Option{goldmark.WithParserOptions(parser.WithAutoHeadingID())}
	return &HTMLWebClientConfig{
		Domain:      "gno.land",
		Highlighter: &noopHighlighter{},
		Markdown:    goldmark.New(mdopts...),
		RPCClient:   client,
	}
}

type HTMLWebClient struct {
	domain      string
	logger      *slog.Logger
	client      *client.RPCClient
	md          goldmark.Markdown
	highlighter Highlighter
}

// NewHTMLClient creates a new instance of WebClient.
// It requires a configured logger and WebClientConfig.
func NewHTMLClient(log *slog.Logger, cfg *HTMLWebClientConfig) *HTMLWebClient {
	return &HTMLWebClient{
		logger:      log,
		domain:      cfg.Domain,
		client:      cfg.RPCClient,
		md:          cfg.Markdown,
		highlighter: cfg.Highlighter,
	}
}

// Functions retrieves a list of function signatures from a
// specified package path.
func (s *HTMLWebClient) Functions(pkgPath string) ([]vm.FunctionSignature, error) {
	const qpath = "vm/qfuncs"

	args := fmt.Sprintf("%s/%s", s.domain, strings.Trim(pkgPath, "/"))
	res, err := s.query(qpath, []byte(args))
	if err != nil {
		return nil, fmt.Errorf("unable to query func list: %w", err)
	}

	var fsigs vm.FunctionSignatures
	if err := amino.UnmarshalJSON(res, &fsigs); err != nil {
		s.logger.Warn("unable to unmarshal function signatures, client is probably outdated")
		return nil, fmt.Errorf("unable to unmarshal function signatures: %w", err)
	}

	return fsigs, nil
}

// SourceFile fetches and writes the source file from a given
// package path and file name to the provided writer. It uses
// Chroma for syntax highlighting source.
func (s *HTMLWebClient) SourceFile(w io.Writer, path, fileName string) (*FileMeta, error) {
	const qpath = "vm/qfile"

	fileName = strings.TrimSpace(fileName)
	if fileName == "" {
		return nil, errors.New("empty filename given") // XXX: Consider creating a specific error variable
	}

	// XXX: Consider moving this into gnoclient
	fullPath := filepath.Join(s.domain, strings.Trim(path, "/"), fileName)

	source, err := s.query(qpath, []byte(fullPath))
	if err != nil {
		// XXX: this is a bit ugly, we should make the keeper return an
		// assertable error.
		if strings.Contains(err.Error(), "not available") {
			return nil, ErrClientPathNotFound
		}

		return nil, err
	}

	fileMeta := FileMeta{
		Lines:  strings.Count(string(source), "\n"),
		SizeKb: float64(len(source)) / 1024.0,
	}

	// Use Chroma for syntax highlighting
	if err := s.highlighter.Format(w, fileName, source); err != nil {
		return nil, err
	}

	return &fileMeta, nil
}

// Sources lists all source files available in a specified
// package path by querying the RPC client.
func (s *HTMLWebClient) Sources(path string) ([]string, error) {
	const qpath = "vm/qfile"

	// XXX: Consider moving this into gnoclient
	pkgPath := strings.Trim(path, "/")
	fullPath := fmt.Sprintf("%s/%s", s.domain, pkgPath)
	res, err := s.query(qpath, []byte(fullPath))
	if err != nil {
		// XXX: this is a bit ugly, we should make the keeper return an
		// assertable error.
		if strings.Contains(err.Error(), "not available") {
			return nil, ErrClientPathNotFound
		}

		return nil, err
	}

	files := strings.Split(strings.TrimSpace(string(res)), "\n")
	return files, nil
}

// RenderRealm renders the content of a realm from a given path
// and arguments into the provided writer. It uses Goldmark for
// Markdown processing to generate HTML content.
func (s *HTMLWebClient) RenderRealm(w io.Writer, pkgPath string, args string) (*RealmMeta, error) {
	const qpath = "vm/qrender"

	pkgPath = strings.Trim(pkgPath, "/")
	data := fmt.Sprintf("%s/%s:%s", s.domain, pkgPath, args)
	rawres, err := s.query(qpath, []byte(data))
	if err != nil {
		return nil, err
	}

	// Use Goldmark for Markdown parsing
	doc := s.md.Parser().Parse(text.NewReader(rawres))
	if err := s.md.Renderer().Render(w, rawres, doc); err != nil {
		return nil, fmt.Errorf("unable to render realm %q: %w", data, err)
	}

	var meta RealmMeta
	meta.Toc, err = md.TocInspect(doc, rawres, md.TocOptions{MaxDepth: 6, MinDepth: 2})
	if err != nil {
		s.logger.Warn("unable to inspect for TOC elements", "error", err)
	}

	return &meta, nil
}

// query sends a query to the RPC client and returns the response
// data.
func (s *HTMLWebClient) query(qpath string, data []byte) ([]byte, error) {
	s.logger.Info("query", "path", qpath, "data", string(data))

	qres, err := s.client.ABCIQuery(qpath, data)
	if err != nil {
		s.logger.Debug("request error", "path", qpath, "data", string(data), "error", err)
		return nil, fmt.Errorf("%w: %s", ErrClientBadRequest, err.Error())
	}

	if err = qres.Response.Error; err != nil {
		if errors.Is(err, vm.InvalidPkgPathError{}) {
			return nil, ErrClientPathNotFound
		}

		s.logger.Error("response error", "path", qpath, "log", qres.Response.Log)
		return nil, fmt.Errorf("%w: %s", ErrClientResponse, err.Error())
	}

	return qres.Response.Data, nil
}
