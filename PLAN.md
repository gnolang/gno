# Gnoland Plan

__NOTE__ This plan is a work in progress.  Contributions are welcome to help
improve the plan. For the latest discussion, see the main README.md and join
the Discord server.

## Gnoland ecosystem

* Create snapshot of the Cosmos hub _COMPLETE_
* Launch testnet with validator set changes
* Put bounties in r/gnoland/bounties
* Use IBC2 to fetch gnolang libraries from another chain

## Technical Plan

### Gnolang

* Get basic Go tests working _COMPLETE_
* Implement minimal toolsuite (transpile, gnodev)
* Tweak/enforce gas limitations
* Implement flat-as-struct supported
* Implement ownership/realm logic; phase 1: no cycles
* Implement example smart contract application
* Implement ownership/realm logic; phase 2: ref-counted cycles
* Implement garbage collection of ref-counted cycles (long term)
* Goroutines and concurrency

#### Concurrency

Initially, we don't need to implement routines because realm package functions
provide all the inter-realm functionality we need to implement rich smart
contract programming systems.  But later, for various reasons including
long-running background jobs, and parallel concurrency, Gno will implement
deterministic concurrency as well.

Determinism is supported by including a deterministic timestamp with each
channel message as well as periodic heartbeat messages even with no sends, so
that select/receive operations can behave deterministically even in the
presence of multiple channels to select from.

### Tendermint & SDK

* Port TendermintClassic w/ AminoX with minimal dependencies _COMPLETE_.
* Port minimal SDK.
* Integrate Gnolang to SDK/BaseApp.

### Proof-of-Contributions

* Phase1: implement tiered membership
* Phase2: implement GNOSH
* Phase3: implement interchain licensing

### Gnode

* TODO: define plan

### IBC2

* TODO: define plan

## Token Plan

The ATOM distribution will be spooned, and a new premine created to incentive new contributors.

* 67% of GNOTs to ATOM holders, with minor modifications
  - Governance voting-based modifications not needed
  - TODO figure out how/whether to exclude/include exchanges.
  - TODO figure out how/whether to exclude/include IBC locked ATOMs.
  - NOTE: ATOM holders staked on the hub do not need to unstake.

* 33% of GNOTs as premine to new contributors:
  - The following distribution is a work in progress:
  - 15% (of 33%) for core contributors responsible for the delivery of Gno's complete objectives.
  - 10% for other contributors and business partnerships.
  -  4% for short term incentivization of Gno's BFT and SDK stack for use by Cosmos projects.
  -  4% for long term incentivization of Gnoland adoption.

## TODO UNSORTED & LONG-TERM

* reentrancy bug example
* data structures example
* privacy-preserving voting
* sybil resistant proof-of-human
* open hardware
* logos browser
