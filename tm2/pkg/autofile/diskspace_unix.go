//go:build !js && !wasm && !windows
// +build !js,!wasm,!windows

package autofile

import (
	"errors"
	"syscall"
)

// availableDiskSpace returns the number of bytes available to unprivileged
// users on the filesystem containing the given path.
func availableDiskSpace(path string) (uint64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(path, &stat); err != nil {
		return 0, err
	}
	// Bavail = blocks available to unprivileged users
	return stat.Bavail * uint64(stat.Bsize), nil
}

// isErrNoSpace reports whether the error indicates that no space is left
// on the device (ENOSPC).
func isErrNoSpace(err error) bool {
	return errors.Is(err, syscall.ENOSPC)
}
