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

type FileMeta struct {
	Lines  int
	SizeKb float64
}

type RealmMeta struct {
	*md.Toc
}

// WebClient is an interface for interacting with web resources.
type Client interface {
	Render(w io.Writer, path string, args string) (*RealmMeta, error)
	SourceFile(w io.Writer, pkgPath, fileName string) (*FileMeta, error)
	Functions(path string) ([]vm.FunctionSignature, error)
	Sources(path string) ([]string, error)
}

type WebClientConfig struct {
	Domain      string
	UnsafeHTML  bool
	RPCClient   *client.RPCClient
	Highlighter Highlighter
	Markdown    goldmark.Markdown
}

func NewDefaultWebClientConfig(client *client.RPCClient) *WebClientConfig {
	// Configure goldmark markdown options
	mdopts := []goldmark.Option{goldmark.WithParserOptions(parser.WithAutoHeadingID())}
	return &WebClientConfig{
		Domain:      "gno.land",
		Highlighter: &noopHighlighter{},
		Markdown:    goldmark.New(mdopts...),
		RPCClient:   client,
	}
}

// Validate checks if all elements of WebClientConfig are not nil.
func (cfg *WebClientConfig) Validate() error {
	if cfg.RPCClient == nil {
		return errors.New("RPCClient must not be nil")
	}

	return nil
}

type WebClient struct {
	domain      string
	logger      *slog.Logger
	client      *client.RPCClient
	md          goldmark.Markdown
	highlighter Highlighter
}

func NewWebClient(log *slog.Logger, cfg *WebClientConfig) *WebClient {
	return &WebClient{
		logger:      log,
		domain:      cfg.Domain,
		client:      cfg.RPCClient,
		md:          cfg.Markdown,
		highlighter: cfg.Highlighter,
	}
}

func (s *WebClient) Functions(pkgPath string) ([]vm.FunctionSignature, error) {
	const qpath = "vm/qfuncs"

	args := fmt.Sprintf("%s/%s", s.domain, strings.Trim(pkgPath, "/"))
	res, err := s.query(qpath, []byte(args))
	if err != nil {
		return nil, fmt.Errorf("unable query funcs list: %w", err)
	}

	var fsigs vm.FunctionSignatures
	if err := amino.UnmarshalJSON(res, &fsigs); err != nil {
		s.logger.Warn("unable to unmarshal fsigs, client is probably outdated ?")
		return nil, fmt.Errorf("unable to unamarshal fsigs: %w", err)
	}

	return fsigs, nil
}

func (s *WebClient) SourceFile(w io.Writer, path, fileName string) (*FileMeta, error) {
	const qpath = "vm/qfile"

	fileName = strings.TrimSpace(fileName) // sanitize filename
	if fileName == "" {
		return nil, errors.New("empty filename given") // XXX -> ErrXXX
	}

	// XXX: move this into gnoclient ?
	path = fmt.Sprintf("%s/%s", s.domain, strings.Trim(path, "/"))
	path = filepath.Join(path, fileName)

	source, err := s.query(qpath, []byte(path))
	if err != nil {
		return nil, err
	}

	// XXX: we should either do this on the front or in the markdown parsing side
	fileMeta := FileMeta{
		Lines:  strings.Count(string(source), "\n"),
		SizeKb: float64(len(source)) / 1024.0,
	}

	if err := s.highlighter.Format(w, fileName, source); err != nil {
		return nil, err
	}

	return &fileMeta, nil
}

func (s *WebClient) Sources(path string) ([]string, error) {
	const qpath = "vm/qfile"

	// XXX: move this into gnoclient
	path = fmt.Sprintf("%s/%s", s.domain, strings.Trim(path, "/"))
	res, err := s.query(qpath, []byte(path))
	if err != nil {
		return nil, err
	}

	files := strings.Split(string(res), "\n")
	return files, nil
}

func (s *WebClient) Render(w io.Writer, pkgPath string, args string) (*RealmMeta, error) {
	const qpath = "vm/qrender"

	pkgPath = strings.Trim(pkgPath, "/")
	data := fmt.Sprintf("%s/%s:%s", s.domain, pkgPath, args)
	rawres, err := s.query(qpath, []byte(data))
	if err != nil {
		return nil, err
	}

	doc := s.md.Parser().Parse(text.NewReader(rawres))
	if err := s.md.Renderer().Render(w, rawres, doc); err != nil {
		return nil, fmt.Errorf("unable render real %q: %w", data, err)
	}

	var meta RealmMeta
	meta.Toc, err = md.TocInspect(doc, rawres, md.TocOptions{MaxDepth: 6, MinDepth: 2})
	if err != nil {
		s.logger.Warn("unable to inspect for toc elements", "err", err)
	}

	return &meta, nil
}

func (s *WebClient) query(qpath string, data []byte) ([]byte, error) {
	s.logger.Info("query", "qpath", qpath, "data", string(data))

	qres, err := s.client.ABCIQuery(qpath, data)
	if err != nil {
		s.logger.Error("request error", "path", qpath, "data", string(data), "error", err)
		return nil, fmt.Errorf("unable to query path %q: %w", qpath, err)
	}
	if qres.Response.Error != nil {
		s.logger.Error("response error", "path", qpath, "log", qres.Response.Log)
		return nil, qres.Response.Error
	}

	return qres.Response.Data, nil
}
