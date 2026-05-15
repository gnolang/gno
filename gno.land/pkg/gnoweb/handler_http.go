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
	"net/url"
	"path"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/feature/state"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/gnolang/gno/gnovm/pkg/doc"
	"github.com/gnolang/gno/tm2/pkg/bech32"
)

const ReadmeFileName = "README.md"

// defaultRequestTimeout bounds every GET when no explicit Timeout is
// configured, so r.Context() always carries a deadline (the page path
// can fan out to many RPC calls — an unbounded request is a DoS vector).
const defaultRequestTimeout = 30 * time.Second

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
	// StateRateLimitPerMinute caps per-IP requests against ?state* URLs.
	// 0 ⇒ defaultStateRateLimitPerMinute. Also used as the token-bucket
	// burst. ADR-004 §Rate limiting.
	StateRateLimitPerMinute int
	// StateRateLimitTrustedProxies — see AppConfig field of the same name.
	StateRateLimitTrustedProxies []string
}

// defaultStateRateLimitPerMinute is the safe-by-default cap applied when
// no explicit value is configured. Matches the ADR-004 reference value.
const defaultStateRateLimitPerMinute = 100

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
	// State is the feature/state handler that owns every ?state* URL.
	// Built in NewHTTPHandler so the wire-in dispatch hook is a single
	// method call (ADR-004 §Decision §1 wire-in).
	State *state.Handler
}

// NewHTTPHandler creates a new HTTPHandler.
func NewHTTPHandler(logger *slog.Logger, cfg *HTTPHandlerConfig) (*HTTPHandler, error) {
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validate error: %w", err)
	}

	h := &HTTPHandler{
		Client:   cfg.ClientAdapter,
		Static:   cfg.Meta,
		Renderer: cfg.Renderer,
		Aliases:  cfg.Aliases,
		Timeout:  cfg.Timeout,
		Logger:   logger,
	}
	rate := cfg.StateRateLimitPerMinute
	if rate <= 0 {
		rate = defaultStateRateLimitPerMinute
	}
	h.State = state.New(state.Deps{
		Client:      cfg.ClientAdapter,
		Highlighter: &rendererSnippetHighlighter{renderer: cfg.Renderer},
		FileFetcher: &clientFileFetcher{client: cfg.ClientAdapter},
		Logger:      logger,
		RateLimit: state.RateLimitConfig{
			PerMinute:      rate,
			Burst:          rate,
			MaxIPs:         10_000,
			TrustedProxies: state.ParseTrustedProxies(cfg.StateRateLimitTrustedProxies),
		},
	})
	return h, nil
}

// clientFileFetcher adapts ClientAdapter to components.FileFetcher (the
// shape state.Deps consumes for frag=source). The state package cannot
// import gnoweb directly (cycle), so the adapter lives here.
type clientFileFetcher struct {
	client ClientAdapter
}

func (f *clientFileFetcher) Fetch(ctx context.Context, pkgPath, fileName string, height int64) ([]byte, error) {
	src, _, err := f.client.File(ctx, pkgPath, fileName, height)
	return src, err
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

	timeout := h.Timeout
	if timeout <= 0 {
		timeout = defaultRequestTimeout
	}
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()
	r = r.WithContext(ctx)

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

	// Apply GnowebPath alias rewrite BEFORE parsing — every downstream
	// dispatch (state, source, package view) needs to see the resolved
	// path. Legacy did this inside prepareIndexBodyView, which the state
	// branch below short-circuits, so an alias-mapped state URL would
	// previously route to the unmapped path and 404.
	if alias, ok := h.Aliases[r.URL.Path]; ok && alias.Kind == GnowebPath {
		r.URL.Path = alias.Value
	}

	// Parse the URL
	gnourl, err := weburl.ParseFromURL(r.URL)
	if err != nil {
		h.Logger.Warn("unable to parse url path", "path", r.URL.Path, "error", err)

		// A `$state&json` request must get a JSON envelope even when the
		// URL fails to parse — honor the JSON-in/JSON-out contract instead
		// of returning an HTML body the client can't decode.
		if isStateJSONRequest(r.URL) {
			writeJSONErrorResponse(w, http.StatusNotFound, "invalid path")
			return
		}

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

	// State explorer (all ?state* URLs). The feature/state.Handler.Handle
	// internally dispatches:
	//   - ?state&json[&oid|&tid]  → JSON API path, writes body directly to w
	//   - ?state&frag=*           → htmx HTML fragment, writes body directly to w
	//   - ?state[&oid=X[&tid=Y]]  → HTML page path, returns *components.View
	//                                so IndexLayout wraps it in gnoweb chrome
	// ADR-004 §Decision §1 wire-in. Body-already-written paths return nil
	// View; page path returns a non-nil View for chrome composition.
	if gnourl.WebQuery.Has("state") {
		status, view := h.State.Handle(r.Context(), w, r, gnourl)
		if view == nil {
			// Direct-write path (json or fragment): body and headers
			// already on w; nothing else to render.
			return
		}
		// Page path: wrap the state body in IndexLayout chrome. Set
		// HeaderData/Title here (mirroring prepareIndexBodyView) so the
		// global header — breadcrumb + Content/State/Source/Actions
		// tabs — renders against this realm instead of inheriting zero
		// values and pointing the tabs at empty URLs.
		indexData.Mode = components.ViewModeRealm
		h.setHeaderForRealm(&indexData, gnourl)
		indexData.BodyView = view
		w.WriteHeader(status)
		if err := components.IndexLayout(indexData).Render(w); err != nil {
			h.Logger.Error("failed to render state page", "error", err)
		}
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

// maxPostFormBytes caps r.Body for the redirect form. The form carries
// short fields (path, height, file); 64 KiB leaves plenty of headroom
// while preventing a 32 MiB Go default from being weaponised.
const maxPostFormBytes = 64 * 1024

// Post processes a POST HTTP request.
func (h *HTTPHandler) Post(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		h.Logger.Debug("request completed",
			"url", r.URL.String(),
			"elapsed", time.Since(start).String())
	}()

	r.Body = http.MaxBytesReader(w, r.Body, maxPostFormBytes)
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

	h.setHeaderForRealm(indexData, gnourl)

	switch {
	case aliasExists && aliasTarget.Kind == StaticMarkdown:
		indexData.HeaderData.Static = true
		return h.GetMarkdownView(gnourl, aliasTarget.Value)
	case gnourl.IsRealm(), gnourl.IsPure(), gnourl.IsUser():
		return h.GetPackageView(ctx, gnourl, indexData)
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
		return GetClientErrorStatusPage(gnourl, err, 0)
	}

	return http.StatusOK, components.RealmView(components.RealmData{
		TocItems:         &components.RealmTOCData{Items: toc.Items},
		ComponentContent: components.NewReaderComponent(&content),
	})
}

// GetPackageView handles package pages, including help, source, directory, and user views.
func (h *HTTPHandler) GetPackageView(ctx context.Context, gnourl *weburl.GnoURL, indexData *components.IndexData) (int, *components.View) {
	// Handle Help page
	if gnourl.WebQuery.Has("help") {
		return h.GetHelpView(ctx, gnourl)
	}

	// State explorer (?state, ?state&oid=X) — page path. JSON and
	// fragment variants were intercepted in Get() before this point.
	// They never reach prepareIndexBodyView. This branch should be
	// unreachable now; left as a defensive fallback in case dispatch
	// order changes.
	if gnourl.WebQuery.Has("state") {
		return http.StatusInternalServerError, components.StatusErrorComponent("state dispatch reached prepareIndexBodyView")
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
		return GetClientErrorStatusPage(gnourl, err, 0)
	}

	var content bytes.Buffer
	meta, err := h.Renderer.RenderRealm(&content, gnourl, raw, RealmRenderContext{
		ChainId: h.Static.ChainId,
		Remote:  h.Static.RemoteHelp,
		Domain:  h.Static.Domain,
	})
	if err != nil {
		h.Logger.Error("unable to render realm", "error", err, "path", gnourl.EncodeURL())
		return GetClientErrorStatusPage(gnourl, err, 0)
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

// MaxUserContributions caps how many contributions /u/<user> renders.
// Each entry costs a bech32 decode, a weburl parse, and a sort comparison;
// an unbounded cap turns a single GET into a 10k-iteration amplifier.
// Exported so external tests assert against the documented cap.
// TODO: paginate via ?page= when a contributor exceeds this cap.
const MaxUserContributions = 200

// buildContributions returns the sorted list of contributions (packages and realms) for a user.
func (h *HTTPHandler) buildContributions(ctx context.Context, username string) ([]components.UserContribution, int, error) {
	prefix := "@" + username

	paths, err := h.Client.ListPaths(ctx, prefix, MaxUserContributions)
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
		return GetClientErrorStatusPage(gnourl, err, 0)
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
		return GetClientErrorStatusPage(gnourl, err, 0)
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
	height := gnourl.Height()

	files, err := h.Client.ListFiles(ctx, pkgPath, height)
	if err != nil {
		h.Logger.Warn("unable to list sources file", "path", gnourl.Path, "error", err)
		return GetClientErrorStatusPage(gnourl, err, height)
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
			return GetClientErrorStatusPage(gnourl, err, 0)
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
		return GetClientErrorStatusPage(gnourl, ErrClientPackageNotFound, 0)
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
	height := gnourl.Height()
	files, err := h.Client.ListFiles(ctx, pkgPath, height)
	if err != nil {
		if !errors.Is(err, ErrClientPackageNotFound) {
			h.Logger.Error("unable to list sources file", "path", pkgPath, "error", err)
			return GetClientErrorStatusPage(gnourl, err, height)
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
	return template.HTML(buf.String()), nil //nolint:gosec
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
		status, _ := GetClientErrorStatusPage(gnourl, err, 0)
		http.Error(w, "not found", status)
		return
	}

	// Send raw file as attachment for download (without HTML formating)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", fileName))
	w.WriteHeader(http.StatusOK)
	w.Write(source) // write raw file
}

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

// isStateJSONRequest reports whether u is a `$state&json` request, parsed
// straight from the raw URL so it works even when weburl.ParseFromURL fails.
// The webargs segment lives after `$` in the path; gnoweb's JSON state API
// is keyed on the `state` + `json` web flags being present there.
func isStateJSONRequest(u *url.URL) bool {
	_, webargs, found := strings.Cut(u.EscapedPath(), "$")
	if !found {
		return false
	}
	q, err := url.ParseQuery(webargs)
	if err != nil {
		return false
	}
	return q.Has("state") && q.Has("json")
}

// writeJSONErrorResponse emits the `{"error":"…"}` envelope used by the
// state JSON API, mirroring feature/state.writeJSONError so a JSON client
// always gets a JSON body even on the gnoweb-side parse-failure path.
func writeJSONErrorResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	body, _ := json.Marshal(map[string]string{"error": message})
	_, _ = w.Write(body)
}

// clientErrorMessage classifies a client error into (status, friendly msg).
// height > 0 short-circuits non-NotFound errors to "block height N is not
// available" — the chain rejects out-of-range heights with a generic RPC
// error that would otherwise surface as a confusing 500, while the height
// is the actual cause and the user controls it via the URL. NotFound wins
// regardless of height (a wrong path is wrong at any block).
func clientErrorMessage(err error, height int64) (int, string) {
	if err == nil {
		return http.StatusOK, ""
	}
	if errors.Is(err, ErrClientPackageNotFound) || errors.Is(err, ErrClientObjectNotFound) {
		return http.StatusNotFound, err.Error()
	}
	if height > 0 {
		return http.StatusBadRequest, fmt.Sprintf("block height %d is not available", height)
	}
	switch {
	case errors.Is(err, ErrClientTimeout):
		return http.StatusRequestTimeout, err.Error()
	case errors.Is(err, ErrClientBadRequest):
		return http.StatusBadRequest, "bad request"
	default:
		// ErrClientResponse + unknown errors. Hide internals.
		return http.StatusInternalServerError, "internal error"
	}
}

// GetClientErrorStatusPage wraps clientErrorMessage into a renderable View.
// `height` is the optional ?height=N pin from the URL — pass 0 when the
// caller does not propagate it to the chain query.
func GetClientErrorStatusPage(_ *weburl.GnoURL, err error, height int64) (int, *components.View) {
	status, msg := clientErrorMessage(err, height)
	if msg == "" {
		return status, nil
	}
	return status, components.StatusErrorComponent(msg)
}

// setHeaderForRealm seeds IndexData.HeadData.Title + IndexData.HeaderData
// from the parsed realm URL. Shared by the state-page wire-in and the
// generic prepareIndexBodyView path so the global header (breadcrumb +
// Content/State/Source/Actions tabs) always renders against the same
// realm. Mode must be set on indexData before calling.
func (h *HTTPHandler) setHeaderForRealm(indexData *components.IndexData, gnourl *weburl.GnoURL) {
	indexData.HeadData.Title = h.Static.Domain + " - " + gnourl.Path
	indexData.HeaderData = components.HeaderData{
		Breadcrumb: generateBreadcrumbPaths(gnourl),
		RealmURL:   *gnourl,
		ChainId:    h.Static.ChainId,
		Remote:     h.Static.RemoteHelp,
		Mode:       indexData.Mode,
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
