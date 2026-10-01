package bptree

// Options configures tree behavior.
type Options struct {
	Sync           bool   // fsync writes
	InitialVersion uint64 // first version number
	FlushThreshold int    // batch flush size in bytes
	FastIndex      bool   // maintain the latest-version fast index (read accelerator)
}

// Option is a functional option for tree construction.
type Option func(*Options)

func DefaultOptions() Options {
	return Options{
		FlushThreshold: 100 * 1024, // 100KB
	}
}

func SyncOption(sync bool) Option {
	return func(o *Options) { o.Sync = sync }
}

func InitialVersionOption(iv uint64) Option {
	return func(o *Options) { o.InitialVersion = iv }
}

func FlushThresholdOption(ft int) Option {
	return func(o *Options) { o.FlushThreshold = ft }
}

// FastIndexOption enables the optional latest-version fast index: a flat
// user-key → version‖value map that accelerates point Gets of present keys
// against committed state (1 read instead of a full tree descent + value
// read) — on committed snapshots and on the clean working tree alike. It is
// an unauthenticated read accelerator — not part of the Merkle root — so it
// can be toggled per-node without affecting the app hash. Default off.
func FastIndexOption(b bool) Option {
	return func(o *Options) { o.FastIndex = b }
}
