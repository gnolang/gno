---
id: validators-overview
---

# Validator Overview

## Introduction

Gno.land is a blockchain powered by the Gno tech stack, which consists of the 
[Gno Language](https://docs.gno.land/concepts/gno-language/) (Gno), [Tendermint2](https://docs.gno.land/concepts/tendermint2/) (TM2), and [GnoVM](https://docs.gno.land/concepts/gnovm/). Unlike existing
[Proof of Stake](https://docs.cosmos.network/v0.46/modules/staking/) (PoS) blockchains in the Cosmos ecosystem, Gno.land runs on 
[Proof of Contribution](https://docs.gno.land/concepts/proof-of-contribution/) (PoC), a novel reputation-based consensus mechanism 
that values expertise and alignment with the project. In PoC, validators are 
selected via governance based on their contribution to the project and technical 
proficiency. The voting power of the network is equally distributed across all 
validators to achieve a high nakamoto coefficient. A portion of all transaction 
fees paid to the network are evenly shared between all validators to provide a 
fair incentive structure.

| **Blockchain**                       | Cosmos                  | Gno.land                    |
| ------------------------------------ | ----------------------- | --------------------------- |
| **Consensus Protocol**               | Comet BFT               | Tendermint2                 |
| **Consensus Mechanism**              | Proof of Stake          | Proof of Contribution       |
| **Requirement**                      | Delegation of Stake     | Contribution                |
| **Voting Power Reward Distribution** | Capital-based           | Evenly-distributed          |
| **Number of Validators**             | 180                     | 20~50 (TBD)                 |
| **Virtual Machine**                  | N/A                     | GnoVM                       |
| **Tokenomics**                       | Inflationary (Dilutive) | Deflationary (Non-dilutive) |

## Hardware Requirements

The following minimum hardware requirements are recommended for running a 
validator node.

- Memory: 16 GB RAM (Recommended: 32 GB)
- CPU: 2 cores (Recommended: 4 cores)
- Disk: 100 GB SSD (Depends on the level of pruning)

:::warn

These hardware requirements are currently approximate based on the Cosmos 
validator specifications. Final requirements will be determined following 
thorough testing and optimization experiments in Testnet 4.

:::

## Good Validators

Validators for Gno.land are trusted to demonstrate professionalism and 
responsibility. Below are best practices that can be expected from a good, 
reliable validator.

#### Ecosystem Contribution

- Contributing to the core development of the Gno.land project
- Providing useful tools or infrastructure services (wallets, explorers, public RPCs, etc.)
- Creating educational materials to guide new members
- Localizing documentation or content to lower language or cultural barriers

#### Quality Infrastructure

- Strong connectivity, CPU, and memory setup
- Exercising technical stability by retaining a high uptime with a robust monitoring system
- Robust contingency plans with failover systems, storage backups, and redundant power supplies
- Geographical distribution of servers

#### Transparency

- Providing regular updates
- Engaging actively in community discussions
- Being accountable for any failures

#### Compliance

- Exercising legal compliance
- Consulting with legal experts to identify regulatory risks
- Conducting internal audits

## Community

Join the official Gno.land community in various channels to receive the latest 
updates about the project and actively communicate with other validators and 
contributors.

- [Gno.land Blog](https://gno.land/r/gnoland/blog)
- [Gno.land Discord](https://discord.gg/w2MpVEunxr)
- [Gno.land Twitter](https://x.com/_gnoland)

:::info

The validator set implementation in Gno.land is abstracted away from the 
consensus mechanism inside the `r/sys/val` realm. The realm is not production 
ready yet, and is still under active development. Proposals and contributions 
to improve and complete the implementation are welcome.

**Links to related efforts:**

- Validator set injection through a Realm [[gnolang/gno #1823]](https://github.com/gnolang/gno/issues/1823)
- Add Validator Set Realm / Package [[gnolang/gno #1824](https://github.com/gnolang/gno/issues/1824)
- Add `/r/sys/vals` [[gnolang/gno #2130]](https://github.com/gnolang/gno/pull/2130)
- Add valset injection through `r/sys/vals` [[gnolang/gno #2229]](https://github.com/gnolang/gno/pull/2229)

:::
