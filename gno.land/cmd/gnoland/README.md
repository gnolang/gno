# `gnoland`

## Install `gnoland`

    $> git clone git@github.com:gnolang/gno.git
    $> cd ./gno/gno.land
    $> make install.gnoland

## Option 1: Run `gnoland` full node for local development

    $> gnoland start

Afterward, you can interact with [`gnokey`](../gnokey) or launch a [`gnoweb`](../gnoweb) interface.


## Option 2: Run a node and sync with a Proof of Authority (POA) network

 - gnoland init

 - Download genesis.json from a trusted source and save it to the root-dir/config directory.

 - Get the peer's node id from the node you trust.

 - Start the node

       $> gnoland start --persistent "node_id@peer_ip_address:port" or add the persistent_peers value in the ./testdir/config/config.toml

## Option 3: Run a node as a Proof of Authority validator starting from genesis state

- Initialize the config and key files.

      $> gnoland init

- Return the node info; we will need it to add to validator info in the genesis.json

      $> gnoland node

      Address: "g14t47gv3v2z3pc23g3zr39mnc99w2cplp0jhqvv"
      Pubkey: "E5IFULgXFdS49ILgvPmO3/8chuSWfbqw3zYXaNEP+60="

- Download genesis.json from a trusted source and save it to the root-dir/config directory.

- Add your validator to the genesis file.

      $> genesis validator add \
       --address g14t47gv3v2z3pc23g3zr39mnc99w2cplp0jhqvv \
       --pub-key E5IFULgXFdS49ILgvPmO3/8chuSWfbqw3zYXaNEP+60= \
       --power 10 \
       --name testvalidator2

- Share the genesis with all trusted validators.

- Get the peer's node id from the archive node you trust.

- Start the node

      $> gnoland start --persistent "node_id@peer_ip_address:port"

  or add the persistent_peers value in the ./testdir/config/config.toml


## Option 4: Run as an archive node starting from genesis state

It's recommended to have at least two POA validator nodes running as archive nodes to bootstrap the network.

Complete the steps in Option 4 and replace the last two steps with

- Retrive node id and give it trusted peers.

      $> gnoland node

- Start the node

      $> gnoland start --prune "nothing"


## Reset `gnoland` node back to genesis state. It's suitable for running test node

    $> gnoland unsafe-reset-all

It removes the database and reset validator state back to genesis state but leaves the genesis.json and config.toml files unchanged.

The `unsafe-reset-all` command is labeled "unsafe" because:

1. It irreversibly deletes all node data, risking data loss.
2. It may lead to double signing or chain forks in production.
3. It resets the `priv_validator_state.json`, and can cause network disruption if uncoordinated.

## Reset `gnoland` node history back to genesis state.

 It removes the datastore and keeps the validator state unchanged. It reduces the risk of double signing and chain fork when we sync history state from the genesis. The validator will not sign a block until the node has synced, passing the state where the validator stopped signing.

    $> gnoland reset-state
