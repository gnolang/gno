//go:build !debug

package gnodebug

// Debug is set to true if the "debug" build tag is passed. It can be used to
// guard statements which should only be compiled to make detailed logging
// possible.
const Debug DebugType = false

// Inlineable stubs for no-debug.
// We write the documentation here as most people in editors (and godoc) will
// not use the build tag.

// Printf formats the given string with the arguments and passes it to [Output].
// Callers should specify a flagName argument, which is the name of the flag
// that should be passed to GNODEBUG to enable printing this specific debug log.
// fmt does not need a trailing newline; if not present, it is automatically
// added.
func (DebugType) Printf(flagName, fmt string, args ...any) {}

// Get returns the value in the GNODEBUG [DebugFlags] of the given flag.
func (DebugType) Get(flagName string) string { return "" }

// Enabled determines whether the given flag in the GNODEBUG [DebugFlags] is set
// to "1".
func (DebugType) Enabled(flagName string) bool { return false }

// Set sets the GNODEBUG [DebugFlags] to the given value.
func (DebugType) Set(flagName, val string) {}
