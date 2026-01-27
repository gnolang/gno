package os

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProtocolAndAddress(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name          string
		fullAddr      string
		expectedProto string
		expectedAddr  string
	}{
		{
			name:          "tcp with scheme",
			fullAddr:      "tcp://mydomain:80",
			expectedProto: "tcp",
			expectedAddr:  "mydomain:80",
		},
		{
			name:          "default to tcp",
			fullAddr:      "mydomain:80",
			expectedProto: "tcp",
			expectedAddr:  "mydomain:80",
		},
		{
			name:          "unix socket with scheme",
			fullAddr:      "unix://mydomain:80",
			expectedProto: "unix",
			expectedAddr:  "mydomain:80",
		},
		{
			name:          "unix socket path",
			fullAddr:      "unix:///tmp/test.sock",
			expectedProto: "unix",
			expectedAddr:  "/tmp/test.sock",
		},
		{
			name:          "tcp localhost with port",
			fullAddr:      "tcp://127.0.0.1:26657",
			expectedProto: "tcp",
			expectedAddr:  "127.0.0.1:26657",
		},
		{
			name:          "tcp all interfaces",
			fullAddr:      "tcp://0.0.0.0:26657",
			expectedProto: "tcp",
			expectedAddr:  "0.0.0.0:26657",
		},
		{
			name:          "empty string defaults to tcp",
			fullAddr:      "",
			expectedProto: "tcp",
			expectedAddr:  "",
		},
		{
			name:          "ipv6 address",
			fullAddr:      "tcp://[::1]:26657",
			expectedProto: "tcp",
			expectedAddr:  "[::1]:26657",
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			proto, addr := ProtocolAndAddress(testCase.fullAddr)

			assert.Equal(t, testCase.expectedProto, proto)
			assert.Equal(t, testCase.expectedAddr, addr)
		})
	}
}
