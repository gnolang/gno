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
