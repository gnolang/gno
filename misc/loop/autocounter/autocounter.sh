#!/usr/bin/env bash

CHAIN_ID=${CHAIN_ID:-""}
MNEMONIC=${MNEMONIC:-""}

echo $MNEMONIC

gnokey maketx addpkg \
            --pkgpath "gno.land/r/albttx/autocounter" \
            --pkgdir "." \
            --gas-fee 10000000ugnot \
            --gas-wanted 800000 \
            --broadcast \
            --chainid $(CHAIN_ID) \
            --remote $(REMOTE) \
            test1
