#!/bin/zsh

gnokey maketx \
call \
--pkgpath "gno.land/r/waymobetta/noob" \
--func "Noob" \
--args "baz" \
--gas-fee 1000000ugnot \
--gas-wanted 2000000 \
--broadcast \
--chainid dev \
--remote localhost:26657 \
demo
