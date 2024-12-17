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

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/sdk/vm" // for error types
)

type StaticMetadata struct {
	Domain     string
	AssetsPath string
	ChromaPath string
	RemoteHelp string
	ChainId    string
	Analytics  bool
}

type WebHandlerConfig struct {
	Meta StaticMetadata

	WebClient       Client
	SourceFormatter Highlighter
}

func (cfg WebHandlerConfig) validate() error {
	if cfg.WebClient == nil {
		return fmt.Errorf("no `Webclient` configured")
	}

	if cfg.SourceFormatter == nil {
		return fmt.Errorf("no `SourceFormatter` configured")
	}

	return nil
}

type WebHandler struct {
	logger *slog.Logger

	Static StaticMetadata
	Client Client
}

func NewWebHandler(logger *slog.Logger, cfg WebHandlerConfig) (*WebHandler, error) {
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validate error: %w", err)
	}

	return &WebHandler{
		Client: cfg.WebClient,
		Static: cfg.Meta,

		logger: logger,
	}, nil
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
	indexData.HeadData.AssetsPath = h.Static.AssetsPath
	indexData.HeadData.ChromaPath = h.Static.ChromaPath
	indexData.FooterData.Analytics = h.Static.Analytics
	indexData.FooterData.AssetsPath = h.Static.AssetsPath

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
		indexData.HeaderData.Breadcrumb.Parts = generateBreadcrumbPaths(gnourl.Path)
		indexData.HeaderData.WebQuery = gnourl.WebQuery

		// Render
		switch gnourl.Kind() {
		case KindRealm, KindPure:
			status, err = h.GetPackagePage(&body, gnourl)
		default:
			h.logger.Debug("invalid page kind", "kind", gnourl.Kind)
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

func (h *WebHandler) GetPackagePage(w io.Writer, gnourl *GnoURL) (status int, err error) {
	h.logger.Info("component render", "path", gnourl.Path, "args", gnourl.Args)

	kind := gnourl.Kind()

	// Display realm help page?
	if kind == KindRealm && gnourl.WebQuery.Has("help") {
		return h.GetHelpPage(w, gnourl)
	}

	// Display package source page?
	switch {
	case gnourl.WebQuery.Has("source"):
		return h.GetSource(w, gnourl)
	case kind == KindPure, gnourl.IsFile(), gnourl.IsDir():
		i := strings.LastIndexByte(gnourl.Path, '/')
		if i < 0 {
			return http.StatusInternalServerError, fmt.Errorf("unable to get ending slash for %q", gnourl.Path)
		}

		// Fill webquery with file infos
		gnourl.WebQuery.Set("source", "") // set source

		file := gnourl.Path[i+1:]
		// If there nothing after the last slash that mean its a
		// directory ...
		if file == "" {
			return h.GetDirectoryPage(w, gnourl)
		}

		// ... else, remaining part is a file
		gnourl.WebQuery.Set("file", file)
		gnourl.Path = gnourl.Path[:i]

		return h.GetSource(w, gnourl)
	}

	// Render content into the content buffer
	var content bytes.Buffer
	meta, err := h.Client.Render(&content, gnourl.Path, gnourl.EncodeArgs())
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

func (h *WebHandler) GetHelpPage(w io.Writer, gnourl *GnoURL) (status int, err error) {
	fsigs, err := h.Client.Functions(gnourl.Path)
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
		ChainId:      h.Static.ChainId,
		// TODO: get chain domain and use that.
		PkgPath:   filepath.Join(h.Static.Domain, gnourl.Path),
		Remote:    h.Static.RemoteHelp,
		Functions: fsigs,
	})
	if err != nil {
		h.logger.Error("unable to render helper", "err", err)
		return http.StatusInternalServerError, components.RenderStatusComponent(w, "internal error")
	}

	return http.StatusOK, nil

}

func (h *WebHandler) GetRealmPage(w io.Writer, gnourl *GnoURL) (status int, err error) {
	fsigs, err := h.Client.Functions(gnourl.Path)
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
		ChainId:      h.Static.ChainId,
		// TODO: get chain domain and use that.
		PkgPath:   filepath.Join(h.Static.Domain, gnourl.Path),
		Remote:    h.Static.RemoteHelp,
		Functions: fsigs,
	})
	if err != nil {
		h.logger.Error("unable to render helper", "err", err)
		return http.StatusInternalServerError, components.RenderStatusComponent(w, "internal error")
	}

	return http.StatusOK, nil
}

func (h *WebHandler) GetSource(w io.Writer, gnourl *GnoURL) (status int, err error) {
	pkgPath := gnourl.Path

	files, err := h.Client.Sources(pkgPath)
	if err != nil {
		h.logger.Error("unable to list sources file", "path", gnourl.Path, "err", err)
		return http.StatusInternalServerError, components.RenderStatusComponent(w, "internal error")
	}

	if len(files) == 0 {
		h.logger.Debug("no files available", "path", gnourl.Path)
		return http.StatusOK, components.RenderStatusComponent(w, "no files available")
	}

	var fileName string
	file := gnourl.WebQuery.Get("file")
	if file == "" {
		fileName = files[0]
	} else if slices.Contains(files, file) {
		fileName = file
	} else {
		h.logger.Error("unable to render source", "file", file, "err", "file does not exist")
		return http.StatusInternalServerError, components.RenderStatusComponent(w, "internal error")
	}

	var source bytes.Buffer
	meta, err := h.Client.SourceFile(&source, pkgPath, fileName)
	if err != nil {
		h.logger.Error("unable to get source file", "file", fileName, "err", err)
		return http.StatusInternalServerError, components.RenderStatusComponent(w, "internal error")
	}

	// XXX: we should either do this on the front or in the markdown parsing side
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
		h.logger.Error("unable to render helper", "err", err)
		return http.StatusInternalServerError, components.RenderStatusComponent(w, "internal error")
	}

	return http.StatusOK, nil
}

func (h *WebHandler) GetDirectoryPage(w io.Writer, gnourl *GnoURL) (status int, err error) {
	pkgPath := gnourl.Path

	files, err := h.Client.Sources(pkgPath)
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

func generateBreadcrumbPaths(path string) []components.BreadcrumbPart {
	split := strings.Split(path, "/")
	parts := []components.BreadcrumbPart{}

	var name string
	for i := range split {
		if name = split[i]; name == "" {
			continue
		}

		parts = append(parts, components.BreadcrumbPart{
			Name: name,
			Path: strings.Join(split[:i+1], "/"),
		})
	}

	return parts
}
