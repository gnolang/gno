#!/bin/zsh

gnokey maketx \
call \
--pkgpath "gno.land/r/waymobetta/gor" \
--func "NewPR" \
--args "waymobetta" \
--args "1" \
--gas-fee 1000000ugnot \
--gas-wanted 2000000 \
--broadcast \
--chainid dev \
--remote localhost:26657 \
main

# gnokey maketx \
# call \
# --pkgpath "gno.land/r/waymobetta/gor" \
# --func "NewPR" \
# --args "moul" \
# --args "2" \
# --gas-fee 1000000ugnot \
# --gas-wanted 2000000 \
# --broadcast \
# --chainid dev \
# --remote localhost:26657 \
# main

# gnokey maketx \
# call \
# --pkgpath "gno.land/r/waymobetta/gor" \
# --func "NewPR" \
# --args "test" \
# --args "3" \
# --gas-fee 1000000ugnot \
# --gas-wanted 2000000 \
# --broadcast \
# --chainid dev \
# --remote localhost:26657 \
# main

gnokey maketx \
call \
--pkgpath "gno.land/r/waymobetta/gor" \
--func "RegisterAddress" \
--args "g1c74t34ukg2lq39nxx5cddlkvjtfrm3zchnzvk7" \
--args "moul" \
--gas-fee 1000000ugnot \
--gas-wanted 2000000 \
--broadcast \
--chainid dev \
--remote localhost:26657 \
main

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
main
