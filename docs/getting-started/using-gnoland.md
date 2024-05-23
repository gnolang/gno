---
id: using-gnoland
---

# Working with Key Pairs

## Overview
In this tutorial, you will learn how to use the `gnoland` tool, which helps you
start a local gno.land node.

## Prerequisites
- **Git**
- **`make` (for running Makefiles)**
- **Go 1.19+**
- **Go Environment Setup**:
    - Make sure `$GOPATH` is well-defined, and `$GOPATH/bin` is added to your `$PATH` variable.

## Installation

To install the `gnoland` binary/tool, clone the Gno monorepo:

```bash
git clone https://github.com/gnolang/gno.git
```

After cloning the repo, go into the `gno.land/` folder, and use the existing
Makefile to install the `gnoland` binary:

```bash
cd gno.land
make install.gnoland
```

To verify that you've installed the binary properly and that you are able to use
it, run the `gnoland` command:

```bash
gnoland
```

If everything was successful, you should get the following output:

```bash
❯ gnoland
USAGE
  <subcommand> [flags] [<arg>...]

starts the gnoland blockchain node.

SUBCOMMANDS
  start    run the full node
  secrets  gno secrets manipulation suite
  config   gno config manipulation suite
  genesis  gno genesis manipulation suite
```

If you do not wish to install the binary globally, you can build it and run it
with the following commands from the `gno.land/` folder:

```bash
cd gno.land
make build.gnoland
```

And finally, run it with `./build gnoland`.

## Starting a chain

To start a gno.land node, you can use the `gnoland start` command. This subcommand
will do two main things:
- A default data directory will be created under `gnoland-data/`
- A genesis file for the node will be created under `genesis.json` 

<todo insert start gif>




