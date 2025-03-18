// PKGPATH: gno.land/r/gnoland/valopers_proposal_test
package valopers_proposal_test

// SEND: 100000000ugnot

import (
	"std"

	"gno.land/r/gnoland/valopers"
	"gno.land/r/gnoland/valopers_proposal"
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/init" // so that the govdao initializer is executed
)

const (
	validMoniker     = "test-1"
	validDescription = "test-1's description"
	validAddress     = std.Address("g1sp8v98h2gadm5jggtzz9w5ksexqn68ympsd68h")
	validPubKey      = "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqwpdwpd0f9fvqla089ndw5g9hcsufad77fml2vlu73fk8q8sh8v72cza5p"
)

func init() {
	c := std.OriginCaller()
	init.InitWithUsers(c)

	// Register a validator and add the proposal
	valopers.Register(validMoniker, validDescription, validAddress, validPubKey)
	pr := valopers_proposal.NewValidatorProposalRequest(validAddress)

	std.TestSetOriginCaller(c)
	std.TestSetOriginPkgAddress(std.Address("g1p84dvfh4wrplyxx4zsmy77a8rxcnjup2j7zv5r")) // gno.land/r/gov/dao
	dao.MustCreateProposal(pr)
}

func main() {
	println(dao.Render(""))
}

// Output:
// # Active Proposals:
// ## Proposal with id: 0
// ### Title: Add valoper test-1 to the valset
// 
// ### Proposed by: g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm
//
// Valoper profile: [test-1](/r/gnoland/valopers:g1sp8v98h2gadm5jggtzz9w5ksexqn68ympsd68h)
//
// ## test-1
// test-1's description
//
// - Address: g1sp8v98h2gadm5jggtzz9w5ksexqn68ympsd68h
// - PubKey: gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqwpdwpd0f9fvqla089ndw5g9hcsufad77fml2vlu73fk8q8sh8v72cza5p
//
// [Profile link](/r/demo/profile:u/g1sp8v98h2gadm5jggtzz9w5ksexqn68ympsd68h)
//
//
// This proposal contains the following metadata:
//
// No metadata available.
//
// ### Proposal Status:
//
// - **Proposal open for votes**
// - Allowed tiers to vote: T1 T2 T3
// - YES PERCENT: 0%
// - NO PERCENT: 0%
// - [Go to votes list](:0/votes).
//
//
