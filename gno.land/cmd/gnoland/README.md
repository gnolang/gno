# `gnoland`

## Install `gnoland`

    $> git clone git@github.com:gnolang/gno.git
    $> cd ./gno/gno.land
    $> make install.gnoland

## Run `gnoland` full node

    $> gnoland start

Afterward, you can interact with [`gnokey`](../gnokey) or launch a [`gnoweb`](../gnoweb) interface.


## Reset `gnoland` node back to genesis state. It's only suitable for testnets.

    $> gnoland unsafe-reset-all 

It removes the database and validator state but leaves the genesis.json and config.toml files unchanged.

The `unsafe-reset-all` command is labeled "unsafe" because:

1. It irreversibly deletes all node data, risking data loss.
2. It may lead to double signing or chain forks in production
3. It resets the `priv_validator_state.json`, and can cause network disruption if uncoordinated.
