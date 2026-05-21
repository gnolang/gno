// Package v3acallee is a callee realm for v3a identity-reflection tests.
// It exports functions that report runtime.Caller() and runtime.Self().
// None of its exported functions take cur realm, rlm realm, or any realm
// parameter — callers invoke them with plain arguments, no cross keyword.
package v3acallee

import "chain/runtime"

var (
	owner string
	value string
)

func init() {
	value = "initial"
}

// WhoAmI returns the executing realm's pkgPath via runtime.Self().
func WhoAmI() string {
	return runtime.Self().PkgPath()
}

// WhoCalledMe returns the calling realm's pkgPath via runtime.Caller().
func WhoCalledMe() string {
	return runtime.Caller().PkgPath()
}

// SetOwner sets the authorized caller. Unguarded — callers use this
// to configure who is allowed to call SetValue.
func SetOwner(pkgPath string) {
	owner = pkgPath
}

// SetValue sets the stored value. Guarded: only the configured owner
// (checked via runtime.Caller().PkgPath()) may set the value.
func SetValue(v string) {
	caller := runtime.Caller().PkgPath()
	if caller != owner {
		panic("unauthorized: " + caller)
	}
	value = v
}

// GetValue returns the current stored value.
func GetValue() string {
	return value
}
