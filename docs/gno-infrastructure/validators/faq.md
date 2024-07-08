---
id: validators-faq
---

# Validators FAQ

## General Concepts

### What is a Gno.land validator?

Gno.land is based on [Tendermint2](https://docs.gno.land/concepts/tendermint2) that relies on a set of validators
selected based on [Proof of Contribution](https://docs.gno.land/concepts/proof-of-contribution) (PoC) to secure the
network. Validators are tasked with participating in consensus by committing new blocks and broadcasting votes.
Validators are compensated with a portion of transaction fees generated in the network. In Gno.land, the voting power of
all validators are equally weighted to achieve a high nakamoto coefficient and fairness.

### What is Tendermint2?

[Tendermint2](https://docs.gno.land/concepts/tendermint2) (TM2) is the consensus protocol that powers Gno.land. TM2 is a
successor of [Tendermint Core](https://github.com/tendermint/tendermint2), a de facto consensus framework for building
Proof of Stake blockchains. The design philosophy of TM2 is to create “complete software” without any vulnerabilities
with development focused on minimalism, dependency removal, and modularity.

### What is Proof of Contribution?

[Proof of Contribution](https://docs.gno.land/concepts/proof-of-contribution) (PoC) is a novel consensus mechanism that
secures Gno.land. PoC weighs expertise and alignment with the project to evaluate the contribution of individuals or
teams who govern and operate the chain. Unlike Proof of Stake (PoS), validators are selected via governance of
Contributors based on their reputation and technical proficiency. The voting power of the network is equally distributed
across all validators for higher decentralization. A portion of all transaction fees paid to the network are evenly
shared between all validators to provide a fair incentive structure.

### How does Gno.land differ from the Cosmos Hub?

In Cosmos Hub, validators are selected based on the amount of staked `ATOM` tokens delegated. This means that anyone
with enough capital can join as a validator only to seek economic incentives without any alignment or technical
expertise. This system leads to an undesirable incentive structure in which validators are rewarded purely based on the
capital delegated, regardless of the quality of their infrastructure or service.

On the contrary, validators in Gno.land must be reviewed and verified to have made significant contributions in order to
join the validator set. This property resembles the validator selection mechanism
in [Proof of Authority](https://openethereum.github.io/Proof-of-Authority-Chains). Furthermore, all validators are
evenly rewarded to ensure that the entire validator set is fairly incentivized to ensure the sustainability of the
network.

### What stage is the Gno.land project in?

Gno.land is currently in Testnet 3, the single-node testnet stage. The next version, Testnet 4, is scheduled to go live
in Q3 2024, which will include a validator set implementation for a multinode environment.

## Becoming a Validator

### How do I join the testnet as a validator?

Out of many official Gno testnets, Testnet4 (`test4`) is the purpose-built network for testing the multi-node validator
environment prior to mainnet launch. Testnet4 is scheduled to go live in Q3 2024 with genesis validators consisting of
the Gno Core Team, partners, and external contributors.

For more information about joining testnet4,
visit [the relevant issue](https://github.com/gnolang/hackerspace/issues/69). For more information about different
testnets, visit [Gno Testnets](https://docs.gno.land/concepts/testnets).

### What are the incentives for running a validator?

Network transaction fees paid on the Gno.land in `GNOT` are collected, from which a portion is directed to reward
validators for their work. All validators fairly receive an equal amount of rewards.

### How many validators will there be in mainnet?

The exact plans for mainnet are still TBD. Based on the latest discussions between contributors, the mainnet will likely
have an inital validator set size of 20~50, which will gradually scale with the development and decentralization of the
Gno.land project.

### How do I make my first contribution?

Gno.land is in active development and external contributions are always welcome! If you’re looking for tasks to begin
with, we suggest you visit
the [Bounties &](https://github.com/orgs/gnolang/projects/35/views/3) [Worx](https://github.com/orgs/gnolang/projects/35/views/3)
board and search for open tasks up for grabs. Start from small challenges and work your way up to the bigger ones. Every
contribution is acknowledged and highly regarded in PoC. We look forward to having you onboard as a new Contributor!

## Technical Guides

### What are the different types of keys?

1. **Tendermint ( Tendermint2 ) Key :** A unique key used for voting in consensus during creation of blocks. A
   Tendermint Key is also often called a Validator Key. It is automatically created when running
   the `gnoland secrets init` command. A validator may check their Tendermint Key by running
   the `gnoland secrets get validator_key` command.

2. **User-owned keys :** A key that is generated when a new account is created using the `gnokey` command. It is used to
   sign transactions.

3. **Node Key :** A key used for communicating with other nodes. It is automatically created when running
   the `gnoland secrets init` command. A validator may check their Node Key by running the `gnoland secrets get node_id`
   command.

### What is a full node and a pruned node?

A full node fully validates transactions and blocks of a blockchain and keeps a full record of all historic activity. A
pruned node is a lighter node that processes only block headers and does not keep all historical data of the blockchain
post-verification. Pruned nodes are less resource intensive in terms of storage costs. Although validators may run
either a full node or a pruned node, it is important to retain enough blocks to be able to validate new blocks.

## Technical References

### How do I generate `genesis.json`?

`genesis.json` is the file that is used to create the initial state of the chain. To generate `genesis.json`, use
the `gnoland genesis generate` command. Refer
to [this section](../../gno-tooling/cli/gnoland.md#gnoland-genesis-generate-flags) for various flags that allow you to
manipulate the file.

:::warning

Editing generated genesis.json manually is extremely dangerous. It may corrupt chain initial state which leads chain to
not start

:::

### How do I add or remove validators from `genesis.json`?

Validators inside `genesis.json` will be included in the validator set at genesis. To manipulate the genesis validator
set, use the `gnoland genesis validator` command with the `add` or `remove` subcommands. Refer
to [this section](../../gno-tooling/cli/gnoland.md#gnoland-genesis-validator-flags) for flags that allow you to
configure the name or the voting power of the validator.

### How do I add the balance information to the `genesis.json`?

You may premine coins to various addresses. To modify the balances of addresses at genesis, use
the `gnoland genesis balances` command with the `add` or `remove` subcommands. Refer
to [this section](../../gno-tooling/cli/gnoland.md#gnoland-genesis-balances-add-flags) for various flags that allow you
to update the entire balance sheet with a file or modify the balance of a single address.

:::info

Not only `ugnot`, but other coins are accepted. However, be aware that coins other than `ugnot` may not work(send, and
etc.) properly.

:::

### How do I initialize `gno secrets`?

The `gno secrets init` command allows you to initialize the private information required to run the validator, including
the validator node's private key, the state, and the node ID. Refer
to [this section](../../gno-tooling/cli/gnoland.md#gnoland-secrets-init-flags) for various flags that allow you to
define the output directory or to overwrite the existing secrets.

### How do I get `gno secrets`?

To retrieve the private information of your validator node, use the `gnoland-secrets-get` command. Refer
to [this section](../../gno-tooling/cli/gnoland.md#gnoland-secrets-get-flags) for a flag that allows you to define the
output directory.

### How do I initialize the gno node configurations?

To initialize the configurations required to run a node, use the `gnoland config init` command. Refer
to [this section](../../gno-tooling/cli/gnoland.md#gnoland-config-init-flags) for various flags that allow you to define
the path or to overwrite the existing configurations.

### How do I get the current gno node configurations?

To retrieve the specific values the current gno node configurations, use the `gnoland config get` command. Refer
to [this section](../../gno-tooling/cli/gnoland.md#gnoland-config-get) for a flag that allows you to define the path to
the configurations file.

### How do I edit the gno node configurations?

To edit the specific value of gno node configurations, use the `gnoland-config set` command. Refer
to [this section](../../gno-tooling/cli/gnoland.md#gnoland-config-set) for a flag that allows you to define the path to
the configurations file.

### How do I initialize and start a new gno chain?

To start an independent gno chain, follow the initialization process available
in [this section](./setting-up-a-new-chain.md).

### How do I connect to an existing gno chain?

To join the validator set of a gno chain, you must first establish a connection. Refer
to [this section](./connect-to-existing-chain.md) for a step-by-step guide on how to connect to an existing gno
chain.
