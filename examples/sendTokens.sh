#!/bin/bash
gnokey maketx send \
-gas-fee="10ugnot" \
-gas-wanted="5000000" \
-broadcast="true" \
-remote="localhost:26657" \
-chainid="dev" \
-to="g1ym25u5nasres8qyew5cr8ftcx69xww0gsxn22m" \
-send="1000000000ugnot" test
