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
	"slices"
	"strings"
	"time"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm" // for error types
)

const DefaultChainDomain = "gno.land"

type StaticMetadata struct {
	AssetsPath string
	ChromaPath string
	RemoteHelp string
	ChainId    string
	Analytics  bool
}

type WebHandlerConfig struct {
	Meta         StaticMetadata
	RenderClient *WebClient
	Formatter    Formatter
}

type WebHandler struct {
	formatter Formatter

	logger *slog.Logger
	static StaticMetadata
	webcli *WebClient
}

func NewWebHandler(logger *slog.Logger, cfg WebHandlerConfig) *WebHandler {
	if cfg.RenderClient == nil {
		logger.Error("no renderer has been defined")
	}

	return &WebHandler{
		formatter: cfg.Formatter,
		webcli:    cfg.RenderClient,
		logger:    logger,
		static:    cfg.Meta,
	}
}

func (h *WebHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("receiving request", "method", r.Method, "path", r.URL.Path)

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.Get(w, r)
}

func (h *WebHandler) Get(w http.ResponseWriter, r *http.Request) {
	var body bytes.Buffer

	start := time.Now()
	defer func() {
		h.logger.Debug("request completed",
			"url", r.URL.String(),
			"elapsed", time.Since(start).String())
	}()

	var indexData components.IndexData
	indexData.HeadData.AssetsPath = h.static.AssetsPath
	indexData.HeadData.ChromaPath = h.static.ChromaPath
	indexData.FooterData.Analytics = h.static.Analytics
	indexData.FooterData.AssetsPath = h.static.AssetsPath

	// Render the page body into the buffer
	var status int
	gnourl, err := ParseGnoURL(r.URL)
	if err != nil {
		h.logger.Warn("page not found", "path", r.URL.Path, "err", err)
		status, err = http.StatusNotFound, components.RenderStatusComponent(&body, "page not found")
	} else {
		// TODO: real data (title & description)
		indexData.HeadData.Title = "gno.land - " + gnourl.Path

		// Header
		indexData.HeaderData.RealmPath = gnourl.Path
		indexData.HeaderData.Breadcrumb = generateBreadcrumbPaths(gnourl)
		indexData.HeaderData.WebQuery = gnourl.WebQuery

		// Render
		switch {
		case gnourl.IsRealm(), gnourl.IsPure():
			status, err = h.renderPackage(&body, gnourl)
		default:
			h.logger.Debug("invalid path: path is neither a pure package or a realm")
			status, err = http.StatusNotFound, components.RenderStatusComponent(&body, "page not found")
		}
	}

	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)

	// NOTE: HTML escaping should have already been done by markdown rendering package
	indexData.Body = template.HTML(body.String()) //nolint:gosec

	// Render the final page with the rendered body
	if err = components.RenderIndexComponent(w, indexData); err != nil {
		h.logger.Error("failed to render index component", "err", err)
	}

	return
}

func (h *WebHandler) renderPackage(w io.Writer, gnourl *GnoURL) (status int, err error) {
	h.logger.Info("component render", "path", gnourl.Path, "args", gnourl.Args)

	// Display realm help page?
	if gnourl.WebQuery.Has("help") {
		return h.renderRealmHelp(w, gnourl)
	}

	// Display package source page?
	switch {
	case gnourl.WebQuery.Has("source"):
		return h.renderRealmSource(w, gnourl)
	case gnourl.IsFile():
		// Fill webquery with file infos
		return h.renderRealmSource(w, gnourl)
	case gnourl.IsDir(), gnourl.IsPure():
		return h.renderRealmDirectory(w, gnourl)
	}

	// Render content into the content buffer
	var content bytes.Buffer
	meta, err := h.webcli.Render(&content, gnourl.Path, gnourl.EncodeArgs())
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
		// NOTE: `content` should have already been escaped by
		Content: template.HTML(content.String()), //nolint:gosec
	})
	if err != nil {
		h.logger.Error("unable to render template", "err", err)
		return http.StatusInternalServerError, components.RenderStatusComponent(w, "internal error")
	}

	// Write the rendered content to the response writer
	return http.StatusOK, nil
}

func (h *WebHandler) renderRealmHelp(w io.Writer, gnourl *GnoURL) (status int, err error) {
	fsigs, err := h.webcli.Functions(gnourl.Path)
	if err != nil {
		h.logger.Error("unable to fetch path functions", "err", err)
		return http.StatusInternalServerError, components.RenderStatusComponent(w, "internal error")
	}

	var selArgs map[string]string
	var selFn string
	if selFn = gnourl.WebQuery.Get("func"); selFn != "" {
		for _, fn := range fsigs {
			if selFn != fn.FuncName {
				continue
			}

			selArgs = make(map[string]string)
			for _, param := range fn.Params {
				selArgs[param.Name] = gnourl.WebQuery.Get(param.Name)
			}

			fsigs = []vm.FunctionSignature{fn}
			break
		}
	}

	// Catch last name of the path
	// XXX: we should probably add a helper within the template
	realmName := filepath.Base(gnourl.Path)
	err = components.RenderHelpComponent(w, components.HelpData{
		SelectedFunc: selFn,
		SelectedArgs: selArgs,
		RealmName:    realmName,
		ChainId:      h.static.ChainId,
		// TODO: get chain domain and use that.
		PkgPath:   filepath.Join(DefaultChainDomain, gnourl.Path),
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

	files, err := h.webcli.Sources(pkgPath)
	if err != nil {
		h.logger.Error("unable to list sources file", "path", gnourl.Path, "err", err)
		return http.StatusInternalServerError, components.RenderStatusComponent(w, "internal error")
	}

	if len(files) == 0 {
		h.logger.Debug("no files available", "path", gnourl.Path)
		return http.StatusOK, components.RenderStatusComponent(w, "no files available")
	}

	file := gnourl.WebQuery.Get("file") // webquery override file
	if file == "" {
		file = gnourl.File
	}

	var fileName string
	if file == "" {
		fileName = files[0] // Default to the first file if none specified
	} else if slices.Contains(files, file) {
		fileName = file // Use specified file if it exists
	} else {
		h.logger.Error("unable to render source", "file", file, "err", "file does not exist")
		return http.StatusInternalServerError, components.RenderStatusComponent(w, "internal error")
	}

	source, err := h.webcli.SourceFile(pkgPath, fileName)
	if err != nil {
		h.logger.Error("unable to get source file", "file", fileName, "err", err)
		return http.StatusInternalServerError, components.RenderStatusComponent(w, "internal error")
	}

	// XXX: we should either do this on the front or in the markdown parsing side
	fileLines := strings.Count(string(source), "\n")
	fileSizeKb := float64(len(source)) / 1024.0
	fileSizeStr := fmt.Sprintf("%.2f Kb", fileSizeKb)

	// Highlight code source
	hsource, err := h.highlightSource(fileName, source)
	if err != nil {
		h.logger.Error("unable to highlight source file", "file", fileName, "err", err)
		return http.StatusInternalServerError, components.RenderStatusComponent(w, "internal error")
	}

	err = components.RenderSourceComponent(w, components.SourceData{
		PkgPath:     gnourl.Path,
		Files:       files,
		FileName:    fileName,
		FileCounter: len(files),
		FileLines:   fileLines,
		FileSize:    fileSizeStr,
		FileSource:  template.HTML(hsource), //nolint:gosec
	})
	if err != nil {
		h.logger.Error("unable to render helper", "err", err)
		return http.StatusInternalServerError, components.RenderStatusComponent(w, "internal error")
	}

	return http.StatusOK, nil
}

func (h *WebHandler) renderRealmDirectory(w io.Writer, gnourl *GnoURL) (status int, err error) {
	pkgPath := gnourl.Path

	files, err := h.webcli.Sources(pkgPath)
	if err != nil {
		h.logger.Error("unable to list sources file", "path", gnourl.Path, "err", err)
		return http.StatusInternalServerError, components.RenderStatusComponent(w, "internal error")
	}

	if len(files) == 0 {
		h.logger.Debug("no files available", "path", gnourl.Path)
		return http.StatusOK, components.RenderStatusComponent(w, "no files available")
	}

	err = components.RenderDirectoryComponent(w, components.DirData{
		PkgPath:     gnourl.Path,
		Files:       files,
		FileCounter: len(files),
	})
	if err != nil {
		h.logger.Error("unable to render directory", "err", err)
		return http.StatusInternalServerError, components.RenderStatusComponent(w, "internal error")
	}

	return http.StatusOK, nil
}

func (h *WebHandler) highlightSource(fileName string, src []byte) ([]byte, error) {
	var lexer chroma.Lexer

	switch strings.ToLower(filepath.Ext(fileName)) {
	case ".gno":
		lexer = lexers.Get("go")
	case ".md":
		lexer = lexers.Get("markdown")
	default:
		lexer = lexers.Get("txt") // file kind not supported, fallback on `.txt`
	}

	if lexer == nil {
		return nil, fmt.Errorf("unsupported lexer for file %q", fileName)
	}

	iterator, err := lexer.Tokenise(nil, string(src))
	if err != nil {
		h.logger.Error("unable to ", "fileName", fileName, "err", err)
	}

	var buff bytes.Buffer
	if err := h.formatter.Format(&buff, iterator); err != nil {
		return nil, fmt.Errorf("unable to format source file %q: %w", fileName, err)
	}

	return buff.Bytes(), nil
}

func generateBreadcrumbPaths(url *GnoURL) components.BreadcrumbData {
	split := strings.Split(url.Path, "/")

	var data components.BreadcrumbData
	var name string
	for i := range split {
		if name = split[i]; name == "" {
			continue
		}

		data.Parts = append(data.Parts, components.BreadcrumbPart{
			Name: name,
			URL:  strings.Join(split[:i+1], "/"),
		})
	}

	if args := url.EncodeArgs(); args != "" {
		data.Args = args
	}

	return data
}
