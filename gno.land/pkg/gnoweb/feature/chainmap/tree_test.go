package chainmap

import (
	"slices"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"
)

func tilesOf(groups []*Group) []Tile {
	var out []Tile
	for _, g := range groups {
		for _, s := range g.Subgroups {
			out = append(out, s.Tiles...)
		}
	}
	return out
}

func groupByKey(groups []*Group, key string) *Group {
	for _, g := range groups {
		if g.Key == key {
			return g
		}
	}
	return nil
}

func TestBuildGroupsByFirstThenSecondSegment(t *testing.T) {
	t.Parallel()

	paths := []string{"/r/zed", "/p/outside/prefix"} // the second is dropped
	for _, sub := range []string{"x", "x", "x", "demo", "demo", "blog"} {
		paths = append(paths, "/r/moul/"+sub+"/p"+strconv.Itoa(len(paths)))
	}
	groups := buildGroups("/r/", paths)

	if len(groups) != 2 || groups[0].Key != "moul" || groups[0].Count != 6 {
		t.Fatalf("groups = %+v, want moul (6) then zed", groups)
	}
	moul := groups[0]
	if moul.zoom != "/r/moul/" {
		t.Errorf("zoom = %q, want /r/moul/", moul.zoom)
	}
	// Six packages: split by the next segment. x holds three and keeps a band;
	// demo and blog are too small and pool together, named below moul.
	subs := make(map[string]*Subgroup)
	for _, sg := range moul.Subgroups {
		subs[sg.Key] = sg
	}
	x, pool := subs["x"], subs[""]
	if len(moul.Subgroups) != 2 || x == nil || pool == nil || len(x.Tiles) != 3 || len(pool.Tiles) != 3 {
		t.Fatalf("moul subgroups = %v, want x (3) and the pooled rest (3)", subs)
	}
	if x.zoom != "/r/moul/x/" {
		t.Errorf("subgroup zoom = %q, want /r/moul/x/", x.zoom)
	}
	if pool.zoom != "" {
		t.Errorf("the pooled subgroup must have no zoom: %q", pool.zoom)
	}
	for _, tl := range pool.Tiles {
		if name := tileName(tl.Path, moul.root+"/"); !strings.Contains(name, "/") {
			t.Errorf("pooled tile %q must be named by its path below the group", name)
		}
	}

	// /r/zed is a package with nothing below it: no zoom, one keyless
	// subgroup, its tile named after it.
	zed := groupByKey(groups, "zed")
	if zed.zoom != "" || len(zed.Subgroups) != 1 || zed.Subgroups[0].Key != "" {
		t.Fatalf("zed = %+v", zed)
	}
	if tl := zed.Subgroups[0].Tiles[0]; tl.Path != "/r/zed" || tileName(tl.Path, zed.root+"/") != "zed" {
		t.Errorf("zed tile = %+v", tl)
	}
}

// Below subgroupMin a group is not split: a split of two or three packages
// only draws frames around single tiles.
func TestSmallGroupsAreNotSplit(t *testing.T) {
	t.Parallel()

	g := buildGroups("/r/", []string{"/r/a/x/1", "/r/a/y/2", "/r/a/z/3"})[0]
	if len(g.Subgroups) != 1 || g.Subgroups[0].Key != "" {
		t.Fatalf("subgroups = %+v, want the group unsplit", g.Subgroups)
	}
	if name := tileName(g.Subgroups[0].Tiles[0].Path, g.root+"/"); name != "x/1" {
		t.Errorf("tile name = %q, want the path below the group", name)
	}
}

// A tile is named below the nearest band drawn over it. When its group's
// band does not fit, the group's key goes into the tile's own label, or
// nothing on the map would say which namespace it belongs to.
func TestTilesNameTheBandThatIsNotDrawn(t *testing.T) {
	t.Parallel()

	labels := func(base string) []string {
		g := buildGroups("/r/", []string{"/r/a/x", "/r/a/y"})[0]
		layoutSubgroups(g.Subgroups, Rect{0, 0, mapWidth, mapHeight}, base)
		var out []string
		for _, tl := range g.Subgroups[0].Tiles {
			out = append(out, tl.Label.Text)
		}
		return out
	}
	if got := labels("/r/a/"); !slices.Equal(got, []string{"x", "y"}) {
		t.Errorf("under the group's band, labels = %q, want x and y", got)
	}
	if got := labels("/r/"); !slices.Equal(got, []string{"a/x", "a/y"}) {
		t.Errorf("with no band, labels = %q, want a/x and a/y", got)
	}
}

// The map is a rendering of the listing, so every listed path under the
// prefix comes out as exactly one tile.
func TestBuildGroupsKeepsEveryPath(t *testing.T) {
	t.Parallel()

	paths := []string{"/r/a/x", "/r/a/y/z", "/r/b", "/r/c/d/e/f", "/r/a/q/1", "/r/a/q/2", "/r/a/q/3", "/r/a/w"}
	if n := len(tilesOf(buildGroups("/r/", paths))); n != len(paths) {
		t.Fatalf("tiles = %d, want %d", n, len(paths))
	}
}

// A key that is itself a package gets no zoom: the map of what lies below it
// is the package page, not a listing.
func TestNoZoomIntoAPackage(t *testing.T) {
	t.Parallel()

	groups := buildGroups("/r/", []string{"/r/pkg", "/r/pkg/sub", "/r/ns/a", "/r/ns/b"})
	if g := groupByKey(groups, "pkg"); g.zoom != "" {
		t.Errorf("zoom into a package: %q", g.zoom)
	}
	if g := groupByKey(groups, "ns"); g.zoom != "/r/ns/" {
		t.Errorf("zoom into a namespace = %q, want /r/ns/", g.zoom)
	}
}

func TestLayoutTilesTheMapAndFitsEveryLabel(t *testing.T) {
	t.Parallel()

	var paths []string
	for i := range 40 {
		paths = append(paths, "/r/moul/x/daily/package-with-a-long-name-"+strconv.Itoa(i))
	}
	paths = append(paths, "/r/gnoland/home", "/r/gnoland/blog", "/r/sys/users")
	groups := buildGroups("/r/", paths)
	layout(groups)

	var area float64
	for _, g := range groups {
		area += g.Rect.W * g.Rect.H
		if g.Head != nil {
			checkFits(t, g.Head.Label.Text, g.Rect.W, headFont)
		}
		for _, s := range g.Subgroups {
			if s.Head != nil {
				checkFits(t, s.Head.Label.Text, s.Rect.W, subFont)
			}
			for _, tl := range s.Tiles {
				if tl.Rect.W <= 0 || tl.Rect.H <= 0 {
					t.Errorf("tile %s has no area", tl.Path)
				}
				if tl.Label != nil {
					checkFits(t, tl.Label.Text, tl.Rect.W, tileFont)
				}
			}
		}
	}
	if full := mapWidth * mapHeight; area < full*0.999 || area > full*1.001 {
		t.Errorf("groups cover %.0f, want the whole map %.0f", area, full)
	}
}

func checkFits(t *testing.T, text string, w, font float64) {
	t.Helper()
	if n := utf8.RuneCountInString(text); float64(n)*font*monoAdvance > w-2*labelPad+0.01 {
		t.Errorf("label %q (%d chars) overflows a box %.1f wide", text, n, w)
	}
	if strings.HasSuffix(text, "…") && utf8.RuneCountInString(text) < minLabelChars {
		t.Errorf("label %q is too short to be worth drawing", text)
	}
}

func TestFit(t *testing.T) {
	t.Parallel()

	room := func(chars int) float64 { return float64(chars)*tileFont*monoAdvance + 2*labelPad }
	cases := []struct {
		text string
		w    float64
		want string
	}{
		{"daily", room(5), "daily"},
		{"governance", room(6), "gover…"},
		{"governance", room(3), ""},
		{"a", room(1), "a"},
	}
	for _, c := range cases {
		if got := fit(c.text, c.w, tileFont); got != c.want {
			t.Errorf("fit(%q, %.1f) = %q, want %q", c.text, c.w, got, c.want)
		}
	}
}

// A group that is one package at its own root carries no band: its tile
// already names it. A one-package group below a folder keeps its band, which
// is the only place that names the folder.
func TestNoBandOverALonePackage(t *testing.T) {
	t.Parallel()

	groups := buildGroups("/r/", []string{"/r/blog", "/r/boards2/v0"})
	layout(groups)
	if g := groupByKey(groups, "blog"); g.Head != nil {
		t.Errorf("blog has a band %q over its own tile", g.Head.Label.Text)
	}
	if g := groupByKey(groups, "boards2"); g.Head == nil || g.Head.Label.Text != "boards2" {
		t.Errorf("boards2 band = %+v, want its name without a count of one", g.Head)
	}
}
