# Glossary of Gno Terms

A reference guide to common terminology used throughout the Gno.land
ecosystem. Terms are listed in alphabetical order.

## A

### ABCI
Application Blockchain Interface - the interface that connects the Tendermint
consensus engine with the Gno application logic.

### ABCI Queries
A set of queries that can be executed to retrieve data from the Gno.land
blockchain without changing state.
See [Querying a network](../users/interact-with-gnokey.md#querying-a-gnoland-network).

### Account Number
A unique number given to each address on a given network, used for transaction
processing and authentication.

### Address
A unique identifier derived from a public key, prefixed with 'g1' on Gno.land
networks, representing an account that can own assets and interact with the
blockchain.

### Adena
A browser extension wallet designed specifically for interacting with Gno.land
networks.
See [Third-party wallets](../users/third-party-wallets.md).

### AVL Tree
Data structure commonly used instead of the native `map` for
key-value storage. It is gas and performance efficient, particularly suited
for large or growing datasets. Deployed under `gno.land/p/nt/avl`.

See [Effective Gno](./effective-gno.md#prefer-avltree-over-map-for-scalable-storage) for usage guidance 
and the [package README](../../examples/gno.land/p/nt/avl/README.md) for technical details.

## B

### Banker
A Tendermint2 module that is natively embedded into the Gno language, via the
`std` package. Used for manipulating Coins within Gno.

### Block
A fundamental unit in a blockchain that contains a collection of validated
transactions, a timestamp, and a cryptographic hash referencing the previous
block. This structure ensures the integrity and immutability of the blockchain.

### Blockchain
A distributed ledger that records totally ordered transactions, linking the
records together via cryptographic hashes. Each block contains transaction data,
a timestamp, and a hash of the previous block, forming a continuous chain of
immutable records.

### Boards
A popular on-chain forum application on Gno.land that demonstrates social
functionality. See [Example Boards](../users/example-boards.md).

## C

### Chain ID
A unique identifier for a blockchain network (e.g., "staging" for the main
Gno.land testnet).

### Coins vs Tokens
In Gno.land, coins are native assets created by the Banker, such as `ugnot`. On
the other hand, tokens are created with packages such as GRC20, GRC721, etc.

### Contract
See [Realm](#realm).

### Crossing
Refers to the way context is switched during interrealm interactions. Check out
the [Interrealm Specification](./gno-interrealm.md).

## D

### dApp
Decentralized Application - an application built on blockchain technology,
typically consisting of smart contracts (realms in Gno.land) and a frontend
interface.

### Deploy
The process of uploading code to the blockchain. On Gno.land, this is done using
the `gnokey maketx addpkg` command or through compatible wallets.

## E

### Ephemeral Package
A temporary, user-executed code package created when using `gnokey maketx run`. 
Ephemeral packages are stored under the `/e/` path with the pattern 
`domain/e/{user-address}/run`. They execute in the user's realm context and 
allow complex interactions that aren't possible with simple function calls. 
Unlike realms and pure packages, ephemeral packages are not permanently stored 
on-chain.

### EOA
See [Externally Owned Account](#externally-owned-account-eoa).

### Externally Owned Account (EOA)
A user account controlled by a private key, as opposed to a smart contract 
account. EOAs can initiate transactions and are the only accounts that can 
sign transactions on Gno.land.

## F

### Faucet Hub
A web service that provides test tokens for Gno.land testnets, allowing
developers to try out realms and test transactions.
Visit the [Gno Faucet Hub](https://faucet.gno.land/).

## G

### Gas
A unit that measures the computational and storage resources required to execute
operations on the blockchain. Used to calculate transaction fees and prevent
spam.
See [Gas Fees](./gas-fees.md) for detailed information.

### Gas Fee
The amount paid per unit of gas, denominated in ugnot. For example,
"1000000ugnot" means 1 GNOT per unit of gas.

### Gas Wanted
The maximum amount of gas a transaction is allowed to consume. If a transaction
exceeds this limit, it fails without changing state.

### Gno
1. The programming language used for writing smart contracts on Gno.land.
2. The broader platform and ecosystem built around the language.

### Gno Debugger
A tool that helps developers test and debug Gno smart contracts by allowing them
to step through code execution. It lets you pause, inspect variables, and see
how your contract behaves before deployment.

### gno-js
A JavaScript/TypeScript client implementation that allows developers to interact
with Gno chains. This client library serves as the primary tool for JavaScript
applications to communicate with and build on top of the Gno blockchain.

### Gno Playground
A simple web interface that lets you write, test, and experiment with your Gno
code to improve your understanding of the Gno language. You can share your code,
run unit tests, deploy your realms and packages, and execute functions.
Try out [Gno Playground](https://play.gno.land/).

### Gno Studio Connect
A tool that provides seamless access to realms, making it simple to explore,
interact, and engage with Gno.land's smart contracts through function calls.
Try out [Gno Studio Connect](https://gno.studio/connect).

### Gno Test
Gno.land's built-in testing framework that enables developers to write and
execute unit tests for their Gno smart contracts (realms). It provides a
structured way to verify contract logic and simulate on-chain execution.

### Gno.land
The blockchain platform and ecosystem built on the Gno language and GnoVM.

### GnoConnect
A protocol that allows external wallets and applications to interact with
Gno.land networks, similar to WalletConnect in Ethereum.

### gnodev
A development tool which provides a local Gno.land node with hot-reloading,
state preservation, and a `gnoweb` interface for testing.
See [Local Development with gnodev](../builders/local-dev-with-gnodev.md).

### gnokey
The official command-line keychain and client for Gno.land, allowing keypair
management, transaction signing and sending queries to Gno.land chains.
See [Interacting with gnokey](../users/interact-with-gnokey.md).

### GnoVM
The virtual machine that interprets Gno, a custom version of Go optimized for
blockchains, featuring automatic state management, full determinism, and
idiomatic Go. Unlike traditional VMs, it interprets the abstract syntax tree
directly rather than using bytecode.

### gnoweb
The web interface component of the Gno ecosystem that allows users to browse Gno
source code, realms, and packages deployed to a Gno.land chain using ABCI
queries.
See [Exploring with gnoweb](../users/explore-with-gnoweb.md).

### GNOT
The native token of Gno.land networks, used for paying transaction fees and
other on-chain operations.

### GRC20
A token standard in Gno.land that defines rules for creating and managing
fungible tokens. It ensures compatibility for transfers, approvals, and
interactions within Gno.land realms.

## K

### Key Pair
A combination of a private key (for signing transactions) and a public key (from
which the address is derived) that represents an account on Gno.land.

## M

### Map
A native Gno data structure for storing key-value pairs, identical to Go maps.
Maps provide O(1) lookup time and type safety, but store all entries in a single
object which can be inefficient for large datasets. For scalable storage in
realms, consider using `avl.Tree` instead.
See [Data Structures](./gno-data-structures.md) and
[Effective Gno](./effective-gno.md#prefer-avltree-over-map-for-scalable-storage).

### Merkleization
The process of organizing data into a Merkle tree structure, allowing efficient
and secure verification of large data sets.

### Mnemonic Phrase
A series of words (usually 12 or 24) that can regenerate a private key. Also
known as a seed phrase or recovery phrase.

## N

### Namespace
A unique identifier that allows users to exclusively deploy contracts under
their own name, similar to usernames on GitHub.
See [Users and Teams](./users-and-teams.md).

## P

### Package
A collection of related code that provides specific functionality, similar to
libraries in other languages. Pure packages don't maintain state.
See [Gno Packages](../resources/gno-packages.md).

### Package Path
A unique identifier of code on the Gno.land blockchain, following the format:
`gno.land/[r|p|e]/[namespace]/[name]`. The package path determines where the code
is stored and how it can be imported or accessed. The first component indicates
the package type: `r` for realms, `p` for pure packages, and `e` for ephemeral packages.

### Portal Loop
Former testing network that was replaced by the Staging network.

### Pure Package
A stateless, importable, and reusable code (library) on the Gno.land
blockchain. Pure packages are stored under the `/p/` path and don't maintain
state.

## R

### Realm
A stateful application or smart contract on the Gno.land blockchain. Realms are
stored under the `/r/` path and can maintain state across transactions.

### Render Function
A special function in Gno with the signature `func Render(path string) string`
that returns HTML-like or Markdown content for displaying in web browsers when
the realm is viewed through gnoweb. The path parameter enables different pages
to be returned based on the path.

## S

### Sequence Number
Number of transactions executed previously by a specific account, also known as
nonce. Used to protect against replay attacks.

### Smart Contract
See [Realm](#realm).

### Staging network
The main Gno.land chain for testing, accessible at [gno.land](https://gno.land).

### Standard Library
Built-in packages that provide core functionality to Gno programs without
requiring imports from the blockchain.
See [Standard Libraries](../resources/gno-stdlibs.md).

### State
The persistent data stored by a realm. Each realm has its own isolated state
that can only be modified by functions within that realm.

## T

### Tendermint
The consensus engine used by Gno.land to secure the network and validate
transactions.

### Tendermint 2
Minimalistic version of Tendermint created by Jae Kwon, used as the consensus
layer for Gno.land.

### Testnet
A blockchain network used for testing purposes, where tokens have no real-world
value. Gno.land currently operates several testnets.

### Transaction
A state-changing action on the Gno.land blockchain, such as transferring tokens
or calling a realm function.

## U

### ugnot
The smallest unit of GNOT. 1 GNOT = 1,000,000 ugnot (micro-GNOT).

### User Registry

A system realm that allows users to register usernames and claim matching
namespaces for deploying code. List of releases found at `gno.land/r/gnoland/users`.
See [Users and Teams](./users-and-teams.md) for details.

### wugnot
Wrapped version of `ugnot`, following the GRC20 standard.
