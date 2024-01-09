#!/bin/bash
gnokey maketx send \
-gas-fee="10ugnot" \
-gas-wanted="5000000" \
-broadcast="true" \
-remote="localhost:26657" \
-chainid="dev" \
-to="g1j8kjkw4afsmzzs4xp9lz8yrglf6ly2apchzpyx" \
-send="1000000000ugnot" test
