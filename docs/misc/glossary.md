# Glossary of Gno Terms

This is a WIP glossary of gno.land terms. This doc is meant to be a collaboration doc.

Items should be 1-2 liners that give the gist of what a concept is, linking further if needed.
Links are either external, or linking to local files within the gno.land docs.

Add in lexicographic order!

### ABCI

Short for Application Blockchain Interface.

### ABCI Queries

A set of queries that can be executed to get data from the gno.land blockchain.
See [Querying a network](../dev-guides/gnokey/querying-a-network.md).

### Account Number

A unique number given to each address on a given network.

### AVL Tree

Data structure in Gno which is commonly used instead of the native `map`.
Found at `gno.land/p/demo/avl`.

### Banker

A Tendermint2 module that is natively embedded into the Gno language, via the
`std` package. Used for manipulating Coins within Gno. See the
[Banker concept page](../concepts/stdlibs/banker.md) for more.

### Block

A block is a fundamental unit in a blockchain that contains a collection of
validated transactions, a timestamp, and a cryptographic hash referencing the
previous block. This structure ensures the integrity and immutability of the
blockchain by linking each block in a secure, chronological chain.

### Blockchain

A distributed ledger that records totally ordered transactions, linking the
records together via cryptographic hashes. Each block contains the transaction
data, a timestamp, and a cryptographic hash of the previous block. The inclusion
of the hash of the previous block forms a continuous chain, since the data stored
in any given block is affected by the data that came before it. Functionally,
this makes each block in a blockchain immutable once written, since a block
cannot be altered without also altering all subsequent blocks.

### Package Path

A unique identifier of code on the gno.land blockchain. See [Package Paths](../concepts/pkg-paths.md).

### Portal Loop

A unique, rolling testnet for gno.land. See [Portal Loop](../concepts/portal-loop.md).

### Pure Package

Stateless, importable, and reusable code (libraries) on the gno.land blockchain.
See [Pure packages](../concepts/packages.md).

### Realm

A stateful application on the gno.land blockchain. See [Realms](../concepts/realms.md).

### Sequence Number

Number of transactions executed previously by a specific account, also known
as nonce. Used to protect against replay attacks.

### Transaction

A state-changing action on the gno.land blockchain.
See [Making transactions](../dev-guides/gnokey/making-transactions.md).

### Tendermint 2

Minimalistic version of Tendermint created by Jae Kwon.

## @contributors, add suggestions below

### Coins vs Tokens

In gno.land, coins are native assets created by the [Banker](../concepts/stdlibs/banker.md), such as `ugnot`. On the other hand, tokens are created with packages such as GRC20, GRC721, etc.

### GRC20

GRC20 is a token standard in gno.land that defines rules for creating and managing fungible tokens. It ensures compatibility for transfers, approvals, and interactions within gno.land realms.

### GNOT

GNOT is the native token of the gno.land blockchain platform, used primarily for paying transaction fees (gas) and participating in the network.

### ugnot

`ugnot` (micrognot) is the smallest unit of GNOT. One GNOT equals 1,000,000 `ugnot`.

### wugnot

Wrapped version of `ugnot`, following the GRC20 standard.

### gnokey

`gnokey` is a CLI keychain and client for gno.land, allowing keypair management, transaction signing and sending queries to gno.land chains. See [gnokey](../dev-guides/gnokey/overview.md).

### gnodev

`gnodev` is a development tool which provides a local gno.land node with hot-reloading, state preservation, and a `gnoweb` interface for testing.

### gnoweb

`gnoweb` is the web interface component of the Gno ecosystem that allows users to browse Gno source code, realms, and packages deployed to a gno.land chain using ABCI queries.

### Render Function

The `Render()` function is a special function in Gno that is meant to return Markdown content that can be rendered in `gnoweb`. The signature of the Render function must always be `func Render(path string) string`. The caller can pass the path argument during execution to enable different code paths while rendering, allowing different markdown pages to be returned.

### gno-js

`gno-js` is a JavaScript/TypeScript client implementation that allows developers to interact with Gno chains. This client library serves as the primary tool for JavaScript applications to communicate with and build on top of the Gno blockchain. See [gno-js](https://github.com/gnolang/gno-js-client).

### GnoVM

GnoVM is a virtual machine that interprets Gno, a custom version of Go optimized for blockchains, featuring automatic state management, full determinism, and idiomatic Go. It works with Tendermint2 and enables smarter, more modular, and transparent appchains with embedded smart-contracts.

### Gno Studio Connect

Gno Studio Connect provides seamless access to realms, making it simple to explore, interact, and engage with gno.land’s smart contracts through function calls. Try out [Gno Studio Connect](https://gno.studio/connect).

### Gno Playground

Gno Playground is a simple web interface that lets you write, test, and experiment with your Gno code to improve your understanding of the Gno language. You can share your code, run unit tests, deploy your realms and packages, and execute functions in your code using the repo. Try out [Gno Playground](https://play.gno.land/).

### Namespaces

Namespaces allow users to exclusively deploy contracts under their own unique identifiers, similar to how GitHub manages users and organizations. Read more [here](../concepts/pkg-paths.md).

### Faucet Hub

The Faucet Hub is a single place that provides test tokens for gno.land testnets, allowing developers to try out realms and test transactions without spending real money. See [Gno Faucet Hub](https://faucet.gno.land/).

### Gno Test

Gno Test is gno.land’s built-in testing framework that enables developers to write and execute unit tests for their Gno smart contracts (realms). It provides a structured way to verify contract logic, simulate on-chain execution, and ensure correctness before deployment. See [Gno Test](https://docs.gno.land/concepts/gno-test/).

### Gno Debugger

The Gno Debugger is a tool that helps developers test and debug Gno smart contracts by allowing them to step through code execution. It lets you pause, inspect variables, and see how your contract behaves before deploying it. See [more](https://gno.land/r/gnoland/blog:p/gno-debugger).
