#! /bin/bash
set -eu

ID=$1
DOCKER_IMAGE=$2
NODEID="$(docker run --rm -e TMHOME=/go/src/github.com/tendermint/classic/test/p2p/data/mach$((ID-1)) $DOCKER_IMAGE tendermint show_node_id)"
echo "$NODEID@172.57.0.$((100+$ID))"
