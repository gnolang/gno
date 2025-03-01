# Introduction to Blockchains

*This article is designed to give software developers, especially those new to blockchain development, an introduction to how blockchains work and why they matter. This document will help you to understand fundamental blockchain concepts. You'll learn what makes blockchains different from traditional databases, how transactions and consensus work, and how gno.land builds on these ideas with a unique developer-focused approach.*

## Blockchain Basics

### What Is a Blockchain?

A blockchain is a digital ledger that stores data across a network of computers (often called nodes). These nodes work together to validate and record new information, which is grouped into blocks. Each block holds a reference to the block before it, forming a chain. If anyone tries to change something in an old block, they’d have to change all subsequent blocks. This is an enormous task that becomes more difficult as the chain grows. This design keeps the history of data secure and transparent.

### Why It Matters

A traditional database is usually managed by a single organization, which can introduce questions of trust and control. A blockchain, by contrast, is typically distributed, decentralized, and maintained by many independent participants. This can reduce reliance on a single authority and can increase transparency. Over time, blockchains have expanded to support more than just cryptocurrency; they can also run on-chain code that performs operations without requiring permission from an overseeing authority.

### A Brief History

- **1982:** David Chaum published a paper on computer systems run by “mutually suspicious” groups, laying the foundation for later ideas.
- **1991:** Stuart Haber and W. Scott Stornetta introduced a method to timestamp and link digital documents, making them tamper-resistant.
- **1992:** Merkle trees were added, which made it more efficient to bundle multiple documents in a single cryptographic block.
- **2009:** Satoshi Nakamoto released the Bitcoin whitepaper, demonstrating a decentralized payment system run by a peer-to-peer network rather than a bank.

Bitcoin captured the public’s attention, and other projects followed. Ethereum introduced smart contracts, which allow code to run directly on the blockchain. gno.land continues this trend with additional ideas about how to reward developers and validators.

---

## Blocks, Transactions, and Fees

### Blocks

A block is a batch of transactions that have been confirmed at a specific point in time. In many blockchains, new blocks are created at regular intervals. Each block contains:

- A reference (hash) to the previous block.
- A list of transactions.
- Some additional data for validation (signatures, timestamps, etc.).

This structure forms the chain. Once a block is added, changing it later is very difficult because it would break the connection to all subsequent blocks.

### Transactions

A transaction is any action that changes the state of the blockchain. It might be a transfer of coins from one account to another or a call to a function in an on-chain program. Every transaction requires a fee. The fee is higher if the transaction demands more processing or storage.

### Gas and Fees

Blockchains measure the cost of a transaction using a gas metric. In gno.land, gas depends on how many virtual machine instructions you run and how much data you read or write. The more gas you consume, the higher your fee. These fees go mainly to validators/miners.

---

## Consensus: How the Network Agrees

Blockchains must ensure all nodes maintain the same state. This is solved by a consensus mechanism. There are many different consensus mechanisms, but most utilize one of two core strategies:

- **Proof of Work (PoW):** Used by Bitcoin. Miners compete to solve a puzzle; the winner adds a block and gets a reward.
- **Proof of Stake (PoS):** Validators lock up coins as collateral. The chance to propose a new block often scales with the amount they stake.

The gno.land blockchain has introduced a new conceptual mechanism, called **Proof of Contribution (PoCo)**. gno.land’s approach, where validators and developers both share in the rewards, encourages code contributions as well as network security.

No matter the mechanism, the idea is for a secure quorum to agree on which transactions are valid and in what order they appear.

---

## Keys, Wallets, and Addresses

### Private Keys and Public Keys

A private key is like a secret password that proves ownership of coins or access to on-chain code. The corresponding public key can be shared with others, letting them verify signatures on transactions.

### Addresses

Addresses are short strings derived from your public key. They identify you in the network. If someone wants to send you coins, they’ll use your address.

### Mnemonics

A mnemonic is a set of 12 or 24 words that can be thought of as the master password. This set of words is used to generate the private key mentioned above. This is often how wallets store and back up your credentials. Thus, it is important to keep your unique mnemonic offline and secure.

### Wallets

A wallet is a software or hardware tool that stores and manages your private keys. It handles signing transactions and tracking balances for your addresses. Your wallet doesn’t literally hold coins; it manages the keys that let you prove you own coins recorded on the blockchain. gno.land has [gnokey](https://docs.gno.land/gno-tooling/cli/gnokey/), a CLI wallet, and there are 3rd party web wallets, such as [Adena](https://adena.app).

---

## On-Chain Code (aka Smart Contracts)

Blockchains gained significant attention when people realized they could run code in a decentralized way. This code is often referred to as a “smart contract.” In gno.land, on-chain programs, which are called *realms*, are written in Gno, which is [similar to Go](https://docs.gno.land/reference/go-gno-compatibility) but with extra rules to ensure deterministic behavior across all nodes.

### What Are On-Chain Programs?

On-chain programs are pieces of code that run on a blockchain and can store arbitrary state in the blockchain’s data store. Think of it like having a database built right into the chain. This means that developers don’t have to spin up or manage an external DB. Each on-chain program can expose an API with functions that users (or other on-chain programs) can call. Every state change in these programs is recorded as a transaction, creating a tamper-resistant history in the chain itself. Collectively, such programs are known as *smart contracts*, though each program is individually its own contract (which is called a “realm” in gno.land).

### Determinism

If it is possible for two nodes to run the same code but arrive at different final states, then consensus between nodes is impossible. Thus, all nodes must reach the same result for each transaction, so blockchain code must be deterministic. Functions that rely on random numbers or local system time aren’t allowed unless they use deterministic approaches. gno.land doesn’t permit direct access to filesystems or external networks, preventing differences in execution across nodes.

### Source Code in gno.land

Unlike some blockchains that store only compiled bytecode, gno.land [stores your full source code](https://gno.land/r/gnoland/home$source). This lets anyone inspect exactly what any smart contract does. To upload code, you create a run transaction with the `.gno` file. After deployment, you can call any exported functions via a call transaction or through web interfaces that interact with the network’s RPC endpoints.

---

## Coins in gno.land

### GNOT

GNOT is the native coin of gno.land, used to pay fees. It’s also what is distributed to validators and developers. gno.land measures balances in “µgnot” (micro-gnot, often denoted as "ugnot" for simplicity), where 1 GNOT = 1,000,000 µgnot.

### Faucets and Test Networks

During development, at some point you will likely use a public test network with a faucet that gives free coins to your address. This helps you try out transactions without risking real funds. For gno.land chains, you can use the [Gno Faucet Hub](https://faucet.gno.land) to get coins. Additionally, there is a command line tool, [`gnofaucet`](https://docs.gno.land/gno-tooling/cli/faucet/gno-tooling-gnofaucet), that is available.

---

## Challenges and Ongoing Work

### Scalability

Many public blockchains can only handle a limited number of transactions per second. This is often caused by a combination of physical constraints created by the performance characteristics of the network, combined with the mathematical requirements of the consensus algorithm. This can cause slowdowns and high fees under heavy usage. Solutions like “Layer 2” networks and different consensus algorithms aim to improve transaction throughput.

### Energy Consumption

Some networks use large amounts of energy to secure their chains (Proof of Work chains). Newer methods like Proof of Stake, or gno.land’s Proof of Contribution, seek to reduce energy impact.

### Regulations

Governments are still forming policies on cryptocurrencies and blockchain tokens. Different countries may have conflicting rules, making cross-border activity more complex. Regulations about user data can also conflict with the permanent, transparent nature of blockchains.

### User Experience and Security

Creating a wallet, storing a mnemonic, and understanding on-chain transactions can feel complex. While the underlying technology is secure if used correctly, user mistakes can lead to lost funds. Also, smart contracts may have coding bugs that attackers can exploit. Developers are always working on better auditing, testing tools, and more straightforward interfaces.

---

## Conclusion

Blockchains began as a method to record transactions without relying on a single source of truth. They’ve evolved to run code in a distributed way, letting developers build applications that operate across many independent nodes. gno.land extends these ideas with an on-chain approach to source code, a developer-friendly language, and a focus on rewarding those who create and secure applications.

There is still plenty of growth to come, from better scaling and energy efficiency to simpler user interfaces. This is an evolving field, and developers play a major role in shaping how these networks expand. If you’re comfortable coding and excited to try something different, gno.land offers a chance to dive in and learn by building.
