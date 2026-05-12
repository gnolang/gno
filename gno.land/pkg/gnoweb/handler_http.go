package gnoweb

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/token"
	"html/template"
	"log/slog"
	"net/http"
	"path"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/gnolang/gno/gnovm/pkg/doc"
	"github.com/gnolang/gno/tm2/pkg/bech32"
)

const ReadmeFileName = "README.md"

// maxStateIDLength bounds attacker-controlled oid/tid query params to
// prevent request→response amplification: without it a small GET can
// flood the upstream RPC payload and inflate the rendered HTML.
const maxStateIDLength = 256

// StaticMetadata holds static configuration for a web handler.
type StaticMetadata struct {
	Domain     string
	AssetsPath string
	ChromaPath string
	RemoteHelp string
	ChainId    string
	Analytics  bool
	BuildTime  string
	Banner     components.BannerData
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

// HTTPHandlerConfig configures an HTTPHandler.
type HTTPHandlerConfig struct {
	Meta          StaticMetadata
	ClientAdapter ClientAdapter
	Renderer      Renderer
	Aliases       map[string]AliasTarget
	Timeout       time.Duration
}

// validate checks if the HTTPHandlerConfig is valid.
func (cfg *HTTPHandlerConfig) validate() error {
	if cfg.ClientAdapter == nil {
		return errors.New("no `ClientAdapter` configured")
	}
	if cfg.Renderer == nil {
		return errors.New("no `Renderer` configured")
	}
	if cfg.Aliases == nil {
		return errors.New("no `Aliases` configured")
	}
	return nil
}

// HTTPHandler processes HTTP requests for gnoweb.
type HTTPHandler struct {
	Logger   *slog.Logger
	Static   StaticMetadata
	Client   ClientAdapter
	Renderer Renderer
	Aliases  map[string]AliasTarget
	Timeout  time.Duration
}

// NewHTTPHandler creates a new HTTPHandler.
func NewHTTPHandler(logger *slog.Logger, cfg *HTTPHandlerConfig) (*HTTPHandler, error) {
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validate error: %w", err)
	}

	return &HTTPHandler{
		Client:   cfg.ClientAdapter,
		Static:   cfg.Meta,
		Renderer: cfg.Renderer,
		Aliases:  cfg.Aliases,
		Timeout:  cfg.Timeout,
		Logger:   logger,
	}, nil
}

// ServeHTTP handles HTTP requests and only allows GET requests.
func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.Logger.Debug("receiving request", "method", r.Method, "path", r.URL.Path)

	switch r.Method {
	case http.MethodGet:
		w.Header().Add("Content-Type", "text/html; charset=utf-8")
		h.Get(w, r)
	case http.MethodPost:
		h.Post(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// Get processes a GET HTTP request and renders the appropriate page.
func (h *HTTPHandler) Get(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		h.Logger.Debug("request completed",
			"url", r.URL.String(),
			"elapsed", time.Since(start).String())
	}()

	if h.Timeout > 0 {
		ctx, cancel := context.WithTimeout(r.Context(), h.Timeout)
		defer cancel()
		r = r.WithContext(ctx)
	}

	// Theme cookie is embedded in the HTML before CSS loads to prevent FOUC.
	theme := readWhitelistedCookie(r, "theme", "light", "dark")

	indexData := components.IndexData{
		HeadData: components.HeadData{
			AssetsPath: h.Static.AssetsPath,
			ChromaPath: h.Static.ChromaPath,
			ChainId:    h.Static.ChainId,
			Remote:     h.Static.RemoteHelp,
			BuildTime:  h.Static.BuildTime,
		},
		FooterData: components.FooterData{
			Analytics:  h.Static.Analytics,
			AssetsPath: h.Static.AssetsPath,
			BuildTime:  h.Static.BuildTime,
		},
		Theme:  theme,
		Banner: h.Static.Banner,
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
		h.ServeSourceDownload(r.Context(), gnourl, w, r)
		return
	}

	// Raw Amino JSON passthrough — stable API surface inherited from
	// ADR-003 (`?state&json` and oid/tid variants). Bypasses the SSR
	// rendering flow entirely.
	if gnourl.WebQuery.Has("state") && gnourl.WebQuery.Has("json") {
		h.ServeStateJSON(r.Context(), gnourl, w)
		return
	}

	// Set the header mode based on the URL type and context
	switch {
	case r.RequestURI == "/": // is home path
		indexData.Mode = components.ViewModeHome
	case gnourl.IsPure():
		indexData.Mode = components.ViewModePackage
	case gnourl.IsUser():
		indexData.Mode = components.ViewModeUser
	default:
		indexData.Mode = components.ViewModeRealm
	}

	var status int
	status, indexData.BodyView = h.prepareIndexBodyView(r, &indexData)

	// Render the final page with the rendered body
	w.WriteHeader(status)
	if err := components.IndexLayout(indexData).Render(w); err != nil {
		h.Logger.Error("failed to render index component", "error", err)
	}
}

// Post processes a POST HTTP request.
func (h *HTTPHandler) Post(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		h.Logger.Debug("request completed",
			"url", r.URL.String(),
			"elapsed", time.Since(start).String())
	}()

	// Parse the form data
	if err := r.ParseForm(); err != nil {
		h.Logger.Error("failed to parse form", "error", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// Parse the URL
	gnourl, err := weburl.ParseFromURL(r.URL)
	if err != nil {
		h.Logger.Warn("unable to parse url path", "path", r.URL.Path, "error", err)
		http.Error(w, "invalid path", http.StatusNotFound)
		return
	}

	// Extract path from hidden form field if present.
	// The value is HTML-escaped in the form and URL-encoded when building the redirect.
	if gnoPath := r.PostForm.Get("__gno_path"); gnoPath != "" {
		gnourl.Args = gnoPath
		// Remove from form data so it's not included in query params
		r.PostForm.Del("__gno_path")
	}

	// Use remaining form data as query
	gnourl.Query = r.PostForm

	// Build redirect URL using EncodeFormURL.
	// url.PathEscape encodes slashes and delimiter characters; the args remain part of
	// the path (e.g. /r/realm:args), not a URL scheme.
	sanitizedRedirectURL := gnourl.EncodeFormURL()

	// Defense-in-depth: validate redirect URL to prevent open redirects,
	// This can happen when path is "/" and file is "evil.domain" -> "//evil.domain"
	if strings.HasPrefix(sanitizedRedirectURL, "//") {
		h.Logger.Warn("blocked unsafe redirect", "url", sanitizedRedirectURL)
		http.Error(w, "invalid redirect", http.StatusBadRequest)
		return
	}

	// Redirect to the new URL
	http.Redirect(w, r, sanitizedRedirectURL, http.StatusSeeOther)
}

// prepareIndexBodyView prepares the data and main view for the index page.
func (h *HTTPHandler) prepareIndexBodyView(r *http.Request, indexData *components.IndexData) (int, *components.View) {
	ctx := r.Context()
	viewMode := readWhitelistedCookie(r, stateViewModeCookie, "tree", "pretty")

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
		Mode:       indexData.Mode,
	}

	switch {
	case aliasExists && aliasTarget.Kind == StaticMarkdown:
		indexData.HeaderData.Static = true
		return h.GetMarkdownView(gnourl, aliasTarget.Value)
	case gnourl.IsRealm(), gnourl.IsPure(), gnourl.IsUser():
		return h.GetPackageView(ctx, gnourl, indexData, viewMode)
	default:
		h.Logger.Debug("invalid path: path is neither a pure package or a realm")
		return http.StatusBadRequest, components.StatusErrorComponent("invalid path")
	}
}

// GetMarkdownView handles rendering of markdown files.
func (h *HTTPHandler) GetMarkdownView(gnourl *weburl.GnoURL, mdContent string) (int, *components.View) {
	var content bytes.Buffer

	// Use Goldmark for Markdown parsing
	toc, err := h.Renderer.RenderRealm(&content, gnourl, []byte(mdContent), RealmRenderContext{
		ChainId: h.Static.ChainId,
		Remote:  h.Static.RemoteHelp,
		Domain:  h.Static.Domain,
	})
	if err != nil {
		h.Logger.Error("unable to render markdown file", "error", err, "path", gnourl.EncodeURL())
		return GetClientErrorStatusPage(gnourl, err)
	}

	return http.StatusOK, components.RealmView(components.RealmData{
		TocItems:         &components.RealmTOCData{Items: toc.Items},
		ComponentContent: components.NewReaderComponent(&content),
	})
}

// GetPackageView handles package pages, including help, source, directory, and user views.
func (h *HTTPHandler) GetPackageView(ctx context.Context, gnourl *weburl.GnoURL, indexData *components.IndexData, viewMode string) (int, *components.View) {
	// Handle Help page
	if gnourl.WebQuery.Has("help") {
		return h.GetHelpView(ctx, gnourl)
	}

	// Handle State explorer page
	if gnourl.WebQuery.Has("state") {
		return h.GetStateView(ctx, gnourl, viewMode)
	}

	// Handle Source page
	if gnourl.WebQuery.Has("source") || gnourl.IsFile() {
		return h.GetSourceView(ctx, gnourl)
	}

	// Handle Source page
	if gnourl.IsDir() || gnourl.IsPure() {
		return h.GetDirectoryView(ctx, gnourl, indexData)
	}

	// Handle User page
	if gnourl.IsUser() {
		return h.GetUserView(ctx, gnourl)
	}

	// Ultimately get realm view
	return h.GetRealmView(ctx, gnourl, indexData)
}

// GetRealmView renders a realm page or returns an error/status if not available.
func (h *HTTPHandler) GetRealmView(ctx context.Context, gnourl *weburl.GnoURL, indexData *components.IndexData) (int, *components.View) {
	// First fecth the realm
	raw, err := h.Client.Realm(ctx, gnourl.Path, gnourl.EncodeArgs())
	switch {
	case err == nil: // ok
	case errors.Is(err, ErrClientRenderNotDeclared):
		// No Render() declared: fall back to directory view (which will show README.md if present)
		return h.GetDirectoryView(ctx, gnourl, indexData)
	case errors.Is(err, ErrClientPackageNotFound):
		// No realm exists here, try to display underlying paths
		return h.GetPathsListView(ctx, gnourl, indexData)
	default:
		h.Logger.Error("unable to fetch realm", "error", err, "path", gnourl.EncodeURL())
		return GetClientErrorStatusPage(gnourl, err)
	}

	var content bytes.Buffer
	meta, err := h.Renderer.RenderRealm(&content, gnourl, raw, RealmRenderContext{
		ChainId: h.Static.ChainId,
		Remote:  h.Static.RemoteHelp,
		Domain:  h.Static.Domain,
	})
	if err != nil {
		h.Logger.Error("unable to render realm", "error", err, "path", gnourl.EncodeURL())
		return GetClientErrorStatusPage(gnourl, err)
	}

	return http.StatusOK, components.RealmView(components.RealmData{
		TocItems: &components.RealmTOCData{
			Items: meta.Items,
		},
		// NOTE: `RenderRealm` should ensure that HTML content is
		// sanitized before rendering
		ComponentContent: components.NewReaderComponent(&content),
	})
}

// buildContributions returns the sorted list of contributions (packages and realms) for a user.
func (h *HTTPHandler) buildContributions(ctx context.Context, username string) ([]components.UserContribution, int, error) {
	prefix := "@" + username

	paths, err := h.Client.ListPaths(ctx, prefix, 10000)
	if err != nil {
		h.Logger.Error("unable to query contributions", "user", username, "error", err)
		return nil, 0, fmt.Errorf("unable to query contributions for user %q: %w", username, err)
	}

	contribs := make([]components.UserContribution, 0, len(paths))
	realmCount := 0
	for _, raw := range paths {
		trimmed := strings.TrimPrefix(raw, h.Static.Domain)
		u, err := weburl.Parse(trimmed)
		if err != nil {
			h.Logger.Warn("bad contribution URL", "path", raw, "error", err)
			continue
		}
		ctype := components.UserContributionTypePackage
		if u.IsRealm() {
			ctype = components.UserContributionTypeRealm
			realmCount++
		}
		contribs = append(contribs, components.UserContribution{
			Title: path.Base(raw),
			URL:   raw,
			Type:  components.UserContributionType(ctype),
			// TODO: size, description, date...
		})
	}

	sort.Slice(contribs, func(i, j int) bool {
		return contribs[i].Title < contribs[j].Title
	})
	return slices.Clip(contribs), realmCount, nil
}

// TODO: Check username from r/sys/users in addition to bech32 address test (username + gno address to be used)
// CreateUsernameFromBech32 creates a shortened version of the username if it's a valid bech32 address.
func CreateUsernameFromBech32(username string) string {
	_, _, err := bech32.Decode(username)
	if err == nil {
		// If it's a valid bech32 address, create a shortened version
		username = username[:4] + "..." + username[len(username)-4:]
	}

	return username
}

// GetUserView returns the user profile view for a given GnoURL.
func (h *HTTPHandler) GetUserView(ctx context.Context, gnourl *weburl.GnoURL) (int, *components.View) {
	username := strings.TrimPrefix(gnourl.Path, "/u/")

	var content bytes.Buffer

	// Render user profile realm
	raw, err := h.Client.Realm(ctx, "/r/"+username+"/home", "")
	if err == nil {
		_, err = h.Renderer.RenderRealm(&content, gnourl, raw, RealmRenderContext{
			ChainId: h.Static.ChainId,
			Remote:  h.Static.RemoteHelp,
			Domain:  h.Static.Domain,
		})
	}

	if content.Len() == 0 {
		h.Logger.Debug("unable to fetch user realm", "username", username, "error", err)
	}

	// Build contributions
	contribs, realmCount, err := h.buildContributions(ctx, username)
	if err != nil {
		h.Logger.Error("unable to build contributions", "error", err)
		return GetClientErrorStatusPage(gnourl, err)
	}

	// Compute package counts
	pkgCount := len(contribs)
	pureCount := pkgCount - realmCount

	// TODO: Check username from r/sys/users in addition to bech32 address test (username + gno address to be used)
	// Try to decode the bech32 address
	username = CreateUsernameFromBech32(username)

	// TODO: get from user r/profile and use placeholder if not set
	handlename := "Gnome " + username

	data := components.UserData{
		Username:      username,
		Handlename:    handlename,
		Contributions: contribs,
		PackageCount:  pkgCount,
		RealmCount:    realmCount,
		PureCount:     pureCount,
		Content:       components.NewReaderComponent(&content),
		// TODO: add bio, pic, links, teams, etc.
	}

	return http.StatusOK, components.UserView(data)
}

func (h *HTTPHandler) GetHelpView(ctx context.Context, gnourl *weburl.GnoURL) (int, *components.View) {
	jdoc, err := h.Client.Doc(ctx, gnourl.Path, 0)
	if err != nil {
		h.Logger.Error("unable to fetch qdoc", "error", err)
		return GetClientErrorStatusPage(gnourl, err)
	}

	// renderDoc renders a markdown documentation string to a Component.
	// Returns nil for empty input; renderer errors degrade to escaped text.
	renderDoc := func(src string) components.Component {
		if strings.TrimSpace(src) == "" {
			return nil
		}
		var buf bytes.Buffer
		if err := h.Renderer.RenderDocumentation(&buf, []byte(src)); err != nil {
			h.Logger.Warn("render doc failed — falling back to escaped plain text",
				"error", err)
			return components.NewReaderComponent(bytes.NewBufferString(template.HTMLEscapeString(src)))
		}
		return components.NewReaderComponent(&buf)
	}

	// Get public non-method funcs
	fsigs := []*doc.JSONFunc{}
	for _, fun := range jdoc.Funcs {
		if !(fun.Type == "" && token.IsExported(fun.Name)) {
			continue
		}

		if len(fun.Params) >= 1 && fun.Params[0].Type == "realm" {
			// Don't make an entry field for "cur realm". The signature will still show it.
			fun.Params = fun.Params[1:]
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

	// Wrap each function with its pre-rendered documentation Component.
	functions := make([]components.HelpFunction, 0, len(fsigs))
	for _, fn := range fsigs {
		functions = append(functions, components.HelpFunction{
			JSONFunc:     fn,
			DocComponent: renderDoc(fn.Doc),
		})
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
		Functions: functions,
		Doc:       renderDoc(jdoc.PackageDoc),
		Domain:    h.Static.Domain,
	})
}

// renderReadme renders the README.md file and returns the component and the raw content
func (h *HTTPHandler) renderReadme(ctx context.Context, gnourl *weburl.GnoURL, pkgPath string) (components.Component, []byte) {
	file, _, err := h.Client.File(ctx, pkgPath, ReadmeFileName, 0)
	if err != nil {
		h.Logger.Warn("fetch README.md", "path", pkgPath, "error", err)
		return nil, nil
	}

	var buf bytes.Buffer
	if _, err := h.Renderer.RenderRealm(&buf, gnourl, file, RealmRenderContext{
		ChainId: h.Static.ChainId,
		Remote:  h.Static.RemoteHelp,
		Domain:  h.Static.Domain,
	}); err != nil {
		h.Logger.Error("render README.md", "error", err)
		return nil, nil
	}
	return components.NewReaderComponent(&buf), file
}

func (h *HTTPHandler) GetSourceView(ctx context.Context, gnourl *weburl.GnoURL) (int, *components.View) {
	pkgPath := gnourl.Path

	files, err := h.Client.ListFiles(ctx, pkgPath)
	if err != nil {
		h.Logger.Warn("unable to list sources file", "path", gnourl.Path, "error", err)
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
	} else {
		// Prefer README.md, then .gno files, otherwise first file
		i := slices.IndexFunc(files, func(f string) bool {
			return f == "README.md" || strings.HasSuffix(f, ".gno")
		})

		if i >= 0 {
			fileName = files[i] // prefer .gno files and README.md
		} else {
			fileName = files[0] // fallback to first file - might be a .toml file
		}
	}

	// Standard file rendering
	var (
		fileSource components.Component
		fileLines  int
		sizeKB     float64
	)

	// Check whether the file is a markdown file
	switch fileName {
	case ReadmeFileName:
		// Try to render README.md with markdown processing
		readmeComp, raw := h.renderReadme(ctx, gnourl, pkgPath)
		if readmeComp != nil && raw != nil {
			fileSource = readmeComp
			fileLines = bytes.Count(raw, []byte("\n")) + 1
			sizeKB = float64(len(raw)) / 1024.0
			break
		}
		// Fall through to default case if markdown rendering fails
		fallthrough

	default:
		// Fetch raw source file
		file, meta, err := h.Client.File(ctx, pkgPath, fileName, 0)
		if err != nil {
			h.Logger.Warn("unable to get source file", "file", fileName, "error", err)
			return GetClientErrorStatusPage(gnourl, err)
		}

		var buff bytes.Buffer
		if err := h.Renderer.RenderSource(&buff, fileName, file); err != nil {
			h.Logger.Error("unable to render source file", "file", fileName, "error", err)
			return http.StatusInternalServerError, components.StatusErrorComponent("rendering error")
		}

		fileSource = components.NewReaderComponent(&buff)
		sizeKB = meta.SizeKB
		fileLines = meta.Lines
	}

	fileSizeStr := fmt.Sprintf("%.2f Kb", sizeKB)

	return http.StatusOK, components.SourceView(components.SourceData{
		PkgPath:      gnourl.Path,
		Files:        files,
		FileName:     fileName,
		FileCounter:  len(files),
		FileLines:    fileLines,
		FileSize:     fileSizeStr,
		FileDownload: gnourl.Path + "$download&file=" + fileName,
		FileSource:   fileSource,
	})
}

func (h *HTTPHandler) GetPathsListView(ctx context.Context, gnourl *weburl.GnoURL, indexData *components.IndexData) (int, *components.View) {
	const limit = 1_000 // XXX: implement pagination

	prefix := path.Join(h.Static.Domain, gnourl.Path) + "/"
	paths, qerr := h.Client.ListPaths(ctx, prefix, limit)
	if qerr != nil {
		h.Logger.Error("unable to query path", "error", qerr, "path", gnourl.EncodeURL())
	} else {
		h.Logger.Debug("query paths", "prefix", prefix, "paths", len(paths))
	}

	if len(paths) == 0 || paths[0] == "" {
		return GetClientErrorStatusPage(gnourl, ErrClientPackageNotFound)
	}

	// Always use explorer mode for paths list
	indexData.Mode = components.ViewModeExplorer

	// Update header mode
	indexData.HeaderData.Mode = indexData.Mode

	return http.StatusOK, components.DirectoryView(
		gnourl.Path,
		paths,
		len(paths),
		components.DirLinkTypeFile,
		indexData.Mode,
	)
}

// GetDirectoryView renders the directory view for a package, showing available files.
func (h *HTTPHandler) GetDirectoryView(ctx context.Context, gnourl *weburl.GnoURL, indexData *components.IndexData) (int, *components.View) {
	pkgPath := strings.TrimSuffix(gnourl.Path, "/")
	files, err := h.Client.ListFiles(ctx, pkgPath)
	if err != nil {
		if !errors.Is(err, ErrClientPackageNotFound) {
			h.Logger.Error("unable to list sources file", "path", pkgPath, "error", err)
			return GetClientErrorStatusPage(gnourl, err)
		}
		return h.GetPathsListView(ctx, gnourl, indexData)
	}

	if len(files) == 0 {
		h.Logger.Debug("no files available", "path", pkgPath)
		return http.StatusOK, components.StatusErrorComponent("no files available")
	}

	if gnourl.IsPure() {
		indexData.Mode = components.ViewModePackage
	}

	// Get README.md file if it exists
	readmeComp, _ := h.renderReadme(ctx, gnourl, pkgPath)
	return http.StatusOK, components.DirectoryView(
		pkgPath,
		files,
		len(files),
		components.DirLinkTypeSource,
		indexData.Mode,
		readmeComp,
	)
}

// GetStateView dispatches state-explorer URLs:
//   - /r/foo$state           → top-level package state (getStatePackageView)
//   - /r/foo$state&oid=ID    → a single stored object's contents (getStateObjectView)
//
// `viewMode` is the saved Pretty/Tree preference (from cookie) so the
// server stamps the right radio `checked` on first paint.
func (h *HTTPHandler) GetStateView(ctx context.Context, gnourl *weburl.GnoURL, viewMode string) (int, *components.View) {
	if oid := gnourl.WebQuery.Get("oid"); oid != "" {
		if len(oid) > maxStateIDLength {
			return http.StatusBadRequest, components.StatusErrorComponent("invalid object id")
		}
		return h.getStateObjectView(ctx, gnourl, oid, viewMode)
	}
	return h.getStatePackageView(ctx, gnourl, viewMode)
}

// ServeStateJSON exposes the chain's raw Amino JSON as a stable API surface
// for external tooling (block explorers, IDE plugins, JS SDKs) that decode
// state in the browser — the use case @gnojs/amino was carved out for.
// Triggered by `?state&json`, `?state&oid=…&json`, or `?state&tid=…&json`.
//
// Bytes flow through unmodified: no decoder, no walker, no fan-out, so none
// of the per-render bounds apply. The only validation is maxStateIDLength on
// attacker-controlled oid/tid to keep request→response amplification bounded.
func (h *HTTPHandler) ServeStateJSON(ctx context.Context, gnourl *weburl.GnoURL, w http.ResponseWriter) {
	height := gnourl.Height()

	var (
		raw []byte
		err error
	)
	switch {
	case gnourl.WebQuery.Has("oid"):
		oid := gnourl.WebQuery.Get("oid")
		if len(oid) > maxStateIDLength {
			writeStateJSONError(w, http.StatusBadRequest, "invalid object id")
			return
		}
		raw, err = h.Client.StateObject(ctx, oid, height)
	case gnourl.WebQuery.Has("tid"):
		tid := gnourl.WebQuery.Get("tid")
		if len(tid) > maxStateIDLength {
			writeStateJSONError(w, http.StatusBadRequest, "invalid type id")
			return
		}
		raw, err = h.Client.StateType(ctx, tid, height)
	default:
		raw, err = h.Client.StatePkg(ctx, gnourl.Path, height)
	}

	if err != nil {
		h.Logger.Error("unable to fetch state json", "error", err, "path", gnourl.EncodeURL(), "height", height)
		status, _ := stateErrorPage(gnourl, err, height)
		writeStateJSONError(w, status, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// Pinned `?height=N` is immutable once the block is finalized;
	// "latest" gets a 1s freshness window matching the ~3s block time.
	// Sets the terrain for the planned nginx/ETag layer.
	if height > 0 {
		w.Header().Set("Cache-Control", "public, max-age=86400, immutable")
	} else {
		w.Header().Set("Cache-Control", "public, max-age=1")
	}
	w.WriteHeader(http.StatusOK)
	w.Write(raw)
}

// writeStateJSONError writes a minimal `{"error":"…"}` envelope so consumers
// can reliably parse failures without sniffing for HTML.
func writeStateJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	body, _ := json.Marshal(map[string]string{"error": msg})
	w.Write(body)
}

// getStatePackageView renders the top-level state of a package or realm.
func (h *HTTPHandler) getStatePackageView(ctx context.Context, gnourl *weburl.GnoURL, viewMode string) (int, *components.View) {
	// Time-travel: `?height=N` pins the view to a historical block.
	// Invalid or missing → 0 = latest.
	height := gnourl.Height()

	// Fetch state JSON + doc index concurrently — they're both
	// per-package lookups and we want one RTT total. Doc errors are
	// non-fatal: the page renders without comments rather than aborting.
	var (
		raw     []byte
		stateErr error
		jdoc    *doc.JSONDocumentation
		wg      sync.WaitGroup
	)
	wg.Add(2)
	go func() {
		defer wg.Done()
		raw, stateErr = h.Client.StatePkg(ctx, gnourl.Path, height)
	}()
	go func() {
		defer wg.Done()
		d, derr := h.Client.Doc(ctx, gnourl.Path, height)
		if derr == nil {
			jdoc = d
		} else {
			h.Logger.Warn("unable to fetch package docs", "error", derr, "path", gnourl.EncodeURL())
		}
	}()
	wg.Wait()
	if stateErr != nil {
		h.Logger.Error("unable to fetch state", "error", stateErr, "path", gnourl.EncodeURL(), "height", height)
		return stateErrorPage(gnourl, stateErr, height)
	}

	nodes, err := components.DecodePkgJSON(raw)
	if err != nil {
		h.Logger.Error("unable to decode state JSON", "error", err, "path", gnourl.EncodeURL())
		return http.StatusInternalServerError, components.StatusErrorComponent("failed to decode state")
	}

	// Project the doc index onto the nodes by name so each card can
	// surface its declaration's source comment.
	if jdoc != nil {
		var vals, funs, typs []components.NamedDoc
		for _, vd := range jdoc.Values {
			for _, v := range vd.Values {
				// Per-entry doc wins over group doc when both are set —
				// `var (A, B int) // …` puts the comment on the group,
				// per-line comments stay on the entry.
				doc := v.Doc
				if doc == "" {
					doc = vd.Doc
				}
				vals = append(vals, components.NamedDoc{Name: v.Name, Doc: doc})
			}
		}
		for _, fn := range jdoc.Funcs {
			funs = append(funs, components.NamedDoc{Name: fn.Name, Doc: fn.Doc})
		}
		for _, t := range jdoc.Types {
			typs = append(typs, components.NamedDoc{Name: t.Name, Doc: t.Doc})
		}
		components.AttachDocs(nodes, vals, funs, typs)
	}

	// Title is just the realm/package name; the template appends a
	// `<span class="b-tag">state</span>` pill (mirrors the action
	// page's `<h1>name<span class="b-tag">package</span></h1>` shape).
	name := path.Base(gnourl.Path)
	if name == "/" || name == "." || name == "" {
		name = gnourl.Path
	}
	label := name

	return http.StatusOK, h.renderStateView(ctx, gnourl, renderStateInput{
		Nodes:    nodes,
		Label:    label,
		Sidebar:  components.BuildPackageSidebar(gnourl.Path, nodes),
		Height:   height,
		RawJSON:  raw,
		ViewMode: viewMode,
	})
}

// getStateObjectView renders the contents of a single stored object,
// reachable via /r/foo$state&oid=<ObjectID>[&tid=<TypeID>]. The page is
// bookmarkable and shareable — that's the whole point of moving away from
// in-place AJAX expansion.
//
// When `tid` is set, qtype_json is fetched in parallel with qobject_json so
// the decoder can label struct fields with their declared names instead of
// positional indices. Both calls run concurrently — the page render takes
// max(qobject, qtype) instead of their sum.
func (h *HTTPHandler) getStateObjectView(ctx context.Context, gnourl *weburl.GnoURL, oid, viewMode string) (int, *components.View) {
	tid := gnourl.WebQuery.Get("tid")
	if len(tid) > maxStateIDLength {
		return http.StatusBadRequest, components.StatusErrorComponent("invalid type id")
	}
	height := gnourl.Height()

	var (
		raw, typeRaw []byte
		objErr       error
		wg           sync.WaitGroup
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		raw, objErr = h.Client.StateObject(ctx, oid, height)
	}()
	if tid != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tr, err := h.Client.StateType(ctx, tid, height)
			if err == nil {
				typeRaw = tr
			} else {
				// Type fetch failure is recoverable — the page renders
				// with positional field indices instead of names.
				h.Logger.Warn("unable to fetch type for state object",
					"error", err, "path", gnourl.EncodeURL(), "tid", tid)
			}
		}()
	}
	wg.Wait()

	if objErr != nil {
		h.Logger.Error("unable to fetch state object", "error", objErr, "path", gnourl.EncodeURL(), "oid", oid, "height", height)
		return stateErrorPage(gnourl, objErr, height)
	}

	decoded, err := components.DecodeObjectFull(raw, typeRaw)
	if err != nil {
		h.Logger.Error("unable to decode state object JSON", "error", err, "path", gnourl.EncodeURL(), "oid", oid)
		return http.StatusInternalServerError, components.StatusErrorComponent("failed to decode state object")
	}

	// Single back-link — renders as a back-arrow icon inside the
	// h1 title (template). The realm context is still surfaced in
	// the sidebar's "Realm" row, so the label is just the realm
	// short name used for aria-label / hover title.
	realmName := path.Base(gnourl.Path)
	if realmName == "" || realmName == "." || realmName == "/" {
		realmName = gnourl.Path
	}
	crumbs := []components.StateCrumb{
		{Label: realmName, Href: components.RealmStateHref(gnourl.Path)},
	}
	// Page title — keep the OID readable on one line by truncating
	// the 40-char hashlet head…tail (the `:N` suffix stays). The full
	// OID remains visible & copy-able from the sidebar's "Object ID"
	// row. The field count moves to the sidebar's `Heading` so the
	// title stays short and wraps cleanly on narrow viewports.
	label := fmt.Sprintf("Object %s", components.TruncOID(oid, 8, 6))

	return http.StatusOK, h.renderStateView(ctx, gnourl, renderStateInput{
		Nodes:        decoded.Nodes,
		Label:        label,
		Crumbs:       crumbs,
		Sidebar:      components.BuildObjectSidebar(gnourl.Path, oid, tid, height, decoded.Info, decoded.Nodes),
		IsObjectPage: true,
		Height:       height,
		RawJSON:      raw,
		ViewMode:     viewMode,
	})
}

// renderStateInput collects everything renderStateView packages into a
// StateData. Carried as a struct because the call sites diverge only
// on a few fields (object vs package view).
type renderStateInput struct {
	Nodes        []components.StateNode
	Label        string
	Crumbs       []components.StateCrumb
	Sidebar      *components.StateSidebar
	IsObjectPage bool
	Height       int64
	RawJSON      []byte
	ViewMode     string
}

// renderStateView runs inline-preview + source-snippet enrichment, then
// packages the StateData. Order matters: preview adds nodes that Enrich
// then walks for hrefs/sources.
func (h *HTTPHandler) renderStateView(ctx context.Context, gnourl *weburl.GnoURL, in renderStateInput) *components.View {
	pkgPath := gnourl.Path
	adapter := &stateAdapter{ctx: ctx, client: h.Client, height: in.Height}
	components.EnrichInlinePreviews(in.Nodes, adapter, adapter)
	components.Enrich(in.Nodes, pkgPath, in.Height, adapter,
		&rendererSnippetHighlighter{renderer: h.Renderer})
	return components.StateView(components.StateData{
		PkgPath:      pkgPath,
		Nodes:        in.Nodes,
		CountLabel:   in.Label,
		Crumbs:       in.Crumbs,
		Sidebar:      in.Sidebar,
		IsObjectPage: in.IsObjectPage,
		Height:       in.Height,
		LatestHref:   template.URL(gnourl.WithoutHeight().EncodeWebURL()),
		ViewMode:     in.ViewMode,
		RawJSON:      string(in.RawJSON),
		KindCounts:   components.ComputeKindCounts(in.Nodes),
	})
}

// stateAdapter satisfies the three fetcher interfaces consumed by the
// orchestrator (FileFetcher, StateObjectFetcher, StateTypeFetcher) with
// one shared (ctx, client, height) carrier.
type stateAdapter struct {
	ctx    context.Context
	client ClientAdapter
	height int64
}

func (a *stateAdapter) Fetch(pkgPath, fileName string) ([]byte, error) {
	src, _, err := a.client.File(a.ctx, pkgPath, fileName, a.height)
	return src, err
}

func (a *stateAdapter) FetchObject(oid string) ([]byte, error) {
	return a.client.StateObject(a.ctx, oid, a.height)
}

func (a *stateAdapter) FetchType(tid string) ([]byte, error) {
	return a.client.StateType(a.ctx, tid, a.height)
}

// rendererSnippetHighlighter adapts HTMLRenderer to SnippetHighlighter —
// wraps chroma-backed RenderSource into a template.HTML producer.
type rendererSnippetHighlighter struct {
	renderer Renderer
}

func (h *rendererSnippetHighlighter) Render(fileName string, source []byte) (template.HTML, error) {
	var buf bytes.Buffer
	if err := h.renderer.RenderSource(&buf, fileName, source); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}



// ServeSourceDownload handles downloading a source file as plain text.
func (h *HTTPHandler) ServeSourceDownload(ctx context.Context, gnourl *weburl.GnoURL, w http.ResponseWriter, r *http.Request) {
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
	source, _, err := h.Client.File(ctx, pkgPath, fileName, 0)
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
	w.Write(source) // write raw file
}

// stateViewModeCookie holds the saved Pretty/Tree preference the JS
// controller writes on toggle; read server-side so the right radio is
// `checked` on first paint (no flicker).
const stateViewModeCookie = "state_view_mode"

// readWhitelistedCookie returns the cookie's value when it matches one
// of `allowed`; otherwise the empty string. Defends downstream code
// against arbitrary cookie payloads.
func readWhitelistedCookie(r *http.Request, name string, allowed ...string) string {
	c, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	for _, a := range allowed {
		if c.Value == a {
			return c.Value
		}
	}
	return ""
}

// stateErrorPage maps a state-query failure to a friendly status page,
// preferring 400 + "block height N is not available" when the failing
// query specified a non-zero height. The chain rejects out-of-range
// heights with a generic RPC error that GetClientErrorStatusPage would
// surface as a confusing 500 — this telegraphs to the user that the
// height (which they control via the URL) is the cause. PackageNotFound
// stays a 404 even at height>0 (path is wrong regardless of height).
func stateErrorPage(gnourl *weburl.GnoURL, err error, height int64) (int, *components.View) {
	// Not-found verdicts are authoritative regardless of height — a wrong
	// path/OID stays wrong whether we pin the height or not.
	if errors.Is(err, ErrClientPackageNotFound) || errors.Is(err, ErrClientObjectNotFound) {
		return http.StatusNotFound, components.StatusErrorComponent(err.Error())
	}
	if height > 0 {
		return http.StatusBadRequest, components.StatusErrorComponent(
			fmt.Sprintf("block height %d is not available", height))
	}
	return GetClientErrorStatusPage(gnourl, err)
}

func GetClientErrorStatusPage(_ *weburl.GnoURL, err error) (int, *components.View) {
	if err == nil {
		return http.StatusOK, nil
	}

	switch {
	case errors.Is(err, ErrClientTimeout):
		return http.StatusRequestTimeout, components.StatusErrorComponent(err.Error())
	case errors.Is(err, ErrClientPackageNotFound),
		errors.Is(err, ErrClientObjectNotFound):
		return http.StatusNotFound, components.StatusErrorComponent(err.Error())
	case errors.Is(err, ErrClientBadRequest):
		return http.StatusBadRequest, components.StatusErrorComponent("bad request")
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
