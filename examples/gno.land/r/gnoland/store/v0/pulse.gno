package store

const (
	pulseEpochs   = 30 // epochs kept by the global counters
	activitySize  = 20 // events kept: room for those of listings hidden or no longer apps
	activityShown = 5  // events served: gnoweb shows the newest
)

// milestones are the star counts announced in the activity feed.
var milestones = []int{10, 50, 100}

// Global per-epoch counters behind the home "pulse".
var (
	pulseListed = newEpochRing(pulseEpochs)
	pulseStars  = newEpochRing(pulseEpochs)
)

// event points at its listing, not at a slug: Hide releases slugs, and an
// event must never follow a slug to another listing.
type event struct {
	Kind    string // listed, updated or milestone
	Listing *listing
	Stars   int // ranked stars at the time
}

// activity holds the last activitySize events in a fixed ring.
var activity eventRing

type eventRing struct {
	events [activitySize]event
	head   int // index of the newest event
	n      int
}

func (r *eventRing) push(e event) {
	r.head = (r.head + 1) % activitySize
	r.events[r.head] = e
	if r.n < activitySize {
		r.n++
	}
}

// newestFirst returns the events, newest first.
func (r *eventRing) newestFirst() []event {
	out := make([]event, r.n)
	for i := range out {
		out[i] = r.events[(r.head-i+activitySize)%activitySize]
	}
	return out
}

// track records a listing event in the activity feed and the pulse
// counters. It is the single gate of both: hidden and unconfirmed listings
// announce nothing.
func track(kind string, l *listing, h int64) {
	if !l.indexed() {
		return
	}
	if kind == "listed" {
		pulseListed.add(h/epochBlocks, 1)
	}
	activity.push(event{kind, l, l.rankedStars.Size()})
}

// epochRing counts events per epoch over the last len(counts) epochs, in
// constant space: slot epoch%len is reused once its epoch is too old.
type epochRing struct {
	epochs []int64
	counts []int
}

func newEpochRing(n int) *epochRing {
	return &epochRing{make([]int64, n), make([]int, n)}
}

// add adds delta to epoch's count. A negative delta for an epoch that has
// left the ring is dropped: that epoch no longer counts anywhere.
func (r *epochRing) add(epoch int64, delta int) {
	i := int(epoch % int64(len(r.counts)))
	if r.epochs[i] != epoch {
		if delta < 0 {
			return
		}
		r.epochs[i], r.counts[i] = epoch, 0
	}
	r.counts[i] += delta
}

// sum returns the total of the window epochs ending at epoch.
func (r *epochRing) sum(epoch int64, window int) int {
	n := 0
	for i, e := range r.epochs {
		if e <= epoch && epoch-e < int64(window) {
			n += r.counts[i]
		}
	}
	return n
}
