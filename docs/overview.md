---
id: overview
slug: /
description: "Gno.land is a Layer 1 blockchain platform that enables the execution of Smart Contracts using an interpreted
version of the Go programming language called Gno."
---

# Overview

## What is Gno.land?

Gno.land is a Layer 1 blockchain platform that enables the execution of Smart Contracts using an interpreted
version of the Go programming language called Gno (Gno for short).

### Key Features and Technology

1. **Interpreted Gno**: Gno.land utilizes the Gno programming language, which is based on Go. It is executed
   through a specialized virtual machine called the GnoVM, purpose-built for blockchain development with built-in
   determinism and a modified standard library. While Gno
   shares similarities with Go in terms of syntax, it currently lacks go routine support. However, this feature is
   planned for future development, ensuring deterministic GnoVM executions.
2. **Consensus Protocol - Tendermint2**: Gno.land achieves consensus between blockchain nodes using the Tendermint2
   consensus protocol. This approach ensures secure and reliable network operation.
3. **Inter-Blockchain Communication (IBC)**: In the future, Gno.land will be able to communicate and exchange data with
   other blockchain networks within the Cosmos ecosystem through the Inter-Blockchain Communication (IBC) protocol.

### Why Go-based?

The decision to base Gno.land's language on Go was influenced by the following factors:

1. **Standard and Secure Language**: Go is a well-established and secure programming language, widely adopted in the
   software development community. By leveraging Go's features, Gno.land benefits from a robust and proven foundation.
2. **User-Friendly**: Go's simplicity and ease of understanding make it beginner-friendly. This accessibility lowers the
   entry barrier for developers to create Smart Contracts on the Gno.land platform.

### How does it compare with Ethereum?

In comparison to Ethereum, Gno.land offers distinct advantages:

1. **Transparent and Auditable Smart Contracts**: Gno.land Smart Contracts are fully transparent and auditable by users
   because the actual source code is uploaded to the blockchain. In contrast, Ethereum requires contracts to be
   precompiled into bytecode, leading to less transparency as bytecode is stored on the blockchain, not the
   human-readable source code.

2. **General-Purpose Language**: Gno.land's Gno is a general-purpose language, similar to Go, extending its
   usability beyond the context of blockchain. In contrast, Solidity is designed specifically for Smart Contracts on the
   Ethereum platform.

## Using the Gno.land Documentation

Gno.land's documentation adopts the [Diataxis](https://diataxis.fr/) framework, ensuring structured and predictable content. It includes:
- A [Getting Started](getting-started/local-setup/local-setup.md) section, covering simple instructions on how to begin your journey into Gno.land.
- Concise how-to guides for specific technical tasks.
- Conceptual explanations, offering context and usage insights.
- Detailed reference sections with implementation specifics.
- Tutorials aimed at beginners to build fundamental skills for developing in Gno.land.
