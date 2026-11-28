//go:build !linux

package privval

import (
	"net"
	"os"
)

// bindUnixSocket creates the socket at address and listens on it. Until
// listenUnixSocket sets its mode, the socket has 0777 minus the umask: macOS,
// for one, refuses fchmod(2) on a socket before bind(2).
func bindUnixSocket(address string, _ os.FileMode) (net.Listener, error) {
	return net.Listen("unix", address)
}
