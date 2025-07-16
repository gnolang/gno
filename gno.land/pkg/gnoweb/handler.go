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
	"sort"
	"strings"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/gnolang/gno/gnovm/pkg/doc"
	"github.com/gnolang/gno/tm2/pkg/bech32"
)

const ReadmeFileName = "README.md"

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

	// Set the header mode based on the URL type and context
	switch {
	case IsHomePath(r.RequestURI):
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
func (h *WebHandler) Post(w http.ResponseWriter, r *http.Request) {
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

	// Use form data as query
	gnourl.Query = r.PostForm

	// Redirect to the new URL
	http.Redirect(w, r, gnourl.EncodeWebURL(), http.StatusSeeOther)
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
		Mode:       indexData.Mode,
	}

	switch {
	case aliasExists && aliasTarget.Kind == StaticMarkdown:
		return h.GetMarkdownView(gnourl, aliasTarget.Value)
	case gnourl.IsRealm(), gnourl.IsPure(), gnourl.IsUser():
		return h.GetPackageView(gnourl, indexData)
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
func (h *WebHandler) GetPackageView(gnourl *weburl.GnoURL, indexData *components.IndexData) (int, *components.View) {
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
		return h.GetDirectoryView(gnourl, indexData)
	}

	// Handle User page
	if gnourl.IsUser() {
		return h.GetUserView(gnourl)
	}

	// Ultimately get realm view
	return h.GetRealmView(gnourl, indexData)
}

func (h *WebHandler) GetRealmView(gnourl *weburl.GnoURL, indexData *components.IndexData) (int, *components.View) {
	var content bytes.Buffer

	meta, err := h.Client.RenderRealm(&content, gnourl, h.MarkdownRenderer)
	switch {
	case err == nil: // ok
	case errors.Is(err, ErrRenderNotDeclared):
		// No Render() declared: fall back to directory view (which will show README.md if present)
		return h.GetDirectoryView(gnourl, indexData)
	case errors.Is(err, ErrClientPathNotFound):
		// No realm exists here, try to display underlying paths
		return h.GetPathsListView(gnourl, indexData)
	default:
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

// buildContributions returns the sorted list of contributions (packages and realms) for a user.
func (h *WebHandler) buildContributions(username string) ([]components.UserContribution, int, error) {
	prefix := "@" + username
	paths, err := h.Client.QueryPaths(prefix, 10000)
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
// createUsernameFromBech32 creates a shortened version of the username if it's a valid bech32 address
func CreateUsernameFromBech32(username string) string {
	_, _, err := bech32.Decode(username)
	if err == nil {
		// If it's a valid bech32 address, create a shortened version
		username = username[:4] + "..." + username[len(username)-4:]
	}

	return username
}

// GetUserView returns the user profile view for a given GnoURL.
func (h *WebHandler) GetUserView(gnourl *weburl.GnoURL) (int, *components.View) {
	username := strings.TrimPrefix(gnourl.Path, "/u/")
	var content bytes.Buffer

	// Render user profile realm
	if _, err := h.Client.RenderRealm(&content, &weburl.GnoURL{Path: "/r/" + username + "/home"}, h.MarkdownRenderer); err != nil {
		h.Logger.Debug("unable to render user realm", "error", err)
	}

	// Build contributions
	contribs, realmCount, err := h.buildContributions(username)
	if err != nil {
		h.Logger.Error("unable to build contributions", "error", err)
		return http.StatusInternalServerError, components.StatusErrorComponent(err.Error())
	}

	// Compute package counts
	pkgCount := len(contribs)
	pureCount := pkgCount - realmCount

	// TODO: Check username from r/sys/users in addition to bech32 address test (username + gno address to be used)
	// Try to decode the bech32 address
	username = CreateUsernameFromBech32(username)

	//TODO: get from user r/profile and use placeholder if not set
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

// renderReadme renders the README.md file and returns the component and the raw content
func (h *WebHandler) renderReadme(gnourl *weburl.GnoURL, pkgPath string) (components.Component, []byte) {
	if !h.Client.HasFile(pkgPath, ReadmeFileName) {
		return nil, nil
	}

	var rawBuffer bytes.Buffer
	_, err := h.Client.SourceFile(&rawBuffer, pkgPath, ReadmeFileName, true)
	if err != nil {
		h.Logger.Error("fetch README.md", "path", pkgPath, "error", err)
		return nil, nil
	}

	raw := rawBuffer.Bytes()
	var buf bytes.Buffer
	if _, err := h.MarkdownRenderer.Render(&buf, gnourl, raw); err != nil {
		h.Logger.Error("render README.md", "error", err)
		return nil, nil
	}
	return components.NewReaderComponent(&buf), raw
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
		// Prefer README.md, then .gno files, otherwise first file
		if i := slices.IndexFunc(files, func(f string) bool {
			return f == "README.md" || strings.HasSuffix(f, ".gno")
		}); i >= 0 {
			fileName = files[i] // prefer .gno files and README.md
		} else {
			fileName = files[0] // fallback to first file - might be a .toml file
		}
	}

	// Standard file rendering
	var (
		source     bytes.Buffer
		fileSource components.Component
		fileLines  int
		sizeKb     float64
	)

	// Check whether the file is a markdown file
	switch fileName {
	case ReadmeFileName:
		// Try to render README.md with markdown processing
		readmeComp, raw := h.renderReadme(gnourl, pkgPath)
		if readmeComp != nil && raw != nil {
			fileSource = readmeComp
			fileLines = bytes.Count(raw, []byte("\n")) + 1
			sizeKb = float64(len(raw)) / 1024.0
			break
		}
		// Fall through to default case if markdown rendering fails
		fallthrough

	default:
		// Fetch raw source file
		meta, err := h.Client.SourceFile(&source, pkgPath, fileName, false)
		if err != nil {
			h.Logger.Error("unable to get source file", "file", fileName, "error", err)
			return GetClientErrorStatusPage(gnourl, err)
		}

		fileSource = components.NewReaderComponent(&source)
		sizeKb = meta.SizeKb
		fileLines = meta.Lines
	}

	fileSizeStr := fmt.Sprintf("%.2f Kb", sizeKb)

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

func (h *WebHandler) GetPathsListView(gnourl *weburl.GnoURL, indexData *components.IndexData) (int, *components.View) {
	const limit = 1_000 // XXX: implement pagination

	prefix := path.Join(h.Static.Domain, gnourl.Path) + "/"
	paths, qerr := h.Client.QueryPaths(prefix, limit)
	if qerr != nil {
		h.Logger.Error("unable to query path", "error", qerr, "path", gnourl.EncodeURL())
	} else {
		h.Logger.Debug("query paths", "prefix", prefix, "paths", len(paths))
	}

	if len(paths) == 0 || paths[0] == "" {
		return GetClientErrorStatusPage(gnourl, ErrClientPathNotFound)
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

func (h *WebHandler) GetDirectoryView(gnourl *weburl.GnoURL, indexData *components.IndexData) (int, *components.View) {
	pkgPath := strings.TrimSuffix(gnourl.Path, "/")
	files, err := h.Client.Sources(pkgPath)
	if err != nil {
		if !errors.Is(err, ErrClientPathNotFound) {
			h.Logger.Error("unable to list sources file", "path", pkgPath, "error", err)
			return GetClientErrorStatusPage(gnourl, err)
		}
		return h.GetPathsListView(gnourl, indexData)
	}
	if len(files) == 0 {
		h.Logger.Debug("no files available", "path", pkgPath)
		return http.StatusOK, components.StatusErrorComponent("no files available")
	}
	// get the README.md file if it exists
	readmeComp, _ := h.renderReadme(gnourl, pkgPath)
	return http.StatusOK, components.DirectoryView(
		pkgPath,
		files,
		len(files),
		components.DirLinkTypeSource,
		indexData.Mode,
		readmeComp,
	)
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
