---
id: using-gnoland
---

## Using `gnoland`

## Overview

In this tutorial, you will learn how to spin up & configure a local gno.land 
node by using the `gnoland` tool, which is the Gno.land blockchain client binary.
`gnoland` is capable of managing node working files, as well as starting the
blockchain client itself.

## Prerequisites

- **Git**
- **`make` (for running Makefiles)**
- **Go 1.21+**
- **Go Environment Setup**:
    - Make sure `$GOPATH` is well-defined, and `$GOPATH/bin` is added to your `$PATH` variable.

## Installation

To install the `gnoland` binary, clone the Gno monorepo:

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
gnoland --help
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

If you do not wish to install the binary globally, you can build it with the
following command from the `gno.land/` folder:

```bash
make build.gnoland
```

And finally, run it with `./build gnoland`.

## Starting a local node

There are two ways to start a local node:
- Using lazy initialization - easier and more simple
- Initializing the node configuration yourself - more configuration options

### Lazy initialization

Lazy initialization provides a simple way to get a node up and running and ready
for use. You can start a local a node by using the `gnoland start` command, with
the included `--lazy` flag:

```bash
gnoland start --lazy
```

This subcommand in combination with the `--lazy` flag will make sure two main
things happen:
- A default data directory for the node is created under `gnoland-data/`,
- A default genesis file for the node is created under `genesis.json`.

![gnoland-start-lazy](../assets/getting-started/using-gnoland/gnoland-start-lazy.gif)

By default, the node will start listening on `localhost:26657`. To test the 
endpoint, we can run the following `curl` command, which will fetch the list of
available endpoints:

```bash
curl --location --request GET 'localhost:26657' 
```

To see what each endpoint is used for, check out the
[Gno RPC Endpoints page](../reference/rpc-endpoints.md).




### Manual initialization

Manual configuration provides a more customizable way to set up a gno.land node.


