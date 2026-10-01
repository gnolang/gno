# GnoVM -- Gnolang Virtual Machine

GnoVM is a virtual machine that interprets Gnolang, a custom version of Golang optimized for blockchains, featuring automatic state management, full determinism, and idiomatic Go.
It works with Tendermint2 and enables smarter, more modular, and transparent appchains with embedded smart-contracts.
It can be used in TendermintCore, forks, and non-Cosmos blockchains.

Read the ["Intro to Gnoland"](https://gno.land/r/gnoland/blog:p/intro) blogpost.

This folder focuses on the VM, language, stdlibs, tests, and tools, independent of the blockchain.
This enables non-web3 developers to contribute without requiring an understanding of the broader context.

## Language Features

* Like interpreted Go, but more ambitious.
* Completely deterministic, for complete accountability.
* Transactional persistence across data realms.
* Designed for concurrent blockchain smart contracts systems.

## Getting started

Install [`gno`](./cmd/gno) and refer to the [`examples`](../examples) folder to start developing contracts.

Check the [Makefile](./Makefile) to enhance GnoVM, Gnolang, and stdlibs.

