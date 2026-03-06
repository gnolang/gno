package validators

import (
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/poa/v0"
	"gno.land/p/sys/validators"
)

func init() {
	// Prepare the initial validator set
	set := []validators.Validator{
		// gnocore-val-01
		{
			Address:     address("g1euw20dwq4yt3zvjl0kl725me0lfrjf5lzaws4z"),
			PubKey:      "gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pqty3jnuspxthzmqyvjgxcwlu90pq8atj8lda7a2wsr2gqmpa47pdj2jvqrc",
			VotingPower: 1,
		},
		// gnocore-val-02
		{
			Address:     address("g1maa9t9ew7v3xj0cmnuyrr7frjguzykqeykjh0n"),
			PubKey:      "gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pqwdr6r6rr5eyrcrmletzk3rpnxvcupppu20tkhh4fzqlnx6erzazvhsf25g",
			VotingPower: 1,
		},
	}

	// The default valset protocol is PoA
	vp = poa.NewPoA(poa.WithInitialSet(set))

	// No changes to apply initially
	changes = avl.NewTree()
}
