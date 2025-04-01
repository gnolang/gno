package std

import (
	"testing"
)

func TestValid(t *testing.T) {
	type test struct {
		inputAddress Address
		expected     bool
	}

	testCases := []test{
		{inputAddress: "g1f4v282mwyhu29afke4vq5r2xzcm6z3ftnugcnv", expected: true},
		{inputAddress: "g127jydsh6cms3lrtdenydxsckh23a8d6emqcvfa", expected: true},
		{inputAddress: "g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq", expected: true},
		{inputAddress: "g14da4n9hcynyzz83q607uu8keuh9hwlv42ra6fa", expected: true},

		// Bech32 doesn't allow '1', 'b', 'i', 'o' for data part
		//
		// 	https://github.com/bitcoin/bips/blob/master/bip-0173.mediawiki#bech32
		{inputAddress: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5", expected: true},
		{inputAddress: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf1", expected: false},
		{inputAddress: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqfb", expected: false},
		{inputAddress: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqfi", expected: false},
		{inputAddress: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqfo", expected: false},

		{inputAddress: "g1u0000000000000000000000000000000000000", expected: false},
		{inputAddress: "", expected: false},
		{inputAddress: "000000000000", expected: false},
		{inputAddress: "0000000000000000000000000000000000000000000000000000000000000000000000", expected: false},
	}

	for _, tc := range testCases {
		result := tc.inputAddress.IsValid()

		if result != tc.expected {
			t.Fatalf("Expected: %t, got: %t", tc.expected, result)
		}
	}
}

func TestDerivePkgAddr(t *testing.T) {
	type test struct {
		inputPath string
		expected  string
	}

	testCases := []test{
		{inputPath: "gno.land/r/gnoland/faucet", expected: "g1ttrq7mp4zy6dssnmgyyktnn4hcj3ys8xhju0n7"},
		{inputPath: "gno.land/r/demo/tamagotchi", expected: "g1a3tu874agjlkrpzt9x90xv3uzncapcn959yte4"},
	}

	for _, tc := range testCases {
		result := DerivePkgAddr(tc.inputPath)
		if result.String() != tc.expected {
			t.Fatalf("Expected: %t, got: %t", tc.expected, result)
		}
	}
}

func TestPubkeyToAddress(t *testing.T) {
	type test struct {
		inputKey string
		expected string
	}

	testCases := []test{
		{inputKey: "gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pqgfakcapzvuux7vug9vagyjmvj6k2w2xu723qnue69z9r7wlt9c6zeyxvt8", expected: "g14vxq5e5pt5sev7rkz2ej438scmxtylnzv5vnkw"},
		{inputKey: "gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pqt3gjrvnd3qslct7hwy9e7uu5vje95tcedhvls3hw4prrg7z8hdzwt7a930", expected: "g1s3ngxp56v8w230qp4p9qf7memq6ra2x4x3zqsn"},
		{inputKey: "gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pqf4s3e9agxz20lqvuqhkwp6xcl0pdmtw907axyxxv4ntnx865u6zvc46z77", expected: "g1cfg8z8fjap628nsh5tf2eu73x7qsq7svemfexr"},
		{inputKey: "gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pqfuuq0n9ytg4f6ryhpkwrm8c8255vgjyua9s7v07nhemfyywf7nz68axq0m", expected: "g1gwdx26dksjwpkhg37j2cgmlyfhv2nfzdut7ae2"},
	}

	for _, tc := range testCases {

		address, err := PubkeyToAddress(tc.inputKey)
		if err != nil {
			panic(err)
		}

		if address.String() != tc.expected {
			t.Fatalf("Expected: %t, got: %t", tc.expected, address)
		}
	}
}
