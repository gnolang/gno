package keyderive

import (
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"
)

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
		{inputKey: "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zp2fyjt37gl0h9pel3tgwwlfcynzllk7gmja2s8khaah5z7ars7qlhc9874", expected: "g19355aeqzalykcamgk83dzz64mj0sp0ydmnngnp"},
		{inputKey: "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqwpdwpd0f9fvqla089ndw5g9hcsufad77fml2vlu73fk8q8sh8v72cza5p", expected: "g1sp8v98h2gadm5jggtzz9w5ksexqn68ympsd68h"},
		{inputKey: "gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pq0skzdkmzu0r9h6gny6eg8c9dc303xrrudee6z4he4y7cs5rnjwmyf40yaj", expected: "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"},
	}

	for _, tc := range testCases {
		derivedAddress, err := Pubkey2Address(tc.inputKey)
		if err != nil {
			panic(err)
		}

		if derivedAddress.String() != tc.expected {
			t.Fatalf("Expected: %t, got: %t", tc.expected, derivedAddress)
		}
	}
}

func TestAddressMatchesKey(t *testing.T) {
	type test struct {
		inputKey    string
		address     string
		shouldMatch bool
	}

	testCases := []test{
		{
			inputKey:    "gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pqgfakcapzvuux7vug9vagyjmvj6k2w2xu723qnue69z9r7wlt9c6zeyxvt8",
			address:     "g14vxq5e5pt5sev7rkz2ej438scmxtylnzv5vnkw",
			shouldMatch: true,
		},
		{
			inputKey:    "gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pqt3gjrvnd3qslct7hwy9e7uu5vje95tcedhvls3hw4prrg7z8hdzwt7a930",
			address:     "g1s3ngxp56v8w230qp4p9qf7memq6ra2x4x3zqsn",
			shouldMatch: true,
		},
		{
			inputKey:    "gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pqt3gjrvnd3qslct7hwy9e7uu5vje95tcedhvls3hw4prrg7z8hdzwt7a930",
			address:     "g1pwxuhltfqxcumjmuquuue6y3f2g3f2d0rcq52x",
			shouldMatch: false,
		},
		{
			inputKey:    "gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pqt3gjrvnd3qslct7hwy9e7uu5vje95tcedhvls3hw4prrg7z8hdzwt7a930",
			address:     "g173vekmq05he23yx9w8028nr4uqujuc3049fxlq",
			shouldMatch: false,
		},
	}

	for _, tc := range testCases {
		match, err := AddressMatchesKey(tc.address, tc.inputKey)
		if err != nil {
			panic(err)
		}

		uassert.Equal(t, tc.shouldMatch, match, ufmt.Sprintf("unexpected result got %v, want %v", tc.shouldMatch, match))
	}
}
