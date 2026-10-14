//go:build !linux && !darwin && !windows

package ledger

import (
	"fmt"
	"runtime"
)

// hidSupported reports whether this platform has a HID backend.
const hidSupported = false

func connectHID() (*hidDevice, error) {
	return nil, fmt.Errorf("no Ledger support on %s", runtime.GOOS)
}
