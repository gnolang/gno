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
			Address:     address("g1wmaglcam7xq3kwvrks6ysyeutf0jc877f4q3nl"),
			PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zpelzea0ep9dclr83vygetymcs5uekam9j53t2sgcefumxth2wagquu3vrk",
			VotingPower: 1,
		},
		// onbloc-val-02
		{
			Address:     address("g1quanjrv29f3zvy8cnsd7s3wjcnhmv4v4n3knhm"),
			PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqz3ewj7qudswvtsex2l73g42v5tnkyguxcgft8tdg3mrpzwzzn4lde0vag",
			VotingPower: 1,
		},
	}

	// The default valset protocol is PoA
	vp = poa.NewPoA(poa.WithInitialSet(set))

	// No changes to apply initially
	changes = avl.NewTree()
}
