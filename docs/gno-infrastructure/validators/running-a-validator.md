---
id: validators-running-a-validator
---

# Running a Validator

## Becoming a Gno.land validator

The Gno.land blockchain is powered by the [Tendermint2](https://docs.gno.land/concepts/tendermint2) (TM2) consensus, which involves committing of new blocks and broadcasting votes by multiple validators selected via governance in [Proof of Contribution](https://docs.gno.land/concepts/proof-of-contribution) (PoC). While traditional Proof of Stake (PoS) blockchains such as the Cosmos Hub required validators to secure a delegation of staked tokens to join the validator set, no bonding of capital is involved in Gno.land. Rather, the validators on Gno.land are expected to demonstrate their technical expertise and alignment with the project by making continuous, meaningful contributions to the project. Furthermore, the voting power and the transaction fee rewards between validators are distributed evenly to achieve higher decentralization. From a technical perspective, the validator set implementation in Gno.land as its abstracted away into the `r/sys/val` realm ([work in progress](https://github.com/gnolang/gno/issues/1824)), as a form of smart-contract, for modularity, whereas existing blockchains include the validator management logic within the consensus layer.

# Start a New Gno Chain and a Validator

- [start a new gno chain and a validator](./start-a-new-gno-chain-and-validator.md)

# Connect to an Existing Gno Chain

- [connect to an existing gno chain](./connect-to-an-existing-gno-chain.md)
