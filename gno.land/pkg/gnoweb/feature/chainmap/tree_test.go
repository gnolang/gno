package chainmap

import (
	"testing"
)

func TestBuildGroupsByNextSegment(t *testing.T) {
	t.Parallel()

	paths := []string{
		"/r/gnoland/home",
		"/r/gnoland/blog",
		"/r/gnoland/blog/admin",
		"/r/alice/game",
		"/r/zed",
		"/p/outside/prefix", // not under /r/: dropped
	}
	groups := buildGroups("/r/", paths)

	if len(groups) != 3 {
		t.Fatalf("groups = %d, want 3 (gnoland, alice, zed)", len(groups))
	}
	// Largest first.
	if g := groups[0]; g.Key != "gnoland" || len(g.Tiles) != 3 {
		t.Fatalf("first group = %s with %d tiles, want gnoland with 3", g.Key, len(g.Tiles))
	}
	if got := groups[0].Tiles[0].Label; got != "blog" {
		t.Errorf("tiles sort by label: first = %q, want blog", got)
	}
	if got := groups[0].ZoomURL; got != "/r/gnoland/$map" {
		t.Errorf("zoom = %q, want /r/gnoland/$map", got)
	}

	// /r/zed is a package with nothing below it: a lone tile, no header, and
	// no zoom, since a map of what is under it would be empty.
	var zed *Group
	for _, g := range groups {
		if g.Key == "zed" {
			zed = g
		}
	}
	if zed == nil || zed.Header || zed.ZoomURL != "" {
		t.Fatalf("zed = %+v, want a headerless group with no zoom", zed)
	}
	if zed.Tiles[0].Label != "zed" || zed.Tiles[0].Path != "/r/zed" {
		t.Errorf("zed tile = %+v", zed.Tiles[0])
	}
}

// The map is a rendering of the listing, so every listed path under the
// prefix must come out as exactly one tile.
func TestBuildGroupsKeepsEveryPath(t *testing.T) {
	t.Parallel()

	paths := []string{"/r/a/x", "/r/a/y/z", "/r/b", "/r/c/d/e/f"}
	tiles := 0
	for _, g := range buildGroups("/r/", paths) {
		tiles += len(g.Tiles)
	}
	if tiles != len(paths) {
		t.Fatalf("tiles = %d, want %d", tiles, len(paths))
	}
}

func TestLayoutFillsTheMap(t *testing.T) {
	t.Parallel()

	groups := buildGroups("/r/", []string{"/r/a/x", "/r/a/y", "/r/a/z", "/r/b/x"})
	layout(groups)

	var area float64
	for _, g := range groups {
		area += g.Box.Width * g.Box.Height
		for _, tl := range g.Tiles {
			if tl.Box.Width <= 0 || tl.Box.Height <= 0 {
				t.Errorf("tile %s has no area: %+v", tl.Path, tl.Box)
			}
		}
	}
	if area < 9999 || area > 10001 { // percent × percent
		t.Errorf("groups cover %.2f%%², want the whole map", area)
	}
}
