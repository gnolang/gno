// PKGPATH: gno.land/r/gnoland/valopers_proposal_test
package valopers_proposal_test

// SEND: 20000000ugnot

import (
	"std"
	"testing"

	"gno.land/r/gnoland/valopers"
	"gno.land/r/gnoland/valopers_proposal"
	"gno.land/r/gov/dao"
	"gno.land/r/gov/dao/v3/init"
)

const (
	validMoniker     = "test-1"
	validDescription = "test-1's description"
	validAddress     = std.Address("g1sp8v98h2gadm5jggtzz9w5ksexqn68ympsd68h")
	otherAddress     = std.Address("g1juz2yxmdsa6audkp6ep9vfv80c8p5u76e03vvh")
	validPubKey      = "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqwpdwpd0f9fvqla089ndw5g9hcsufad77fml2vlu73fk8q8sh8v72cza5p"
)

func init() {
	c := std.OriginCaller()
	init.InitWithUsers(c)
}

func main() {
	// Register a validator
	valopers.Register(validMoniker, validDescription, validAddress, validPubKey)
	// Try to make a proposal for a non-existing validator
	defer func() {
		r := recover()
		println("r: ", r)
	}()
	pr := valopers_proposal.NewValidatorProposalRequest(otherAddress)

	testing.SetOriginPkgAddress(std.Address("g1p84dvfh4wrplyxx4zsmy77a8rxcnjup2j7zv5r")) // gno.land/r/gov/dao
	dao.MustCreateProposal(pr)
}

// Output:
// r:  valoper does not exist
