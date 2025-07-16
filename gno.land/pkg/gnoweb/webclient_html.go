package gnoweb

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	gopath "path"
	"slices"
	"strings"

	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm" // for error types
	"github.com/gnolang/gno/gnovm/pkg/doc"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
)

var (
	chromaDefaultStyle   = styles.Get("friendly")
	chromaDefaultOptions = []chromahtml.Option{
		chromahtml.WithLineNumbers(true),
		chromahtml.WithLinkableLineNumbers(true, "L"),
		chromahtml.WithClasses(true),
		chromahtml.ClassPrefix("chroma-"),
	}
)

type HTMLWebClientConfig struct {
	Domain            string
	RPCClient         *client.RPCClient
	ChromaStyle       *chroma.Style
	ChromaHTMLOptions []chromahtml.Option
}

// NewDefaultHTMLWebClientConfig initializes a WebClientConfig with default settings.
func NewDefaultHTMLWebClientConfig(client *client.RPCClient) *HTMLWebClientConfig {
	return &HTMLWebClientConfig{
		Domain:            "gno.land",
		ChromaHTMLOptions: chromaDefaultOptions,
		ChromaStyle:       chromaDefaultStyle,
		RPCClient:         client,
	}
}

type HTMLWebClient struct {
	Formatter *chromahtml.Formatter

	domain      string
	logger      *slog.Logger
	client      *client.RPCClient
	chromaStyle *chroma.Style
}

var _ WebClient = (*HTMLWebClient)(nil)

// NewHTMLClient creates a new instance of WebClient.
// It requires a configured logger and WebClientConfig.
func NewHTMLClient(log *slog.Logger, cfg *HTMLWebClientConfig) *HTMLWebClient {
	return &HTMLWebClient{
		// XXX: Possibly consider exporting this in a single interface logic.
		// For now it's easier to manager all this in one place
		Formatter: chromahtml.New(cfg.ChromaHTMLOptions...),

		logger:      log,
		domain:      cfg.Domain,
		client:      cfg.RPCClient,
		chromaStyle: cfg.ChromaStyle,
	}
}

// Doc retrieves the JSON doc suitable for printing from a
// specified package path.
func (s *HTMLWebClient) Doc(pkgPath string) (*doc.JSONDocumentation, error) {
	const qpath = "vm/qdoc"

	args := fmt.Sprintf("%s/%s", s.domain, strings.Trim(pkgPath, "/"))
	res, err := s.query(qpath, []byte(args))
	if err != nil {
		return nil, fmt.Errorf("unable to query qdoc: %w", err)
	}

	jdoc := &doc.JSONDocumentation{}
	if err := amino.UnmarshalJSON(res, jdoc); err != nil {
		s.logger.Warn("unable to unmarshal qdoc, client is probably outdated")
		return nil, fmt.Errorf("unable to unmarshal qdoc: %w", err)
	}

	return jdoc, nil
}

// SourceFile fetches and writes the source file from a given
// package path and file name to the provided writer. It uses
// Chroma for syntax highlighting or Raw style source.
func (s *HTMLWebClient) SourceFile(w io.Writer, path, fileName string, isRaw bool) (*FileMeta, error) {
	const qpath = "vm/qfile"

	fileName = strings.TrimSpace(fileName)
	if fileName == "" {
		return nil, errors.New("empty filename given") // XXX: Consider creating a specific error variable
	}

	// XXX: Consider moving this into gnoclient
	fullPath := gopath.Join(s.domain, strings.Trim(path, "/"), fileName)

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

	if isRaw {
		// Use raw syntax for source
		if _, err := w.Write(source); err != nil {
			return nil, err
		}
	} else {
		// Use Chroma for syntax highlighting
		if err := s.FormatSource(w, fileName, source); err != nil {
			return nil, err
		}
	}

	return &fileMeta, nil
}

// HasFile checks if fileName exists in the list of source files for pkgPath.
func (s *HTMLWebClient) HasFile(pkgPath, fileName string) bool {
	files, err := s.Sources(pkgPath)
	if err != nil {
		return false
	}
	return slices.Contains(files, fileName)
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

// Sources lists all source files available in a specified
// package path by querying the RPC client.
func (s *HTMLWebClient) QueryPaths(prefix string, limit int) ([]string, error) {
	const qpath = "vm/qpaths"

	// XXX: Consider moving this into gnoclient
	res, err := s.query(qpath, []byte(prefix))
	if err != nil {
		return nil, err
	}

	// update the paths to be relative to the root instead of the domain
	paths := strings.Split(strings.TrimSpace(string(res)), "\n")
	for i, path := range paths {
		paths[i] = strings.TrimPrefix(path, s.domain)
	}

	return paths, nil
}

// RenderRealm renders the content of a realm from a given path
// and arguments into the provided writer. It uses Goldmark for
// Markdown processing to generate HTML content.
func (s *HTMLWebClient) RenderRealm(w io.Writer, u *weburl.GnoURL, cr ContentRenderer) (*RealmMeta, error) {
	const qpath = "vm/qrender"

	pkgPath := strings.Trim(u.Path, "/")
	data := fmt.Sprintf("%s/%s:%s", s.domain, pkgPath, u.EncodeArgs())

	rawres, err := s.query(qpath, []byte(data))
	if err != nil {
		return nil, err
	}

	var meta RealmMeta
	if meta.Toc, err = cr.Render(w, u, rawres); err != nil {
		return nil, fmt.Errorf("unable to render realm: %w", err)
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

		if errors.Is(err, vm.NoRenderDeclError{}) {
			return nil, ErrRenderNotDeclared
		}

		s.logger.Debug("query response error", "path", qpath, "log", qres.Response.Log)
		return nil, fmt.Errorf("%w: %s", ErrClientResponse, err.Error())
	}

	s.logger.Debug("response query", "path", qpath, "data", qres.Response.Data)
	return qres.Response.Data, nil
}

func (s *HTMLWebClient) FormatSource(w io.Writer, fileName string, src []byte) error {
	var lexer chroma.Lexer

	// Determine the lexer to be used based on the file extension.
	switch strings.ToLower(gopath.Ext(fileName)) {
	case ".gno":
		lexer = lexers.Get("go")
	case ".md":
		lexer = lexers.Get("markdown")
	case ".mod":
		lexer = lexers.Get("gomod")
	default:
		lexer = lexers.Get("txt") // Unsupported file type, default to plain text.
	}

	if lexer == nil {
		return fmt.Errorf("unsupported lexer for file %q", fileName)
	}

	iterator, err := lexer.Tokenise(nil, string(src))
	if err != nil {
		return fmt.Errorf("unable to tokenise %q: %w", fileName, err)
	}

	if err := s.Formatter.Format(w, s.chromaStyle, iterator); err != nil {
		return fmt.Errorf("unable to format source file %q: %w", fileName, err)
	}

	return nil
}

func (s *HTMLWebClient) WriteFormatterCSS(w io.Writer) error {
	return s.Formatter.WriteCSS(w, s.chromaStyle)
}

func init() {
	// Register custom go.mod (and gno.mod) lexer
	lexers.Register(chroma.MustNewLexer(
		&chroma.Config{
			Name:      "Go/Gno module file",
			Aliases:   []string{"go-mod", "gomod", "gno-mod", "gnomod"},
			Filenames: []string{"go.mod", "gno.mod"},
			MimeTypes: []string{"text/x-go-mod"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {
					{Pattern: `\s+`, Type: chroma.Text, Mutator: nil},
					{Pattern: `//[^\n\r]*`, Type: chroma.CommentSingle, Mutator: nil},
					// Keywords
					{Pattern: `\b(module|require|replace|exclude|gn?o) \b`, Type: chroma.Keyword, Mutator: nil},
					{Pattern: `\b(indirect)\b`, Type: chroma.NameDecorator, Mutator: nil},
					// Quoted version strings
					{Pattern: `"[^"]*"`, Type: chroma.LiteralString, Mutator: nil},
					// Versions (v1.2.3, v0.0.0-yyyymmddhhmmss-abcdefabcdef) as well as 0.9 version
					{
						Pattern: `\b(v?\d+\.\d+(?:\.\d+)?(?:-[0-9]{14}-[a-f0-9]+)?)\b`,
						Type:    chroma.LiteralNumber, Mutator: nil,
					},
					// Module paths: module/example.com, example.com/something
					{
						Pattern: `[a-zA-Z0-9._~\-\/]+(\.[a-zA-Z]{2, Type:})+(/[^\s]+)?`,
						Type:    chroma.NameNamespace, Mutator: nil,
					},
					// Operator (=> in replace directive)
					{Pattern: `=>`, Type: chroma.Operator, Mutator: nil},
					// Everything else (bare words, etc.)
					{Pattern: `[^\s"//]+`, Type: chroma.Text, Mutator: nil},
					{Pattern: `\n`, Type: chroma.Text, Mutator: nil},
				},
			}
		},
	))
}
