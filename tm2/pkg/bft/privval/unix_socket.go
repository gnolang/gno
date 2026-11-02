package privval

import (
	"fmt"
	"net"
	"os"
)

// chmodSocket sets a unix:// listener's socket mode. A test swaps it to make
// the call fail.
var chmodSocket = os.Chmod

// listenUnixSocket listens on a unix socket at address and sets the socket's
// file mode to mode. The mode is the listener's only access control: a local
// user who can write to the socket can connect and become the signer. If the
// mode cannot be set, the listener is closed, which removes the socket, and
// the error is returned.
func listenUnixSocket(address string, mode os.FileMode) (net.Listener, error) {
	ln, err := bindUnixSocket(address, mode)
	if err != nil {
		return nil, err
	}
	if err := chmodSocket(address, mode); err != nil {
		_ = ln.Close()
		// The *PathError from os.Chmod names the socket.
		return nil, fmt.Errorf("tmkms_listener: unable to set the unix socket's mode to %04o: %w", mode, err)
	}
	return ln, nil
}
