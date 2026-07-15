package ifaceext

type Str interface{ S() string }
type Sec interface{ sec() int } // unexported method: identity is pkgpath-qualified

type SecImpl struct{}

func (SecImpl) sec() int { return 7 }

// UseSec selects sec from within its origin package: legitimate in Go no
// matter how the caller spelled the interface (the anonymous embed here
// flattens sec with the same ifaceext-qualified identity).
func UseSec(x interface{ Sec }) int { return x.sec() }

// SecStr conflicts with Sec: same sec, different signature.
type SecStr interface{ sec() string }

// SecPtrImpl satisfies Sec only through a pointer.
type SecPtrImpl struct{}

func (*SecPtrImpl) sec() int { return 8 }
