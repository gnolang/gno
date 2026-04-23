package bptree

// Options configures tree behavior.
//
// The struct is intentionally narrow: only fields that the tree
// actually consults live here. Previous incarnations carried
// FlushThreshold and AsyncPruning knobs that no code read; leaving
// them defined invited callers to tune values that had no effect. If
// those features land, they should re-appear here wired to their
// implementations.
type Options struct {
	Sync           bool   // fsync writes
	InitialVersion uint64 // first version number

	// FastNodeCacheSize sets the capacity (in entries) of the latest-view
	// key→value cache that skips tree descent on GetHit/Has. Zero leaves
	// the default (DefaultFastNodeCacheSize) in place; a negative value
	// disables the cache entirely. Only covers the MutableTree's current
	// working-view reads — GetImmutable snapshots do not consult it.
	FastNodeCacheSize int

	// InlineValueThreshold is the byte-length cutoff at which a value
	// stored via Set is written inline into the leaf rather than via an
	// external ValueKey indirection. The named InlineThreshold type
	// makes intent visible at call sites — see its doc for the value
	// semantics (InlineDisabled, DefaultInlineValueThreshold, or an
	// explicit byte count up to MaxInlineValueThreshold).
	//
	// The configured threshold is silently clamped to
	// MaxInlineValueThreshold at construction so that a misconfigured
	// option cannot produce a leaf whose serialised form exceeds the
	// reader's per-leaf budget (maxLeafReadBytes = 256 KiB). The clamp
	// applies to direct struct-literal writers as well — see
	// resolveInlineThreshold.
	InlineValueThreshold InlineThreshold
}

// Option is a functional option for tree construction.
type Option func(*Options)

func DefaultOptions() Options {
	return Options{}
}

func SyncOption(sync bool) Option {
	return func(o *Options) { o.Sync = sync }
}

func InitialVersionOption(iv uint64) Option {
	return func(o *Options) { o.InitialVersion = iv }
}

// FastNodeCacheSizeOption configures the latest-view fast-node cache
// size. Pass a negative value to disable the cache; zero leaves the
// default in place.
func FastNodeCacheSizeOption(n int) Option {
	return func(o *Options) { o.FastNodeCacheSize = n }
}

// DefaultFastNodeCacheSize is the number of entries held by the
// MutableTree's fast-node cache when FastNodeCacheSize is unset. Sized
// to comfortably cover a hot working set of keys without dominating
// heap under typical gno.land workloads (avg value < 256 B).
const DefaultFastNodeCacheSize = 10000

// InlineValueThresholdOption configures the cutoff at which values are
// stored inline in the leaf (<= threshold) vs via a ValueKey indirection
// (> threshold). Pass DefaultInlineValueThreshold (or another positive
// byte count) to enable inlining at that cutoff; pass InlineDisabled
// to opt out entirely. Values above MaxInlineValueThreshold are
// silently clamped so a full inline-value leaf cannot overflow the
// reader's per-leaf budget; see MaxInlineValueThreshold for the
// rationale.
func InlineValueThresholdOption(n InlineThreshold) Option {
	if n > MaxInlineValueThreshold {
		n = MaxInlineValueThreshold
	}
	return func(o *Options) { o.InlineValueThreshold = n }
}
