package core

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	rpctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/p2p"
	p2pcfg "github.com/gnolang/gno/tm2/pkg/p2p/config"
)

var mu sync.Mutex

func TestUnsafeDialSeeds(t *testing.T) {
	sw := p2p.MakeSwitch(p2pcfg.DefaultP2PConfig(), 1, "testing", "123.123.123",
		func(n int, sw *p2p.Switch) *p2p.Switch { return sw })
	err := sw.Start()
	require.NoError(t, err)

	mu.Lock()
	defer sw.Stop()
	mu.Unlock()

	logger = log.NewNoopLogger()

	mu.Lock()
	p2pPeers = sw
	mu.Unlock()

	testCases := []struct {
		seeds []string
		isErr bool
	}{
		{[]string{}, true},
		{[]string{"g1m6kmam774klwlh4dhmhaatd7al02m0h0jwnyc6@127.0.0.1:41198"}, false},
		{[]string{"127.0.0.1:41198"}, true},
	}

	for _, tc := range testCases {
		res, err := UnsafeDialSeeds(&rpctypes.Context{}, tc.seeds)
		if tc.isErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.NotNil(t, res)
		}
	}
}

func TestUnsafeDialPeers(t *testing.T) {
	sw := p2p.MakeSwitch(p2pcfg.DefaultP2PConfig(), 1, "testing", "123.123.123",
		func(n int, sw *p2p.Switch) *p2p.Switch { return sw })
	err := sw.Start()
	require.NoError(t, err)
	defer sw.Stop()

	logger = log.NewNoopLogger()

	mu.Lock()
	p2pPeers = sw
	mu.Unlock()

	testCases := []struct {
		peers []string
		isErr bool
	}{
		{[]string{}, true},
		{[]string{"g1m6kmam774klwlh4dhmhaatd7al02m0h0jwnyc6@127.0.0.1:41198"}, false},
		{[]string{"127.0.0.1:41198"}, true},
	}

	for _, tc := range testCases {
		res, err := UnsafeDialPeers(&rpctypes.Context{}, tc.peers, false)
		if tc.isErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.NotNil(t, res)
		}
	}
}
