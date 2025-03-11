// PKGPATH: gno.land/r/gnoland/valopers_test
package valopers_test

// SEND: 20000000ugnot

import (
	"std"
	"strings"

	"gno.land/p/demo/ufmt"
	"gno.land/r/gnoland/valopers"
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
}

func main() {
	// Simulate clicking on the validator
	println(valopers.Render(validAddress.String()))
}

// Output:
// Valoper's details:
// ## test-1
// test-1's description
//
// - Address: g1sp8v98h2gadm5jggtzz9w5ksexqn68ympsd68h
// - PubKey: gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqwpdwpd0f9fvqla089ndw5g9hcsufad77fml2vlu73fk8q8sh8v72cza5p
// - P2P Addresses: [
//
// 	- g1sp8v98h2gadm5jggtzz9w5ksexqn68ympsd68h@tcp://0.0.0.0:26656,
// 	- g1sp8v98h2gadm5jggtzz9w5ksexqn68ympsd68h@udp://192.168.1.1:8080
//
// ]
//
// [Profile link](/r/demo/profile:u/g1sp8v98h2gadm5jggtzz9w5ksexqn68ympsd68h)
//
