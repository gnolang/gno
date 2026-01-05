package types

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func BenchmarkNetAddress_String(b *testing.B) {
	key := GenerateNodeKey()

	na, err := NewNetAddressFromString(NetAddressString(key.ID(), "127.0.0.1:0"))
	require.NoError(b, err)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = na.String()
	}
}

func TestNewNetAddress(t *testing.T) {
	t.Parallel()

	t.Run("invalid TCP address", func(t *testing.T) {
		t.Parallel()

		var (
			key     = GenerateNodeKey()
			address = "127.0.0.1:8080"
		)

		udpAddr, err := net.ResolveUDPAddr("udp", address)
		require.NoError(t, err)

		_, err = NewNetAddress(key.ID(), udpAddr)
		require.Error(t, err)
	})

	t.Run("invalid ID", func(t *testing.T) {
		t.Parallel()

		var (
			id      = "" // zero ID
			address = "127.0.0.1:8080"
		)

		tcpAddr, err := net.ResolveTCPAddr("tcp", address)
		require.NoError(t, err)

		_, err = NewNetAddress(ID(id), tcpAddr)
		require.Error(t, err)
	})

	t.Run("valid net address", func(t *testing.T) {
		t.Parallel()

		var (
			key     = GenerateNodeKey()
			address = "127.0.0.1:8080"
		)

		tcpAddr, err := net.ResolveTCPAddr("tcp", address)
		require.NoError(t, err)

		addr, err := NewNetAddress(key.ID(), tcpAddr)
		require.NoError(t, err)

		assert.Equal(t, fmt.Sprintf("%s@%s", key.ID(), address), addr.String())
	})
}

func TestNewNetAddressFromString(t *testing.T) {
	t.Parallel()

	t.Run("valid net address", func(t *testing.T) {
		t.Parallel()

		testTable := []struct {
			name     string
			addr     string
			expected string
		}{
			{"no protocol", "g1m6kmam774klwlh4dhmhaatd7al02m0h0jwnyc6@127.0.0.1:8080", "g1m6kmam774klwlh4dhmhaatd7al02m0h0jwnyc6@127.0.0.1:8080"},
			{"tcp input", "tcp://g1m6kmam774klwlh4dhmhaatd7al02m0h0jwnyc6@127.0.0.1:8080", "g1m6kmam774klwlh4dhmhaatd7al02m0h0jwnyc6@127.0.0.1:8080"},
			{"udp input", "udp://g1m6kmam774klwlh4dhmhaatd7al02m0h0jwnyc6@127.0.0.1:8080", "g1m6kmam774klwlh4dhmhaatd7al02m0h0jwnyc6@127.0.0.1:8080"},
			{"no protocol", "g1m6kmam774klwlh4dhmhaatd7al02m0h0jwnyc6@127.0.0.1:8080", "g1m6kmam774klwlh4dhmhaatd7al02m0h0jwnyc6@127.0.0.1:8080"},
			{"tcp input", "tcp://g1m6kmam774klwlh4dhmhaatd7al02m0h0jwnyc6@127.0.0.1:8080", "g1m6kmam774klwlh4dhmhaatd7al02m0h0jwnyc6@127.0.0.1:8080"},
			{"udp input", "udp://g1m6kmam774klwlh4dhmhaatd7al02m0h0jwnyc6@127.0.0.1:8080", "g1m6kmam774klwlh4dhmhaatd7al02m0h0jwnyc6@127.0.0.1:8080"},
			{"correct nodeId w/tcp", "tcp://g1m6kmam774klwlh4dhmhaatd7al02m0h0jwnyc6@127.0.0.1:8080", "g1m6kmam774klwlh4dhmhaatd7al02m0h0jwnyc6@127.0.0.1:8080"},
		}

		for _, testCase := range testTable {
			t.Run(testCase.name, func(t *testing.T) {
				t.Parallel()

				addr, err := NewNetAddressFromString(testCase.addr)
				require.NoError(t, err)

				assert.Equal(t, testCase.expected, addr.String())
			})
		}
	})

	t.Run("invalid net address", func(t *testing.T) {
		t.Parallel()

		testTable := []struct {
			name string
			addr string
		}{
			{"no node id and no protocol", "127.0.0.1:8080"},
			{"no node id w/ tcp input", "tcp://127.0.0.1:8080"},
			{"no node id w/ udp input", "udp://127.0.0.1:8080"},

			{"malformed tcp input", "tcp//g1m6kmam774klwlh4dhmhaatd7al02m0h0jwnyc6@127.0.0.1:8080"},
			{"malformed udp input", "udp//g1m6kmam774klwlh4dhmhaatd7al02m0h0jwnyc6@127.0.0.1:8080"},

			{"invalid host", "notahost"},
			{"invalid port", "127.0.0.1:notapath"},
			{"invalid host w/ port", "notahost:8080"},
			{"just a port", "8082"},
			{"non-existent port", "127.0.0:8080000"},

			{"too short nodeId", "deadbeef@127.0.0.1:8080"},
			{"too short, not hex nodeId", "this-isnot-hex@127.0.0.1:8080"},
			{"not bech32 nodeId", "xxxm6kmam774klwlh4dhmhaatd7al02m0h0hdap9l@127.0.0.1:8080"},

			{"too short nodeId w/tcp", "tcp://deadbeef@127.0.0.1:8080"},
			{"too short notHex nodeId w/tcp", "tcp://this-isnot-hex@127.0.0.1:8080"},
			{"not bech32 nodeId w/tcp", "tcp://xxxxm6kmam774klwlh4dhmhaatd7al02m0h0hdap9l@127.0.0.1:8080"},

			{"no node id", "tcp://@127.0.0.1:8080"},
			{"no node id or IP", "tcp://@"},
			{"tcp no host, w/ port", "tcp://:26656"},
			{"empty", ""},
			{"node id delimiter 1", "@"},
			{"node id delimiter 2", " @"},
			{"node id delimiter 3", " @ "},
		}

		for _, testCase := range testTable {
			t.Run(testCase.name, func(t *testing.T) {
				t.Parallel()

				addr, err := NewNetAddressFromString(testCase.addr)

				assert.Nil(t, addr)
				assert.Error(t, err)
			})
		}
	})
}

func TestNewNetAddressFromStrings(t *testing.T) {
	t.Parallel()

	t.Run("invalid addresses", func(t *testing.T) {
		t.Parallel()

		var (
			keys = generateKeys(t, 10)
			strs = make([]string, 0, len(keys))
		)

		for index, key := range keys {
			if index%2 != 0 {
				strs = append(
					strs,
					fmt.Sprintf("%s@:8080", key.ID()), // missing host
				)

				continue
			}

			strs = append(
				strs,
				fmt.Sprintf("%s@127.0.0.1:8080", key.ID()),
			)
		}

		// Convert the strings
		addrs, errs := NewNetAddressFromStrings(strs)

		assert.Len(t, errs, len(keys)/2)
		assert.Equal(t, len(keys)/2, len(addrs))

		for index, addr := range addrs {
			assert.Contains(t, addr.String(), keys[index*2].ID())
		}
	})

	t.Run("valid addresses", func(t *testing.T) {
		t.Parallel()

		var (
			keys = generateKeys(t, 10)
			strs = make([]string, 0, len(keys))
		)

		for _, key := range keys {
			strs = append(
				strs,
				fmt.Sprintf("%s@127.0.0.1:8080", key.ID()),
			)
		}

		// Convert the strings
		addrs, errs := NewNetAddressFromStrings(strs)

		assert.Len(t, errs, 0)
		assert.Equal(t, len(keys), len(addrs))

		for index, addr := range addrs {
			assert.Contains(t, addr.String(), keys[index].ID())
		}
	})
}

func TestNewNetAddressFromIPPort(t *testing.T) {
	t.Parallel()

	var (
		host = "127.0.0.1"
		port = uint16(8080)
	)

	addr := NewNetAddressFromIPPort(net.ParseIP(host), port)

	assert.Equal(
		t,
		fmt.Sprintf("%s:%d", host, port),
		addr.String(),
	)
}

func TestNetAddress_Local(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name    string
		addr    string
		isLocal bool
	}{
		{
			"local loopback",
			"127.0.0.1:8080",
			true,
		},
		{
			"local loopback, zero",
			"0.0.0.0:8080",
			true,
		},
		{
			"non-local address",
			"200.100.200.100:8080",
			false,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			key := GenerateNodeKey()

			addr, err := NewNetAddressFromString(
				fmt.Sprintf(
					"%s@%s",
					key.ID(),
					testCase.addr,
				),
			)
			require.NoError(t, err)

			assert.Equal(t, testCase.isLocal, addr.Local())
		})
	}
}

func TestNetAddressResolveIP(t *testing.T) {
	t.Parallel()

	t.Run("updates IP from IP hostname", func(t *testing.T) {
		t.Parallel()

		var (
			key      = GenerateNodeKey()
			expected = "127.0.0.2"
		)

		addr := &NetAddress{
			ID:       key.ID(),
			Hostname: expected,
			IP:       net.ParseIP("127.0.0.1"),
			Port:     8080,
		}

		err := addr.ResolveIP(context.Background())
		require.NoError(t, err)

		assert.Equal(t, expected, addr.IP.String())
	})

	t.Run("resolves hostname to IP", func(t *testing.T) {
		t.Parallel()

		key := GenerateNodeKey()

		addr := &NetAddress{
			ID:       key.ID(),
			Hostname: "localhost",
			Port:     8080,
		}

		err := addr.ResolveIP(context.Background())
		require.NoError(t, err)

		require.NotNil(t, addr.IP)
		assert.NotEmpty(t, addr.IP.String())
	})
}

func TestNetAddressFromStringHostnamePreserved(t *testing.T) {
	t.Parallel()

	key := GenerateNodeKey()

	addr, err := NewNetAddressFromString(fmt.Sprintf("%s@localhost:8080", key.ID()))
	require.NoError(t, err)

	assert.Equal(t, "localhost", addr.Hostname)
	require.NotNil(t, addr.IP)
}

func TestNetAddress_Routable(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name       string
		addr       string
		isRoutable bool
	}{
		{
			"local loopback",
			"127.0.0.1:8080",
			false,
		},
		{
			"routable address",
			"gno.land:80",
			true,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			key := GenerateNodeKey()

			addr, err := NewNetAddressFromString(
				fmt.Sprintf(
					"%s@%s",
					key.ID(),
					testCase.addr,
				),
			)
			require.NoError(t, err)

			assert.Equal(t, testCase.isRoutable, addr.Routable())
		})
	}
}
