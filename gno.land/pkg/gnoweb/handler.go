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
	var body bytes.Buffer

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

	status, err := h.renderPage(&body, r, &indexData)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)

	// NOTE: HTML escaping should have already been done by markdown rendering package
	indexData.Body = template.HTML(body.String()) //nolint:gosec

	// Render the final page with the rendered body
	if err = components.RenderIndexComponent(w, indexData); err != nil {
		h.Logger.Error("failed to render index component", "err", err)
	}
}

// renderPage renders the page into the given buffer and prepares the index data.
func (h *WebHandler) renderPage(body *bytes.Buffer, r *http.Request, indexData *components.IndexData) (int, error) {
	gnourl, err := ParseGnoURL(r.URL)
	if err != nil {
		h.Logger.Warn("unable to parse url path", "path", r.URL.Path, "err", err)
		return http.StatusNotFound, components.RenderStatusComponent(body, "invalid path")
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
		return h.GetPackagePage(body, gnourl)
	default:
		h.Logger.Debug("invalid path: path is neither a pure package or a realm")
		return http.StatusBadRequest, components.RenderStatusComponent(body, "invalid path")
	}
}

// GetPackagePage handles package pages.
func (h *WebHandler) GetPackagePage(w io.Writer, gnourl *GnoURL) (int, error) {
	h.Logger.Info("component render", "path", gnourl.Path, "args", gnourl.Args)

	// Handle Help page
	if gnourl.WebQuery.Has("help") {
		return h.GetHelpPage(w, gnourl)
	}

	// Handle Source page
	if gnourl.WebQuery.Has("source") || gnourl.IsFile() {
		return h.GetSourcePage(w, gnourl)
	}

	// Handle Source page
	if gnourl.IsDir() || gnourl.IsPure() {
		return h.GetDirectoryPage(w, gnourl)
	}

	// Ultimately render realm content
	return h.renderRealmContent(w, gnourl)
}

// renderRealmContent renders the content of a realm.
func (h *WebHandler) renderRealmContent(w io.Writer, gnourl *GnoURL) (int, error) {
	var content bytes.Buffer
	meta, err := h.Client.RenderRealm(&content, gnourl.Path, gnourl.EncodeArgs())
	if err != nil {
		h.Logger.Error("unable to render realm", "err", err, "path", gnourl.EncodeArgs())
		return renderClientErrorStatusPage(w, gnourl, err)
	}

	err = components.RenderRealmComponent(w, components.RealmData{
		TocItems: &components.RealmTOCData{
			Items: meta.Toc.Items,
		},
		// NOTE: `RenderRealm` should ensure that HTML content is
		// sanitized before rendering
		Content: template.HTML(content.String()), //nolint:gosec
	})
	if err != nil {
		h.Logger.Error("unable to render template", "err", err)
		return http.StatusInternalServerError, components.RenderStatusComponent(w, "internal error")
	}

	return http.StatusOK, nil
}

// GetHelpPage renders the help page.
func (h *WebHandler) GetHelpPage(w io.Writer, gnourl *GnoURL) (int, error) {
	fsigs, err := h.Client.Functions(gnourl.Path)
	if err != nil {
		h.Logger.Error("unable to fetch path functions", "err", err)
		return renderClientErrorStatusPage(w, gnourl, err)
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
	err = components.RenderHelpComponent(w, components.HelpData{
		SelectedFunc: selFn,
		SelectedArgs: selArgs,
		RealmName:    realmName,
		ChainId:      h.Static.ChainId,
		// TODO: get chain domain and use that.
		PkgPath:   filepath.Join(h.Static.Domain, gnourl.Path),
		Remote:    h.Static.RemoteHelp,
		Functions: fsigs,
	})
	if err != nil {
		h.Logger.Error("unable to render helper", "err", err)
		return http.StatusInternalServerError, components.RenderStatusComponent(w, "internal error")
	}

	return http.StatusOK, nil
}

// GetSource renders the source page.
func (h *WebHandler) GetSourcePage(w io.Writer, gnourl *GnoURL) (int, error) {
	pkgPath := gnourl.Path
	files, err := h.Client.Sources(pkgPath)
	if err != nil {
		h.Logger.Error("unable to list sources file", "path", gnourl.Path, "err", err)
		return renderClientErrorStatusPage(w, gnourl, err)
	}

	if len(files) == 0 {
		h.Logger.Debug("no files available", "path", gnourl.Path)
		return http.StatusOK, components.RenderStatusComponent(w, "no files available")
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
		return renderClientErrorStatusPage(w, gnourl, err)
	}

	fileSizeStr := fmt.Sprintf("%.2f Kb", meta.SizeKb)
	err = components.RenderSourceComponent(w, components.SourceData{
		PkgPath:     gnourl.Path,
		Files:       files,
		FileName:    fileName,
		FileCounter: len(files),
		FileLines:   meta.Lines,
		FileSize:    fileSizeStr,
		FileSource:  template.HTML(source.String()), //nolint:gosec
	})
	if err != nil {
		h.Logger.Error("unable to render helper", "err", err)
		return http.StatusInternalServerError, components.RenderStatusComponent(w, "internal error")
	}

	return http.StatusOK, nil
}

// GetDirectoryPage renders the directory page.
func (h *WebHandler) GetDirectoryPage(w io.Writer, gnourl *GnoURL) (int, error) {
	pkgPath := strings.TrimSuffix(gnourl.Path, "/")

	files, err := h.Client.Sources(pkgPath)
	if err != nil {
		h.Logger.Error("unable to list sources file", "path", gnourl.Path, "err", err)
		return renderClientErrorStatusPage(w, gnourl, err)
	}

	if len(files) == 0 {
		h.Logger.Debug("no files available", "path", gnourl.Path)
		return http.StatusOK, components.RenderStatusComponent(w, "no files available")
	}

	err = components.RenderDirectoryComponent(w, components.DirData{
		PkgPath:     gnourl.Path,
		Files:       files,
		FileCounter: len(files),
	})
	if err != nil {
		h.Logger.Error("unable to render directory", "err", err)
		return http.StatusInternalServerError, components.RenderStatusComponent(w, "not found")
	}

	return http.StatusOK, nil
}

func renderClientErrorStatusPage(w io.Writer, _ *GnoURL, err error) (int, error) {
	if err == nil {
		return http.StatusOK, nil
	}

	switch {
	case errors.Is(err, ErrClientPathNotFound):
		return http.StatusNotFound, components.RenderStatusComponent(w, err.Error())
	case errors.Is(err, ErrClientBadRequest):
		return http.StatusInternalServerError, components.RenderStatusComponent(w, "bad request")
	case errors.Is(err, ErrClientResponse):
		fallthrough // XXX: for now fallback as internal error
	default:
		return http.StatusInternalServerError, components.RenderStatusComponent(w, "internal error")
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
