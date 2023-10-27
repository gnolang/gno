package rpcclient

import (
	"testing"

	"github.com/jaekwon/testify/assert"
)

func Test_parseRemoteAddr(t *testing.T) {
	tt := []struct {
		remoteAddr              string
		network, s, errContains string
	}{
		{"127.0.0.1", "tcp", "127.0.0.1", ""},
		{"https://example.com", "https", "example.com", ""},
		{"wss://[::1]", "wss", "[::1]", ""},
		// no error cases - they cannot happen!
	}

	for _, tc := range tt {
		n, s, err := parseRemoteAddr(tc.remoteAddr)
		if tc.errContains != "" {
			_ = assert.NotNil(t, err) && assert.Contains(t, err.Error(), tc.errContains)
		}
		assert.NoError(t, err)
		assert.Equal(t, n, tc.network)
		assert.Equal(t, s, tc.s)
	}
}

// Following tests check that we correctly translate http/https to tcp,
// and other protocols are left intact from parseRemoteAddr()

func Test_makeHTTPDialer(t *testing.T) {
	dl := makeHTTPDialer("https://.")
	_, err := dl("hello", "world")
	if assert.NotNil(t, err) {
		e := err.Error()
		assert.Contains(t, e, "dial tcp:", "should convert https to tcp")
		assert.Contains(t, e, "address .:", "should have parsed the address (as incorrect)")
	}
}

func Test_makeHTTPDialer_noConvert(t *testing.T) {
	dl := makeHTTPDialer("udp://.")
	_, err := dl("hello", "world")
	if assert.NotNil(t, err) {
		e := err.Error()
		assert.Contains(t, e, "dial udp:", "udp protocol should remain the same")
		assert.Contains(t, e, "address .:", "should have parsed the address (as incorrect)")
	}
}
