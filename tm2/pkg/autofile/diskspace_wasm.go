//go:build js && wasm
// +build js,wasm

package autofile

// availableDiskSpace is a no-op stub for wasm targets where filesystem
// space checking is not supported. It returns diskSpaceUnsupported and a
// nil error, which causes the caller to skip space-based checks.
func availableDiskSpace(_ string) (uint64, error) {
	return diskSpaceUnsupported, nil
}

// isErrNoSpace always returns false on wasm targets.
func isErrNoSpace(_ error) bool {
	return false
}
