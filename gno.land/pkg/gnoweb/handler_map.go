package gnoweb

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"path"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/feature/chainmap"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/feature/state"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

// chainmapDeps wires feature/chainmap. The indexer is the one configured for
// search: *indexer.Client answers both features, so the map needs no flag of
// its own. A nil cfg.Indexer leaves the feature without one, which is the
// switch.
func chainmapDeps(cfg *HTTPHandlerConfig, logger *slog.Logger, trustedProxies []*net.IPNet) chainmap.Deps {
	deps := chainmap.Deps{
		Imports: importReader{client: cfg.ClientAdapter, domain: cfg.Meta.Domain},
		Domain:  cfg.Meta.Domain,
		Logger:  logger,
	}
	if cfg.Indexer == nil {
		return deps
	}
	idx, ok := cfg.Indexer.(chainmap.Indexer)
	if !ok {
		// Not an error a deployment can hit with *indexer.Client, but a
		// silent switch-off would hide it.
		logger.Warn("indexer cannot back the map and dependency pages", "type", fmt.Sprintf("%T", cfg.Indexer))
		return deps
	}
	deps.Indexer = idx
	// Its own bucket, and a narrow one: an importer lookup is a whole-chain
	// indexer scan plus a node read per candidate, so the general rate would
	// let one address run hundreds a minute. Cached answers are not charged.
	deps.Limiter = state.NewIPLimiter(state.RateLimitConfig{
		PerMinute:      chainmap.LookupsPerMinute,
		Burst:          chainmap.LookupsBurst,
		TrustedProxies: trustedProxies,
	})
	return deps
}

// importReader adapts ClientAdapter to chainmap.ImportReader: a package's
// non-test imports are what vm/qdoc reports. A missing package is both
// chainmap's ErrNotLive and the client's not-found, so either side reads it.
type importReader struct {
	client ClientAdapter
	domain string
}

func (r importReader) Imports(ctx context.Context, pkgPath string) ([]string, error) {
	d, err := r.client.Doc(ctx, strings.TrimPrefix(pkgPath, r.domain), 0)
	if errors.Is(err, ErrClientPackageNotFound) {
		return nil, fmt.Errorf("%w: %w", chainmap.ErrNotLive, err)
	}
	if err != nil {
		return nil, err
	}
	return d.Imports, nil
}

// GetDepsView renders a package's dependencies: what it imports, and what
// imports it. The answer costs an indexer scan per package, so it is not a
// page for crawlers to walk.
func (h *HTTPHandler) GetDepsView(r *http.Request, gnourl *weburl.GnoURL, indexData *components.IndexData) (int, *components.View) {
	indexData.HeadData.NoIndex = true
	pkgPath := strings.TrimSuffix(gnourl.Path, "/")
	status, view, err := h.ChainMap.DepsView(r, pkgPath, displayPackageName(pkgPath))
	if err != nil {
		return GetClientErrorStatusView(gnourl, err, 0)
	}
	return status, view
}

// listPaths lists the package paths below gnourl, capped at maxListedPaths.
// One more than is shown is asked for, so a listing stopping at the cap can
// tell "exactly the cap" from "more than the cap".
func (h *HTTPHandler) listPaths(ctx context.Context, gnourl *weburl.GnoURL) (paths []string, truncated bool) {
	prefix := path.Join(h.Static.Domain, gnourl.Path) + "/"
	paths, qerr := h.Client.ListPaths(ctx, prefix, maxListedPaths+1)
	if qerr != nil {
		h.Logger.Error("unable to query path", "error", qerr, "path", gnourl.EncodeURL())
	} else {
		h.Logger.Debug("query paths", "prefix", prefix, "paths", len(paths))
	}
	if len(paths) == 0 || paths[0] == "" {
		return nil, false
	}
	if len(paths) > maxListedPaths {
		return paths[:maxListedPaths], true
	}
	return paths, false
}

// mapRequest asks renderListing for the map rather than the list. up is the
// listing one level up, or "" when there is none to zoom out to.
type mapRequest struct{ up string }

// renderListing renders a listing as a list, or as a map when m is set. The
// two render this one listing, so they cannot disagree about what exists.
func (h *HTTPHandler) renderListing(ctx context.Context, gnourl *weburl.GnoURL, indexData *components.IndexData, paths []string, truncated bool, m *mapRequest) (int, *components.View) {
	if len(paths) == 0 {
		// Both the realm view and the source view funnel here when nothing is
		// live at the path, so this is the one place that has to distinguish
		// "never submitted" from "submitted, not approved yet".
		if view := h.pendingApprovalView(ctx, gnourl); view != nil {
			return http.StatusNotFound, view
		}
		return GetClientErrorStatusView(gnourl, ErrClientPackageNotFound, 0)
	}

	indexData.Mode = components.ViewModeExplorer
	indexData.HeaderData.Mode = indexData.Mode

	// A listing too small for a map offers none, and draws a list even when
	// its map is asked for.
	mappable := len(paths) >= chainmap.MinPackages
	indexData.HeaderData.MapTab = mappable
	if mappable && m != nil {
		parts := h.ChainMap.Map(ctx, chainmap.Listing{Path: gnourl.Path, Paths: paths, Up: m.up})
		return http.StatusOK, components.ExplorerView(gnourl.Path, paths, truncated, &parts)
	}
	return http.StatusOK, components.ExplorerView(gnourl.Path, paths, truncated, nil)
}

// GetMapView draws the listing below a path. A map exists only where the list
// does, which is where no package lives: on a package its Directory tab would
// open the package, not the listing. The same holds one level up, so the
// zoom-out link is offered only where the parent is not a package either. Both
// checks run beside the listing query.
func (h *HTTPHandler) GetMapView(ctx context.Context, gnourl *weburl.GnoURL, indexData *components.IndexData) (int, *components.View) {
	root := strings.TrimSuffix(gnourl.Path, "/")
	isPkg := func(p string) <-chan error {
		ch := make(chan error, 1)
		go func() {
			_, err := h.Client.ListFiles(ctx, p, 0)
			ch <- err
		}()
		return ch
	}

	pkgErr := isPkg(root)
	// Below a kind's root ("/r/foo") the parent is that root, never a
	// package; deeper, it has to be asked.
	var up string
	var upErr <-chan error
	switch parent := path.Dir(root); {
	case strings.Count(root, "/") == 2:
		up = parent + "/"
	case strings.Count(root, "/") > 2:
		up, upErr = parent+"/", isPkg(parent)
	}
	paths, truncated := h.listPaths(ctx, gnourl)

	switch err := <-pkgErr; {
	case err == nil:
		return http.StatusNotFound, components.StatusErrorComponent("This path is a package, not a listing: open it, or map the path above it.")
	case errors.Is(err, ErrClientPackageNotFound):
	case ctx.Err() != nil:
		// The reader left or the request ran out of time: nothing to report.
		return GetClientErrorStatusView(gnourl, ctx.Err(), 0)
	default:
		// The node could not say: drawing the listing could draw one below a
		// package.
		h.Logger.Warn("map: unable to tell a package from a listing", "path", root, "error", err)
		return http.StatusBadGateway, components.StatusErrorComponent("The node could not tell whether this path is a package. Try again in a moment.")
	}
	// A parent the node could not place gets no link: a dead one is worse
	// than none, and the breadcrumb still leads up.
	if upErr != nil && !errors.Is(<-upErr, ErrClientPackageNotFound) {
		up = ""
	}
	return h.renderListing(ctx, gnourl, indexData, paths, truncated, &mapRequest{up: up})
}
