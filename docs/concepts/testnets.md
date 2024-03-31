---
id: testnets
---

# Gno.land Testnets

This page documents all Gno.land testnets, what their properties are, and how
they are meant to be used. For testnet configuration, visit the 
[reference section](../reference/testnets.md).

Gno.land testnets are categorized by 3 main points:
- Persistence of state
- Timeliness of on-chain code
- Intended purpose

Below you can find a breakdown of each existing testnet by these categories.

## Portal Loop

Portal Loop is an always up-to-date rolling testnet. It is meant to be used as 
a nightly build of the Gno tech stack. The home page of gno.land is the `gnoweb`
render of the Portal Loop testnet. 

- Persistence of state:
  - State is kept on a best-effort basis 
  - Transactions that are affected by breaking changes will be discarded
- Timeliness of on-chain code:
  - Packages & realms which are available in the `examples/` folder on the [Gno
monorepo](https://github.com/gnolang/gno) exist on the Portal Loop in matching 
state - they are refreshed with every new commit to the `master` branch.
- Intended purpose
  - Giving access the latest version of Gno for fast development & demoing

For more information on how the Portal Loop works, and how you can best utilize it, 
check out the [concept page](./portal-loop.md).

View the Portal Loop network configuration [here]. // add rpc config 

## Staging

Staging is a testnet that is reset about every hour.

- Persistence of state:
    - State is fully discarded
- Timeliness of on-chain code:
    - With every reset, the latest commit of the Gno tech stack is applied 
- Intended purpose
    - Demoing, single-use code in a staging environment, testing automation which
uploads code to the chain, etc.

View the staging network configuration [here]. // add rpc config

## Test4 (upcoming)

Test4 (name subject to change) is an upcoming, permanent, multi-node testnet. To follow test4 progress,
view the test4 milestone [here](https://github.com/gnolang/gno/milestone/4).

## TestX

These testnets are deprecated and currently serve as archives of previous progress.


## Test3 (archive)
