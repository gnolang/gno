# Installation

XXX: add right now, or maybe just add a comment to show how to install gnokey alternatively, like by downloading official releases

## Overview

In this tutorial, you will learn how to set up the Gno development environment
locally, so you can get up and running writing Gno code.

Gno is best developed in the local environment, due to its similarities with Go.
You will download and install all the necessary tooling, and validate that it
is correctly configured to run on your machine. A unix-based system is recommended
for Gno development.

## Prerequisites

- **Git**
- **`make` (for running Makefiles)**
- **Go 1.22+**
- **[A properly configured](https://go.dev/doc/install) Go environment**[^1]

## 1. Cloning the repository

To get started with a local gno.land development environment, you must clone the
GitHub repository somewhere on disk:

```bash
git clone https://github.com/gnolang/gno.git
```

## 2. Installing the required tools

There are three tools that should be used for getting started with Gno development:
- `gno` - the GnoVM binary
- `gnodev` - the Gno development solution stack
- `gnokey` - the Gno [key pair manager & client](../../dev-guides/gnokey/overview.md)

To install all three  tools, simply run the following in the root of the repo:
```bash
make install
```

## 3. Verifying installation

### `gno`

`gno` provides ample functionality to the user, among which is running,
transpiling, testing and building `.gno` files.

To verify the `gno` binary is installed system-wide, you can run:

```bash
gno --help
```

Alternatively, if you don't want to have the binary callable system-wide, you
can run the binary directly:

```bash
cd gnovm
go run ./cmd/gno --help
```

### `gnodev`

`gnodev` is the go-to Gno development helper tool - it comes with a built-in
gno.land node, a `gnoweb` server to display the state of your smart contracts
(realms), and a watcher system to actively track changes in your code.

To verify that the `gnodev` binary is installed system-wide, you can run:

```bash
gnodev --help
```

Alternatively, if you don't want to have the binary callable system-wide, you
can run the binary directly:

```bash
cd contribs/gnodev
go run ./cmd/gnodev --help
```

### `gnokey`

`gnokey` is the gno.land key pair management CLI tool. It allows you to create
key pairs, sign transactions, and broadcast them to gno.land chains. Read more
about `gnokey` [here](../../dev-guides/gnokey/overview.md).

To verify that the `gnokey` binary is installed system-wide, you can run:

```bash
gnokey --help
```

Alternatively, if you don't want to have the binary callable system-wide, you
can run the binary directly:

```bash
cd gno.land
go run ./cmd/gnokey --help
```

## Conclusion

That's it 🎉

You have successfully built out and installed the necessary tools for Gno
development!

In the upcoming tutorials, you will gain a better understanding of how they are 
used to develop gno.land apps locally. 

[^1]: If your Go environment is not set up properly, you will get `unknown command`
errors when running Gno binaries, since your terminal is not aware of their location.
