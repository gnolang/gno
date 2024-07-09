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
		// core-val-1
		{
			Address:     std.Address("g19v2h4pn6lrf8pwvhn8h0anek0cpt2tmhye4vkv"),
			PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqm6x7rlcrp96dz60tn4ws5a6mt34aptmj4qrzxv59fcnlrrd59q52vvvrj",
			VotingPower: 1,
		},
		// core-val-2
		{
			Address:     std.Address("g13us7swtc9hq550y9v4z6vcarak9vf8nqdvcqj4"),
			PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zp6z4cv32j5sjte5keucvfj6f44m9ctmaj4seqgg4vmekd4jdpzd6hzf0n7",
			VotingPower: 1,
		},
		// core-val-3
		{
			Address:     std.Address("g1ectm6algkfw3qnjmjvx7hacmh358t36ggj5lqv"),
			PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zpq4pfx2yfnaqy4pf6xwz2mjx8fvw8d3rmfxxdega46g0z3ak2t47wj27kf",
			VotingPower: 1,
		},
		// core-val-4
		{
			Address:     std.Address("g1tcxls3ylnrwrq95j33xpyuct4l370ra7jca4kj"),
			PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zp8lg7lfxj2lp68txvkh2mrpjkmgatpcgpsmlw53vssx9m2zgtfmnz5teuj",
			VotingPower: 1,
		},
		// devx-val-1
		{
			Address:     std.Address("g1mxguhd5zacar64txhfm0v7hhtph5wur5hx86vs"),
			PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqz6fwulsygvu9xypka3zqxhkxllm467e3adphmj6y44vn3yy8qq34vxnse",
			VotingPower: 1,
		},
		// devx-val-2
		{
			Address:     std.Address("g1t9ctfa468hn6czff8kazw08crazehcxaqa2uaa"),
			PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zpsq650w975vqsf6ajj5x4wdzfnrh64kmw7sljqz7wts6k0p6l36d0huls3",
			VotingPower: 1,
		},
		// onbloc-val-1
		{
			Address:     std.Address("g1gav33elude7prcdctpjekze7ft9l8qdjxqaj6d"),
			PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zp7pjfz35u8pm3jld0rsyw4ctphzql7n9jr59vrylq5l4f8qh35nw3pagcx",
			VotingPower: 1,
		},
	}

	// The default valset protocol is PoA
	vp = poa.NewPoA(poa.WithInitialSet(set))

	// No changes to apply initially
	changes = avl.NewTree()
}
