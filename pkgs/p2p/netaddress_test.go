package p2p

import (
	"encoding/hex"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/classic/crypto"
)

func TestAddress2ID(t *testing.T) {
	idbz, _ := hex.DecodeString("deadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
	id := crypto.AddressFromBytes(idbz).ID()
	assert.Equal(t, crypto.ID("c1m6kmam774klwlh4dhmhaatd7al02m0h0hdap9l"), id)

	idbz, _ = hex.DecodeString("deadbeefdeadbeefdeadbeefdeadbeefdead0000")
	id = crypto.AddressFromBytes(idbz).ID()
	assert.Equal(t, crypto.ID("c1m6kmam774klwlh4dhmhaatd7al026qqq9xk0hx"), id)
}

func TestNewNetAddress(t *testing.T) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:8080")
	require.Nil(t, err)

	assert.Panics(t, func() {
		NewNetAddress("", tcpAddr)
	})

	idbz, _ := hex.DecodeString("deadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
	id := crypto.AddressFromBytes(idbz).ID()
	// ^-- is "c1m6kmam774klwlh4dhmhaatd7al02m0h0hdap9l"

	addr := NewNetAddress(id, tcpAddr)
	assert.Equal(t, "c1m6kmam774klwlh4dhmhaatd7al02m0h0hdap9l@127.0.0.1:8080", addr.String())

	assert.NotPanics(t, func() {
		NewNetAddress("", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8000})
	}, "Calling NewNetAddress with UDPAddr should not panic in testing")
}

func TestNewNetAddressFromString(t *testing.T) {
	testCases := []struct {
		name     string
		addr     string
		expected string
		correct  bool
	}{
		{"no node id and no protocol", "127.0.0.1:8080", "", false},
		{"no node id w/ tcp input", "tcp://127.0.0.1:8080", "", false},
		{"no node id w/ udp input", "udp://127.0.0.1:8080", "", false},

		{"no protocol", "c1m6kmam774klwlh4dhmhaatd7al02m0h0hdap9l@127.0.0.1:8080", "c1m6kmam774klwlh4dhmhaatd7al02m0h0hdap9l@127.0.0.1:8080", true},
		{"tcp input", "tcp://c1m6kmam774klwlh4dhmhaatd7al02m0h0hdap9l@127.0.0.1:8080", "c1m6kmam774klwlh4dhmhaatd7al02m0h0hdap9l@127.0.0.1:8080", true},
		{"udp input", "udp://c1m6kmam774klwlh4dhmhaatd7al02m0h0hdap9l@127.0.0.1:8080", "c1m6kmam774klwlh4dhmhaatd7al02m0h0hdap9l@127.0.0.1:8080", true},
		{"malformed tcp input", "tcp//c1m6kmam774klwlh4dhmhaatd7al02m0h0hdap9l@127.0.0.1:8080", "", false},
		{"malformed udp input", "udp//c1m6kmam774klwlh4dhmhaatd7al02m0h0hdap9l@127.0.0.1:8080", "", false},

		// {"127.0.0:8080", false},
		{"invalid host", "notahost", "", false},
		{"invalid port", "127.0.0.1:notapath", "", false},
		{"invalid host w/ port", "notahost:8080", "", false},
		{"just a port", "8082", "", false},
		{"non-existent port", "127.0.0:8080000", "", false},

		{"too short nodeId", "deadbeef@127.0.0.1:8080", "", false},
		{"too short, not hex nodeId", "this-isnot-hex@127.0.0.1:8080", "", false},
		{"not bech32 nodeId", "xxxm6kmam774klwlh4dhmhaatd7al02m0h0hdap9l@127.0.0.1:8080", "", false},

		{"too short nodeId w/tcp", "tcp://deadbeef@127.0.0.1:8080", "", false},
		{"too short notHex nodeId w/tcp", "tcp://this-isnot-hex@127.0.0.1:8080", "", false},
		{"not bech32 nodeId w/tcp", "tcp://xxxxm6kmam774klwlh4dhmhaatd7al02m0h0hdap9l@127.0.0.1:8080", "", false},
		{"correct nodeId w/tcp", "tcp://c1m6kmam774klwlh4dhmhaatd7al02m0h0hdap9l@127.0.0.1:8080", "c1m6kmam774klwlh4dhmhaatd7al02m0h0hdap9l@127.0.0.1:8080", true},

		{"no node id", "tcp://@127.0.0.1:8080", "", false},
		{"no node id or IP", "tcp://@", "", false},
		{"tcp no host, w/ port", "tcp://:26656", "", false},
		{"empty", "", "", false},
		{"node id delimiter 1", "@", "", false},
		{"node id delimiter 2", " @", "", false},
		{"node id delimiter 3", " @ ", "", false},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			addr, err := NewNetAddressFromString(tc.addr)
			if tc.correct {
				if assert.Nil(t, err, tc.addr) {
					assert.Equal(t, tc.expected, addr.String())
				}
			} else {
				assert.NotNil(t, err, tc.addr)
			}
		})
	}
}

func TestNewNetAddressFromStrings(t *testing.T) {
	addrs, errs := NewNetAddressFromStrings([]string{
		"127.0.0.1:8080",
		"c1m6kmam774klwlh4dhmhaatd7al02m0h0hdap9l@127.0.0.1:8080",
		"c1m6kmam774klwlh4dhmhaatd7al02m0h0hdap9l@127.0.0.2:8080"})
	assert.Len(t, errs, 1)
	assert.Equal(t, 2, len(addrs))
}

func TestNewNetAddressFromIPPort(t *testing.T) {
	addr := NewNetAddressFromIPPort("", net.ParseIP("127.0.0.1"), 8080)
	assert.Equal(t, "127.0.0.1:8080", addr.String())
}

func TestNetAddressProperties(t *testing.T) {
	// TODO add more test cases
	testCases := []struct {
		addr     string
		valid    bool
		local    bool
		routable bool
	}{
		{"c1m6kmam774klwlh4dhmhaatd7al02m0h0hdap9l@127.0.0.1:8080", true, true, false},
		{"c1m6kmam774klwlh4dhmhaatd7al02m0h0hdap9l@ya.ru:80", true, false, true},
	}

	for _, tc := range testCases {
		addr, err := NewNetAddressFromString(tc.addr)
		require.Nil(t, err)

		err = addr.Validate()
		if tc.valid {
			assert.NoError(t, err)
		} else {
			assert.Error(t, err)
		}
		assert.Equal(t, tc.local, addr.Local())
		assert.Equal(t, tc.routable, addr.Routable())
	}
}

func TestNetAddressReachabilityTo(t *testing.T) {
	// TODO add more test cases
	testCases := []struct {
		addr         string
		other        string
		reachability int
	}{
		{"c1m6kmam774klwlh4dhmhaatd7al02m0h0hdap9l@127.0.0.1:8080", "c1m6kmam774klwlh4dhmhaatd7al02m0h0hdap9l@127.0.0.1:8081", 0},
		{"c1m6kmam774klwlh4dhmhaatd7al02m0h0hdap9l@ya.ru:80", "c1m6kmam774klwlh4dhmhaatd7al02m0h0hdap9l@127.0.0.1:8080", 1},
	}

	for _, tc := range testCases {
		addr, err := NewNetAddressFromString(tc.addr)
		require.Nil(t, err)

		other, err := NewNetAddressFromString(tc.other)
		require.Nil(t, err)

		assert.Equal(t, tc.reachability, addr.ReachabilityTo(other))
	}
}
