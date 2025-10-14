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
			Address:     address("g1ek7ftha29qv4ahtv7jzpc0d57lqy7ynzklht7t"),
			PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zq7kdacwkqquf7j9ywd6c3yaj4vsytkan526knd7nt2z5v4q2mgc6hfkfyu",
			VotingPower: 1,
		},
		// gnocore-val-02
		{
			Address:     address("g1xwlzxpuh4v3l9fjrl4cq2ylzpk7mhj0598xdy7"),
			PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zpnxghmy95x979vyw43073prpzqah9k9sez2mxsfkyfxge02hqczhqy4r4j",
			VotingPower: 1,
		},
		// gnodevx-val-01
		{
			Address:     address("g1zeawgdp4nh6j6phmnsjmssqrz4ukruh43anckl"),
			PubKey:      "gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pqwhdpj834lxwenxja6tanrz2rn3a4rpcxxqxhh23f77dsftyhqrgxpghjwm",
			VotingPower: 1,
		},
		// gnodevx-val-02
		{
			Address:     address("g1uajmkwz6w7juqxhu64qecfm6kr4r8us70culre"),
			PubKey:      "gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pqfc9agrkxsm7d4ry6zz823f7ulkkc3u3nqv6avz99tm45nmfj4ug54c38c0",
			VotingPower: 1,
		},
		// onbloc-val-01
		{
			Address:     address("g1kntcjkfplj0z44phajajwqkx5q4ry5yaft5q2h"),
			PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqxkeawgn93f8jm7wumknjwysg5mm3shyeyh8fky6pwytpcf5uyqk4vwshh",
			VotingPower: 1,
		},
		// onbloc-val-02
		{
			Address:     address("g1j306jcl4qyhgjw78shl3ajp88vmvdcf7m7ntm2"),
			PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqvpv84g5n8zz7k2m2yenuv73e9r77zxzxzrucj064gzykj2puyvszw8kxc",
			VotingPower: 1,
		},
	}

	// The default valset protocol is PoA
	vp = poa.NewPoA(poa.WithInitialSet(set))

	// No changes to apply initially
	changes = avl.NewTree()
}
