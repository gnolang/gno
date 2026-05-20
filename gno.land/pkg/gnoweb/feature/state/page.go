package state

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"path"
	"strconv"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"github.com/gnolang/gno/gnovm/pkg/doc"
	"golang.org/x/sync/errgroup"
)

// servePage renders the full HTML page for `?state` and `?state&oid=…`.
// Returns (http.StatusOK, nil) on success after writing the body directly;
// on error returns the mapped status plus a renderable status view so the
// gnoweb wire-in can present it through its standard chrome.
func (h *Handler) servePage(ctx context.Context, w http.ResponseWriter, r *http.Request, u *weburl.GnoURL) (int, *components.View) {
	height, err := ValidateHeightFromURL(u)
	if err != nil {
		return http.StatusBadRequest, components.StatusErrorComponent("invalid height")
	}

	oid := u.WebQuery.Get("oid")
	if oid != "" {
		if err := ValidateOID(oid); err != nil {
			return http.StatusBadRequest, components.StatusErrorComponent("invalid object id")
		}
		return h.serveObjectPage(ctx, w, u, oid, height)
	}
	return h.servePackagePage(ctx, w, r, u, height)
}

// servePackagePage handles `?state` — fetches StatePkg + Doc in parallel
// and renders the pretty/tree views. Previews stay lazy (hx-trigger=
// revealed in the template) so the SSR path itself does 2 RPCs.
// htmx search requests are served as a fragment (cards + OOB sidebar)
// instead of the full page; r.Header `HX-Request` distinguishes the two.
func (h *Handler) servePackagePage(ctx context.Context, w http.ResponseWriter, r *http.Request, u *weburl.GnoURL, height int64) (int, *components.View) {
	offset, err := ValidateOffset(u.WebQuery.Get("offset"))
	if err != nil {
		return http.StatusBadRequest, components.StatusErrorComponent("invalid offset")
	}
	limit, err := ValidateLimit(u.WebQuery.Get("limit"))
	if err != nil {
		return http.StatusBadRequest, components.StatusErrorComponent("invalid limit")
	}
	// Canonical lives in $webargs; ?query fallback mirrors u.Height().
	searchRaw := u.WebQuery.Get("search")
	if searchRaw == "" {
		searchRaw = u.Query.Get("search")
	}
	search, err := ValidateSearch(searchRaw)
	if err != nil {
		return http.StatusBadRequest, components.StatusErrorComponent("invalid search")
	}

	var (
		raw  []byte
		jdoc *doc.JSONDocumentation
	)
	// One-way fail-fast: a fatal StatePkg failure cancels gctx and short-
	// circuits the Doc sibling (freeing the RPC slot, saving a round trip).
	// Doc is best-effort — its errors and panics are swallowed so the page
	// still renders without doc comments. Decoding of `raw` happens after
	// g.Wait below, gated by parsePackage/decodePackageSlice's own panic
	// recovery so a malformed payload never crashes the process.
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() (gerr error) {
		defer recoverToErr(h.deps.Logger, "statepkg", &gerr, "path", u.EncodeURL())
		raw, gerr = h.deps.Client.StatePkg(gctx, u.Path, height)
		return gerr
	})
	g.Go(func() error {
		defer recoverFetcher(h.deps.Logger, "doc", "path", u.EncodeURL())
		d, derr := h.deps.Client.Doc(gctx, u.Path, height)
		if derr != nil {
			h.deps.Logger.Warn("unable to fetch package docs", "error", derr, "path", u.EncodeURL())
			return nil
		}
		jdoc = d
		return nil
	})

	if stateErr := g.Wait(); stateErr != nil {
		h.deps.Logger.Error("unable to fetch state", "error", stateErr, "path", u.EncodeURL(), "height", height)
		status, msg := mapClientError(stateErr, height)
		return status, components.StatusErrorComponent(msg)
	}

	// Parse once; peek-kind powers the sidebar TOC without re-decoding.
	resp, err := parsePackage(raw)
	if err != nil {
		h.deps.Logger.Error("unable to decode state JSON", "error", err, "path", u.EncodeURL())
		return http.StatusInternalServerError, components.StatusErrorComponent("failed to decode state")
	}
	realmTotal := min(len(resp.Names), len(resp.Values))
	anchors := computeAnchors(resp.Names)
	// Bound the peek loop to what the sidebar can actually render.
	peekEnd := min(realmTotal, maxSidebarTOC)
	allKinds := make([]string, peekEnd)
	allTypes := make([]string, peekEnd)
	for i := 0; i < peekEnd; i++ {
		allKinds[i], allTypes[i] = peekTopLevelKind(resp.Values[i])
	}
	pageLimit := limit
	if pageLimit <= 0 {
		pageLimit = maxTopLevelDecls
	}
	var (
		indices  []int
		setTotal int
	)
	if search != "" {
		matches := filterIndices(resp.Names, search)
		setTotal = len(matches)
		// Offset past the filtered total would render a blank page; reset.
		if offset >= setTotal {
			offset = 0
		}
		start, end := clampSliceWindow(offset, pageLimit, setTotal)
		indices = matches[start:end]
	} else {
		setTotal = realmTotal
		start, end := clampSliceWindow(offset, pageLimit, realmTotal)
		indices = make([]int, 0, end-start)
		for i := start; i < end; i++ {
			indices = append(indices, i)
		}
	}
	nodes, err := decodePackageSlice(ctx, resp, RenderConfig{MaxChildrenPerNode: maxChildrenPerNode, MaxDecodeDepth: maxDecodeDepth}, indices)
	if err != nil {
		h.deps.Logger.Error("unable to decode state JSON", "error", err, "path", u.EncodeURL())
		return http.StatusInternalServerError, components.StatusErrorComponent("failed to decode state")
	}
	for i, idx := range indices {
		if i < len(nodes) && idx < len(anchors) {
			nodes[i].Anchor = anchors[idx]
		}
	}

	var docIndex template.JS = "{}"
	if jdoc != nil {
		vals, funs, typs := flattenDocs(jdoc)
		AttachDocs(nodes, vals, funs, typs)
		docIndex = marshalDocIndexJSON(vals, funs, typs, h.deps.Logger)
	}

	// Pretty literal: pretty hrefs stay canonical; the tree container
	// builds its own view=tree hrefs inline (see _nodes.html) so the
	// CSS-only toggle never leaks fragments between containers.
	viewMode := CanonicalViewMode(u.WebQuery.Get("view"))
	hp := heightParam(height)
	EnrichLinks(nodes, u.Path, hp, ViewModePretty)

	// Sidebar mirrors the visible cards: full realm by default, filtered
	// subset under search.
	sidebarNames := resp.Names[:realmTotal]
	sidebarAnchors := anchors[:realmTotal]
	sidebarKinds := allKinds
	sidebarTypes := allTypes
	sidebarOffset := offset
	sidebarLimit := pageLimit
	if search != "" {
		sidebarNames = make([]string, 0, len(indices))
		sidebarAnchors = make([]string, 0, len(indices))
		sidebarKinds = make([]string, 0, len(indices))
		sidebarTypes = make([]string, 0, len(indices))
		for _, idx := range indices {
			sidebarNames = append(sidebarNames, resp.Names[idx])
			sidebarAnchors = append(sidebarAnchors, anchors[idx])
			if idx < len(allKinds) {
				sidebarKinds = append(sidebarKinds, allKinds[idx])
			}
			if idx < len(allTypes) {
				sidebarTypes = append(sidebarTypes, allTypes[idx])
			}
		}
		sidebarOffset = 0
		sidebarLimit = len(indices)
	}
	sidebar, truncated := BuildPackageSidebarFull(u.Path, sidebarNames, sidebarAnchors, sidebarKinds, sidebarTypes, sidebarOffset, sidebarLimit, hp)
	// Empty realm → no sidebar at all. Filter-yields-zero keeps the shell.
	if realmTotal == 0 {
		sidebar = nil
	}

	data := StateData{
		PkgPath:          u.Path,
		Nodes:            nodes,
		CountLabel:       shortPackageLabel(u.Path),
		Sidebar:          sidebar,
		SidebarTruncated: truncated,
		SidebarTotal:     realmTotal,
		KindCounts:       ComputeKindCounts(nodes),
		Height:           height,
		HeightParam:      hp,
		LatestHref:       template.URL(u.WithoutHeight().EncodeWebURL()), //nolint:gosec
		ListHref:         template.URL(u.EncodeWebURL()),                 //nolint:gosec
		DocIndexJSON:     docIndex,
		ViewMode:         viewMode,
		Pagination:       buildPagination(u.Path, hp, viewMode, setTotal, offset, pageLimit),
		SearchQuery:      search,
	}

	if r != nil && r.Header.Get("HX-Request") != "" {
		return h.writeSearchFragment(w, height, offset, data)
	}

	return h.writePage(w, height, false, data)
}

// serveObjectPage handles `?state&oid=…` (and optional `&tid=…`). Mirrors
// getStateObjectView semantics with the slim DecodeObject path.
func (h *Handler) serveObjectPage(ctx context.Context, w http.ResponseWriter, u *weburl.GnoURL, oid string, height int64) (int, *components.View) {
	tid := u.WebQuery.Get("tid")
	if tid != "" {
		if err := ValidateTID(tid); err != nil {
			return http.StatusBadRequest, components.StatusErrorComponent("invalid type id")
		}
	}

	var raw, typeRaw []byte
	// One-way fail-fast (mirrors servePackagePage): a fatal StateObject
	// failure cancels the optional StateType sibling. StateType is best-
	// effort (positional fallback on miss) so its errors and panics are
	// swallowed. DecodeObjectFull below carries its own panic recovery
	// so amino panics on hostile chain bytes never escape the handler.
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() (err error) {
		defer recoverToErr(h.deps.Logger, "stateobject", &err, "path", u.EncodeURL(), "oid", oid)
		raw, err = h.deps.Client.StateObject(gctx, oid, height)
		return err
	})
	if tid != "" {
		g.Go(func() error {
			defer recoverFetcher(h.deps.Logger, "statetype", "path", u.EncodeURL(), "tid", tid)
			tr, err := h.deps.Client.StateType(gctx, tid, height)
			if err != nil {
				h.deps.Logger.Warn("unable to fetch type for state object",
					"error", err, "path", u.EncodeURL(), "tid", tid)
				return nil
			}
			typeRaw = tr
			return nil
		})
	}

	if objErr := g.Wait(); objErr != nil {
		h.deps.Logger.Error("unable to fetch state object", "error", objErr, "path", u.EncodeURL(), "oid", oid, "height", height)
		status, msg := mapClientError(objErr, height)
		return status, components.StatusErrorComponent(msg)
	}

	// DecodeObjectFull keeps positional parity with the legacy path so
	// sidebar Info fields (Hash, OwnerID, RefCount, ...) still surface.
	decoded, err := DecodeObjectFull(raw, typeRaw, DefaultPageRenderConfig())
	if err != nil {
		h.deps.Logger.Error("unable to decode state object JSON", "error", err, "path", u.EncodeURL(), "oid", oid)
		return http.StatusInternalServerError, components.StatusErrorComponent("failed to decode state object")
	}

	nodes := decoded.Nodes

	// See servePackagePage for the literal-ViewMode rationale.
	viewMode := CanonicalViewMode(u.WebQuery.Get("view"))
	hp := heightParam(height)
	EnrichLinks(nodes, u.Path, hp, ViewModePretty)

	crumbs := []StateCrumb{{Label: shortPackageLabel(u.Path), Href: RealmStateHref(u.Path)}}

	data := StateData{
		PkgPath:      u.Path,
		Nodes:        nodes,
		CountLabel:   fmt.Sprintf("Object %s", TruncOID(oid, 8, 6)),
		Crumbs:       crumbs,
		Sidebar:      BuildObjectSidebar(u.Path, oid, tid, height, decoded.Info, nodes),
		KindCounts:   ComputeKindCounts(nodes),
		Height:       height,
		HeightParam:  hp,
		LatestHref:   template.URL(u.WithoutHeight().EncodeWebURL()), //nolint:gosec
		ListHref:     template.URL(u.EncodeWebURL()),                 //nolint:gosec
		DocIndexJSON: "{}",
		ViewMode:     viewMode,
	}

	return h.writePage(w, height, true, data)
}

// shortPackageLabel returns the last URL segment of a package path, falling
// back to the full path for root-like inputs ("/", ".", "").
func shortPackageLabel(pkgPath string) string {
	name := path.Base(pkgPath)
	if name == "/" || name == "." || name == "" {
		return pkgPath
	}
	return name
}

// writePage stamps cache/SEO headers and returns the page View so the
// gnoweb wire-in can compose it inside IndexLayout (header, breadcrumb,
// footer). The body itself is rendered later by IndexLayout via the
// returned *components.View.
func (h *Handler) writePage(w http.ResponseWriter, height int64, noindex bool, data StateData) (int, *components.View) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", cacheControlForHeight(height))
	w.Header().Set("Vary", "HX-Request")
	if noindex {
		w.Header().Set("X-Robots-Tag", "noindex, nofollow")
	}
	return http.StatusOK, NewPageView(data)
}

// writeSearchFragment writes the partial response for an HX-Request search.
// Body is written here — caller must not write a View.
func (h *Handler) writeSearchFragment(w http.ResponseWriter, height int64, offset int, data StateData) (int, *components.View) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", cacheControlForHeight(height))
	// Splits cache entries for partial vs full-page responses.
	w.Header().Set("Vary", "HX-Request")
	w.Header().Set("HX-Push-Url", string(canonicalStateURL(data.PkgPath, data.HeightParam, data.ViewMode, data.SearchQuery, offset)))
	w.WriteHeader(http.StatusOK)
	if err := SearchFragmentTemplate.ExecuteTemplate(w, "searchFragment", data); err != nil {
		h.deps.Logger.Error("search fragment template execute failed", "err", err)
	}
	return http.StatusOK, nil
}

// heightParam returns the decimal string stamped into every fragment hx-get
// URL so fragments inherit the parent page's concrete height during
// stale-while-revalidate windows. Empty for latest (height=0).
func heightParam(height int64) string {
	if height <= 0 {
		return ""
	}
	return strconv.FormatInt(height, 10)
}

// flattenDocs projects a JSONDocumentation into the three flat (Name, Doc)
// slices AttachDocs consumes. Mirrors the legacy handler_http.go projection.
func flattenDocs(jdoc *doc.JSONDocumentation) (vals, funs, typs []NamedDoc) {
	if jdoc == nil {
		return
	}
	for _, vd := range jdoc.Values {
		for _, v := range vd.Values {
			d := v.Doc
			if d == "" {
				d = vd.Doc
			}
			vals = append(vals, NamedDoc{Name: v.Name, Doc: d})
		}
	}
	for _, fn := range jdoc.Funcs {
		funs = append(funs, NamedDoc{Name: fn.Name, Doc: fn.Doc})
	}
	for _, t := range jdoc.Types {
		typs = append(typs, NamedDoc{Name: t.Name, Doc: t.Doc})
	}
	return
}

// marshalDocIndexJSON builds the inline `{name: doc}` map the
// controller-state.ts hydration layer reads on htmx:afterSwap.
func marshalDocIndexJSON(vals, funs, typs []NamedDoc, logger *slog.Logger) template.JS {
	index := make(map[string]string, len(vals)+len(funs)+len(typs))
	for _, group := range [][]NamedDoc{vals, funs, typs} {
		for _, d := range group {
			if d.Doc != "" {
				index[d.Name] = d.Doc
			}
		}
	}
	if len(index) == 0 {
		return template.JS("{}")
	}
	b, err := json.Marshal(index)
	if err != nil {
		logger.Error("marshal doc index failed", "error", err)
		return template.JS("{}")
	}
	return template.JS(b) //nolint:gosec // JSON object intended for <script type="application/json"> embed
}
