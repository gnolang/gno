// PKGPATH: gno.land/r/gnoland/valopers_test
// SEND: 20000000ugnot

package valopers_test

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/r/gnoland/valopers"
)

var g1user = testutils.TestAddress("g1user") // g1vuch2um9wf047h6lta047h6lta047h6l2ewm6w

const (
	validMoniker     = "test-1"
	validDescription = "test-1's description"
	validAddress     = std.Address("g1sp8v98h2gadm5jggtzz9w5ksexqn68ympsd68h")
	validPubKey      = "gpub1pggj7ard9eg82cjtv4u52epjx56nzwgjyg9zqwpdwpd0f9fvqla089ndw5g9hcsufad77fml2vlu73fk8q8sh8v72cza5p"
)

func init() {
	testing.SetOriginCaller(g1user)

	// Register a validator and add the proposal
	valopers.Register(cross, validMoniker, validDescription, validAddress, validPubKey)
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
// To add your validator node to the registry, use the [**Register**](/r/gnoland/valopers$help&func=Register) function with the following parameters:
//
// - **Moniker** (Validator Name)
//   - Must be **human-readable**
//   - **Max length**: **32 characters**
//   - **Allowed characters**: Letters, numbers, spaces, hyphens (**-**), and underscores (**_**)
//   - **No special characters** at the beginning or end
//
// - **Description** (Introduction & Validator Details)
//   - **Max length**: **2048 characters**
//   - Must include answers to the questions listed below
//
// - **Validator Address**
//   - Your validator node’s address
//
// - **Validator Public Key**
//   - Your validator node’s public key
//
// ### ✍️ Required Information for the Description
//
// Please provide detailed answers to the following questions to ensure transparency and improve your chances of being accepted:
//
// 1. The name of your validator
// 2. Networks you are currently validating and your total AuM (assets under management)
// 3. Links to your **digital presence** (website, social media, etc.). Please include your Discord handle to be added to our main comms channel, the gno.land valoper Discord channel.
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
// Once you're satisfied with your **valoper** profile, you need to notify GovDAO; only a GovDAO member can submit a proposal to add you to the validator set.
//
// If you are a GovDAO member, you can nominate yourself by executing the following function: [**r/gnoland/valopers_proposal.ProposeNewValidator**](/r/gnoland/valopers_proposal$help&func=ProposeNewValidator)
//
// This will initiate a governance process where **GovDAO** members will vote on your proposal.
//
// ---
//
// 🚀 **Register now and become a part of gno.land’s validator ecosystem!**
//
// Read more: [How to become a testnet validator](https://gnops.io/articles/guides/become-testnet-validator/) <!-- XXX: replace with a r/gnops/blog:xxx link -->
//
// Disclaimer: Please note, registering your validator profile and/or validating on testnets does not guarantee a validator slot on the gno.land beta mainnet. However, active participation and contributions to testnets will help establish credibility and may improve your chances for future validator acceptance. The initial validator amount and valset will ultimately be selected through GovDAO governance proposals and acceptance.
//
// ---
//
//
//
//  * [test-1](/r/gnoland/valopers:g1sp8v98h2gadm5jggtzz9w5ksexqn68ympsd68h) - [profile](/r/demo/profile:u/g1vuch2um9wf047h6lta047h6lta047h6l2ewm6w)
//
//
