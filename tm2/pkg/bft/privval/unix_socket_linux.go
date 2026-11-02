package privval

import (
	"context"
	"net"
	"os"
	"syscall"
)

// bindUnixSocket creates the socket at address and listens on it. The mode is
// set on the socket before bind(2), which then creates the file with mode
// minus the umask: the socket is never more open than mode, even before
// listenUnixSocket sets the exact mode.
func bindUnixSocket(address string, mode os.FileMode) (net.Listener, error) {
	lc := net.ListenConfig{
		// Control runs after the socket is created and before it is bound.
		Control: func(_, _ string, c syscall.RawConn) error {
			var err error
			if cerr := c.Control(func(fd uintptr) {
				err = syscall.Fchmod(int(fd), uint32(mode.Perm()))
			}); cerr != nil {
				return cerr
			}
			return err
		},
	}
	return lc.Listen(context.Background(), "unix", address)
}
