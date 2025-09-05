package validators

import (
	"std"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/poa"
	"gno.land/p/sys/validators"
)

func init() {
	// Prepare the initial validator set
	set := []validators.Validator{
		// gnocore-val-01
		{
			Address:     std.Address("g1cc9x4xyf3n3jygf3fvz462k9w0qaklda39mh2c"),
			PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqgd6ya9uc5auw4a2y8ms4jyqvdvu453jruseclyz6xnk5qltxvvnwjtkrs",
			VotingPower: 1,
		},
		// gnocore-val-02
		{
			Address:     std.Address("g1854g5lq87x5f75etzmg2wcd2mdmvj96hr469sa"),
			PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zpkljjay9cnqmtdcmndxd355ggga8s0tmmu0g5fdplu592wju6spfzngswj",
			VotingPower: 1,
		},
		// onbloc-val-01
		{
			Address:     std.Address("g15zkeyz2gwrjluqj6eremllrh6nx7mt4tlz8f32"),
			PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zpkhggjkjpvuuf4vmdv5lws9f3c22j6qkys3djt44dsyqwqc9padxlkhfd0",
			VotingPower: 1,
		},
		// onbloc-val-02
		{
			Address:     std.Address("g1v7wl7qlakzku5mrafmgntfuvd7xjrluhuhwewp"),
			PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqvl9cfzralqscxunw7cus2dmjwveskwcldjwnpr29zps6d5tlv5e2q2443",
			VotingPower: 1,
		},
	}

	// The default valset protocol is PoA
	vp = poa.NewPoA(poa.WithInitialSet(set))

	// No changes to apply initially
	changes = avl.NewTree()
}
