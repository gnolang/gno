package privval

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/crypto/ed25519"
	"github.com/gnolang/gno/tm2/pkg/errors"
)

func getDialerTestCases(t *testing.T) []dialerTestCase {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(fmt.Errorf("no open ports? error listening to 127.0.0.1:0: %w", err))
	}
	unixFilePath, err := testUnixAddr()
	require.NoError(t, err)
	ul, err := net.Listen("unix", unixFilePath)
	if err != nil {
		panic(err)
	}

	return []dialerTestCase{
		{
			listener: l,
			dialer:   DialTCPFn(l.Addr().String(), testTimeoutReadWrite, ed25519.GenPrivKey()),
		},
		{
			listener: ul,
			dialer:   DialUnixFn(unixFilePath),
		},
	}
}

func TestIsConnTimeoutForFundamentalTimeouts(t *testing.T) {
	t.Parallel()

	// Generate a networking timeout
	tcpAddr := "127.0.0.1:34985"
	dialer := DialTCPFn(tcpAddr, time.Millisecond, ed25519.GenPrivKey())
	_, err := dialer()
	assert.Error(t, err)
	assert.True(t, IsConnTimeout(err))
}

func TestIsConnTimeoutForWrappedConnTimeouts(t *testing.T) {
	t.Parallel()

	tcpAddr := "127.0.0.1:34985"
	dialer := DialTCPFn(tcpAddr, time.Millisecond, ed25519.GenPrivKey())
	_, err := dialer()
	assert.Error(t, err)
	err = errors.Wrap(ErrConnectionTimeout, err.Error())
	assert.True(t, IsConnTimeout(err))
}
