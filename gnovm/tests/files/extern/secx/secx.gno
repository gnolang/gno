package secx

type Sec interface{ sec() int }

type SecImpl struct{}

func (SecImpl) sec() int { return 7 }

type SecImpl2 struct{}

func (SecImpl2) sec() int { return 8 }

// Cnt is a value-receiver impl reading mutable state (snapshot axis).
type Cnt struct{ N int }

func (c Cnt) sec() int { return c.N }

// RecImpl records each sec run (observable through a deferred call).
type RecImpl struct{ Ran int }

func (r *RecImpl) sec() int { r.Ran++; return 7 }

func UseSec(x Sec) int { return x.sec() }

// MakeGetter forms the method value inside sec's origin package.
func MakeGetter(x Sec) func() int { return x.sec }
