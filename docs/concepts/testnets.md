---
id: testnets
---

# Gno Testnets

This page documents all Gno.land testnets, what their properties are, and how
they are meant to be used. For testnet configuration, visit the 
[reference section](../reference/testnets.md).

Gno.land testnets are categorized by 4 main points:
- **Persistence of state**
  - Is the state and transaction history persisted?
- **Timeliness of code**
  - How up-to-date are Gno language features and demo packages & realms?
- **Intended purpose**
  - When should this testnet be used?
- **Versioning strategy**
  - How is this testnet versioned?

Below you can find a breakdown of each existing testnet by these categories.

## Portal Loop
Portal Loop is an always up-to-date rolling testnet. It is meant to be used as 
a nightly build of the Gno tech stack. The home page of gno.land is the `gnoweb`
render of the Portal Loop testnet. 

- **Persistence of state:**
  - State is kept on a best-effort basis 
  - Transactions that are affected by breaking changes will be discarded
- **Timeliness of code:**
  - Packages & realms which are available in the `examples/` folder on the [Gno
monorepo](https://github.com/gnolang/gno) exist on the Portal Loop in matching 
state - they are refreshed with every new commit to the `master` branch.
- **Intended purpose**
  - Providing access the latest version of Gno for fast development & demoing
- **Versioning strategy**:
  - Portal Loop infrastructure is managed within the `misc/loop` folder

For more information on the Portal Loop, and how it can be best utilized, 
check out the [Portal Loop concept page](./portal-loop.md).

View the Portal Loop network configuration [here]. // add rpc config 

## Staging
Staging is a testnet that is reset once every 60 minutes.

- **Persistence of state:**
    - State is fully discarded
- **Timeliness of code:**
  - With every reset, the latest commit of the Gno tech stack is applied, including
  the demo packages and realms
- **Intended purpose**
  - Demoing, single-use code in a staging environment, testing automation which
  uploads code to the chain, etc.
- **Versioning strategy**:
  - Staging is reset every 60 minutes to match the latest monorepo commit

View the staging network configuration [here]. // add rpc config

## Test4 (upcoming)
Test4 (name subject to change) is an upcoming, permanent, multi-node testnet. 
To follow test4 progress, view the test4 milestone
[here](https://github.com/gnolang/gno/milestone/4).
Once it is complete, it will have the following properties:

- **Persistence of state:**
  - State is fully persisted unless there are breaking changes in a new release,
where persistence partly depends on implementing a migration strategy
- **Timeliness of code:**
  - 
- **Intended purpose**
  - Running a full node, testing validator coordination, deploying stable Gno 
dApps, creating tools that require persisted state & transaction history
- **Versioning strategy**:
  - Test4 will be the first testnet to be release-based, following releases of
the Gno tech stack. 

View the staging network configuration [here]. // add rpc config

## TestX
These testnets are deprecated and currently serve as archives of previous progress.

### Test3
Test3 is the most recent persistent Gno testnet. It is still being used, but 
most packages, such as the AVL package, are outdated.

- **Persistence of state:**
  - State is fully preserved
- **Timeliness of code:**
  - Test3 is at commit XYZ, and it can contain new on-chain code
- **Intended purpose**
  - Running a full node, building an indexer, showing demos, persisting history  
- **Versioning strategy**:
  - There is no versioning strategy for test3. It will stay the way it is, until
the team chooses to shut it down.

Since Gno.land is designed with open-source in mind, anyone can see currently 
available code by browsing the [test3 homepage](https://test3.gno.land/). 

Test3 is a single-node testnet, ran by the Gno core team. There is no plan to 
upgrade test3 to a multi-node testnet. View the staging network
configuration [here]. // add rpc config

### Test2 (archive)
// add launch date?

The second Gno testnet. Find archive data [here](https://github.com/gnolang/tx-exports/tree/main/test2.gno.land).

### Test1 (archive)
// add launch date?

The first Gno testnet. Find archive data [here](https://github.com/gnolang/tx-exports/tree/main/test1.gno.land).


