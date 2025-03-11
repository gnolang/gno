// PKGPATH: gno.land/r/gnoland/valopers_proposal_test
package valopers_proposal_test

// SEND: 20000000ugnot

import (
	"std"
	"strings"

	"gno.land/p/demo/ufmt"
	"gno.land/r/gnoland/valopers"
	"gno.land/r/gnoland/valopers_proposal"
	_ "gno.land/r/gov/dao/init" // so that the govdao initializer is executed
	govdao "gno.land/r/gov/dao/v2"
)

const (
	validMoniker     = "test-1"
	validDescription = "test-1's description"
	validAddress     = std.Address("g1sp8v98h2gadm5jggtzz9w5ksexqn68ympsd68h")
	validPubKey      = "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqwpdwpd0f9fvqla089ndw5g9hcsufad77fml2vlu73fk8q8sh8v72cza5p"
)

var (
	validListeningAddresses = []string{
		"tcp://0.0.0.0:26656",
		"udp://192.168.1.1:8080",
		"http://example.com:80",
		"https://secure.com:443",
		"ws://localhost:8080",
		"without-protocol.com:443",
	}

	validP2PAddresses = []string{ufmt.Sprintf("%s@%s", validAddress, validListeningAddresses[0]), ufmt.Sprintf("%s@%s", validAddress, validListeningAddresses[1])}
)

func init() {
	// Register a validator and add the proposal
	valopers.Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
	valopers_proposal.MakeProposal(validAddress)
}

func main() {
	println(govdao.Render(""))
}

// Output:
// # GovDAO Proposals
//
// ## [Prop #0 - Add valoper test-1 (Address: g1sp8v98h2g...](/r/gov/dao/v2:0)
//
// **Status: ACTIVE**
//
// **Author: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm**
//
//
