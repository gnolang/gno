package bptree

// Options configures tree behavior.
type Options struct {
	Sync           bool   // fsync writes
	InitialVersion uint64 // first version number
	FlushThreshold int    // batch flush size in bytes
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
