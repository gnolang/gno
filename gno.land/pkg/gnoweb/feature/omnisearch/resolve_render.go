package omnisearch

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"
)

// Bounds on the rendered-content search. One candidate costs one Render on
// the node — a page view — so the field must be narrowed first, then capped.
const (
	maxRenderCandidates = 8
	renderConcurrency   = 4
	renderSnippetRadius = 60
)

// errRenderNeedsNarrowing states the one precondition.
var errRenderNeedsNarrowing = errors.New(
	"rendered-content search needs a narrower field: add author:<name> or in:/r/<path>")

// renderSelector searches what realms display. Chain data, not indexer: the
// node executes Render(), so a hit is what a reader would see.
func renderSelector() *Selector {
	return &Selector{
		Name:  "render",
		Hint:  "render:<text>",
		Label: "Rendered content",
		Scope: ScopeGlobal,
		// Executing Render costs what serving a page view costs, so this one
		// never runs on a keystroke — only on a committed navigation.
		PageOnly: true,
		MinTerm:  3,
		resolve: func(ctx context.Context, h *Handler, q *Query, term string) ([]Result, error) {
			return h.resolveRender(ctx, q, term)
		},
	}
}

func (h *Handler) resolveRender(ctx context.Context, q *Query, term string) ([]Result, error) {
	candidates, err := h.renderCandidates(ctx, q)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return nil, nil
	}

	// Over the original bytes: an offset into a lowercased copy can be
	// wrong, since ToLower is not length-preserving for every rune.
	re, err := regexp.Compile(`(?i)` + regexp.QuoteMeta(term))
	if err != nil {
		return nil, err
	}

	var (
		mu   sync.Mutex
		hits []Result
	)

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(renderConcurrency)
	for _, pkgPath := range candidates {
		g.Go(func() (err error) {
			// errgroup does not recover and net/http's does not reach here:
			// a panic on hostile chain data would take the process down.
			defer func() {
				if r := recover(); r != nil {
					h.deps.Logger.Error("omnisearch: render fetcher panic recovered",
						"path", pkgPath, "panic", fmt.Sprintf("%.512s", r))
					err = nil
				}
			}()

			body, err := h.deps.Client.Realm(gctx, pkgPath, "")
			if err != nil {
				// Not a search failure: the other candidates still answer.
				h.deps.Logger.Debug("omnisearch: render failed", "path", pkgPath, "error", err)
				return nil
			}
			loc := re.FindIndex(body)
			if loc == nil {
				return nil
			}

			mu.Lock()
			defer mu.Unlock()
			hits = append(hits, Result{
				Title:  pkgPath,
				Detail: snippetAround(body, loc[0], loc[1]),
				Href:   safePathHref(pkgPath),
				Tags:   []string{"rendered"},
			})
			return nil
		})
	}
	// Wait never reports an error: every goroutine swallows its own failure,
	// because a realm that does not render is not a search failure. errgroup
	// is here for SetLimit and the shared context, not for propagation.
	_ = g.Wait()
	return capResults(hits), nil
}

// renderCandidates is the field this search may render: `in:` names one
// realm, `author:` costs one listing and is then capped.
func (h *Handler) renderCandidates(ctx context.Context, q *Query) ([]string, error) {
	if q.PkgPath != "" {
		return []string{q.PkgPath}, nil
	}

	author, ok := q.Get(FilterAuthor)
	if !ok || author == "" {
		return nil, errRenderNeedsNarrowing
	}

	paths, _, _, err := h.deps.Directory.Paths(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]string, 0, maxRenderCandidates)
	for _, p := range paths {
		rel := strings.TrimPrefix(p, h.deps.Domain)
		if !strings.EqualFold(namespaceOf(rel), author) || !isSafeRelPath(rel) {
			continue
		}
		out = append(out, rel)
		if len(out) == maxRenderCandidates {
			break
		}
	}
	return out, nil
}

// snippetAround returns the match with context either side. Bounds are
// clamped: a snippet is not worth a panic.
func snippetAround(body []byte, from, to int) string {
	start := max(0, min(from, len(body))-renderSnippetRadius)
	end := min(len(body), max(to, 0)+renderSnippetRadius)
	if start >= end {
		return ""
	}

	s := strings.Join(strings.Fields(string(body[start:end])), " ")
	if start > 0 {
		s = "…" + s
	}
	if end < len(body) {
		s += "…"
	}
	return s
}
