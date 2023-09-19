# Teritori contracts

## Packages

### gno.land/p/demo/teritori/dao_interfaces

Interfaces for the modular DAOs, partial port of [DA0-DA0](https://github.com/DA0-DA0/dao-contracts/tree/7776858e780f1ce9f038a3b06cce341dd41d2189)

Also contains the MessageRegistry object, ExecutableMessage interface and MessageHandler interface which are used to decode and execute proposals actions

### gno.land/p/demo/teritori/dao_core

Implementation of the DAO core interface, partial port of [dao-dao-core](https://github.com/DA0-DA0/dao-contracts/tree/7776858e780f1ce9f038a3b06cce341dd41d2189/contracts/dao-dao-core)

### gno.land/p/demo/teritori/dao_proposal_single

Proposal module implementation that supports binary proposals, users can only vote Yes or No

### gno.land/p/demo/teritori/voting_group

Voting module that queries a group to get the voting power of users

### gno.land/p/demo/ujson

Simple-to-use JSON encoding and decoding without reflection

### gno.land/p/demo/teritori/flags_index

Provides generic flagging features, allows to:
- Flag an ID
- List most flagged IDs
- Check if an address has flagged a particular ID
- Clear flags on an ID

### gno.land/p/demo/teritori/havl

Provides an avl-like interface that allows to query state at any height, the implementation is a very simple wrapper around an avl and is very inefficient

### gno.land/p/demo/markdown_utils

Used by some other packages to properly nest markdown renders

### gno.land/p/demo/teritori/binutils

Binary utilities we used before having JSON encoding to transfer complex object between UI and chain

### gno.land/p/demo/teritori/utf16

Golang's utf16 library, used by ujson

## Realms

### gno.land/r/demo/teritori/dao_registry

Provides the list of existing DAOs

It also provides the DAOs infos (name, description, pfp) until we use the [dedicated profile realm](https://github.com/gnolang/gno/pull/181)

See the [live demo](https://app.teritori.com/orgs?network=gno-teritori) 

### gno.land/r/demo/teritori/dao_realm

Example DAO with a single choice proposal module and a voting group module

Supported proposals actions:
- Add/remove members from the voting group
- Create and moderate boards
- Mint Toris (grc20)
- Update proposal settings (threshold/quorum)

### gno.land/r/demo/teritori/groups

Fork of `gno.land/r/demo/groups` with the following changes:
- allows non-EOA to create groups (so a DAO can create and manage a group)
- add `ExecutableMessage`s implementations to support adding and removing members via proposals
- use `havl` to store members weight so we can query it at any height (this is needed by the `dao_voting_group` module)

### gno.land/r/demo/teritori/modboards

Fork of `gno.land/r/demo/boards` with the following changes:
- allows non-EOA to create boards (so a DAO can create and moderate a board)
- allow users to flag content and display most flagged posts (using `gno.land/p/demo/teritori/flags_index`)
- add `ExecutableMessage`s implementations to support creating boards an deleting threads/posts via proposals

### gno.land/r/demo/teritori/social_feeds

Social feed contract that strives to have feature-parity with Teritori's cosmwasm social feed

It supports content flagging and moderation by DAOs

See the [live demo](https://app.teritori.com/feed?network=gno-teritori)

### gno.land/r/demo/teritori/social_feeds_dao

Example of a DAO that can moderate social feeds

### gno.land/r/demo/teritori/escrow

Escrow contract, will be used in grants manager and freelance marketplace

### gno.land/r/demo/teritori/gnodaos

Exploration of a realm managing monolitic DAOs

This can't be used as is because we can't have a single realm handle multiple DAOs due to the PrevRealm auth model

Since this supports NoWithVeto we will probably refactor this code into a "cosmos-like" proposal module

### gno.land/r/demo/teritori/justicedao

Exploration based on `gno.land/r/demo/teritori/gnodaos` that randomly (using vrf realm) selects members that are allowed to vote

### gno.land/r/demo/teritori/vrf

Chainlink-style VRF contract

### gno.land/r/demo/teritori/tori

GRC20 instance that can be managed by a DAO