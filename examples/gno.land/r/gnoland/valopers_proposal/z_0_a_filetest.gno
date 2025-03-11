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
)

const (
	validMoniker     = "test-1"
	validDescription = "test-1's description"
	validAddress     = std.Address("g1sp8v98h2gadm5jggtzz9w5ksexqn68ympsd68h")
	otherAddress     = std.Address("g1juz2yxmdsa6audkp6ep9vfv80c8p5u76e03vvh")
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
}

func main() {
	// Register a validator
	valopers.Register(validMoniker, validDescription, validAddress, validPubKey, strings.Join(validP2PAddresses, ","))
	// Try to make a proposal for a non-existing validator
	defer func() {
		r := recover()
		println("r: ", r)
	}()
	valopers_proposal.MakeProposal(otherAddress)
}

// Output:
// r:  valoper does not exist
