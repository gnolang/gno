package ifaceext

type Str interface{ S() string }
type Sec interface{ sec() int } // unexported method: identity is pkgpath-qualified

type SecImpl struct{}

func (SecImpl) sec() int { return 7 }

// UseSec selects sec from within its origin package: legitimate in Go no
// matter how the caller spelled the interface (the anonymous embed here
// flattens sec with the same ifaceext-qualified identity).
func UseSec(x interface{ Sec }) int { return x.sec() }

// SecStr declares a same-spelled unexported method with a different signature,
// so two ifaceext interfaces can contribute a conflicting sec.
type SecStr interface{ sec() string }

// SecPtrImpl satisfies Sec only through a pointer.
type SecPtrImpl struct{}

func (*SecPtrImpl) sec() int { return 8 }

// SecStrImpl provides ifaceext's sec() string.
type SecStrImpl struct{}

func (SecStrImpl) sec() string { return "s" }

// FieldSec carries sec as a field rather than a method, so a lookup finds a
// non-function where Sec wants one.
type FieldSec struct{ sec int }
