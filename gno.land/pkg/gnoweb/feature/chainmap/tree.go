package chainmap

import (
	"cmp"
	"slices"
	"strings"
)

// mapAspect is the map's width over its height. The stylesheet draws the map
// at this ratio, so the squares the layout computes stay square on screen.
const mapAspect = 1.6

// hueClasses are the group colours, written out whole so the stylesheet
// purge, which keeps only class names it finds in the sources, sees them.
var hueClasses = [...]string{
	"b-map__group--hue0", "b-map__group--hue1", "b-map__group--hue2",
	"b-map__group--hue3", "b-map__group--hue4", "b-map__group--hue5",
}

// Group is one box of the map: the packages sharing the next path segment.
type Group struct {
	Key   string
	Tiles []Tile
	Box   Box
	// HueClass tells neighbouring groups apart; decoration only.
	HueClass string

	// Header is false for a group holding nothing but the package named by its
	// key, whose tile already says everything a header would, and for a group
	// too small to carry one.
	Header bool

	// ZoomURL is the map of what lies below the key. Empty when nothing does:
	// a zoom there would list no package and answer not found.
	ZoomURL string

	// Calls sums the tiles' calls; meaningful only when the map has activity.
	Calls int
}

// Tile is one package.
type Tile struct {
	Path  string // gnoweb-relative, the package's page
	Label string
	Box   Box
	// Title is the tooltip and accessible name: the full path, which a small
	// tile clips or omits, and the activity when there is any.
	Title string
	// ShowLabel is false for a tile too small to show a readable name.
	ShowLabel bool

	Calls, Callers int
	// Unknown marks a tile whose count the indexer could not settle.
	Unknown bool
	// ShadeClass is the activity colour, empty when the map shows none.
	ShadeClass string
}

// buildGroups groups the listed paths by their first segment below prefix and
// lays the result out. prefix ends with a slash, paths are gnoweb-relative
// like prefix. A path outside prefix is dropped rather than misfiled.
func buildGroups(prefix string, paths []string) []*Group {
	byKey := make(map[string]*Group)
	var groups []*Group
	for _, p := range paths {
		rel, ok := strings.CutPrefix(p, prefix)
		if !ok || rel == "" {
			continue
		}
		key, rest, nested := strings.Cut(rel, "/")
		g := byKey[key]
		if g == nil {
			g = &Group{Key: key}
			byKey[key] = g
			groups = append(groups, g)
		}
		label := key
		if nested {
			label = rest
			g.ZoomURL = prefix + key + "/$map"
		}
		g.Tiles = append(g.Tiles, Tile{Path: p, Label: label})
	}

	// Largest first, as squarify requires; ties by name so a reload draws the
	// same map.
	slices.SortFunc(groups, func(a, b *Group) int {
		if c := cmp.Compare(len(b.Tiles), len(a.Tiles)); c != 0 {
			return c
		}
		return cmp.Compare(a.Key, b.Key)
	})
	for i, g := range groups {
		slices.SortFunc(g.Tiles, func(a, b Tile) int { return cmp.Compare(a.Label, b.Label) })
		g.HueClass = hueClasses[i%len(hueClasses)]
		g.Header = len(g.Tiles) > 1 || g.ZoomURL != ""
	}
	return groups
}

// Below these sizes, in percent of the map's width and height, a name cannot
// be read at the stylesheet's font size on a desktop-wide map. The tile still
// carries it as its accessible name and tooltip.
const (
	minLabelWidth  = 4.0
	minLabelHeight = 3.0
	minHeadWidth   = 6.0
	minHeadHeight  = 5.0
)

// layout places groups on the map and tiles in their group. Every package
// weighs the same: the map says how many packages there are, which is the one
// size the chain answers for a whole listing at once.
func layout(groups []*Group) {
	frame := rect{0, 0, mapAspect, 1}
	weights := make([]float64, len(groups))
	for i, g := range groups {
		weights[i] = float64(len(g.Tiles))
	}
	for i, gr := range squarify(weights, frame) {
		g := groups[i]
		g.Box = percent(gr, frame)

		ones := make([]float64, len(g.Tiles))
		for j := range ones {
			ones[j] = 1
		}
		for j, tr := range squarify(ones, gr) {
			t := &g.Tiles[j]
			t.Box = percent(tr, gr)
			onMap := percent(tr, frame)
			t.ShowLabel = onMap.Width >= minLabelWidth && onMap.Height >= minLabelHeight
		}
		g.Header = g.Header && g.Box.Width >= minHeadWidth && g.Box.Height >= minHeadHeight
	}
}
