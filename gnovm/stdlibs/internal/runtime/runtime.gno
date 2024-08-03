// Package runtime contains some runtime information, which is used by packages
// like "chain" which cannot import "runtime" to avoid circular dependencies.
package runtime

// IsTesting returns whether the caller's caller is a function within a
// _test.gno or a _filetest.gno file.
func IsTesting() bool // injected

// IsRuntime returns whether the caller's caller is the "runtime" package.
func IsRuntime() bool // injected
