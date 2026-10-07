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
	// Metric is what the tiles are shaded by; the zero value is calls.
	Metric Metric
}

// Metric is what a map's tiles are shaded by, chosen with ?color= on $map.
type Metric string

const (
	MetricCalls Metric = "calls"
	MetricGas   Metric = "gas"
)

// ParseMetric reads the color= value of a map URL. Anything but "gas" is
// calls, so an unknown value draws the default map rather than an error.
func ParseMetric(s string) Metric {
	if s == string(MetricGas) {
		return MetricGas
	}
	return MetricCalls
}

// query is what a map link appends after $map to keep this metric.
func (m Metric) query() string {
	if m == MetricGas {
		return "&color=gas"
	}
	return ""
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

	// Busiest are the realms ranking highest on the metric, set with
	// activity.
	Busiest []Tile

	// Metric is what the tiles are shaded by, and Query what a map link
	// appends after $map to keep it.
	Metric Metric
	Query  string
}

// Gas reports whether the tiles are shaded by gas, for the template.
func (d *MapData) Gas() bool { return d.Metric == MetricGas }

// SwitchURL is this map under the other metric, for the key's switch.
func (d *MapData) SwitchURL() string {
	if d.Gas() {
		return d.Root + "$map"
	}
	return d.Root + "$map" + MetricGas.query()
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

// ScaleStep is one colour of the key: a shade and the range it stands for.
type ScaleStep struct {
	Class, Range string
}

// Map draws a listing: the figure for the listing's body and its key for the
// listing's rail. The page around them is the directory view's own.
func (h *Handler) Map(ctx context.Context, l Listing) components.MapParts {
	root := strings.TrimSuffix(l.Path, "/") + "/"
	metric := l.Metric
	if metric != MetricGas {
		metric = MetricCalls
	}
	data := &MapData{Root: root, Up: l.Up, Groups: buildGroups(root, l.Paths), Metric: metric, Query: metric.query()}
	layout(data.Groups)
	data.eachHead(func(h *Head) { h.Query = data.Query })

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
	a, err := h.loadActivity(ctx, activityWait)
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
	busiest, lightest := 0, 0
	data.eachTile(func(t *Tile) {
		full := h.deps.Domain + t.Path
		t.Calls, t.Callers, t.Gas = a.Calls[full], a.Callers[full], a.Gas[full]
		t.Amount = data.Metric.amount(t)
		if v := t.value(data.Metric); v > 0 {
			busiest = max(busiest, v)
			if lightest == 0 || v < lightest {
				lightest = v
			}
		}
	})
	data.eachTile(func(t *Tile) {
		// Partial: a zero may be calls in the part that went unread.
		t.Unknown = a.Partial && t.Calls == 0
		l := level(t.value(data.Metric), busiest)
		if data.Gas() {
			l = gasLevel(t.value(MetricGas), lightest, busiest)
		}
		t.ShadeClass = shadeClasses[l]
		if t.Unknown {
			t.ShadeClass = unknownShadeClass
		}
	})
	if data.Gas() {
		data.Scale = gasScale(lightest, busiest, a.Partial)
	} else {
		data.Scale = scale(busiest, a.Partial)
	}
}

// loadActivity returns the activity aggregate, waiting up to wait for a
// refresh. Before any aggregate exists it does not wait: the first scan takes
// seconds, and a page would only wait to say it is pending. The scan carries
// on for the next reader either way.
func (h *Handler) loadActivity(ctx context.Context, wait time.Duration) (*Activity, error) {
	if !h.activity.has(activityKey) {
		wait = coldWait
	}
	wctx, cancel := context.WithTimeout(ctx, wait)
	defer cancel()
	return h.activity.get(wctx, activityKey, func(ctx context.Context) (*Activity, error) {
		return computeActivity(ctx, h.deps.Indexer, h.closedBands)
	})
}

const (
	// activityKey is the aggregate's one cache key.
	activityKey = "activity"
	// coldWait is how long a reader waits when no aggregate exists yet:
	// long enough to start the scan, not to wait for it.
	coldWait = 10 * time.Millisecond
)

// gasLevel maps gas, in thousands, to a colour step on a log scale from the
// lightest realm on the map to the heaviest. Gas sits in a narrow band (a call
// costs hundreds of thousands), so a scale from zero would put every realm in
// the top shades.
func gasLevel(v, lightest, heaviest int) int {
	switch {
	case v <= 0:
		return 0
	case heaviest <= lightest:
		return len(shadeClasses) - 1
	}
	f := math.Log(float64(v)/float64(lightest)) / math.Log(float64(heaviest)/float64(lightest))
	return 1 + min(len(shadeClasses)-2, int(f*float64(len(shadeClasses)-1)))
}

// gasScale is the key's colour steps under gas: each shade with its upper
// bound, since rounded gas ranges would print the same number where they
// meet. A bound that reads like the previous one is skipped.
func gasScale(lightest, heaviest int, partial bool) []ScaleStep {
	steps := []ScaleStep{{Class: shadeClasses[0], Range: "0"}}
	if partial {
		steps[0].Range = "0 (counted)"
	}
	last := len(shadeClasses) - 1
	prev := ""
	for l := 1; l <= last && heaviest > 0; l++ {
		bound := float64(heaviest)
		if l < last && heaviest > lightest {
			bound = float64(lightest) * math.Pow(float64(heaviest)/float64(lightest), float64(l)/float64(last))
		}
		label := "≤ " + components.FormatGas(int64(bound)*1_000)
		if label == prev {
			continue
		}
		prev = label
		steps = append(steps, ScaleStep{Class: shadeClasses[l], Range: label})
	}
	if partial {
		steps = append(steps, ScaleStep{Class: unknownShadeClass, Range: "unknown"})
	}
	return steps
}

// value is what t is shaded by under m: its calls, or its gas in thousands,
// rounded up so that any gas at all draws a shade. Thousands keep a map whose
// heaviest realm used under a million apart; ranking and labels use the
// measured gas (rank, amount).
func (t *Tile) value(m Metric) int {
	if m == MetricGas {
		return int((t.Gas + 999) / 1_000)
	}
	return t.Calls
}

// rank is what t is ordered by under m, as measured.
func (t *Tile) rank(m Metric) int64 {
	if m == MetricGas {
		return t.Gas
	}
	return int64(t.Calls)
}

// amount writes t's measure under m for the busiest list.
func (m Metric) amount(t *Tile) string {
	if m == MetricGas {
		return components.FormatGas(t.Gas)
	}
	return strconv.Itoa(t.Calls)
}

// scale is the key's colour steps for a map whose busiest realm has busiest
// calls: each shade with the calls it stands for. A shade no count reaches is
// skipped; unknown comes last when the window is partial.
func scale(busiest int, partial bool) []ScaleStep {
	steps := []ScaleStep{{Class: shadeClasses[0], Range: "0"}}
	if partial {
		steps[0].Range = "0 (counted)"
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
			steps = append(steps, ScaleStep{Class: shadeClasses[l], Range: strconv.Itoa(lo)})
		default:
			steps = append(steps, ScaleStep{Class: shadeClasses[l], Range: strconv.Itoa(lo) + "–" + strconv.Itoa(hi)})
		}
	}
	if partial {
		steps = append(steps, ScaleStep{Class: unknownShadeClass, Range: "unknown"})
	}
	return steps
}

// busiest returns up to n tiles ranking highest on the map's metric, highest
// first; none at zero.
func busiest(data *MapData, n int) []Tile {
	var top []Tile
	data.eachTile(func(t *Tile) {
		if t.rank(data.Metric) > 0 {
			top = append(top, *t)
		}
	})
	slices.SortFunc(top, func(a, b Tile) int {
		if c := cmp.Compare(b.rank(data.Metric), a.rank(data.Metric)); c != 0 {
			return c
		}
		return cmp.Compare(a.Path, b.Path)
	})
	return top[:min(n, len(top))]
}

// eachHead calls fn on every name band of the map.
func (d *MapData) eachHead(fn func(*Head)) {
	for _, g := range d.Groups {
		if g.Head != nil {
			fn(g.Head)
		}
		for _, s := range g.Subgroups {
			if s.Head != nil {
				fn(s.Head)
			}
		}
	}
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
	if data.Gas() {
		return fmt.Sprintf("%s · %s gas over %s", t.Path, components.FormatGas(t.Gas), calls)
	}
	return fmt.Sprintf("%s · %s by %s", t.Path, calls, plural(t.Callers, "account"))
}

func plural(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return fmt.Sprintf("%d %ss", n, noun)
}
