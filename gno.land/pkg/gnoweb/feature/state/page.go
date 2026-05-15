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
	height := u.Height()

	oid := u.WebQuery.Get("oid")
	if oid != "" {
		if err := ValidateOID(oid); err != nil {
			return http.StatusBadRequest, components.StatusErrorComponent("invalid object id")
		}
		return h.serveObjectPage(ctx, w, u, oid, height)
	}
	return h.servePackagePage(ctx, w, u, height)
}

// servePackagePage handles `?state` — fetches StatePkg + Doc in parallel
// and renders the pretty/tree views. Previews stay lazy (hx-trigger=
// revealed in the template) so the SSR path itself does 2 RPCs.
func (h *Handler) servePackagePage(ctx context.Context, w http.ResponseWriter, u *weburl.GnoURL, height int64) (int, *components.View) {
	var (
		raw  []byte
		jdoc *doc.JSONDocumentation
	)
	var g errgroup.Group
	// Both fetches decode attacker-controlled chain data — recover so an
	// amino panic surfaces as a clean 500 (StatePkg) or a doc-less page
	// (Doc), never as a process crash.
	g.Go(func() (err error) {
		defer recoverToErr(h.deps.Logger, "statepkg", &err, "path", u.EncodeURL())
		raw, err = h.deps.Client.StatePkg(ctx, u.Path, height)
		return err
	})
	g.Go(func() error {
		defer recoverFetcher(h.deps.Logger, "doc", "path", u.EncodeURL())
		d, derr := h.deps.Client.Doc(ctx, u.Path, height)
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

	nodes, err := DecodePackage(ctx, raw, RenderConfig{MaxChildrenPerNode: maxChildrenPerNode, MaxDecodeDepth: maxDecodeDepth})
	if err != nil {
		h.deps.Logger.Error("unable to decode state JSON", "error", err, "path", u.EncodeURL())
		return http.StatusInternalServerError, components.StatusErrorComponent("failed to decode state")
	}

	var docIndex template.JS = "{}"
	if jdoc != nil {
		vals, funs, typs := flattenDocs(jdoc)
		AttachDocs(nodes, vals, funs, typs)
		docIndex = marshalDocIndexJSON(vals, funs, typs, h.deps.Logger)
	}

	// Refs hydrate lazily via hx-trigger="revealed" (state/node-details
	// in page.html) — SSR render is 2 RPCs; fragment requests pay the
	// rest, one per ref scrolled into view.
	viewMode := CanonicalViewMode(u.WebQuery.Get("view"))
	EnrichLinks(nodes, u.Path, heightParam(height), viewMode)

	data := StateData{
		PkgPath:      u.Path,
		Nodes:        nodes,
		CountLabel:   shortPackageLabel(u.Path),
		Sidebar:      BuildPackageSidebar(u.Path, nodes),
		KindCounts:   ComputeKindCounts(nodes),
		Height:       height,
		HeightParam:  heightParam(height),
		LatestHref:   template.URL(u.WithoutHeight().EncodeWebURL()), //nolint:gosec
		DocIndexJSON: docIndex,
		ViewMode:     viewMode,
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
	var g errgroup.Group
	// See servePackagePage: same panic-recover discipline. StateObject
	// fatal → 500, StateType non-fatal → type-less render.
	g.Go(func() (err error) {
		defer recoverToErr(h.deps.Logger, "stateobject", &err, "path", u.EncodeURL(), "oid", oid)
		raw, err = h.deps.Client.StateObject(ctx, oid, height)
		return err
	})
	if tid != "" {
		g.Go(func() error {
			defer recoverFetcher(h.deps.Logger, "statetype", "path", u.EncodeURL(), "tid", tid)
			tr, err := h.deps.Client.StateType(ctx, tid, height)
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

	// Refs hydrate lazily — see servePackagePage.
	viewMode := CanonicalViewMode(u.WebQuery.Get("view"))
	EnrichLinks(nodes, u.Path, heightParam(height), viewMode)

	crumbs := []StateCrumb{{Label: shortPackageLabel(u.Path), Href: RealmStateHref(u.Path)}}

	data := StateData{
		PkgPath:      u.Path,
		Nodes:        nodes,
		CountLabel:   fmt.Sprintf("Object %s", TruncOID(oid, 8, 6)),
		Crumbs:       crumbs,
		Sidebar:      BuildObjectSidebar(u.Path, oid, tid, height, decoded.Info, nodes),
		KindCounts:   ComputeKindCounts(nodes),
		Height:       height,
		HeightParam:  heightParam(height),
		LatestHref:   template.URL(u.WithoutHeight().EncodeWebURL()), //nolint:gosec
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
	if height > 0 {
		w.Header().Set("Cache-Control", "public, max-age=86400, immutable")
	} else {
		w.Header().Set("Cache-Control", "public, max-age=1")
	}
	if noindex {
		w.Header().Set("X-Robots-Tag", "noindex, nofollow")
	}
	return http.StatusOK, NewPageView(data)
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
