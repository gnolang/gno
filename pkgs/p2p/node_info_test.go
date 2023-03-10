package p2p

import (
	"fmt"
	"net"
	"testing"

	"github.com/gnolang/gno/pkgs/crypto/ed25519"
	"github.com/gnolang/gno/pkgs/versionset"
	"github.com/stretchr/testify/assert"
)

func TestNodeInfoValidate(t *testing.T) {
	// empty fails
	ni := NodeInfo{}
	assert.Error(t, ni.Validate())

	channels := make([]byte, maxNumChannels)
	for i := 0; i < maxNumChannels; i++ {
		channels[i] = byte(i)
	}
	dupChannels := make([]byte, 5)
	copy(dupChannels, channels[:5])
	dupChannels = append(dupChannels, testCh)

	nonAscii := "¢§µ"
	emptyTab := fmt.Sprintf("\t")
	emptySpace := fmt.Sprintf("  ")

	testCases := []struct {
		testName         string
		malleateNodeInfo func(*NodeInfo)
		expectErr        bool
	}{
		{"Too Many Channels", func(ni *NodeInfo) { ni.Channels = append(channels, byte(maxNumChannels)) }, true}, //nolint: gocritic
		{"Duplicate Channel", func(ni *NodeInfo) { ni.Channels = dupChannels }, true},
		{"Good Channels", func(ni *NodeInfo) { ni.Channels = ni.Channels[:5] }, false},

		{"Nil NetAddress", func(ni *NodeInfo) { ni.NetAddress = nil }, true},
		{"Zero NetAddress ID", func(ni *NodeInfo) { ni.NetAddress.ID = "" }, true},
		{"Invalid NetAddress IP", func(ni *NodeInfo) { ni.NetAddress.IP = net.IP([]byte{0x00}) }, true},

		{"Non-ASCII Version", func(ni *NodeInfo) { ni.Version = nonAscii }, true},
		{"Empty tab Version", func(ni *NodeInfo) { ni.Version = emptyTab }, true},
		{"Empty space Version", func(ni *NodeInfo) { ni.Version = emptySpace }, true},
		{"Empty Version", func(ni *NodeInfo) { ni.Version = "" }, false},

		{"Non-ASCII Moniker", func(ni *NodeInfo) { ni.Moniker = nonAscii }, true},
		{"Empty tab Moniker", func(ni *NodeInfo) { ni.Moniker = emptyTab }, true},
		{"Empty space Moniker", func(ni *NodeInfo) { ni.Moniker = emptySpace }, true},
		{"Empty Moniker", func(ni *NodeInfo) { ni.Moniker = "" }, true},
		{"Good Moniker", func(ni *NodeInfo) { ni.Moniker = "hey its me" }, false},

		{"Non-ASCII TxIndex", func(ni *NodeInfo) { ni.Other.TxIndex = nonAscii }, true},
		{"Empty tab TxIndex", func(ni *NodeInfo) { ni.Other.TxIndex = emptyTab }, true},
		{"Empty space TxIndex", func(ni *NodeInfo) { ni.Other.TxIndex = emptySpace }, true},
		{"Empty TxIndex", func(ni *NodeInfo) { ni.Other.TxIndex = "" }, false},
		{"Off TxIndex", func(ni *NodeInfo) { ni.Other.TxIndex = "off" }, false},

		{"Non-ASCII RPCAddress", func(ni *NodeInfo) { ni.Other.RPCAddress = nonAscii }, true},
		{"Empty tab RPCAddress", func(ni *NodeInfo) { ni.Other.RPCAddress = emptyTab }, true},
		{"Empty space RPCAddress", func(ni *NodeInfo) { ni.Other.RPCAddress = emptySpace }, true},
		{"Empty RPCAddress", func(ni *NodeInfo) { ni.Other.RPCAddress = "" }, false},
		{"Good RPCAddress", func(ni *NodeInfo) { ni.Other.RPCAddress = "0.0.0.0:26657" }, false},
	}

	nodeKey := NodeKey{PrivKey: ed25519.GenPrivKey()}
	name := "testing"

	// test case passes
	ni = testNodeInfo(nodeKey.ID(), name)
	ni.Channels = channels
	assert.NoError(t, ni.Validate())

	for _, tc := range testCases {
		ni := testNodeInfo(nodeKey.ID(), name)
		ni.Channels = channels
		tc.malleateNodeInfo(&ni)
		err := ni.Validate()
		if tc.expectErr {
			assert.Error(t, err, tc.testName)
		} else {
			assert.NoError(t, err, tc.testName)
		}
	}
}

func TestNodeInfoCompatible(t *testing.T) {
	nodeKey1 := NodeKey{PrivKey: ed25519.GenPrivKey()}
	nodeKey2 := NodeKey{PrivKey: ed25519.GenPrivKey()}
	name := "testing"

	var newTestChannel byte = 0x2

	// test NodeInfo is compatible
	ni1 := testNodeInfo(nodeKey1.ID(), name)
	ni2 := testNodeInfo(nodeKey2.ID(), name)
	assert.NoError(t, ni1.CompatibleWith(ni2))

	// add another channel; still compatible
	ni2.Channels = []byte{newTestChannel, testCh}
	assert.NoError(t, ni1.CompatibleWith(ni2))

	// wrong NodeInfo type is not compatible
	_, netAddr := CreateRoutableAddr()
	ni3 := NodeInfo{NetAddress: netAddr}
	assert.Error(t, ni1.CompatibleWith(ni3))

	testCases := []struct {
		testName         string
		malleateNodeInfo func(*NodeInfo)
	}{
		{"Bad block version", func(ni *NodeInfo) {
			ni.VersionSet.Set(versionset.VersionInfo{Name: "Block", Version: "badversion"})
		}},
		{"Wrong block version", func(ni *NodeInfo) {
			ni.VersionSet.Set(versionset.VersionInfo{Name: "Block", Version: "v999.999.999-wrong"})
		}},
		{"Wrong network", func(ni *NodeInfo) { ni.Network += "-wrong" }},
		{"No common channels", func(ni *NodeInfo) { ni.Channels = []byte{newTestChannel} }},
	}

	for i, tc := range testCases {
		t.Logf("case #%v", i)
		ni := testNodeInfo(nodeKey2.ID(), name)
		tc.malleateNodeInfo(&ni)
		fmt.Printf("case #%v\n", i)
		assert.Error(t, ni1.CompatibleWith(ni))
	}
}
