package chainmap

import (
	"cmp"
	"math"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"
)

// The map is drawn in a fixed coordinate space the SVG scales to its width,
// so text sizes are known in the same units as the boxes and every label is
// fitted here, to the character. The template draws the viewBox and every
// font size from these values, so nothing else has to agree with them.
const (
	mapWidth  = 960.0
	mapHeight = 600.0
)

// Label metrics, in map units. Labels are drawn in the monospace face, whose
// advance is 0.6 em. The map scales down with the content column, so these
// sit a step above body text sizes to stay readable there.
const (
	headBand    = 26.0 // a group's name band
	headFont    = 14.0
	subBand     = 22.0 // a subgroup's name band
	subFont     = 12.5
	tileFont    = 12.5
	labelPad    = 6.0
	monoAdvance = 0.6

	// minLabelChars is the shortest label worth drawing: below it a name is
	// an ellipsis and a letter, which reads as a rendering fault.
	minLabelChars = 4
)

// subgroupMin is how many packages a group needs before it is split by the
// next path segment, and bandMin how many a subgroup needs to keep a band of
// its own. Smaller subgroups are pooled with the group's loose packages, named
// by their path below the group: a band over one or two tiles only repeats
// their names.
const (
	subgroupMin = 6
	bandMin     = 3
)

// hueClasses are the group colours, written out whole so the stylesheet
// purge, which keeps only class names it finds in the sources, sees them.
var hueClasses = [...]string{
	"b-map__group--hue0", "b-map__group--hue1", "b-map__group--hue2", "b-map__group--hue3",
}

// Rect is a box in map units.
type Rect struct{ X, Y, W, H float64 }

// Label is a name fitted to the box it sits in, where to draw it, and the
// font size it was fitted to.
type Label struct {
	Text       string
	X, Y, Size float64
}

// Head is the name band over the top of a group or subgroup: a link to that
// box's own map when it has one.
type Head struct {
	// Class is the band's block class: one partial draws both kinds.
	Class string
	Band  Rect
	Label Label
	// ZoomURL is the box's own map, and ZoomPath what it maps; both empty
	// when the box has no map of its own.
	ZoomURL, ZoomPath string
}

// Group is the packages sharing the first path segment below the map's root.
type Group struct {
	Key string
	// root is the path the key names, the map's root plus the key.
	root     string
	Count    int
	Rect     Rect
	HueClass string
	// Head is the name band; nil when the group is too small to carry one.
	Head *Head
	// ZoomURL is the map of what lies below the key; empty when the key is
	// itself a package, whose map would carry a List tab opening the package.
	ZoomURL   string
	Subgroups []*Subgroup
}

// Subgroup is the packages of a group sharing its next path segment. A group
// too small to split holds one subgroup with no key.
type Subgroup struct {
	Key string
	// root is the path the key names; the group's root when Key is empty.
	root    string
	Rect    Rect
	Head    *Head
	ZoomURL string
	Tiles   []Tile
}

// Tile is one package.
type Tile struct {
	Path  string // gnoweb-relative, the package's page
	Rect  Rect
	Label *Label
	// Title is the tooltip and accessible name: the full path, which the
	// label shortens, and the activity when there is any.
	Title string

	Calls, Callers int
	// Unknown marks a tile whose count the indexer could not settle.
	Unknown bool
	// ShadeClass is the activity colour, empty when the map shows none.
	ShadeClass string
}

// buildGroups groups the listed paths below prefix, which ends with a slash,
// by their first and then their second path segment. A path outside prefix
// is dropped rather than misfiled.
func buildGroups(prefix string, paths []string) []*Group {
	byKey := make(map[string][]string)
	for _, p := range paths {
		rel, ok := strings.CutPrefix(p, prefix)
		if !ok || rel == "" {
			continue
		}
		key, _, _ := strings.Cut(rel, "/")
		byKey[key] = append(byKey[key], p)
	}

	groups := make([]*Group, 0, len(byKey))
	for key, members := range byKey {
		slices.Sort(members)
		root := prefix + key
		g := &Group{Key: key, root: root, Count: len(members)}
		if !slices.Contains(members, root) {
			g.ZoomURL = root + "/$map"
		}
		g.Subgroups = subgroups(root, members, len(members) >= subgroupMin)
		groups = append(groups, g)
	}
	// Largest first, as squarify requires; ties by name so a reload draws the
	// same map.
	slices.SortFunc(groups, func(a, b *Group) int {
		if c := cmp.Compare(b.Count, a.Count); c != 0 {
			return c
		}
		return cmp.Compare(a.Key, b.Key)
	})
	for i, g := range groups {
		g.HueClass = hueClasses[i%len(hueClasses)]
	}
	return groups
}

// subgroups splits a group's members, all at or below root, by their next
// segment, or keeps them together when split is false. The package at root
// itself, if any, sits in the subgroup with no key.
func subgroups(root string, members []string, split bool) []*Subgroup {
	// Count each next segment first: only one with bandMin packages or more
	// becomes a subgroup of its own.
	count := make(map[string]int)
	if split {
		for _, p := range members {
			if k, _, _ := strings.Cut(strings.TrimPrefix(p, root+"/"), "/"); p != root {
				count[k]++
			}
		}
	}

	byKey := make(map[string]*Subgroup)
	var subs []*Subgroup
	for _, p := range members {
		key := ""
		if k, _, _ := strings.Cut(strings.TrimPrefix(p, root+"/"), "/"); p != root && count[k] >= bandMin {
			key = k
		}
		s := byKey[key]
		if s == nil {
			s = &Subgroup{Key: key, root: root}
			if key != "" {
				s.root += "/" + key
			}
			byKey[key] = s
			subs = append(subs, s)
		}
		s.Tiles = append(s.Tiles, Tile{Path: p})
	}
	for _, s := range subs {
		if s.Key != "" && !slices.ContainsFunc(s.Tiles, func(t Tile) bool { return t.Path == s.root }) {
			s.ZoomURL = s.root + "/$map"
		}
	}
	slices.SortFunc(subs, func(a, b *Subgroup) int {
		if c := cmp.Compare(len(b.Tiles), len(a.Tiles)); c != 0 {
			return c
		}
		return cmp.Compare(a.Key, b.Key)
	})
	return subs
}

// tileName is what a tile at p is called under the band naming base: its path
// below it, or its last segment when it is the package the band names.
func tileName(p, base string) string {
	if rest, ok := strings.CutPrefix(p, base); ok {
		return rest
	}
	return baseName(p)
}

// baseName is the last segment of a path.
func baseName(p string) string { return p[strings.LastIndex(p, "/")+1:] }

// layout places groups on the map, subgroups in their group and tiles in
// their subgroup, and fits every label. Every package weighs the same: the
// map says how many packages there are, the one size the chain answers for a
// whole listing at once.
func layout(groups []*Group) {
	weights := make([]float64, len(groups))
	for i, g := range groups {
		weights[i] = float64(g.Count)
	}
	for i, r := range squarify(weights, rect{0, 0, mapWidth, mapHeight}) {
		g := groups[i]
		g.Rect = toRect(r)
		inner := r
		// Tiles are named below the nearest band drawn over them: without the
		// group's, the key is part of their name, or nothing on the map says it.
		base := strings.TrimSuffix(g.root, g.Key)
		// A group that is one package at its own root is named by its tile;
		// a band would only say it twice.
		if g.Count > 1 || g.ZoomURL != "" {
			name := g.Key
			if g.Count > 1 {
				name += " · " + strconv.Itoa(g.Count)
			}
			if g.Head = newHead("b-map__head", name, g.ZoomURL, r, headBand, headFont); g.Head != nil {
				inner = rect{r.X, r.Y + headBand, r.W, r.H - headBand}
				base = g.root + "/"
			}
		}
		layoutSubgroups(g.Subgroups, inner, base)
	}
}

// layoutSubgroups places subgroups in r and their tiles in them. base is the
// path the group's band names, ending with a slash.
func layoutSubgroups(subs []*Subgroup, r rect, base string) {
	weights := make([]float64, len(subs))
	for i, s := range subs {
		weights[i] = float64(len(s.Tiles))
	}
	for i, sr := range squarify(weights, r) {
		s := subs[i]
		s.Rect = toRect(sr)
		inner, tileBase := sr, base
		if s.Key != "" {
			if s.Head = newHead("b-map__sub", s.Key, s.ZoomURL, sr, subBand, subFont); s.Head != nil {
				inner, tileBase = rect{sr.X, sr.Y + subBand, sr.W, sr.H - subBand}, s.root+"/"
			}
		}
		ones := make([]float64, len(s.Tiles))
		for j := range ones {
			ones[j] = 1
		}
		for j, tr := range squarify(ones, inner) {
			t := &s.Tiles[j]
			t.Rect = toRect(tr)
			if tr.H < tileFont+2*labelPad {
				continue
			}
			if text := fit(tileName(t.Path, tileBase), tr.W, tileFont); text != "" {
				t.Label = &Label{Text: text, X: round1(tr.X + labelPad), Y: round1(tr.Y + tileFont + labelPad - 2), Size: tileFont}
			}
		}
	}
}

// newHead bands the top of r with name, or returns nil when r is too short
// to give a band and room below it, or too narrow for a readable name.
func newHead(class, name, zoomURL string, r rect, band, font float64) *Head {
	if r.H < 2*band {
		return nil
	}
	text := fit(name, r.W, font)
	if text == "" {
		return nil
	}
	return &Head{
		Class:    class,
		Band:     toRect(rect{r.X, r.Y, r.W, band}),
		Label:    Label{Text: text, X: round1(r.X + labelPad), Y: round1(r.Y + band/2 + font*0.35), Size: font},
		ZoomURL:  zoomURL,
		ZoomPath: strings.TrimSuffix(zoomURL, "$map"),
	}
}

// fit returns text as it fits a box of width w at the given font size,
// shortened with an ellipsis when it does not, or "" when fewer than
// minLabelChars would.
func fit(text string, w, font float64) string {
	// The epsilon keeps a box exactly n characters wide from rounding to n-1.
	room := int(math.Floor((w-2*labelPad)/(font*monoAdvance) + 1e-9))
	if utf8.RuneCountInString(text) <= room {
		return text
	}
	if room < minLabelChars {
		return ""
	}
	return string([]rune(text)[:room-1]) + "…"
}

func round1(v float64) float64 { return math.Round(v*10) / 10 }

func toRect(r rect) Rect {
	return Rect{X: round1(r.X), Y: round1(r.Y), W: round1(r.W), H: round1(r.H)}
}
