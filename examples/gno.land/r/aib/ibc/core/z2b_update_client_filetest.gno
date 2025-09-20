package main

import (
	"crypto/sha256"
	"time"

	"gno.land/p/aib/ibc/types"
	"gno.land/r/aib/ibc/core"
	"gno.land/r/aib/ibc/core/testing"
)

// UpdateClient w/o existing client
func main() {
	var (
		clientID      = "07-tendermint-1"
		chainID       = "chain-id-2"
		apphash       = sha256.Sum256([]byte("apphash"))
		latestHeight  = types.Height{2, 3}
		newHeight     = uint64(5)
		newTimestamp  = time.Now()
		valset        = testing.GenValset()
		trustedValset = valset // no change in valset between the 2 heights
		msgHeader     = testing.GenMsgHeader(
			chainID, newTimestamp, apphash[:],
			newHeight, latestHeight, valset, trustedValset,
		)
	)
	core.UpdateClient(cross, clientID, msgHeader)
}

// Error:
// client 07-tendermint-1 not found
