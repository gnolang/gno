---
id: using-gnoland
---

## Using `gnoland`

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

If you do not wish to install the binary globally, you can build and run it
with the following command from the `gno.land/` folder:

```bash
make build.gnoland
```

And finally, run it with `./build gnoland`.

## Starting a local node

By using the `gnoland start` command, you can start a local a gno.land node.
This subcommand will make sure two main things happen:
- A default data directory is created under `gnoland-data/`,
- A genesis file for the node will is under `genesis.json`.

By default, the node will start listening on `localhost:26657`.

// insert gnoland start gif

## Configuring the chain

The `gnoland` tool comes with a `config` subcommand that lets you create and 
customize a configuration file for the node. 

### Initializing the config file

To create the config file, you can run the following command:

```bash
gnoland config init
```

By default, a `config.toml` file will be created in the default directory,
which can be configured by using the following flag:

```bash
gnoland config init -config-path ./config.toml
```

### Setting a value in the config

Apart from editing the `config.toml` file manually, you can set a specific value
in your config file by using the `set` subcommand:

```bash
gnoland config set <key> <value>
```

For example, we can change  

### Reading a config value

You can access all the config values in your `config.toml` file with the `get`
subcommand:

```bash
gnoland config get <key>
```

## Generating node secrets

By using the `secrets` subcommand, you can initialize your validator keypair,
as well as a file which will store 




