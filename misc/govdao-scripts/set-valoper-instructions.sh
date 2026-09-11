#!/usr/bin/env bash
# Update the r/gnops/valopers registration instructions via govDAO proposal.
#
# Pushes the embedded registration guide from gnolang/gno PR #5842 to the
# live realm, then creates the govDAO proposal, votes YES, and executes it.
# The realm's instructions string is set at genesis by init.gno, so an
# already-deployed realm can only be updated through this governance path.
#
# The Go program below reconstructs the instructions exactly as PR #5842's
# init.gno does (same txlink link construction), so the on-chain value
# matches the PR byte-for-byte. PR #5842 is the source of truth — init.gno
# on this branch still carries the pre-PR text.
#
# The proposal stores the string twice (realm variable + proposal
# description) and it is large, so if the tx runs out of gas, raise
# GAS_WANTED.
#
# Usage:
#   ./set-valoper-instructions.sh
#
# Environment: see README.md.
set -eo pipefail

GNOKEY_NAME="${GNOKEY_NAME:?GNOKEY_NAME is required}"
CHAIN_ID="${CHAIN_ID:?CHAIN_ID is required}"
REMOTE="${REMOTE:?REMOTE is required}"
GAS_WANTED="${GAS_WANTED:-50000000}"
GAS_FEE="${GAS_FEE:-1000000ugnot}"

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

# Quoted heredoc: the Go raw string contains backticks, so the body must
# reach gnokey verbatim without shell expansion.
cat >"$TMPDIR/set_instructions.gno" <<'GOEOF'
package main

import (
	"gno.land/p/moul/txlink"
	"gno.land/r/gnops/valopers/proposal"
	"gno.land/r/gov/dao"
)

func main(cur realm) {
	newInstructions := `
# Welcome to the **Valopers** realm

## 📌 Purpose of this Contract

The **Valopers** contract is designed to maintain a registry of **validator profiles**. This registry provides essential information to **GovDAO members**, enabling them to make informed decisions when voting on the inclusion of new validators into the **valset**.

By registering your validator profile, you contribute to a transparent and well-informed governance process within **gno.land**.

---

## 📝 How to Register Your Validator Node

To add your validator node to the registry, use the [**Register**](` + txlink.Realm("gno.land/r/gnops/valopers").Call("Register") + `) function with the following parameters:

- **Moniker** (Validator Name)
  - Must be **human-readable**
  - **Max length**: **32 characters**
  - **Allowed characters**: Letters, numbers, spaces, hyphens (**-**), and underscores (**_**)
  - **No special characters** at the beginning or end

- **Description** (Introduction & Validator Details)
  - **Max length**: **2048 characters**
  - Must include answers to the questions listed below

- **Server Type** (Infrastructure Type)
  - Must be one of the following values:
    - **cloud**: For validators running on cloud infrastructure (AWS, GCP, Azure, etc.)
    - **on-prem**: For validators running on on-premises infrastructure
    - **data-center**: For validators running in dedicated data centers

- **Operator Address**
  - The ` + "`g1...`" + ` address of your operator account (from your ` + "`gnokey`" + ` keyring)
  - **Must be controlled by the signer** of this transaction — the realm rejects the call if the signer doesn't control that address

- **Validator Consensus Public Key**
  - Your validator node's consensus public key, in the ` + "`gpub1...`" + ` format
  - Retrieve it by running: ` + "`gnoland secrets get validator_key`" + `

### ✍️ Required Information for the Description

Please provide detailed answers to the following questions to ensure transparency and improve your chances of being accepted:

1. The name of your validator
2. Networks you are currently validating and your total AuM (assets under management)
3. Links to your **digital presence** (website, social media, etc.). Please include your Discord handle to be added to our main comms channel, the gno.land valoper Discord channel.
4. Contact details
5. Why are you interested in validating on **gno.land**?
6. What contributions have you made or are willing to make to **gno.land**?

---

## 🔄 Updating Your Validator Information

After registration, you can update your validator details using the **update functions** provided by the contract.

---

## 📢 Submitting a Proposal to Join the Validator Set

Once you're satisfied with your **valoper** profile, you need to notify GovDAO; only a GovDAO member can submit a proposal to add you to the validator set.

If you are a GovDAO member, you can nominate yourself by executing the following function: [**r/gnops/valopers/proposal.ProposeNewValidator**](` + txlink.Realm("gno.land/r/gnops/valopers/proposal").Call("ProposeNewValidator") + `)

This will initiate a governance process where **GovDAO** members will vote on your proposal.

---

🚀 **Register now and become a part of gno.land’s validator ecosystem!**

Read more: [How to become a testnet validator](https://gnops.io/articles/guides/become-testnet-validator/) <!-- XXX: replace with a r/gnops/blog:xxx link -->

Disclaimer: Please note, registering your validator profile and/or validating on testnets does not guarantee a validator slot on the gno.land beta mainnet. However, active participation and contributions to testnets will help establish credibility and may improve your chances for future validator acceptance. The initial validator amount and valset will ultimately be selected through GovDAO governance proposals and acceptance.

---

`
	pr := proposal.ProposeNewInstructionsProposalRequest(cross(cur), newInstructions)
	pid := dao.MustCreateProposal(cross(cur), pr)
	dao.MustVoteOnProposalSimple(cross(cur), int64(pid), "YES")
	dao.ExecuteProposal(cross(cur), pid)
}
GOEOF

echo "Updating r/gnops/valopers instructions via govDAO proposal"
echo "  Source: gnolang/gno PR #5842"
echo "  Key:    ${GNOKEY_NAME}"
echo "  Chain:  ${CHAIN_ID}"
echo "  Remote: ${REMOTE}"
echo ""

gnokey maketx run \
  -gas-wanted "$GAS_WANTED" \
  -gas-fee "$GAS_FEE" \
  -broadcast \
  -chainid "$CHAIN_ID" \
  -remote "$REMOTE" \
  "$GNOKEY_NAME" \
  "$TMPDIR/set_instructions.gno"

echo ""
echo "Done — valoper instructions updated."
