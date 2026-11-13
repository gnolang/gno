package privval

import (
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The socket is created with its mode, not with 0777 minus the umask: with a
// zero umask, any local user could otherwise connect between bind and chmod
// and become the signer. Not parallel: the umask belongs to the whole
// process, and Go runs parallel top-level tests only once the sequential ones
// are done.
func TestBindUnixSocketCreatesTheSocketWithItsMode(t *testing.T) {
	old := syscall.Umask(0)
	t.Cleanup(func() { syscall.Umask(old) })

	sockPath := shortSocketPath(t)
	ln, err := bindUnixSocket(sockPath, 0o660)
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })

	info, err := os.Stat(sockPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o660).String(), info.Mode().Perm().String())
}
