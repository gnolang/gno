package main

import (
	"crypto/sha256"
	"time"

	tmtesting "gno.land/p/aib/ibc/lightclient/tendermint/testing"
	"gno.land/p/aib/ibc/types"

	// Import to trigger the init and register the app
	_ "gno.land/r/aib/ibc/apps/transfer"
	"gno.land/r/aib/ibc/core"
)

// init: ensure app is registered
func main() {
	var (
		chainID        = "chain-id-2"
		clientState    = tmtesting.NewClientState(chainID, types.NewHeight(2, 2))
		apphash        = sha256.Sum256([]byte("apphash"))
		valset         = tmtesting.GenValset()
		consensusState = tmtesting.GenConsensusState(time.Now(), apphash[:], valset.Hash())
	)
	clientID := core.CreateClient(cross, clientState, consensusState)
	core.RegisterCounterparty(cross, clientID, [][]byte{[]byte("prefix1"), []byte("prefix2")}, "07-tendermint-2")

	println("----------- assert render apps")
	println(core.Render("apps"))
}

// Output:
// ----------- assert render apps
// [{"port_id":"transfer","pkg_path":"gno.land/r/aib/ibc/apps/transfer","address":"g1tp3gk4quumurav4858hjfdy6hxtyffwmnxyr00"}]
