package ifaceext

type Str interface{ S() string }
type Sec interface{ sec() int } // unexported method: identity is pkgpath-qualified
