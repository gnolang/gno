package gnoweb

import (
	"bytes"
	"errors"
	"fmt"
	"go/token"
	"log/slog"
	"net/http"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/gnolang/gno/gnovm/pkg/doc"
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

type AliasKind int

const (
	GnowebPath AliasKind = iota
	StaticMarkdown
)

type AliasTarget struct {
	Value string
	Kind  AliasKind
}

// WebHandlerConfig configures a WebHandler.
type WebHandlerConfig struct {
	Meta             StaticMetadata
	WebClient        WebClient
	MarkdownRenderer *MarkdownRenderer
	Aliases          map[string]AliasTarget
}

// validate checks if the WebHandlerConfig is valid.
func (cfg *WebHandlerConfig) validate() error {
	if cfg.WebClient == nil {
		return errors.New("no `WebClient` configured")
	}
	if cfg.MarkdownRenderer == nil {
		return errors.New("no `MarkdownRenderer` configured")
	}
	if cfg.Aliases == nil {
		return errors.New("no `Aliases` configured")
	}
	return nil
}

// IsHomePath checks if the given path is the home path.
func IsHomePath(path string) bool {
	return path == "/"
}

// WebHandler processes HTTP requests.
type WebHandler struct {
	Logger           *slog.Logger
	Static           StaticMetadata
	Client           WebClient
	MarkdownRenderer *MarkdownRenderer
	Aliases          map[string]AliasTarget
}

// NewWebHandler creates a new WebHandler.
func NewWebHandler(logger *slog.Logger, cfg *WebHandlerConfig) (*WebHandler, error) {
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validate error: %w", err)
	}

	return &WebHandler{
		Client:           cfg.WebClient,
		Static:           cfg.Meta,
		MarkdownRenderer: cfg.MarkdownRenderer,
		Aliases:          cfg.Aliases,
		Logger:           logger,
	}, nil
}

// ServeHTTP handles HTTP requests.
func (h *WebHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Debug("receiving request", "method", r.Method, "path", r.URL.Path)

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Add("Content-Type", "text/html; charset=utf-8")
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
			ChainId:    h.Static.ChainId,
			Remote:     h.Static.RemoteHelp,
		},
		FooterData: components.FooterData{
			Analytics:  h.Static.Analytics,
			AssetsPath: h.Static.AssetsPath,
		},
	}

	// Parse the URL
	gnourl, err := weburl.ParseFromURL(r.URL)
	if err != nil {
		h.Logger.Warn("unable to parse url path", "path", r.URL.Path, "error", err)

		indexData.HeadData.Title = "gno.land — invalid path"
		indexData.BodyView = components.StatusErrorComponent("invalid path")
		w.WriteHeader(http.StatusNotFound)
		if err := components.IndexLayout(indexData).Render(w); err != nil {
			h.Logger.Error("failed to render error view", "error", err)
		}
		return
	}

	// Handle download request outside of component rendering flow.
	if gnourl.WebQuery.Has("download") {
		h.GetSourceDownload(gnourl, w, r)
		return
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
	aliasTarget, aliasExists := h.Aliases[r.URL.Path]

	// If the alias target exists and is a gnoweb path, replace the URL path with it.
	if aliasExists && aliasTarget.Kind == GnowebPath {
		r.URL.Path = aliasTarget.Value
	}

	gnourl, err := weburl.ParseFromURL(r.URL)
	if err != nil {
		h.Logger.Warn("invalid gno url path", "path", r.URL.Path, "error", err)
		return http.StatusNotFound, components.StatusErrorComponent("invalid path")
	}

	indexData.HeadData.Title = h.Static.Domain + " - " + gnourl.Path
	indexData.HeaderData = components.HeaderData{
		Breadcrumb: generateBreadcrumbPaths(gnourl),
		RealmURL:   *gnourl,
		ChainId:    h.Static.ChainId,
		Remote:     h.Static.RemoteHelp,
		IsHome:     IsHomePath(r.RequestURI),
	}

	switch {
	case aliasExists && aliasTarget.Kind == StaticMarkdown:
		return h.GetMarkdownView(gnourl, aliasTarget.Value)
	case gnourl.IsRealm(), gnourl.IsPure():
		return h.GetPackageView(gnourl)
	default:
		h.Logger.Debug("invalid path: path is neither a pure package or a realm")
		return http.StatusBadRequest, components.StatusErrorComponent("invalid path")
	}
}

// GetMarkdownView handles markdown files.
func (h *WebHandler) GetMarkdownView(gnourl *weburl.GnoURL, mdContent string) (int, *components.View) {
	var content bytes.Buffer

	// Use Goldmark for Markdown parsing
	toc, err := h.MarkdownRenderer.Render(&content, gnourl, []byte(mdContent))
	if err != nil {
		h.Logger.Error("unable to render markdown file", "error", err, "path", gnourl.EncodeURL())
		return GetClientErrorStatusPage(gnourl, err)
	}

	return http.StatusOK, components.RealmView(components.RealmData{
		TocItems:         &components.RealmTOCData{Items: toc.Items},
		ComponentContent: components.NewReaderComponent(&content),
	})
}

// GetPackageView handles package pages.
func (h *WebHandler) GetPackageView(gnourl *weburl.GnoURL) (int, *components.View) {
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

func (h *WebHandler) GetRealmView(gnourl *weburl.GnoURL) (int, *components.View) {
	var content bytes.Buffer

	meta, err := h.Client.RenderRealm(&content, gnourl, h.MarkdownRenderer)
	if err != nil {
		if errors.Is(err, ErrRenderNotDeclared) {
			return http.StatusOK, components.StatusNoRenderComponent(gnourl.Path)
		}

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

func (h *WebHandler) GetHelpView(gnourl *weburl.GnoURL) (int, *components.View) {
	jdoc, err := h.Client.Doc(gnourl.Path)
	if err != nil {
		h.Logger.Error("unable to fetch qdoc", "error", err)
		return GetClientErrorStatusPage(gnourl, err)
	}

	// Get public non-method funcs
	fsigs := []*doc.JSONFunc{}
	for _, fun := range jdoc.Funcs {
		if !(fun.Type == "" && token.IsExported(fun.Name)) {
			continue
		}

		fsigs = append(fsigs, fun)
	}

	// Get selected function
	selArgs := make(map[string]string)
	selFn := gnourl.WebQuery.Get("func")
	selSend := gnourl.WebQuery.Get(".send")
	if selFn != "" {
		for _, fn := range fsigs {
			if selFn != fn.Name {
				continue
			}

			for _, param := range fn.Params {
				selArgs[param.Name] = gnourl.WebQuery.Get(param.Name)
			}

			fsigs = []*doc.JSONFunc{fn}
			break
		}
	}

	realmName := path.Base(gnourl.Path)
	return http.StatusOK, components.HelpView(components.HelpData{
		SelectedFunc: selFn,
		SelectedArgs: selArgs,
		SelectedSend: selSend,
		RealmName:    realmName,
		// TODO: get chain domain and use that.
		ChainId:   h.Static.ChainId,
		PkgPath:   path.Join(h.Static.Domain, gnourl.Path),
		Remote:    h.Static.RemoteHelp,
		Functions: fsigs,
		Doc:       jdoc.PackageDoc,
	})
}

func (h *WebHandler) GetSourceView(gnourl *weburl.GnoURL) (int, *components.View) {
	pkgPath := gnourl.Path
	files, err := h.Client.Sources(pkgPath)
	if err != nil {
		h.Logger.Error("unable to list sources file", "path", gnourl.Path, "error", err)
		return GetClientErrorStatusPage(gnourl, err)
	}

	if len(files) == 0 {
		h.Logger.Debug("no files available", "path", gnourl.Path)
		return http.StatusOK, components.StatusErrorComponent("no files available")
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
	meta, err := h.Client.SourceFile(&source, pkgPath, fileName, false)
	if err != nil {
		h.Logger.Error("unable to get source file", "file", fileName, "error", err)
		return GetClientErrorStatusPage(gnourl, err)
	}

	fileSizeStr := fmt.Sprintf("%.2f Kb", meta.SizeKb)
	return http.StatusOK, components.SourceView(components.SourceData{
		PkgPath:      gnourl.Path,
		Files:        files,
		FileName:     fileName,
		FileCounter:  len(files),
		FileLines:    meta.Lines,
		FileSize:     fileSizeStr,
		FileDownload: gnourl.Path + "$download&file=" + fileName,
		FileSource:   components.NewReaderComponent(&source),
	})
}

func (h *WebHandler) GetDirectoryView(gnourl *weburl.GnoURL) (int, *components.View) {
	pkgPath := strings.TrimSuffix(gnourl.Path, "/")
	files, err := h.Client.Sources(pkgPath)
	if err != nil {
		h.Logger.Error("unable to list sources file", "path", gnourl.Path, "error", err)
		return GetClientErrorStatusPage(gnourl, err)
	}

	if len(files) == 0 {
		h.Logger.Debug("no files available", "path", gnourl.Path)
		return http.StatusOK, components.StatusErrorComponent("no files available")
	}

	return http.StatusOK, components.DirectoryView(components.DirData{
		PkgPath:     gnourl.Path,
		Files:       files,
		FileCounter: len(files),
	})
}

func (h *WebHandler) GetSourceDownload(gnourl *weburl.GnoURL, w http.ResponseWriter, r *http.Request) {
	pkgPath := gnourl.Path

	var fileName string
	if gnourl.IsFile() { // check path file from path first
		fileName = gnourl.File
	} else if file := gnourl.WebQuery.Get("file"); file != "" {
		fileName = file
	}

	if fileName == "" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// Get source file
	var source bytes.Buffer
	_, err := h.Client.SourceFile(&source, pkgPath, fileName, true)
	if err != nil {
		h.Logger.Error("unable to get source file", "file", fileName, "error", err)
		status, _ := GetClientErrorStatusPage(gnourl, err)
		http.Error(w, "not found", status)
		return
	}

	// Send raw file as attachment for download (without HTML formating)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", fileName))
	w.WriteHeader(http.StatusOK)
	source.WriteTo(w)
}

func GetClientErrorStatusPage(_ *weburl.GnoURL, err error) (int, *components.View) {
	if err == nil {
		return http.StatusOK, nil
	}

	switch {
	case errors.Is(err, ErrClientPathNotFound):
		return http.StatusNotFound, components.StatusErrorComponent(err.Error())
	case errors.Is(err, ErrClientBadRequest):
		return http.StatusInternalServerError, components.StatusErrorComponent("bad request")
	case errors.Is(err, ErrClientResponse):
		fallthrough // XXX: for now fallback as internal error
	default:
		return http.StatusInternalServerError, components.StatusErrorComponent("internal error")
	}
}

func generateBreadcrumbPaths(url *weburl.GnoURL) components.BreadcrumbData {
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

	// Add args
	if url.Args != "" {
		argSplit := strings.Split(url.Args, "/")
		nonEmptyArgs := slices.DeleteFunc(argSplit, func(a string) bool {
			return a == ""
		})

		for i := range nonEmptyArgs {
			data.ArgParts = append(data.ArgParts, components.BreadcrumbPart{
				Name: nonEmptyArgs[i],
				URL:  url.Path + ":" + strings.Join(nonEmptyArgs[:i+1], "/"),
			})
		}
	}

	// Add query params
	for key, values := range url.Query {
		for _, v := range values {
			data.Queries = append(data.Queries, components.QueryParam{
				Key:   key,
				Value: v,
			})
		}
	}

	return data
}
