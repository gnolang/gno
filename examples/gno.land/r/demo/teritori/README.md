# Teritori contracts

## Packages

### gno.land/p/demo/teritori/dao_interfaces

Interfaces for the modular DAOs, partial port of [DA0-DA0](https://github.com/DA0-DA0/dao-contracts/tree/7776858e780f1ce9f038a3b06cce341dd41d2189)

Also contains the MessageRegistry object, ExecutableMessage interface and MessageHandler interface which are used to decode and execute proposals actions

### gno.land/p/demo/teritori/dao_core

Implementation of the DAO core interface, partial port of [dao-dao-core](https://github.com/DA0-DA0/dao-contracts/tree/7776858e780f1ce9f038a3b06cce341dd41d2189/contracts/dao-dao-core)

### gno.land/p/demo/teritori/dao_proposal_single

Proposal module implementation that supports binary proposals, users can only vote Yes or No

### gno.land/p/demo/teritori/dao_voting_group

Voting module that manages a mapping from address to voting power

### gno.land/p/demo/teritori/havl

Provides an avl-like interface that allows to query state at any height, the implementation is a very simple wrapper around an avl and is very inefficient

### gno.land/p/demo/markdown_utils

Used by some other packages to properly nest markdown renders

## Realms

### gno.land/r/demo/teritori/dao_registry

Provides the list of existing DAOs

It also provides the DAOs infos (name, description, pfp) until we use the [dedicated profile realm](https://github.com/gnolang/gno/pull/181)

See the [live demo](https://app.teritori.com/orgs?network=gno-portal) 

### gno.land/r/demo/teritori/dao_realm

Example DAO with a single choice proposal module and a voting group module

Supported proposals actions:
- Add/remove members from the voting group
- Create and moderate boards
- Mint Toris (grc20)
- Update proposal settings (threshold/quorum)

### gno.land/r/demo/teritori/tori

GRC20 instance that can be managed by a DAO