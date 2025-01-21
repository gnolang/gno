package gnoweb

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm" // For error types
)

// StaticMetadata holds static configuration for a web handler.
type StaticMetadata struct {
	Domain     string
	AssetsPath string
	ChromaPath string
	RemoteHelp string
	ChainId    string
	Analytics  bool
}

// WebHandlerConfig configures a WebHandler.
type WebHandlerConfig struct {
	Meta      StaticMetadata
	WebClient WebClient
}

// validate checks if the WebHandlerConfig is valid.
func (cfg WebHandlerConfig) validate() error {
	if cfg.WebClient == nil {
		return errors.New("no `WebClient` configured")
	}
	return nil
}

// WebHandler processes HTTP requests.
type WebHandler struct {
	Logger *slog.Logger
	Static StaticMetadata
	Client WebClient
}

// PageData groups layout, component, and dev mode information.
type PageData struct {
	Layout       string
	Component    string
	IsDevmodView bool
}

// NewWebHandler creates a new WebHandler.
func NewWebHandler(logger *slog.Logger, cfg WebHandlerConfig) (*WebHandler, error) {
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validate error: %w", err)
	}

	return &WebHandler{
		Client: cfg.WebClient,
		Static: cfg.Meta,
		Logger: logger,
	}, nil
}

// ServeHTTP handles HTTP requests.
func (h *WebHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Debug("receiving request", "method", r.Method, "path", r.URL.Path)

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.Get(w, r)
}

// Get processes a GET HTTP request.
func (h *WebHandler) Get(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		h.Logger.Debug("request completed",
			"url", r.URL.String(),
			"elapsed", time.Since(start).String())
	}()

	indexData := components.IndexData{
		HeadData: components.HeadData{
			AssetsPath: h.Static.AssetsPath,
			ChromaPath: h.Static.ChromaPath,
		},
		FooterData: components.FooterData{
			Analytics:  h.Static.Analytics,
			AssetsPath: h.Static.AssetsPath,
		},
	}

	var status int
	status, indexData.View = h.renderPage(r, &indexData)

	w.WriteHeader(status)

	// Render the final page with the rendered body
	if err := components.RenderIndexLayout(w, indexData); err != nil {
		h.Logger.Error("failed to render index component", "err", err)
	}
}

// renderPage renders the page into the given buffer and prepares the index data.
func (h *WebHandler) renderPage(r *http.Request, indexData *components.IndexData) (int, *components.View) {
	gnourl, err := ParseGnoURL(r.URL)
	if err != nil {
		h.Logger.Warn("unable to parse url path", "path", r.URL.Path, "err", err)
		// components.ContentData{"full", "error", false}
		return http.StatusNotFound, components.RenderStatusComponent("invalid path")
	}

	breadcrumb := generateBreadcrumbPaths(gnourl)
	indexData.HeadData.Title = h.Static.Domain + " - " + gnourl.Path
	indexData.HeaderData = components.HeaderData{
		RealmPath:  gnourl.Encode(EncodePath | EncodeArgs | EncodeQuery | EncodeNoEscape),
		Breadcrumb: breadcrumb,
		WebQuery:   gnourl.WebQuery,
	}

	switch {
	case gnourl.IsRealm(), gnourl.IsPure():
		return h.GetPackagePage(gnourl)
	default:
		h.Logger.Debug("invalid path: path is neither a pure package or a realm")
		return http.StatusBadRequest, components.RenderStatusComponent("invalid path")
	}
}

// GetPackagePage handles package pages.
func (h *WebHandler) GetPackagePage(gnourl *GnoURL) (int, *components.View) {
	// Handle Help page
	if gnourl.WebQuery.Has("help") {
		return h.GetHelpPage(gnourl)
	}

	// Handle Source page
	if gnourl.WebQuery.Has("source") || gnourl.IsFile() {
		return h.GetSourcePage(gnourl)
	}

	// Handle Source page
	if gnourl.IsDir() || gnourl.IsPure() {
		return h.GetDirectoryPage(gnourl)
	}

	// Ultimately render realm content
	return h.renderRealmContent(gnourl)
}

// renderRealmContent renders the content of a realm.
func (h *WebHandler) renderRealmContent(gnourl *GnoURL) (int, *components.View) {
	var content bytes.Buffer

	meta, err := h.Client.RenderRealm(&content, gnourl.Path, gnourl.EncodeArgs())
	if err != nil {
		h.Logger.Error("unable to render realm", "err", err, "path", gnourl.EncodeArgs())
		return renderClientErrorStatusPage(gnourl, err)
	}

	return http.StatusOK, components.RenderRealmView(components.RealmData{
		TocItems: &components.RealmTOCData{
			Items: meta.Toc.Items,
		},

		// NOTE: `RenderRealm` should ensure that HTML content is
		// sanitized before rendering
		Content: components.NewReaderComponent(&content),
	})
}

// GetHelpPage renders the help page.
func (h *WebHandler) GetHelpPage(gnourl *GnoURL) (int, *components.View) {
	fsigs, err := h.Client.Functions(gnourl.Path)
	if err != nil {
		h.Logger.Error("unable to fetch path functions", "err", err)
		return renderClientErrorStatusPage(gnourl, err)
	}

	selArgs := make(map[string]string)
	selFn := gnourl.WebQuery.Get("func")
	if selFn != "" {
		for _, fn := range fsigs {
			if selFn != fn.FuncName {
				continue
			}

			for _, param := range fn.Params {
				selArgs[param.Name] = gnourl.WebQuery.Get(param.Name)
			}

			fsigs = []vm.FunctionSignature{fn}
			break
		}
	}

	realmName := filepath.Base(gnourl.Path)
	return http.StatusOK, components.RenderHelpView(components.HelpData{
		SelectedFunc: selFn,
		SelectedArgs: selArgs,
		RealmName:    realmName,
		// TODO: get chain domain and use that.
		ChainId:   h.Static.ChainId,
		PkgPath:   filepath.Join(h.Static.Domain, gnourl.Path),
		Remote:    h.Static.RemoteHelp,
		Functions: fsigs,
	})
}

// GetSource renders the source page.
func (h *WebHandler) GetSourcePage(gnourl *GnoURL) (int, *components.View) {
	pkgPath := gnourl.Path
	files, err := h.Client.Sources(pkgPath)
	if err != nil {
		h.Logger.Error("unable to list sources file", "path", gnourl.Path, "err", err)
		return renderClientErrorStatusPage(gnourl, err)
	}

	if len(files) == 0 {
		h.Logger.Debug("no files available", "path", gnourl.Path)
		return http.StatusOK, components.RenderStatusComponent("no files available")
	}

	var fileName string
	if gnourl.IsFile() { // check path file from path first
		fileName = gnourl.File
	} else if file := gnourl.WebQuery.Get("file"); file != "" {
		fileName = file
	}

	if fileName == "" {
		fileName = files[0] // fallback on the first file if
	}

	var source bytes.Buffer
	meta, err := h.Client.SourceFile(&source, pkgPath, fileName)
	if err != nil {
		h.Logger.Error("unable to get source file", "file", fileName, "err", err)
		status, renderErr := renderClientErrorStatusPage(gnourl, err)
		return status, renderErr
	}

	fileSizeStr := fmt.Sprintf("%.2f Kb", meta.SizeKb)
	return http.StatusOK, components.RenderSourceView(components.SourceData{
		PkgPath:     gnourl.Path,
		Files:       files,
		FileName:    fileName,
		FileCounter: len(files),
		FileLines:   meta.Lines,
		FileSize:    fileSizeStr,
		FileSource:  template.HTML(source.String()), //nolint:gosec
	})
}

// GetDirectoryPage renders the directory page.
func (h *WebHandler) GetDirectoryPage(gnourl *GnoURL) (int, *components.View) {
	pkgPath := strings.TrimSuffix(gnourl.Path, "/")
	files, err := h.Client.Sources(pkgPath)
	if err != nil {
		h.Logger.Error("unable to list sources file", "path", gnourl.Path, "err", err)
		return renderClientErrorStatusPage(gnourl, err)
	}

	if len(files) == 0 {
		h.Logger.Debug("no files available", "path", gnourl.Path)
		return http.StatusOK, components.RenderStatusComponent("no files available")
	}

	return http.StatusOK, components.RenderDirectoryView(components.DirData{
		PkgPath:     gnourl.Path,
		Files:       files,
		FileCounter: len(files),
	})
}

func renderClientErrorStatusPage(_ *GnoURL, err error) (int, *components.View) {
	if err == nil {
		return http.StatusOK, nil
	}

	switch {
	case errors.Is(err, ErrClientPathNotFound):
		return http.StatusNotFound, components.RenderStatusComponent(err.Error())
	case errors.Is(err, ErrClientBadRequest):
		return http.StatusInternalServerError, components.RenderStatusComponent("bad request")
	case errors.Is(err, ErrClientResponse):
		fallthrough // XXX: for now fallback as internal error
	default:
		return http.StatusInternalServerError, components.RenderStatusComponent("internal error")
	}
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
