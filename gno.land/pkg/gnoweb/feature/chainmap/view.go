package chainmap

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
)

// activityWait is how long a map page waits for the activity aggregate before
// rendering without it. The refresh it started carries on for the next load.
const activityWait = 3 * time.Second

// Listing is a directory listing: the one the list view renders too.
type Listing struct {
	// Path is the listing's root, gnoweb-relative, e.g. "/r/gnoland".
	Path string
	// Paths are the gnoweb-relative package paths below Path.
	Paths []string
	// Up is the listing one level up, ending with a slash: the way back out
	// of a zoom. Empty when there is none, at a kind's root such as /r/ or
	// where the path above is a package, which has no listing to map.
	Up string
}

// Why a map shows no activity, which the key says. A map never shows a count
// it does not have, so "not counted yet" and "could not count" are said, not
// drawn.
const (
	ActivityPending     = "pending"
	ActivityUnavailable = "unavailable"
	// ActivityNotCalled is a map of pure packages, which nothing calls.
	ActivityNotCalled = "not-called"
)

// MapData is the render payload for templates/map.html.
type MapData struct {
	// Root is the listing's root, ending with a slash.
	Root string
	// Up is the listing one level up; see Listing.Up.
	Up     string
	Groups []*Group

	// Window is the activity the tiles are coloured with, nil when there is
	// none; Activity then says why, or is empty when no indexer is
	// configured and the key says nothing about activity at all.
	Window   *Activity
	Activity string
	// Scale is the colour steps the tiles use, for the key.
	Scale []ScaleStep

	// Indexer is the provenance footer, set whenever the indexer answered.
	Indexer *components.IndexerStatus

	// Busiest are the most called realms on the map, set with activity.
	Busiest []Tile
}

// busiestCount is how many of the most called realms the key lists.
const busiestCount = 5

// MinPackages is the smallest listing worth a map: below it a map shows
// nothing the list does not, so the listing offers no map at all.
const MinPackages = 10

// ViewBox is the SVG's coordinate space, the one the layout fitted to.
func (*MapData) ViewBox() string {
	return fmt.Sprintf("0 0 %g %g", mapWidth, mapHeight)
}

// ScaleStep is one colour of the key: a shade and the calls it stands for.
type ScaleStep struct {
	Class, Calls string
}

// Map draws a listing: the figure for the listing's body and its key for the
// listing's rail. The page around them is the directory view's own.
func (h *Handler) Map(ctx context.Context, l Listing) components.MapParts {
	root := strings.TrimSuffix(l.Path, "/") + "/"
	data := &MapData{Root: root, Up: l.Up, Groups: buildGroups(root, l.Paths)}
	layout(data.Groups)

	if h.activity != nil {
		h.addActivity(ctx, data)
	}
	data.eachTile(func(t *Tile) { t.Title = tileTitle(t, data) })
	if data.Window != nil {
		data.Busiest = busiest(data, busiestCount)
	}
	return components.MapParts{
		Figure: &pageComponent{name: "chainmap/figure", data: data},
		Key:    &pageComponent{name: "chainmap/key", data: data},
	}
}

// addActivity colours the tiles with the window's calls, or says why it
// cannot.
func (h *Handler) addActivity(ctx context.Context, data *MapData) {
	if !strings.HasPrefix(data.Root, "/r/") {
		data.Activity = ActivityNotCalled
		return
	}
	wctx, cancel := context.WithTimeout(ctx, activityWait)
	defer cancel()
	a, err := h.activity.get(wctx, "activity", func(ctx context.Context) (*Activity, error) {
		return computeActivity(ctx, h.deps.Indexer, h.closedBands)
	})
	if err != nil {
		data.Activity = ActivityUnavailable
		if errors.Is(err, ErrPending) {
			data.Activity = ActivityPending
		} else {
			h.deps.Logger.Warn("map: activity unavailable", "error", err)
		}
		return
	}

	data.Window = a
	data.Indexer = &components.IndexerStatus{URL: h.deps.Indexer.URL(), LastBlock: a.To}

	// The scale is the busiest realm on this map, so a zoomed-in map still
	// tells its own realms apart.
	busiest := 0
	data.eachTile(func(t *Tile) {
		full := h.deps.Domain + t.Path
		t.Calls, t.Callers = a.Calls[full], a.Callers[full]
		busiest = max(busiest, t.Calls)
	})
	data.eachTile(func(t *Tile) {
		// Partial: a zero may be calls in the part that went unread.
		t.Unknown = a.Partial && t.Calls == 0
		t.ShadeClass = shadeClasses[level(t.Calls, busiest)]
		if t.Unknown {
			t.ShadeClass = unknownShadeClass
		}
	})
	data.Scale = scale(busiest, a.Partial)
}

// scale is the key's colour steps for a map whose busiest realm has busiest
// calls: each shade with the range of calls it stands for, skipping a shade
// no count reaches. Unknown comes last when the window is partial.
func scale(busiest int, partial bool) []ScaleStep {
	steps := []ScaleStep{{Class: shadeClasses[0], Calls: "0"}}
	if partial {
		steps[0].Calls = "0 (counted)"
	}
	// first[l] is the smallest count drawn at level l or above.
	last := len(shadeClasses)
	first := make([]int, last+1)
	first[last] = busiest + 1
	for l := 1; l < last; l++ {
		first[l] = firstAtLevel(l, busiest)
	}
	for l := 1; l < last; l++ {
		lo, hi := first[l], first[l+1]-1
		switch {
		case lo > hi:
			continue
		case lo == hi:
			steps = append(steps, ScaleStep{Class: shadeClasses[l], Calls: strconv.Itoa(lo)})
		default:
			steps = append(steps, ScaleStep{Class: shadeClasses[l], Calls: strconv.Itoa(lo) + "–" + strconv.Itoa(hi)})
		}
	}
	if partial {
		steps = append(steps, ScaleStep{Class: unknownShadeClass, Calls: "unknown"})
	}
	return steps
}

// busiest returns up to n tiles with the most calls, most first; none without
// calls.
func busiest(data *MapData, n int) []Tile {
	var top []Tile
	data.eachTile(func(t *Tile) {
		if t.Calls > 0 {
			top = append(top, *t)
		}
	})
	slices.SortFunc(top, func(a, b Tile) int {
		if c := cmp.Compare(b.Calls, a.Calls); c != 0 {
			return c
		}
		return cmp.Compare(a.Path, b.Path)
	})
	return top[:min(n, len(top))]
}

// eachTile calls fn on every tile of the map.
func (d *MapData) eachTile(fn func(*Tile)) {
	for _, g := range d.Groups {
		for _, s := range g.Subgroups {
			for i := range s.Tiles {
				fn(&s.Tiles[i])
			}
		}
	}
}

// shadeClasses are the activity steps, indexed by level, written out whole
// for the stylesheet purge. unknownShadeClass is never the colour of zero: a
// count the indexer could not settle is not a realm nobody called.
var shadeClasses = [...]string{
	"b-map__tile--l0", "b-map__tile--l1", "b-map__tile--l2",
	"b-map__tile--l3", "b-map__tile--l4",
}

const unknownShadeClass = "b-map__tile--unknown"

// level maps calls to a colour step on a log scale: 0 for none, then 1 to 4
// up to the busiest. Call counts span orders of magnitude, so a linear scale
// would paint every realm but the busiest the same.
func level(calls, busiest int) int {
	if calls <= 0 || busiest <= 0 {
		return 0
	}
	f := math.Log1p(float64(calls)) / math.Log1p(float64(busiest))
	return 1 + min(3, int(f*4))
}

// firstAtLevel is the smallest count from 1 that level draws at l or above,
// or busiest+1 when none does: level's inverse, nudged to absorb rounding.
func firstAtLevel(l, busiest int) int {
	c := int(math.Ceil(math.Expm1(float64(l-1) / 4 * math.Log1p(float64(busiest)))))
	c = max(c, 1)
	for c > 1 && level(c-1, busiest) >= l {
		c--
	}
	for c <= busiest && level(c, busiest) < l {
		c++
	}
	return c
}

// tileTitle is the tile's tooltip. It repeats the label in full, since a
// small tile truncates it.
func tileTitle(t *Tile, data *MapData) string {
	switch {
	case data.Window == nil:
		return t.Path
	case t.Unknown:
		return t.Path + " · calls unknown: part of the window could not be read"
	case t.Calls == 0:
		return t.Path + " · no calls"
	}
	calls := plural(t.Calls, "call")
	if data.Window.Partial {
		calls = "at least " + calls
	}
	return fmt.Sprintf("%s · %s by %s", t.Path, calls, plural(t.Callers, "account"))
}

func plural(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return fmt.Sprintf("%d %ss", n, noun)
}
