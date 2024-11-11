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
			Address:     std.Address("g1qn3jwvdpva622j3fyudqy65zstnqx2wnqhrs3s"),
			PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zpndqtjh5dcsnd0gcez3frs3w6rsttmlekj4cyywegyh0n8uprwvj5n8688",
			VotingPower: 1,
		},
		// core-val-2
		{
			Address:     std.Address("g1gtu9czw9qavrtdnf936usvwjwyjz0x0jk243au"),
			PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zq4y0ppxhxazdwxhnsxxzdmh9rxht888n4fl0mskwcpq7y2404dm2h0lamk",
			VotingPower: 1,
		},
		// core-val-3
		{
			Address:     std.Address("g19emxxnzzfa0pkffvthrss5drgccjnwj8mdme4f"),
			PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zq288fe7pd2yy3h2h8qjh0elu3pxuamf3wpa9qt9s6jja0r3k64ue4mh636",
			VotingPower: 1,
		},
		// core-val-4
		{
			Address:     std.Address("g1hyxtsgjr5zt06jcx4z0xenn3u442ad2xgzu7lp"),
			PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zpy4mst534500z7k6xk5u7c9ex8zs44rjjhmxaxtw9zzjv82qkfhkhx2rfs",
			VotingPower: 1,
		},
		// core-val-5
		{
			Address:     std.Address("g1l072ma0vfhx7s4vpevfvuxd6wzkv5ztt7gh99w"),
			PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqtvz3g6nvu3d6wdz97w7jdw2sjc65us5u8gj8pm4mhasw7zxakjhjn9qkm",
			VotingPower: 1,
		},
		// core-val-6
		{
			Address:     std.Address("g1uwqd3284kuzm56auwyc9d87jf3953tp9pnt506"),
			PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zp8xm09ura7mwyntee78cl64hgzq0x75f05tv7fkxpqvc797j37hsr3vgjg",
			VotingPower: 1,
		},
		// berty-val-1
		{
			Address:     std.Address("g1ut590acnamvhkrh4qz6dz9zt9e3hyu499u0gvl"),
			PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zq2gncppkfzmx7s22mn60mf0uxzzpl23yx97hwmwm8yc6lupepqqnlexfch",
			VotingPower: 1,
		},
		// onbloc-val-1
		{
			Address:     std.Address("g1arkzjfrte9l97v9q2qye07v0lw07039gaa3hfy"),
			PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zpkvdy7n9744qay76fzekpu9l6g3mp4hzhqjmp6k2as72ghlzc546ju3a09",
			VotingPower: 1,
		},
		// onbloc-val-2
		{
			Address:     std.Address("g1x0m33nyne064xdx7tvlfcjwd4xkajjar6h523z"),
			PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zp6s70v4wurhg699w6f9emkwxdlm2eyf2uv64annj47npq85tjeucedmky9",
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
		// devx-val-3
		{
			Address:     std.Address("g1sll4rtvrepdyzcvg5ml0kjtl7fnwgcsxgg9s5q"),
			PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zplr4zg2smgha4n9huwcywm6pnkuny2x2j44kk4srxcf0rrmpql3035k8s2",
			VotingPower: 1,
		},
		// devx-val-4
		{
			Address:     std.Address("g1aa5pp94eaextkump38766hpdra74xtfh805msv"),
			PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqe85el3ardhel5vruywsdjw0vj2zjyhqhsyhcnuh0dy8xhuj8mxjg5h7uw",
			VotingPower: 1,
		},
		// tori-val-1
		{
			Address:     std.Address("g1r2lwzu0y0na4686a0lz4f2zqxlffqkfm7lqqqp"),
			PubKey:      "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zq2quztlp2pffjsun3jeqyesru8rx9yc6tfj9na3hnw9qgn4zlrpul5mhd0",
			VotingPower: 1,
		},
		// aib-val-1
		{
			Address:     std.Address("g1gu6wrz7xcavjtk2dudsfl586qrz5g4ahhhz2j3"),
			PubKey:      "gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pqgruqtkktk0xagtfk8v75c5weuvwf8y5fkdgh9j9cmeh8tfcy6vnxjycdcq",
			VotingPower: 1,
		},
		// aib-val-2
		{
			Address:     std.Address("g1x7rewh0w7u7yrmsmadq6w6t3jwh7ec6ql02klh"),
			PubKey:      "gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pqvgyt5sllr0gqwr0fhr4kacr7p7l239e8cm9p0cyaxtctpqgzdtfcq8us06",
			VotingPower: 1,
		},
		// aib-val-3
		{
			Address:     std.Address("g1l8j7ts0gmghag7zmnatq5ta5xg83ylyxnmaxlh"),
			PubKey:      "gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pq0rwfjzenlzgj73m2ppy6tz427ajukmcj6vj2ttdet22p3pggrdrk0qhpkh",
			VotingPower: 1,
		},
	}

	// The default valset protocol is PoA
	vp = poa.NewPoA(poa.WithInitialSet(set))

	// No changes to apply initially
	changes = avl.NewTree()
}
