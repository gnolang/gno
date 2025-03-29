package gnoweb

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
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
	gnourl, err := weburl.ParseGnoURL(r.URL)
	if err != nil {
		h.Logger.Warn("unable to parse url path", "path", r.URL.Path, "error", err)
		return http.StatusNotFound, components.StatusErrorComponent("invalid path")
	}

	breadcrumb := generateBreadcrumbPaths(gnourl)
	indexData.HeadData.Title = h.Static.Domain + " - " + gnourl.Path
	indexData.HeaderData = components.HeaderData{
		Breadcrumb: breadcrumb,
		RealmURL:   *gnourl,
		ChainId:    h.Static.ChainId,
		Remote:     h.Static.RemoteHelp,
	}

	switch {
	case gnourl.IsRealm(), gnourl.IsPure() || gnourl.IsUser():
		return h.GetPackageView(gnourl)
	default:
		h.Logger.Debug("invalid path: path is neither a pure package or a realm")
		return http.StatusBadRequest, components.StatusErrorComponent("invalid path")
	}
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

	if gnourl.IsUser() {
		return h.GetUserView(gnourl)
	}

	// Ultimately get realm view
	return h.GetRealmView(gnourl)
}

func (h *WebHandler) GetRealmView(gnourl *weburl.GnoURL) (int, *components.View) {
	var content bytes.Buffer

	meta, err := h.Client.RenderRealm(&content, gnourl.Path, gnourl.EncodeArgs())
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

// GetUserView returns the user profile view for a given GnoURL.
func (h *WebHandler) GetUserView(gnourl *weburl.GnoURL) (int, *components.View) {
	username := strings.TrimPrefix(gnourl.Path, "/u/")
	handlename := username

	contributions := []components.UserContribution{
		{
			Title:       "gno-blogpost",
			Description: "Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
			URL:         "/r/blog/posts/example",
			Type:        components.UserContributionTypeRealm,
			Date:        time.Now().Add(-18 * time.Hour),
			Size:        120,
		},
		{
			Title:       "gno-utils",
			Description: "A collection of utility functions for Gno development.",
			URL:         "/p/utils",
			Type:        components.UserContributionTypePackage,
			Date:        time.Now().Add(-27 * time.Hour),
			Size:        79,
		},
		{
			Title:       "gno-pizza",
			Description: "A pizza delivery service.",
			URL:         "/p/pizza",
			Type:        components.UserContributionTypeRealm,
			Date:        time.Now().Add(-14 * 24 * time.Hour),
			Size:        100,
		},
		{
			Title:       "gno-trading",
			Description: "A trading service.",
			URL:         "/p/trading",
			Type:        components.UserContributionTypePackage,
			Date:        time.Now().Add(-3 * 30 * 24 * time.Hour),
			Size:        37,
		},
		{
			Title:       "gno-orders",
			Description: "An order management system.",
			URL:         "/p/orders",
			Type:        components.UserContributionTypePackage,
			Date:        time.Now().Add(-2 * 365 * 24 * time.Hour),
			Size:        23,
		},
		{
			Title:       "gno-payments",
			Description: "A payment processing system.",
			URL:         "/p/payments",
			Type:        components.UserContributionTypePackage,
			Date:        time.Now().Add(-5 * 365 * 24 * time.Hour),
			Size:        17,
		},
	}

	// Trier les contributions par date décroissante
	slices.SortFunc(contributions, func(a, b components.UserContribution) int {
		return b.Date.Compare(a.Date)
	})

	// TODO: get user data from chain
	return http.StatusOK, components.UserView(components.UserData{
		Username:      username,
		Handlename:    handlename,
		Bio:           "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Pellentesque vehicula a lectus ac porta. Pellentesque nunc massa, ultricies vitae nunc a, vulputate malesuada justo.",
		Contributions: contributions,
		Teams:         make([]struct{}, 8),
		Links: []components.UserLink{
			{
				Type: components.UserLinkTypeGithub,
				URL:  "https://github.com/" + username,
			},
			{
				Type: components.UserLinkTypeTwitter,
				URL:  "https://twitter.com/" + username,
			},
			{
				Type: components.UserLinkTypeDiscord,
				URL:  "https://discord.com/" + username,
			},
			{
				Type: components.UserLinkTypeTelegram,
				URL:  "https://t.me/" + username,
			},
			{
				Type: components.UserLinkTypeLinkedin,
				URL:  "https://linkedin.com/" + username,
			},
			{
				Type: components.UserLinkTypeLink,
				URL:  "https://example.com/" + username,
			},
		},
	})
}

func (h *WebHandler) GetHelpView(gnourl *weburl.GnoURL) (int, *components.View) {
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

	realmName := path.Base(gnourl.Path)
	return http.StatusOK, components.HelpView(components.HelpData{
		SelectedFunc: selFn,
		SelectedArgs: selArgs,
		RealmName:    realmName,
		// TODO: get chain domain and use that.
		ChainId:   h.Static.ChainId,
		PkgPath:   path.Join(h.Static.Domain, gnourl.Path),
		Remote:    h.Static.RemoteHelp,
		Functions: fsigs,
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
