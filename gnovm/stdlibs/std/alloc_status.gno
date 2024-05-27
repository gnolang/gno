package std

/*
	USAGE:

	pseudo code:
	a.gno
	a := std.CreateallocTracker()
	a.Register("thisFunc")
	func()...
	a.Finish("thisFunc")

*/

type AllocTracker interface {
	Register(fn string) *allocTracker
	Finish() *allocTracker
	GetUsage() int64
	GetUsageString() string
}

type allocTracker struct {
	Descriptor                          string
	IsRegistered                        bool
	IsFinished                          bool
	StartAlloc, FinishAlloc, AllocUsage int64
}

// Create an instance of alloc tracker
func CreateATInstance() *allocTracker {
	return &allocTracker{
		Descriptor:   "",
		IsFinished:   false,
		IsRegistered: false,
		StartAlloc:   0,
		FinishAlloc:  0,
		AllocUsage:   0,
	}
}

// Register to the tracker
func (a *allocTracker) Register(des string) *allocTracker {
	a.Descriptor = des
	a.IsRegistered = true
	a.IsFinished = false
	a.StartAlloc = GetCurrAllocatedMem()
	return a
}

// Finish calculates the AllocUsage
func (a *allocTracker) Finish() *allocTracker {
	a.IsFinished = true
	a.FinishAlloc = GetCurrAllocatedMem()
	a.AllocUsage = a.FinishAlloc - a.StartAlloc
	return a
}

// GetResult does Finish() and return usage
func (a *allocTracker) GetResult() int64 {
	a.Finish()
	return a.AllocUsage
}

func (a *allocTracker) GetAllocUsage() int64 {
	return a.AllocUsage
}

func (a *allocTracker) Validate() (bool, string) {
	if !a.IsRegistered {
		return false, "this tracker is not registered"
	}
	if a.AllocUsage < 0 {
		return false, "bad call Validate()"
	}
	return true, ""
}
