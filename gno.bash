#!/bin/bash
# Just a little hacky thing that will let you join the gno network


# Clone gno
# git clone https://github.com/gnolang/gno
# Enter gno folder
# cd gno
# checkout the version we should be running.  We can update this script with the commit hash of the version users should be running for now. 
git checkout 452e8b03b7ae7b0c45f2ad9263f5e9c180ad1c7e
# compile gno
make
# move the gnoland executable into the gno folder so it can run contracts.  The language stuff is in the repo root. 
cp build/gnoland ./gno
# run it for 20 seconds, as a lonely node
timeout 20 ./gno
# fetch the genesis that Jae is using today
curl http://gno.land:36657/genesis | jq .result.genesis > testdir/config/genesis.json
# Remove the data on your disk 
rm -rf testdir/data/blockstore.db testdir/data/gnolang.db testdir/data/state.db
# Get priv_validator_state
wget -O testdir/data/priv_validator_state.json https://gist.github.com/faddat/97d6a586fb0407a8c9d103b635fbe196/raw/943bb4f298d532efaa48854ba1bb9c0709e451ab/priv_validator_state.json
# get a config file that has Jae's node
wget -O testdir/config/config.toml https://gist.github.com/faddat/97d6a586fb0407a8c9d103b635fbe196/raw/0254a01ec678051cd46d03dc8cf5f5c44e55baef/config.toml
# GNO
./gno
