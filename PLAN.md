# Gnoland Plan

__NOTE__ This plan is a work in progress.  Contributions are welcome to help
improve the plan.  Read the plan for details on incentives for contributions.

## Technical Plan

### Gnolang

* Get basic Go tests working _COMPLETE_
* Implement ownership/realm logic; phase 1: no cycles
* Implement example smart contract application
* Implement ownership/realm logic; phase 2: ref-counted cycles
* Implement garbage collection of ref-counted cycles (long term)

### Tendermint & SDK

* Port TendermintClassic w/ AminoX with minimal dependencies _COMPLETE_
* Port minimal SDK
* Integrate Gnolang to SDK/BaseApp

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

### Contributing

* The key question to consider is, how to attract the best developers. TODO: create proposal for plan.
* Besides developers, there are various roles that must be filled. TODO: create categories.

### TODO UNSORTED

* reentrancy bug example
* data structures example
