---
id: gno-tooling-gnoland
---

# gnoland

## Overview

`gnoland` is the Gno.land blockchain client binary, which is capable of managing node working files, as well
as starting the blockchain client itself.

## `gnoland init`

`gnoland init` is supposed to initialize the node's working directory in the given path. The node's data directory is
comprised initially from the node's secrets and config (default values).

It is meant to be an initial step in starting the gno blockchain client, as the client itself cannot run without secrets
data like private keys, and a configuration. When the blockchain client is started, it will initialize on its own
relevant DB working directories inside the node directory.

```shell
gnoland init --help

USAGE
  init [flags]

initializes the node directory containing the secrets and configuration files

FLAGS
  -data-dir gnoland-data  the path to the node's data directory
  -force=false            overwrite existing data, if any
```

### Example usage

#### Generating fresh secrets / config

To initialize the node secrets and configuration to `./example-node-data`, run the following command:

```shell
gnoland init --data-dir ./example-node-data
```

This will initialize the following directory structure:

```shell
.
└── example-node-data/
    ├── secrets/
    │   ├── priv_validator_state.json
    │   ├── node_key.json
    │   └── priv_validator_key.json
    └── config/
       └── config.toml
```

#### Overwriting the secrets / config

In case there is an already existing node directory at the given path, you will need to provide an additional `--force`
flag to enable data overwrite.

:::warning Back up any secrets

Running `gnoland init` will generate completely new node secrets (validator private key, node p2p key), so make sure
you back up any existing secrets (located at `<node-dir>/secrets`) if you intend to overwrite them, in case you don't
want to lose them.

:::

Following up from the previous example where our desired node directory is `example-node-data` - to
initialize a completely new node data directory, with overwriting any existing data, run the following command:

```shell
gnoland init --data-dir ./example-node-data --force
```
