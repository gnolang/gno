---
id: validators-faq
---

# FAQ

### What is a Gno.land validator?

Gno.land is based on [Tendermint2](https://docs.gno.land/concepts/tendermint2) that relies on a set of validators selected based on [Proof of Contribution](https://docs.gno.land/concepts/proof-of-contribution) (PoC) to secure the network. Validators are tasked with participating in consensus by committing new blocks and broadcasting votes. Validators are compensated with a portion of transaction fees generated in the network. In Gno.land, the voting power of all validators are equally weighted to achieve a high nakamoto coefficient and fairness.

### What is Tendermint2?

[Tendermint2](https://docs.gno.land/concepts/tendermint2) (TM2) is the consensus protocol that powers Gno.land. TM2 is a successor of [Tendermint Core](https://github.com/tendermint/tendermint2), a de facto consensus framework for building Proof of Stake blockchains. The design philosophy of TM2 is to create “complete software” without any vulnerabilities with development focused on minimalism, dependency removal, and modularity.

### What is Proof of Contribution?

[Proof of Contribution](https://docs.gno.land/concepts/proof-of-contribution) (PoC) is a novel consensus mechanism that secures Gno.land. PoC weighs expertise and alignment with the project to evaluate the contribution of individuals or teams who govern and operate the chain. Unlike Proof of Stake (PoS), validators are selected via governance of Contributors based on their reputation and technical proficiency. The voting power of the network is equally distributed across all validators for higher decentralization. A portion of all transaction fees paid to the network are evenly shared between all validators to provide a fair incentive structure.

### How does Gno.land differ from the Cosmos Hub?

In Cosmos Hub, validators are selected based on the amount of staked `ATOM` tokens delegated. This means that anyone with enough capital can join as a validator only to seek economic incentives without any alignment or technical expertise. This system leads to an undesirable incentive structure in which validators are rewarded purely based on the capital delegated, regardless of the quality of their infrastructure or service. On the contrary, validators in Gno.land are reviewed and verified to have made significant contributions, in order to join the validator set. All validators are evenly rewarded to ensure that the entire validator set is fairly incentivized.

### What is a full node and a pruned node?

A full node fully validates transactions and blocks of a blockchain and keeps a full record of all historic activity. A pruned node is a lighter node that processes only block headers and does not keep all historical data of the blockchain post-verification. Pruned nodes are less resource intensive in terms of storage costs. Although validators may run either a full node or a pruned node, it is important to retain enough blocks to be able to validate new blocks.

### How do I join the testnet as a validator?

Out of many official Gno testnets, Testnet4 (`test4`) is the purpose-built network for testing the multi-node validator environment prior to mainnet launch. Testnet4 is scheduled to go live in Q2 2024 with genesis validators consisting of the Gno Core Team, partners, and external contributors.

For more information about joining testnet4, visit (add link to the criteria issue/discussion). For more information about different testnets, visit [the relevant issue](https://github.com/gnolang/hackerspace/issues/69).

### What are the incentives for running a validator?

Network transaction fees paid on the Gno.land in `GNOT` are collected, from which a portion is directed to reward validators for their work. All validators fairly receive an equal amount of rewards.

### What are the different types of keys?

1. **Tendermint ( Tendermint2 ) Key :** A unique key used for voting in consensus during creation of blocks. A Tendermint Key is also often called a Validator Key. It is automatically created when running the `gnoland secrets init` command. A validator may check their Tendermint Key by running the `gnoland secrets get ValidatorPrivateKey` command.

2. **User-owned keys :** A key that is generated when a new account is created using the `gnokey` command. It is used to sign transactions.

### What stage is the Gno.land project in?

Gno.land is currently in Testnet 3, the single-node testnet stage. The next version, Testnet 4, is scheduled to go live in Q2 2024, which will include a validator set implementation for a multinode environment.

### How many validators will there be in mainnet?

The exact plans for mainnet are still TBD. Based on the latest discussions between contributors, the mainnet will likely have a validator set size of 20~50, which will gradually scale with the development and decentralization of the Gno.land project.

### How do I make my first contribution?

Gno.land is in active development and external contributions are always welcome! If you’re looking for tasks to begin with, we suggest you visit the [Bounties & Worx](https://github.com/orgs/gnolang/projects/35/views/3) board and search for open tasks up for grabs. Start from small challenges and work your way up to the bigger ones. Every contribution is acknowledged and highly regarded in PoC. We look forward to having you onboard as a new Contributor!
