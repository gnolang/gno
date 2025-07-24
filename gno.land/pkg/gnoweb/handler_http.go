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

// HTTPHandlerConfig configures an HTTPHandler.
type HTTPHandlerConfig struct {
	Meta          StaticMetadata
	ClientAdapter ClientAdapter
	Renderer      Renderer
	Aliases       map[string]AliasTarget
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

	// Use form data as query
	gnourl.Query = r.PostForm

	// Redirect to the new URL
	http.Redirect(w, r, gnourl.EncodeWebURL(), http.StatusSeeOther)
}

// prepareIndexBodyView prepares the data and main view for the index page.
func (h *HTTPHandler) prepareIndexBodyView(r *http.Request, indexData *components.IndexData) (int, *components.View) {
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

// GetMarkdownView handles rendering of markdown files.
func (h *HTTPHandler) GetMarkdownView(gnourl *weburl.GnoURL, mdContent string) (int, *components.View) {
	var content bytes.Buffer

	// Use Goldmark for Markdown parsing
	toc, err := h.Renderer.RenderRealm(&content, gnourl, []byte(mdContent))
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
func (h *HTTPHandler) GetPackageView(gnourl *weburl.GnoURL, indexData *components.IndexData) (int, *components.View) {
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

// GetRealmView renders a realm page or returns an error/status if not available.
func (h *HTTPHandler) GetRealmView(gnourl *weburl.GnoURL, indexData *components.IndexData) (int, *components.View) {
	// First fecth the realm
	raw, err := h.Client.Realm(gnourl.Path, gnourl.EncodeArgs())
	switch {
	case err == nil: // ok
	case errors.Is(err, ErrClientRenderNotDeclared):
		// No Render() declared: fall back to directory view (which will show README.md if present)
		return h.GetDirectoryView(gnourl, indexData)
	case errors.Is(err, ErrClientPackageNotFound):
		// No realm exists here, try to display underlying paths
		return h.GetPathsListView(gnourl, indexData)
	default:
		h.Logger.Error("unable to fetch realm", "error", err, "path", gnourl.EncodeURL())
		return GetClientErrorStatusPage(gnourl, err)
	}

	var content bytes.Buffer
	meta, err := h.Renderer.RenderRealm(&content, gnourl, raw)
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
func (h *HTTPHandler) buildContributions(username string) ([]components.UserContribution, int, error) {
	prefix := "@" + username

	paths, err := h.Client.ListPaths(prefix, 10000)
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
func (h *HTTPHandler) GetUserView(gnourl *weburl.GnoURL) (int, *components.View) {
	username := strings.TrimPrefix(gnourl.Path, "/u/")

	var content bytes.Buffer

	// Render user profile realm
	raw, err := h.Client.Realm("/r/"+username+"/home", "")
	if err == nil {
		_, err = h.Renderer.RenderRealm(&content, gnourl, raw)
	}

	if content.Len() == 0 {
		h.Logger.Debug("unable to fetch user realm", "username", username, "error", err)
	}

	// Build contributions
	contribs, realmCount, err := h.buildContributions(username)
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

func (h *HTTPHandler) GetHelpView(gnourl *weburl.GnoURL) (int, *components.View) {
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
		Domain:    h.Static.Domain,
	})
}

// renderReadme renders the README.md file and returns the component and the raw content
func (h *HTTPHandler) renderReadme(gnourl *weburl.GnoURL, pkgPath string) (components.Component, []byte) {
	file, _, err := h.Client.File(pkgPath, ReadmeFileName)
	if err != nil {
		h.Logger.Warn("fetch README.md", "path", pkgPath, "error", err)
		return nil, nil
	}

	var buf bytes.Buffer
	if _, err := h.Renderer.RenderRealm(&buf, gnourl, file); err != nil {
		h.Logger.Error("render README.md", "error", err)
		return nil, nil
	}
	return components.NewReaderComponent(&buf), file
}

func (h *HTTPHandler) GetSourceView(gnourl *weburl.GnoURL) (int, *components.View) {
	pkgPath := gnourl.Path
	files, err := h.Client.ListFiles(pkgPath)
	if err != nil {
		h.Logger.Warn("unable to list sources file", "path", gnourl.Path, "error", err)
		return GetClientErrorStatusPage(gnourl, err)
	}
	if len(files) == 0 {
		h.Logger.Debug("no files available", "path", gnourl.Path)
		return http.StatusOK, components.StatusErrorComponent("no files available")
	}

	fileName := ""
	if gnourl.IsFile() {
		fileName = gnourl.File
	} else if file := gnourl.WebQuery.Get("file"); file != "" {
		fileName = file
	}

	if fileName == "" {
		return h.getSourceOverviewView(pkgPath, gnourl, files)
	}
	return h.getSourceFileView(pkgPath, gnourl, files, fileName)
}

// getSourceOverview construit la vue overview (README, fonctions, fichiers)
func (h *HTTPHandler) getSourceOverviewView(pkgPath string, gnourl *weburl.GnoURL, files []string) (int, *components.View) {
	readmeComp, _ := h.renderReadme(gnourl, pkgPath)
	jdoc, _ := h.Client.Doc(pkgPath)
	var functions []*doc.JSONFunc
	var docString string
	var consts []*doc.JSONValueDecl
	var vars []*doc.JSONValueDecl
	var types []*doc.JSONType
	var dirs []string
	if jdoc != nil {
		functions = jdoc.Funcs
		docString = jdoc.PackageDoc
		types = jdoc.Types
		for _, v := range jdoc.Values {
			if v.Const {
				consts = append(consts, v)
			} else {
				vars = append(vars, v)
			}
		}
	}
	// List dirs
	prefix := h.Static.Domain + pkgPath + "/"
	limit := 1_000
	paths, err := h.Client.ListPaths(prefix, limit)
	if err != nil {
		h.Logger.Warn("ListPaths failed", "pkgPath", pkgPath, "error", err)
		paths = nil
	}

	for _, p := range paths {
		dir := strings.TrimPrefix(p, pkgPath+"/")
		if dir != "" && !strings.Contains(dir, "/") && dir != "." {
			dirs = append(dirs, dir)
		}
	}
	h.Logger.Debug("Dirs for overview", "pkgPath", pkgPath, "dirs", dirs)
	return http.StatusOK, components.OverviewView(components.OverviewData{
		PkgPath:     pkgPath,
		Readme:      readmeComp,
		Functions:   functions,
		Doc:         docString,
		Files:       files,
		FileCounter: len(files),
		Consts:      consts,
		Vars:        vars,
		Types:       types,
		Dirs:        dirs,
	})
}

// getSourceFile construit la vue source pour un fichier donné
func (h *HTTPHandler) getSourceFileView(pkgPath string, gnourl *weburl.GnoURL, files []string, fileName string) (int, *components.View) {
	var (
		fileSource components.Component
		fileLines  int
		sizeKB     float64
	)

	switch fileName {
	case ReadmeFileName:
		readmeComp, raw := h.renderReadme(gnourl, pkgPath)
		if readmeComp != nil && raw != nil {
			fileSource = readmeComp
			fileLines = bytes.Count(raw, []byte("\n")) + 1
			sizeKB = float64(len(raw)) / 1024.0
			break
		}
		fallthrough
	default:
		file, meta, err := h.Client.File(pkgPath, fileName)
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

func (h *HTTPHandler) GetPathsListView(gnourl *weburl.GnoURL, indexData *components.IndexData) (int, *components.View) {
	const limit = 1_000 // XXX: implement pagination

	prefix := path.Join(h.Static.Domain, gnourl.Path) + "/"

	paths, qerr := h.Client.ListPaths(prefix, limit)

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
func (h *HTTPHandler) GetDirectoryView(gnourl *weburl.GnoURL, indexData *components.IndexData) (int, *components.View) {
	pkgPath := strings.TrimSuffix(gnourl.Path, "/")
	files, err := h.Client.ListFiles(pkgPath)
	if err != nil {
		if !errors.Is(err, ErrClientPackageNotFound) {
			h.Logger.Error("unable to list sources file", "path", pkgPath, "error", err)
			return GetClientErrorStatusPage(gnourl, err)
		}
		return h.GetPathsListView(gnourl, indexData)
	}

	if len(files) == 0 {
		h.Logger.Debug("no files available", "path", pkgPath)
		return http.StatusOK, components.StatusErrorComponent("no files available")
	}

	if gnourl.IsPure() {
		indexData.Mode = components.ViewModePackage
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

// GetSourceDownload handles downloading a source file as plain text.
func (h *HTTPHandler) GetSourceDownload(gnourl *weburl.GnoURL, w http.ResponseWriter, r *http.Request) {
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
	source, _, err := h.Client.File(pkgPath, fileName)
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

func GetClientErrorStatusPage(_ *weburl.GnoURL, err error) (int, *components.View) {
	if err == nil {
		return http.StatusOK, nil
	}

	switch {
	case errors.Is(err, ErrClientPackageNotFound):
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
