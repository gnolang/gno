#/bin/bash

cd testdir/data
rm -rf blockstore.db gnolang.db state.db cs.wal
cp priv_validator_state.json.orig priv_validator_state.json 
cd ../../
gnoland start 
