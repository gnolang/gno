package store

const pulseEpochs = 30 // epochs kept by the global counters

// Global per-epoch counters behind the home "pulse": raw counts, which
// gnoweb words. The store keeps no feed of events: gnoweb tells the latest
// from the listings' own dates, and an indexer follows the emitted events.
var (
	pulseListed = newEpochRing(pulseEpochs)
	pulseStars  = newEpochRing(pulseEpochs)
)

// epochRing counts events per epoch over the last len(counts) epochs, in
// constant space: slot epoch%len holds the count of the one epoch of the
// window (last-len, last] that maps to it, so a slot is reused once its
// epoch is too old. One slice, not one per field: each is a stored object.
type epochRing struct {
	counts []int
	last   int64 // newest epoch counted
}

func newEpochRing(n int) epochRing {
	return epochRing{counts: make([]int, n)}
}

// add adds delta to epoch's count. A negative delta for an epoch that has
// left the ring is dropped: that epoch no longer counts anywhere.
func (r *epochRing) add(epoch int64, delta int) {
	n := int64(len(r.counts))
	if delta > 0 && epoch <= r.last-n { // only a test moves the clock back: restart
		r.last = epoch - n
	}
	if epoch > r.last {
		from := r.last + 1
		if from < epoch-n+1 {
			from = epoch - n + 1
		}
		if from < 0 {
			from = 0
		}
		for e := from; e <= epoch; e++ { // the epochs entering the window start at 0
			r.counts[e%n] = 0
		}
		r.last = epoch
	}
	if epoch <= r.last-n {
		return
	}
	r.counts[epoch%n] += delta
}

// sum returns the total of the window epochs ending at epoch.
func (r *epochRing) sum(epoch int64, window int) int {
	total := 0
	for e := r.last; e >= 0 && r.last-e < int64(len(r.counts)); e-- {
		if e <= epoch && epoch-e < int64(window) {
			total += r.counts[e%int64(len(r.counts))]
		}
	}
	return total
}
