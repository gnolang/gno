package gnoweb

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/markdown"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/gnolang/gno/gnovm/pkg/doc"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// MockPackage represents a mock package with files and function signatures.
type MockPackage struct {
	Path      string
	Domain    string
	Files     map[string]string // filename -> body
	Functions []*doc.JSONFunc
}

// MockWebClient is a mock implementation of the Client interface.
type MockWebClient struct {
	Packages map[string]*MockPackage // path -> package
	markdown goldmark.Markdown
}

func NewMockWebClient(pkgs ...*MockPackage) *MockWebClient {
	mpkgs := make(map[string]*MockPackage)
	for _, pkg := range pkgs {
		mpkgs[pkg.Path] = pkg
	}

	return &MockWebClient{
		Packages: mpkgs,
		markdown: goldmark.New(
			goldmark.WithExtensions(
				markdown.GnoExtension,
			),
		),
	}
}

// RenderRealm simulates rendering a package by writing its content to the writer.
func (m *MockWebClient) RenderRealm(w io.Writer, u *weburl.GnoURL) (*RealmMeta, error) {
	pkg, exists := m.Packages[u.Path]
	if !exists {
		return nil, ErrClientPathNotFound
	}

	if !pkgHasRender(pkg) {
		return nil, ErrRenderNotDeclared
	}

	// Return the production format [domain]/path:args
	fmt.Fprintf(w, "[%s]/%s:%s", pkg.Domain, strings.Trim(u.Path, "/"), u.Args)

	// Return a dummy RealmMeta for simplicity
	return &RealmMeta{}, nil
}

// SourceFile simulates retrieving a source file's metadata.
func (m *MockWebClient) SourceFile(w io.Writer, pkgPath, fileName string, isRaw bool) (*FileMeta, error) {
	pkg, exists := m.Packages[pkgPath]
	if !exists {
		return nil, ErrClientPathNotFound
	}

	if body, ok := pkg.Files[fileName]; ok {
		if w != nil {
			_, err := w.Write([]byte(body))
			if err != nil {
				return nil, err
			}
		}
		return &FileMeta{
			Lines:  len(bytes.Split([]byte(body), []byte("\n"))),
			SizeKb: float64(len(body)) / 1024.0,
		}, nil
	}

	return nil, ErrClientPathNotFound
}

// Doc simulates retrieving function docs from a package.
func (m *MockWebClient) Doc(path string) (*doc.JSONDocumentation, error) {
	pkg, exists := m.Packages[path]
	if !exists {
		return nil, ErrClientPathNotFound
	}

	return &doc.JSONDocumentation{Funcs: pkg.Functions}, nil
}

// Sources simulates listing all source files in a package.
func (m *MockWebClient) Sources(path string) ([]string, error) {
	pkg, exists := m.Packages[path]
	if !exists {
		return nil, ErrClientPathNotFound
	}

	fileNames := make([]string, 0, len(pkg.Files))
	for file := range pkg.Files {
		fileNames = append(fileNames, file)
	}

	// Sort for consistency
	sort.Strings(fileNames)

	return fileNames, nil
}

// RenderMd simulates rendering a markdown file.
func (m *MockWebClient) RenderMd(w io.Writer, u *weburl.GnoURL, fileName string) (*RealmMeta, error) {
	pkg, exists := m.Packages[u.Path]
	if !exists {
		return nil, ErrClientPathNotFound
	}

	if body, ok := pkg.Files[fileName]; ok {
		// Parse and render the markdown
		doc := m.markdown.Parser().Parse(text.NewReader([]byte(body)))
		if err := m.markdown.Renderer().Render(w, []byte(body), doc); err != nil {
			return nil, fmt.Errorf("unable to render markdown: %w", err)
		}

		return &RealmMeta{}, nil
	}

	return nil, ErrClientPathNotFound
}

// ParseMarkdown parses and renders Markdown content using Goldmark.
func (m *MockWebClient) ParseMarkdown(w io.Writer, rawContent []byte, ctxOpts ...parser.ParseOption) (ast.Node, error) {
	doc := m.markdown.Parser().Parse(text.NewReader(rawContent), ctxOpts...)
	if err := m.markdown.Renderer().Render(w, rawContent, doc); err != nil {
		return nil, fmt.Errorf("unable to render markdown: %w", err)
	}
	return doc, nil
}

// FormatSource simulates formatting source code with syntax highlighting.
func (m *MockWebClient) FormatSource(w io.Writer, fileName string, source []byte) error {
	// For testing, we just write the source as-is with a CSS class
	fmt.Fprintf(w, "<pre class=\"chroma-\">%s</pre>", source)
	return nil
}

// WriteFormatterCSS simulates writing CSS for syntax highlighting.
func (m *MockWebClient) WriteFormatterCSS(w io.Writer) error {
	// For testing, we just write a minimal CSS
	_, err := w.Write([]byte(".chroma- { background-color: #f8f8f8; }"))
	return err
}

func pkgHasRender(pkg *MockPackage) bool {
	if len(pkg.Functions) == 0 {
		return false
	}

	for _, fn := range pkg.Functions {
		if fn.Name == "Render" &&
			len(fn.Params) == 1 &&
			len(fn.Results) == 1 &&
			fn.Params[0].Type == "string" &&
			fn.Results[0].Type == "string" {
			return true
		}
	}

	return false
}

// ABCIResponse is a mock type for testing
type ABCIResponse struct {
	Data []byte
}

// ABCIQueryResponse is a mock type for testing
type ABCIQueryResponse struct {
	Response ABCIResponse
}
