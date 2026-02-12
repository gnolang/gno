package validators

import (
	"gno.land/p/nt/avl"
	"gno.land/p/nt/poa"
	"gno.land/p/sys/validators"
)

func init() {
	// Prepare the initial validator set
	set := []validators.Validator{
		// gnocore-val-01
		{
			Address:     address("g1e8j30772p4fvjd27eqq0e50vs74ekst50f5hz6"),
			PubKey:      "gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pqdw9mje8pwpd8p6pu3h0ke9mxp3s7l5vr6jtsffyd9wazpgxufwrssj367n",
			VotingPower: 1,
		},
		// gnocore-val-02
		{
			Address:     address("g18k5nz67cg03lrdkwnf9qmdfzf0ptaq7lxefe4y"),
			PubKey:      "gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pq0mxx3hhmy5hsl9kv8ne2ch46l8uk3kpnwc8kr900066ejz3zlhqcj6wjna",
			VotingPower: 1,
		},
	}

	// The default valset protocol is PoA
	vp = poa.NewPoA(poa.WithInitialSet(set))

	// No changes to apply initially
	changes = avl.NewTree()
}
