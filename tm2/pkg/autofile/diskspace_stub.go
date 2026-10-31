//go:build windows || (js && wasm)

package autofile

// availableDiskSpace is a stub for platforms where filesystem space checking
// is not yet supported. It returns diskSpaceUnsupported and a nil error,
// which causes the caller to skip space-based checks.
//
// TODO(windows): implement using GetDiskFreeSpaceEx via golang.org/x/sys/windows.
func availableDiskSpace(_ string) (uint64, error) {
	return diskSpaceUnsupported, nil
}

// isErrNoSpace always returns false on unsupported platforms.
//
// TODO(windows): implement using Windows error codes.
func isErrNoSpace(_ error) bool {
	return false
}
