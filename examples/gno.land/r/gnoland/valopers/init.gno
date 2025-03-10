package valopers

import (
	"gno.land/p/demo/avl"
	"gno.land/p/moul/authz"
)

func init() {
	valopers = avl.NewTree()

	auth = authz.NewWithAuthority(
		authz.NewContractAuthority(
			"gno.land/r/gov/dao/v2",
			func(title string, action authz.PrivilegedAction) error {
				return action()
			},
		),
	)

	instructions = `
# Welcome to the **Valopers** realm

## 📌 Purpose of this Contract

The **Valopers** contract is designed to maintain a registry of **validator profiles**. This registry provides essential information to **GovDAO members**, enabling them to make informed decisions when voting on the inclusion of new validators into the **valset**.

By registering your validator profile, you contribute to a transparent and well-informed governance process within **gno.land**.

---

## 📝 How to Register Your Validator Node

To add your validator node to the registry, use the **Register** function with the following parameters:

- **Moniker**: A human-readable name for your validator.
- **Description**: A detailed introduction in Markdown, including answers to the questions below.
- **Validator Address**: The address of your validator node.
- **Validator Public Key**: Your validator node’s public key.
- **P2P Addresses**: A comma-separated list of peer-to-peer addresses in the format **[nodeID]@[listeningAddress]:[port]**

### ✍️ Required Information for the Description

Please provide detailed answers to the following questions to ensure transparency and improve your chances of being accepted:

1. Which networks are you currently validating?
2. Provide links to your **digital presence** (website, social media, contact details, etc.).
3. What assets do you manage?
4. Why are you interested in validating on **gno.land**?
5. What contributions have you made or are willing to make to **gno.land**?

---

## 🔄 Updating Your Validator Information

After registration, you can update your validator details using the **update functions** provided by the contract.

---

## 📢 Submitting a Proposal to Join the Validator Set

Once you're satisfied with your **valoper** profile, you can propose your inclusion into the validator set. To do so, execute the following function: **r/gnoland/valopers_proposal.MakeProposal**

This will initiate a governance process where **GovDAO** members will vote on your proposal.

---

🚀 **Register now and become a part of gno.land’s validator ecosystem!**

---

`
}
