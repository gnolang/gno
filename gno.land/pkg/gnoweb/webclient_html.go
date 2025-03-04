package gnoweb

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	gopath "path"
	"strings"

	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	md "github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm" // for error types
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/yuin/goldmark"
	markdown "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

var (
	chromaDefaultStyle = styles.Get("friendly")
	chromaDarkStyle    = styles.Get("github-dark")
)

type HTMLWebClientConfig struct {
	Domain            string
	RPCClient         *client.RPCClient
	ChromaStyle       *chroma.Style
	ChromaHTMLOptions []chromahtml.Option
	GoldmarkOptions   []goldmark.Option
}

// NewDefaultHTMLWebClientConfig initializes a WebClientConfig with default settings.
// It sets up goldmark Markdown parsing options and default domain and highlighter.
func NewDefaultHTMLWebClientConfig(client *client.RPCClient) *HTMLWebClientConfig {
	chromaOptions := []chromahtml.Option{
		chromahtml.WithLineNumbers(true),
		chromahtml.WithLinkableLineNumbers(true, "L"),
		chromahtml.WithClasses(true),
		chromahtml.ClassPrefix("chroma-"),
	}

	goldmarkOptions := []goldmark.Option{
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithExtensions(
			md.NewMetadata(),
			markdown.NewHighlighting(
				markdown.WithFormatOptions(chromaOptions...),
			),
			extension.Strikethrough,
			extension.Table,
		),
	}

	return &HTMLWebClientConfig{
		Domain:            "gno.land",
		GoldmarkOptions:   goldmarkOptions,
		ChromaHTMLOptions: chromaOptions,
		ChromaStyle:       chromaDefaultStyle,
		RPCClient:         client,
	}
}

type HTMLWebClient struct {
	Markdown  goldmark.Markdown
	Formatter *chromahtml.Formatter

	domain      string
	logger      *slog.Logger
	client      *client.RPCClient
	chromaStyle *chroma.Style
}

// NewHTMLClient creates a new instance of WebClient.
// It requires a configured logger and WebClientConfig.
func NewHTMLClient(log *slog.Logger, cfg *HTMLWebClientConfig) *HTMLWebClient {
	return &HTMLWebClient{
		// XXX: Possibly consider exporting this in a single interface logic.
		// For now it's easier to manager all this in one place
		Markdown:  goldmark.New(cfg.GoldmarkOptions...),
		Formatter: chromahtml.New(cfg.ChromaHTMLOptions...),

		logger:      log,
		domain:      cfg.Domain,
		client:      cfg.RPCClient,
		chromaStyle: cfg.ChromaStyle,
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

	// Use Chroma for syntax highlighting
	if err := s.FormatSource(w, fileName, source); err != nil {
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

// extractHeadMeta extracts optional head metadata from the provided metaData map
// and returns a HeadMeta struct. All fields ("Title", "Description", "Canonical")
// are optional; if a field is not present or not a string, it will be empty.
func extractHeadMeta(metaData map[string]interface{}) HeadMeta {
	hm := HeadMeta{}
	if title, ok := metaData["Title"].(string); ok {
		hm.Title = title
	}
	if desc, ok := metaData["Description"].(string); ok {
		hm.Description = desc
	}
	if canonical, ok := metaData["Canonical"].(string); ok {
		hm.Canonical = canonical
	}
	return hm
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
	context := parser.NewContext()
	doc := s.Markdown.Parser().Parse(text.NewReader(rawres), parser.WithContext(context))
	metaData := md.Get(context)

	if err := s.Markdown.Renderer().Render(w, rawres, doc); err != nil {
		return nil, fmt.Errorf("unable to render realm %q: %w", data, err)
	}

	var meta RealmMeta
	meta.Toc, err = md.TocInspect(doc, rawres, md.TocOptions{MaxDepth: 6, MinDepth: 2})
	if err != nil {
		s.logger.Warn("unable to inspect for TOC elements", "error", err)
	}

	meta.Head = extractHeadMeta(metaData)

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

		s.logger.Error("response error", "path", qpath, "log", qres.Response.Log)
		return nil, fmt.Errorf("%w: %s", ErrClientResponse, err.Error())
	}

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
	if err := s.Formatter.WriteCSS(w, s.chromaStyle); err != nil {
		return err
	}

	// Generate CSS for chroma dark mode
	var darkCSS bytes.Buffer
	if err := s.Formatter.WriteCSS(&darkCSS, chromaDarkStyle); err != nil {
		return err
	}

	darkCSSStr := darkCSS.String()
	fmt.Fprintf(w, "\n/* Styles for dark mode */\n")

	lines := strings.Split(darkCSSStr, "\n")
	for i, line := range lines {
		// Add .dark class to chroma styles
		if strings.Contains(line, "{") && strings.Contains(line, ".chroma") {
			parts := strings.SplitN(line, "{", 2)
			if len(parts) != 2 {
				continue
			}

			selectors := strings.Split(parts[0], ",")
			for j, selector := range selectors {
				selectors[j] = ".dark " + strings.TrimSpace(selector)
			}

			lines[i] = strings.Join(selectors, ", ") + " {" + parts[1]
		}
	}

	fmt.Fprint(w, strings.Join(lines, "\n"))

	return nil
}
