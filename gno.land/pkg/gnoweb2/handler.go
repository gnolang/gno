package gnoweb

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb2/components"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm" // for error types
	"github.com/gnolang/gno/gno.land/pkg/service"
)

const chromaPath = "/_chroma/style.css"

var chromaStyle = styles.Get("friendly")

func init() {
	if chromaStyle == nil {
		panic("unable to get chroma style")
	}
}

type StaticMetadata struct {
	AssetsPath string
	RemoteHelp string
	ChaindID   string
}

type Formater interface {
}

type WebHandlerConfig struct {
	Meta         StaticMetadata
	RenderClient *service.WebRenderClient
}

type WebHandler struct {
	formatter *html.Formatter

	logger    *slog.Logger
	static    StaticMetadata
	rendercli *service.WebRenderClient

	// bufferPool is used to reuse Buffer instances
	// to reduce memory allocations and improve performance.
	bufferPool sync.Pool
}

func NewWebHandler(logger *slog.Logger, cfg WebHandlerConfig) *WebHandler {
	if cfg.RenderClient == nil {
		logger.Error("no renderer has been defined")
	}

	// Setup Formatter
	formatter := html.New(
		html.WithClasses(true),
		html.ClassPrefix("chroma-"),
		html.WithLineNumbers(true),
	)

	return &WebHandler{
		formatter: formatter,
		rendercli: cfg.RenderClient,
		logger:    logger.WithGroup("web"),
		static:    cfg.Meta,
		// Initialize the pool with bytes.Buffer factory
		bufferPool: sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		},
	}
}

func (h *WebHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("receiving request", "method", r.Method, "path", r.URL.Path)
	if r.Method == http.MethodGet {
		if r.URL.Path == chromaPath {
			w.Header().Set("Content-Type", "text/css")
			// XXX: Move this elsewhere, it should fail with the prefix
			if err := h.formatter.WriteCSS(w, chromaStyle); err != nil {
				h.logger.Error("unable to write css", "err", err)
				http.NotFound(w, r)
			}
			return
		}

		h.Get(w, r)
		return
	}

	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

type PathKind string

func generateBreadcrumbPaths(path string) []components.BreadcrumbPart {
	split := strings.Split(path, "/")
	parts := []components.BreadcrumbPart{}

	for i := range split {
		name := split[i]
		if name == "" {
			continue
		}

		parts = append(parts, components.BreadcrumbPart{
			Name: split[i],
			Path: strings.Join(split[:i+1], "/"),
		})
	}

	return parts
}

func (h *WebHandler) Get(w http.ResponseWriter, r *http.Request) {
	var headData components.HeadData // Placeholder for later metadata

	body := h.getBuffer()
	defer h.putBuffer(body)

	var status int

	gnourl, err := ParseGnoURL(r.URL)
	if err != nil {
		h.logger.Error("unable to render body", "url", r.URL.String(), "err", err)
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}

	// Render the page body into the buffer
	switch gnourl.Kind {
	case KindRealm:
		status, err = h.renderRealm(body, gnourl)
	case KindPure:
		status, err = http.StatusNotFound, components.RenderStatusComponent(w, "page not found")
	case KindUser:
		fallthrough
	default:
		http.Error(w, "not supported", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(status)

	var indexData components.IndexData

	indexData.HeadData = headData
	indexData.HeaderData.RealmPath = gnourl.Path
	indexData.HeaderData.Breadcrumb.Parts = generateBreadcrumbPaths(gnourl.Path)
	indexData.HeaderData.WebQuery = gnourl.WebQuery

	indexData.Body = template.HTML(body.String())
	// Render the final page with the rendered body
	err = components.RenderIndexComponent(w, indexData)

	if err != nil {
		h.logger.Error("failed to render index component", "err", err)
	}

	return
}

func (h *WebHandler) renderRealmHelp(w io.Writer, gnourl *GnoURL) (status int, err error) {
	fsigs, err := h.rendercli.Functions(gnourl.Path)
	if err != nil {
		h.logger.Error("unable to fetch path functions", "err", err)
		return http.StatusInternalServerError, components.RenderStatusComponent(w, "internal error")
	}

	realmName := "<unknown>"
	if i := strings.LastIndexByte(gnourl.Path, '/'); i > 0 {
		realmName = gnourl.Path[i+1:]
	}

	err = components.RenderHelpComponent(w, components.HelpData{
		RealmName: realmName,
		ChainId:   h.static.ChaindID,
		PkgPath:   gnourl.FullPath,
		Remote:    h.static.RemoteHelp,
		Functions: fsigs,
	})
	if err != nil {
		h.logger.Error("unable to render helper", "err", err)
		return http.StatusInternalServerError, components.RenderStatusComponent(w, "internal error")
	}

	return http.StatusOK, nil
}

func (h *WebHandler) renderRealmSource(w io.Writer, gnourl *GnoURL) (status int, err error) {
	pkgPath := gnourl.Path

	files, err := h.rendercli.Sources(pkgPath)
	if err != nil {
		h.logger.Error("unable to list sources file", "path", gnourl.Path, "err", err)
		return http.StatusInternalServerError, components.RenderStatusComponent(w, "internal error")
	}

	if len(files) == 0 {
		h.logger.Debug("no file(s) available", "path", gnourl.Path)
		return http.StatusOK, components.RenderStatusComponent(w, "no files available")
	}

	var fileName string
	file := gnourl.WebQuery.Get("file")
	if file == "" {
		fileName = files[0]
	} else if contains(files, file) {
		fileName = file
	} else {
		h.logger.Error("unable to render source", "file", file, "err", "file does not exist")
		return http.StatusInternalServerError, components.RenderStatusComponent(w, "internal error")
	}

	source, err := h.rendercli.SourceFile(pkgPath, fileName)
	if err != nil {
		h.logger.Error("unable to get source file", "file", fileName, "err", err)
		return http.StatusInternalServerError, components.RenderStatusComponent(w, "internal error")
	}

	// if theme := gnourl.WebQuery.Get("theme"); theme != "" {
	// 	n, err := strconv.ParseInt(theme, 10, 32)
	// 	if err != nil {
	// 		cstyle = styles.Get(theme)
	// 	} else {
	// 		listName := make([]string, len(styles.Registry))
	// 		for name := range styles.Registry {
	// 			listName = append(listName, name)
	// 		}

	// 		sort.Slice(listName, func(i, j int) bool {
	// 			return strings.Compare(listName[i], listName[j]) > 0
	// 		})

	// 		cstyle = styles.Get(listName[n])
	// 	}
	// }

	hsource, err := h.highlightSource(chromaStyle, fileName, source)
	if err != nil {
		h.logger.Error("unable to highlight source file", "file", fileName, "err", err)
		return http.StatusInternalServerError, components.RenderStatusComponent(w, "internal error")
	}

	err = components.RenderSourceComponent(w, components.SourceData{
		PkgPath:    gnourl.Path,
		Files:      files,
		FileName:   fileName,
		FileSource: template.HTML(hsource),
	})
	if err != nil {
		h.logger.Error("unable to render helper", "err", err)
		return http.StatusInternalServerError, components.RenderStatusComponent(w, "internal error")
	}

	return http.StatusOK, nil
}

func contains(files []string, file string) bool {
	for _, f := range files {
		if f == file {
			return true
		}
	}
	return false
}

func (h *WebHandler) renderRealm(w io.Writer, gnourl *GnoURL) (status int, err error) {
	h.logger.Info("component render", "path", gnourl.Path, "args", gnourl.PathArgs)

	// Display realm help page
	if gnourl.WebQuery.Has("help") {
		return h.renderRealmHelp(w, gnourl)
	}

	// Display realm source page
	if gnourl.WebQuery.Has("source") {
		return h.renderRealmSource(w, gnourl)
	}

	// Render content into the content buffer
	content := h.getBuffer()
	defer h.putBuffer(content)

	meta, err := h.rendercli.Render(content, gnourl.Path, gnourl.EncodeArgs())
	if err != nil {
		if errors.Is(err, vm.InvalidPkgPathError{}) {
			return http.StatusNotFound, components.RenderStatusComponent(w, "not found")
		}

		h.logger.Error("unable to render markdown", "err", err)
		return http.StatusInternalServerError, components.RenderStatusComponent(w, "internal error")
	}

	err = components.RenderRealmComponent(w, components.RealmData{
		TocItems: &components.RealmTOCData{
			Items: meta.Items,
		},
		Content: template.HTML(content.String()),
	})

	if err != nil {
		h.logger.Error("unable to render template", "err", err)
		return http.StatusInternalServerError, components.RenderStatusComponent(w, "internal error")
	}

	// Write the rendered content to the response writer
	return http.StatusOK, nil
}

func (h *WebHandler) highlightSource(style *chroma.Style, fileName string, src []byte) ([]byte, error) {
	var lexer chroma.Lexer
	switch strings.ToLower(filepath.Ext(fileName)) {
	case ".gno":
		lexer = lexers.Get("go")
	case ".md":
		lexer = lexers.Get("markdown")
	case ".mod":
		lexer = lexers.Get("gomod")
	default:
		return nil, fmt.Errorf("unsupported extension for highlighting source file: %q", fileName)
	}

	if lexer == nil {
		return nil, fmt.Errorf("unsuported lexer for file %q", fileName)
	}

	iterator, err := lexer.Tokenise(nil, string(src))
	if err != nil {
		h.logger.Error("unable to ", "fileName", fileName, "err", err)
	}

	var buff bytes.Buffer
	if style.Name != chromaStyle.Name {
		if err := html.New().Format(&buff, style, iterator); err != nil {
			return nil, fmt.Errorf("unable to format theme: %w", err)
		}
		return buff.Bytes(), nil
	}

	if err := h.formatter.Format(&buff, style, iterator); err != nil {
		return nil, fmt.Errorf("unable to format source file %q: %w", fileName, err)
	}

	return buff.Bytes(), nil
}

// getBuffer retrieves a buffer from the sync.Pool
func (h *WebHandler) getBuffer() *bytes.Buffer {
	return h.bufferPool.Get().(*bytes.Buffer)
}

// putBuffer resets and puts a buffer back into the sync.Pool
func (h *WebHandler) putBuffer(buf *bytes.Buffer) {
	buf.Reset()
	h.bufferPool.Put(buf)
}
