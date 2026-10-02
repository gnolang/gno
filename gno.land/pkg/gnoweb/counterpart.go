package gnoweb

import (
	"context"
	gopath "path"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

// A realm and the pure packages it is built on usually share a project
// directory in the same namespace (/r/alice/golf/game next to
// /p/alice/golf/physics), but nothing on chain links them. The counterpart
// link in the header crosses from one side to the other.

// maxCounterpartPaths caps the listing of the other side's project directory.
// Past it the link still points at the project, only the count is a floor.
const maxCounterpartPaths = 100

// counterpartGrace is how long a rendered page waits for a lookup still in
// flight. The lookup runs alongside the page's own queries, so a node that is
// slow to list paths costs a missing link, never a slower page.
const counterpartGrace = 300 * time.Millisecond

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
// root: the twin itself when it exists, the only package when there is one,
// and otherwise the deepest directory above the twin that holds any, as a
// listing. n is how many packages that target covers; zero means no link.
func counterpartTarget(twin, root string, paths []string) (target string, n int) {
	members := make([]string, 0, len(paths))
	for _, p := range paths {
		if !isUnder(p, root) || !(weburl.GnoURL{Path: p}).IsValidPath() || strings.Contains(p, "//") {
			continue
		}
		if p == twin {
			return twin, 1
		}
		members = append(members, p)
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
		case n > 1:
			return dir, n
		case dir == root:
			return "", 0
		}
	}
}

// isUnder reports whether p is dir or lies below it.
func isUnder(p, dir string) bool {
	return p == dir || strings.HasPrefix(p, dir+"/")
}

// counterpartLink builds the header link for a lookup result.
func counterpartLink(target, root string, n int) *components.HeaderLink {
	label, icon := "Package", "ico-pure"
	if strings.HasPrefix(root, "/r/") {
		label, icon = "Realm", "ico-realm"
	}
	if n > 1 {
		label += "s"
	}
	return &components.HeaderLink{
		Label: label,
		URL:   target,
		Icon:  icon,
		Title: counterpartTitle(target, n),
	}
}

func counterpartTitle(target string, n int) string {
	switch {
	case n == 1:
		return target
	case n >= maxCounterpartPaths:
		return target + " (" + strconv.Itoa(maxCounterpartPaths) + "+ packages)"
	default:
		return target + " (" + strconv.Itoa(n) + " packages)"
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

	ctx, cancel := context.WithCancel(ctx)
	done := make(chan *components.HeaderLink, 1)
	go func() {
		paths, err := h.Client.ListPaths(ctx, gopath.Join(h.Static.Domain, root), maxCounterpartPaths)
		if err != nil {
			h.Logger.Debug("counterpart lookup failed", "root", root, "error", err)
			done <- nil
			return
		}
		if len(paths) > maxCounterpartPaths {
			paths = paths[:maxCounterpartPaths]
		}
		target, n := counterpartTarget(twin, root, paths)
		if n == 0 {
			done <- nil
			return
		}
		done <- counterpartLink(target, root, n)
	}()

	return func() *components.HeaderLink {
		defer cancel()
		select {
		case link := <-done:
			return link
		case <-time.After(counterpartGrace):
			return nil
		}
	}
}
