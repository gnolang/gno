//go:build windows
// +build windows

package autofile

// availableDiskSpace is a stub for Windows targets where syscall.Statfs is not
// available. It returns diskSpaceUnsupported and a nil error, which causes the
// caller to skip space-based checks.
// TODO: implement using GetDiskFreeSpaceEx via golang.org/x/sys/windows.
func availableDiskSpace(_ string) (uint64, error) {
	return diskSpaceUnsupported, nil
}

// isErrNoSpace always returns false on Windows targets.
// TODO: implement using Windows error codes.
func isErrNoSpace(_ error) bool {
	return false
}
