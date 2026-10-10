package gnoweb

import (
	"context"
	gopath "path"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
	"golang.org/x/sync/singleflight"
)

// A realm and the pure packages it is built on often share a project
// directory in the same namespace (/r/alice/golf/game next to
// /p/alice/golf/physics), but nothing on chain links them: the only tie is the
// path, so the header calls the other side "matching", never related.

// maxCounterpartPaths caps the listing of the other side's project directory.
// Past it the link still points at the project, only the count is a floor.
const maxCounterpartPaths = 100

// counterpartGrace is how long a rendered page waits for a lookup still in
// flight. The lookup runs alongside the page's own queries, so a node that is
// slow to list paths costs a missing link, never a slower page.
const counterpartGrace = 300 * time.Millisecond

// counterpartTimeout bounds a lookup on its own: a page that stops waiting
// does not cancel it, so a slow answer still fills the cache.
const counterpartTimeout = 5 * time.Second

// counterpartTTL is how long a project listing is reused. Most pages have no
// counterpart, so caching the empty answers matters as much as the others; a
// package deployed meanwhile shows up within the TTL.
const counterpartTTL = time.Minute

// maxCounterpartEntries bounds the cache; past it, new roots go uncached
// until entries expire.
const maxCounterpartEntries = 4096

// counterpartRoots returns, for a realm or pure package path, the same path on
// the other side (twin) and the project directory it belongs to there (root).
// ok is false for anything else, and for a bare namespace: /u/<name> is where
// a namespace is viewed whole.
func counterpartRoots(pkgPath string) (twin, root string, ok bool) {
	kind, rest, _ := strings.Cut(strings.TrimPrefix(pkgPath, "/"), "/")
	switch kind {
	case "r":
		kind = "p"
	case "p":
		kind = "r"
	default:
		return "", "", false
	}

	segs := strings.Split(strings.Trim(rest, "/"), "/")
	if len(segs) < 2 || slices.Contains(segs, "") {
		return "", "", false
	}

	return "/" + kind + "/" + strings.Join(segs, "/"),
		"/" + kind + "/" + segs[0] + "/" + segs[1],
		true
}

// counterpartTarget picks the page the link opens, given the paths live under
// root, and n, how many packages that page lists; zero means no link.
//
// The twin opens directly unless it has siblings (v0 next to v2); without a
// twin, the only package opens when there is one, and otherwise the deepest
// directory above the twin that holds any. A directory gnoweb shows as a
// listing only when no package lives there, and that listing holds its whole
// subtree, so n counts the subtree. A directory that is itself a package
// opens that package instead, so the link then points at its overview's
// Directories section, which lists direct children only: it holds the twin's
// siblings, or, on the twinless walk, the package's own children when two or
// more match. With fewer the walk names that package alone, since no page
// lists what lies deeper below it.
func counterpartTarget(twin, root string, paths []string) (target string, n int) {
	members := make([]string, 0, len(paths))
	hasTwin := false
	for _, p := range paths {
		if !isUnder(p, root) || !(weburl.GnoURL{Path: p}).IsValidPath() || strings.Contains(p, "//") {
			continue
		}
		hasTwin = hasTwin || p == twin
		members = append(members, p)
	}

	if hasTwin {
		dir := gopath.Dir(twin)
		siblings := countChildren(members, dir)
		switch {
		case siblings == 1:
			return twin, 1
		case slices.Contains(members, dir):
			return dir + "$source#subpackages", siblings
		default:
			return dir, countUnder(members, dir)
		}
	}

	for dir := twin; ; dir = gopath.Dir(dir) {
		var last string
		for _, m := range members {
			if isUnder(m, dir) {
				n++
				last = m
			}
		}
		switch {
		case n == 1:
			return last, 1
		case n > 1 && slices.Contains(members, dir):
			if children := countChildren(members, dir); children > 1 {
				return dir + "$source#subpackages", children
			}
			return dir, 1
		case n > 1:
			return dir, n
		case dir == root:
			return "", 0
		}
	}
}

// countChildren counts the members lying directly below dir.
func countChildren(members []string, dir string) int {
	n := 0
	for _, m := range members {
		if gopath.Dir(m) == dir {
			n++
		}
	}
	return n
}

// countUnder counts the members lying below dir.
func countUnder(members []string, dir string) int {
	n := 0
	for _, m := range members {
		if isUnder(m, dir) {
			n++
		}
	}
	return n
}

// isUnder reports whether p is dir or lies below it.
func isUnder(p, dir string) bool {
	return p == dir || strings.HasPrefix(p, dir+"/")
}

// counterpartLink builds the header link for a lookup result. cut reports a
// listing that reached maxCounterpartPaths, where n may miss paths past it.
func counterpartLink(target, root string, n int, cut bool) *components.HeaderLink {
	kind, icon := "package", "ico-pure"
	if strings.HasPrefix(root, "/r/") {
		kind, icon = "realm", "ico-realm"
	}
	label := "Matching " + kind
	if n > 1 {
		count := strconv.Itoa(n)
		if cut {
			count += "+"
		}
		label = count + " matching " + kind + "s"
	}
	path, _, _ := strings.Cut(target, "$")
	return &components.HeaderLink{
		Label: label,
		URL:   target,
		Path:  path,
		Icon:  icon,
	}
}

// startCounterpart looks the other side up in the background and returns a
// function that collects the link, or nil when there is none, it failed, or
// it did not arrive within counterpartGrace of being asked for.
func (h *HTTPHandler) startCounterpart(ctx context.Context, gnourl *weburl.GnoURL) func() *components.HeaderLink {
	twin, root, ok := counterpartRoots(gnourl.Path)
	if !ok {
		return func() *components.HeaderLink { return nil }
	}

	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), counterpartTimeout)
	done := make(chan *components.HeaderLink, 1)
	go func() {
		defer cancel()
		paths, err := h.counterparts.get(root, func() ([]string, error) {
			paths, err := h.Client.ListPaths(ctx, gopath.Join(h.Static.Domain, root), maxCounterpartPaths)
			if len(paths) > maxCounterpartPaths {
				paths = paths[:maxCounterpartPaths]
			}
			return paths, err
		})
		if err != nil {
			h.Logger.Debug("counterpart lookup failed", "root", root, "error", err)
			done <- nil
			return
		}
		target, n := counterpartTarget(twin, root, paths)
		if n == 0 {
			done <- nil
			return
		}
		done <- counterpartLink(target, root, n, len(paths) >= maxCounterpartPaths && isUnder(paths[len(paths)-1], root))
	}()

	return func() *components.HeaderLink {
		select {
		case link := <-done:
			return link
		case <-time.After(counterpartGrace):
			return nil
		}
	}
}

// counterpartCache keeps each project listing for counterpartTTL and
// coalesces concurrent lookups of the same root. Errors are not kept: a node
// that failed once may answer the next request.
type counterpartCache struct {
	mu      sync.Mutex
	entries map[string]counterpartEntry
	sf      singleflight.Group
	sweep   time.Time        // when a full cache next has an entry to drop
	now     func() time.Time // for tests; time.Now when nil
}

type counterpartEntry struct {
	paths   []string
	expires time.Time
}

func (c *counterpartCache) get(root string, list func() ([]string, error)) ([]string, error) {
	if paths, ok := c.lookup(root); ok {
		return paths, nil
	}
	v, err, _ := c.sf.Do(root, func() (any, error) {
		if paths, ok := c.lookup(root); ok {
			return paths, nil
		}
		paths, err := list()
		if err == nil {
			c.store(root, paths)
		}
		return paths, err
	})
	paths, _ := v.([]string)
	return paths, err
}

func (c *counterpartCache) clock() time.Time {
	if c.now != nil {
		return c.now()
	}
	return time.Now()
}

func (c *counterpartCache) lookup(root string) ([]string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[root]
	if !ok || !c.clock().Before(e.expires) {
		return nil, false
	}
	return e.paths, true
}

// store keeps the listing unless the cache is full of live entries.
func (c *counterpartCache) store(root string, paths []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := c.clock()
	if c.entries == nil {
		c.entries = make(map[string]counterpartEntry)
	}
	if len(c.entries) >= maxCounterpartEntries {
		if now.Before(c.sweep) {
			return
		}
		c.sweep = now.Add(counterpartTTL)
		for k, e := range c.entries {
			if !now.Before(e.expires) {
				delete(c.entries, k)
			} else if e.expires.Before(c.sweep) {
				c.sweep = e.expires
			}
		}
		if len(c.entries) >= maxCounterpartEntries {
			return
		}
	}
	c.entries[root] = counterpartEntry{paths: paths, expires: now.Add(counterpartTTL)}
}
