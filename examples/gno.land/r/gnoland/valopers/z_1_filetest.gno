// PKGPATH: gno.land/r/gnoland/valopers_test
package valopers_test

// SEND: 20000000ugnot

import (
	"std"

	"gno.land/r/gnoland/valopers"
)

const (
	validMoniker     = "test-1"
	validDescription = "test-1's description"
	validAddress     = std.Address("g1sp8v98h2gadm5jggtzz9w5ksexqn68ympsd68h")
	validPubKey      = "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqwpdwpd0f9fvqla089ndw5g9hcsufad77fml2vlu73fk8q8sh8v72cza5p"
)

func init() {
	// Register a validator and add the proposal
	valopers.Register(validMoniker, validDescription, validAddress, validPubKey)
}

func main() {
	println(valopers.Render(""))
}

// Output:
//
// # Welcome to the **Valopers** realm
//
// ## 📌 Purpose of this Contract
//
// The **Valopers** contract is designed to maintain a registry of **validator profiles**. This registry provides essential information to **GovDAO members**, enabling them to make informed decisions when voting on the inclusion of new validators into the **valset**.
//
// By registering your validator profile, you contribute to a transparent and well-informed governance process within **gno.land**.
//
// ---
//
// ## 📝 How to Register Your Validator Node
//
// To add your validator node to the registry, use the **Register** function with the following parameters:
//
// - **Moniker**: A human-readable name for your validator.
// - **Description**: A detailed introduction in Markdown, including answers to the questions below.
// - **Validator Address**: The address of your validator node.
// - **Validator Public Key**: Your validator node’s public key.
//
// ### ✍️ Required Information for the Description
//
// Please provide detailed answers to the following questions to ensure transparency and improve your chances of being accepted:
//
// 1. The name of your validator
// 2. Networks you are currently validating and your total AuM (assets under management)
// 3. Links to your **digital presence** (website, social media, etc.).
// 4. Contact details
// 5. Why are you interested in validating on **gno.land**?
// 6. What contributions have you made or are willing to make to **gno.land**?
//
// ---
//
// ## 🔄 Updating Your Validator Information
//
// After registration, you can update your validator details using the **update functions** provided by the contract.
//
// ---
//
// ## 📢 Submitting a Proposal to Join the Validator Set
//
// Once you're satisfied with your **valoper** profile, you can propose your inclusion into the validator set. To do so, execute the following function: **r/gnoland/valopers_proposal.MakeProposal**
//
// This will initiate a governance process where **GovDAO** members will vote on your proposal.
//
// ---
//
// 🚀 **Register now and become a part of gno.land’s validator ecosystem!**
//
// ---
//
//
//
//  * [test-1](/r/gnoland/valopers:g1sp8v98h2gadm5jggtzz9w5ksexqn68ympsd68h) - [profile](/r/demo/profile:u/g1p9elmfxvctlkypargf7wruch5vchuysqr2xg2q)
//
//
