package validators

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/nt/poa"
	"gno.land/p/sys/validators"
)

func init() {
	// Prepare the initial validator set
	set := []validators.Validator{
		// gnocore-val-01
		{
			Address:     std.Address("g1mt5d6l56tf8r3u8mehnhgp54mqlx3lu5qdysum"),
			PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zpv5dep4jymxp9p3xkd5e3cdk63wqygw00nsm402nxh593rkuan75gkr7e6",
			VotingPower: 1,
		},
		// gnocore-val-02
		{
			Address:     std.Address("g1mmrdx5j4v878uqttp44t96d3rlw8rls8fgze24"),
			PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zpvklu0t8l37fh4lf8kap8jcrsk3akkvm5nsu6m2v74agp45xt9pd7h5ana",
			VotingPower: 1,
		},
		// gnocore-val-03
		{
			Address:     std.Address("g16wh3t370fctrukvvdslsz9uc76tfpy0ggrm5hr"),
			PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zp0768mk3fg7dkprgkl5twxzhn2hlef27mzt46qea8z4qv0ltwstuy9qk62",
			VotingPower: 1,
		},
		// gnocore-val-04
		{
			Address:     std.Address("g1y7d86mffahwy7c8s4j0nvgwhkg4n4tl7ek4w8z"),
			PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zq82uuwf5yvrfpyhp5zh5hspz7wvee4jpsh37muczqrpatuxr3vpwyhjdsy",
			VotingPower: 1,
		},
	}

	// The default valset protocol is PoA
	vp = poa.NewPoA(poa.WithInitialSet(set))

	// No changes to apply initially
	changes = avl.NewTree()
}
