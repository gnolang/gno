#!/bin/zsh

gnokey maketx \
call \
--pkgpath "gno.land/r/waymobetta/gor" \
--func "RegisterUsername" \
--args "moul" \
--args "g1c74t34ukg2lq39nxx5cddlkvjtfrm3zchnzvk7" \
--gas-fee 1000000ugnot \
--gas-wanted 2000000 \
--broadcast \
--chainid dev \
--remote localhost:26657 \
demo
