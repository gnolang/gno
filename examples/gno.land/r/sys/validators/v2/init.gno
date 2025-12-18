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
			Address:     address("g1gmg597aa85gk6u3wz3aluyxmctgfq9ld2fda7w"),
			PubKey:      "gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pqvsdv9yhul20sq9zfk08r3u72ysqdntu59xcez0ju2ydjguya9fusuc8h27",
			VotingPower: 1,
		},
		// gnocore-val-02
		{
			Address:     address("g1y8uw54dytc3twhv0vr4h8erghh7vxkczxvz9x0"),
			PubKey:      "gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pq06uuq3asdxgs5f438y863wkly5flazwtaz83ldvh33pmdecd0s0jyndfnl",
			VotingPower: 1,
		},
		// onbloc-val-01
		{
			Address:     address("g10c6v7u2gk3vecu33wr68kkdnh2swrg9n7klpus"),
			PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zp08ma9r9jg457ge5r5r3l5lkrg6h5e2h2l449gdxg6uc39a5n4nlksu78n",
			VotingPower: 1,
		},
		// onbloc-val-02
		{
			Address:     address("g16w9dydfwhavfe4vdll9fa8anuw4pu7509fqps8"),
			PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zpfxr5m293n6fkfjlx5yqk2q9kjpsdz3ax8at8vaaw63aye3kgueft6tmkr",
			VotingPower: 1,
		},
	}

	// The default valset protocol is PoA
	vp = poa.NewPoA(poa.WithInitialSet(set))

	// No changes to apply initially
	changes = avl.NewTree()
}
