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
			Address:     std.Address("g1wsa9j6nel8ltt6q2lmf78585ymyfh5nsvhaxa3"),
			PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqma223maxmnw4f42kfqppvgyn8dr8wu7mhtdm6lcq64303a3vlln8xdmms",
			VotingPower: 1,
		},
		// gnocore-val-02
		{
			Address:     std.Address("g13762rd7y8s7jcc6uc4lyxv269hguchhpyzaamt"),
			PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zq5ndww8w6qrxgfdeastcx2lsuuk5r8w9jckkgevylq6duw59d54n935fq2",
			VotingPower: 1,
		},
		// onbloc-val-01
		{
			Address:     std.Address("g14cppfre9hsvu6p4scttuyu7mj082lfwxl7hvz9"),
			PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqe87d7lc0c4l4yaa8a94fucfre8882n8556l9z5220zjaaaqj7k5cl5sud",
			VotingPower: 1,
		},
		// onbloc-val-02
		{
			Address:     std.Address("g1927k3s7q9ujla04r5zy7q5m3gl84wsrart6663"),
			PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zq762adl0tcvdn54d6nzqf68r9wrywn7zj87v92mk3qpr436mevpvc63wsz",
			VotingPower: 1,
		},
	}

	// The default valset protocol is PoA
	vp = poa.NewPoA(poa.WithInitialSet(set))

	// No changes to apply initially
	changes = avl.NewTree()
}
