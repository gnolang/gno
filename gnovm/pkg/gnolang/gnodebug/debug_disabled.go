//go:build !debug

package gnodebug

const Debug DebugType = false

// Inlineable stubs for no-debug.

func (DebugType) Printf(kind, fmt string, args ...any) {}
func (DebugType) Get(flagName string) string           { return "" }
func (DebugType) Enabled(flagName string) bool         { return false }
func (DebugType) Set(flagName, val string)             {}
