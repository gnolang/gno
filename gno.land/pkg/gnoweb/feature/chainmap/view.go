package chainmap

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
)

// MapViewType tags the body view. Local because nothing in components renders
// this feature.
const MapViewType components.ViewType = "map-view"

// activityWait is how long a map page waits for the activity aggregate before
// rendering without it. The refresh it started carries on for the next load.
const activityWait = 3 * time.Second

// Listing is a directory listing: the one the list view renders too.
type Listing struct {
	// Path is the listing's root, gnoweb-relative, e.g. "/r/gnoland".
	Path string
	// Paths are the gnoweb-relative package paths below Path.
	Paths []string
	// Truncated is set when there were more paths than were listed.
	Truncated bool
}

// Activity states the legend tells apart. A map never shows a count it does
// not have, so "not counted yet" and "could not count" are said, not drawn.
const (
	ActivityShown       = "shown"
	ActivityPending     = "pending"
	ActivityUnavailable = "unavailable"
	// ActivityNotCalled is a map of pure packages, which nothing calls.
	ActivityNotCalled = "not-called"
)

// MapData is the render payload for templates/map.html.
type MapData struct {
	Path      string
	Count     int
	Truncated bool
	Groups    []*Group

	// Activity is empty when no indexer is configured: the legend then says
	// nothing about activity at all.
	Activity string
	// Window is set when Activity is ActivityShown.
	Window *Activity

	// Indexer is the provenance footer, set whenever the indexer answered.
	Indexer *components.IndexerStatus
}

// MapView builds the map of a listing.
func (h *Handler) MapView(ctx context.Context, l Listing) *components.View {
	root := strings.TrimSuffix(l.Path, "/")
	data := MapData{
		Path:      root + "/",
		Count:     len(l.Paths),
		Truncated: l.Truncated,
		Groups:    buildGroups(root+"/", l.Paths),
	}
	layout(data.Groups)

	if h.activity != nil {
		h.addActivity(ctx, &data)
	}
	for _, g := range data.Groups {
		for i := range g.Tiles {
			g.Tiles[i].Title = tileTitle(g.Tiles[i], data)
		}
	}
	return &components.View{Type: MapViewType, Component: &pageComponent{name: "renderMap", data: data}}
}

// addActivity colours the tiles with the window's calls, or says why it
// cannot.
func (h *Handler) addActivity(ctx context.Context, data *MapData) {
	if !strings.HasPrefix(data.Path, "/r/") {
		data.Activity = ActivityNotCalled
		return
	}
	wctx, cancel := context.WithTimeout(ctx, activityWait)
	defer cancel()
	a, err := h.activity.get(wctx, "activity", func(ctx context.Context) (*Activity, error) {
		return computeActivity(ctx, h.deps.Indexer)
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

	data.Activity, data.Window = ActivityShown, a
	data.Indexer = &components.IndexerStatus{URL: h.deps.Indexer.URL(), LastBlock: a.To}

	// The scale is the busiest realm on this map, so a zoomed-in map still
	// tells its own realms apart.
	busiest := 0
	for _, g := range data.Groups {
		for i := range g.Tiles {
			t := &g.Tiles[i]
			full := h.deps.Domain + t.Path
			t.Calls, t.Callers = a.Calls[full], a.Callers[full]
			busiest = max(busiest, t.Calls)
		}
	}
	for _, g := range data.Groups {
		for i := range g.Tiles {
			t := &g.Tiles[i]
			// Partial: a zero may be calls in the part that went unread.
			t.Unknown = a.Partial && t.Calls == 0
			t.ShadeClass = shadeClasses[level(t.Calls, busiest)]
			if t.Unknown {
				t.ShadeClass = unknownShadeClass
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

// tileTitle is the tile's tooltip. It repeats the label in full, since a
// small tile truncates it.
func tileTitle(t Tile, data MapData) string {
	switch {
	case data.Activity != ActivityShown:
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
