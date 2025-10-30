package main

import (
	"gno.land/r/aib/ibc/core"
)

func main() {
	var (
		clientID                 = ""
		counterpartyMerklePrefix = [][]byte{
			[]byte("iavlStoreKey"),
			[]byte("storeKey"),
		}
		counterpartyClientID = ""
	)
	core.RegisterCounterparty(cross, clientID, counterpartyMerklePrefix, counterpartyClientID)
}

// Error:
// invalid counterparty client id
