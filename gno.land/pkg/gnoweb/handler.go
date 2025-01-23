package gnoweb

import (
	"bytes"
	"errors"
	"fmt"
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

	w.Header().Add("Content-Type", "text/html")
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
	status, indexData.BodyView = h.prepareIndexBodyView(r, &indexData)

	// Render the final page with the rendered body
	w.WriteHeader(status)
	if err := components.IndexLayout(indexData).Render(w); err != nil {
		h.Logger.Error("failed to render index component", "error", err)
	}
}

// prepareIndexBodyView prepares the data and main view for the index.
func (h *WebHandler) prepareIndexBodyView(r *http.Request, indexData *components.IndexData) (int, *components.View) {
	gnourl, err := ParseGnoURL(r.URL)
	if err != nil {
		h.Logger.Warn("unable to parse url path", "path", r.URL.Path, "error", err)
		return http.StatusNotFound, components.StatusComponent("invalid path")
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
		return h.GetPackageView(gnourl)
	default:
		h.Logger.Debug("invalid path: path is neither a pure package or a realm")
		return http.StatusBadRequest, components.StatusComponent("invalid path")
	}
}

// GetPackageView handles package pages.
func (h *WebHandler) GetPackageView(gnourl *GnoURL) (int, *components.View) {
	// Handle Help page
	if gnourl.WebQuery.Has("help") {
		return h.GetHelpView(gnourl)
	}

	// Handle Source page
	if gnourl.WebQuery.Has("source") || gnourl.IsFile() {
		return h.GetSourceView(gnourl)
	}

	// Handle Source page
	if gnourl.IsDir() || gnourl.IsPure() {
		return h.GetDirectoryView(gnourl)
	}

	// Ultimately get realm view
	return h.GetRealmView(gnourl)
}

func (h *WebHandler) GetRealmView(gnourl *GnoURL) (int, *components.View) {
	var content bytes.Buffer

	meta, err := h.Client.RenderRealm(&content, gnourl.Path, gnourl.EncodeArgs())
	if err != nil {
		h.Logger.Error("unable to render realm", "error", err, "path", gnourl.EncodeURL())
		return GetClientErrorStatusPage(gnourl, err)
	}

	return http.StatusOK, components.RealmView(components.RealmData{
		TocItems: &components.RealmTOCData{
			Items: meta.Toc.Items,
		},

		// NOTE: `RenderRealm` should ensure that HTML content is
		// sanitized before rendering
		ComponentContent: components.NewReaderComponent(&content),
	})
}

func (h *WebHandler) GetHelpView(gnourl *GnoURL) (int, *components.View) {
	fsigs, err := h.Client.Functions(gnourl.Path)
	if err != nil {
		h.Logger.Error("unable to fetch path functions", "error", err)
		return GetClientErrorStatusPage(gnourl, err)
	}

	// Get selected function
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
	return http.StatusOK, components.HelpView(components.HelpData{
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

func (h *WebHandler) GetSourceView(gnourl *GnoURL) (int, *components.View) {
	pkgPath := gnourl.Path
	files, err := h.Client.Sources(pkgPath)
	if err != nil {
		h.Logger.Error("unable to list sources file", "path", gnourl.Path, "error", err)
		return GetClientErrorStatusPage(gnourl, err)
	}

	if len(files) == 0 {
		h.Logger.Debug("no files available", "path", gnourl.Path)
		return http.StatusOK, components.StatusComponent("no files available")
	}

	var fileName string
	if gnourl.IsFile() { // check path file from path first
		fileName = gnourl.File
	} else if file := gnourl.WebQuery.Get("file"); file != "" {
		fileName = file
	}

	if fileName == "" {
		fileName = files[0] // fallback on the first file
	}

	var source bytes.Buffer
	meta, err := h.Client.SourceFile(&source, pkgPath, fileName)
	if err != nil {
		h.Logger.Error("unable to get source file", "file", fileName, "error", err)
		return GetClientErrorStatusPage(gnourl, err)
	}

	fileSizeStr := fmt.Sprintf("%.2f Kb", meta.SizeKb)
	return http.StatusOK, components.SourceView(components.SourceData{
		PkgPath:     gnourl.Path,
		Files:       files,
		FileName:    fileName,
		FileCounter: len(files),
		FileLines:   meta.Lines,
		FileSize:    fileSizeStr,
		FileSource:  components.NewReaderComponent(&source),
	})
}

func (h *WebHandler) GetDirectoryView(gnourl *GnoURL) (int, *components.View) {
	pkgPath := strings.TrimSuffix(gnourl.Path, "/")
	files, err := h.Client.Sources(pkgPath)
	if err != nil {
		h.Logger.Error("unable to list sources file", "path", gnourl.Path, "error", err)
		return GetClientErrorStatusPage(gnourl, err)
	}

	if len(files) == 0 {
		h.Logger.Debug("no files available", "path", gnourl.Path)
		return http.StatusOK, components.StatusComponent("no files available")
	}

	return http.StatusOK, components.DirectoryView(components.DirData{
		PkgPath:     gnourl.Path,
		Files:       files,
		FileCounter: len(files),
	})
}

func GetClientErrorStatusPage(_ *GnoURL, err error) (int, *components.View) {
	if err == nil {
		return http.StatusOK, nil
	}

	switch {
	case errors.Is(err, ErrClientPathNotFound):
		return http.StatusNotFound, components.StatusComponent(err.Error())
	case errors.Is(err, ErrClientBadRequest):
		return http.StatusInternalServerError, components.StatusComponent("bad request")
	case errors.Is(err, ErrClientResponse):
		fallthrough // XXX: for now fallback as internal error
	default:
		return http.StatusInternalServerError, components.StatusComponent("internal error")
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
