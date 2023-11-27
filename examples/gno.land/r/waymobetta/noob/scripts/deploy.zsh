#!/bin/zsh

gnokey maketx \
addpkg \
--pkgpath "gno.land/r/waymobetta/noob" \
--pkgdir "." \
--deposit 100000000ugnot \
--gas-fee 1000000ugnot \
--gas-wanted 2000000 \
--broadcast \
--chainid dev \
--remote localhost:26657 \
demo
