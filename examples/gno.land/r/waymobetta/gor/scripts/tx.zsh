#!/bin/zsh

gnokey maketx \
call \
--pkgpath "gno.land/r/waymobetta/gor" \
--func "NewPR" \
--args 1 \
--args 100 \
--args 100 \
--args 1 \
--args 50 \
--args 0 \
--args 100 \
--args 50 \
--args "waymobetta" \
--args "dao" \
--args "feature" \
--gas-fee 1000000ugnot \
--gas-wanted 2000000 \
--broadcast \
--chainid dev \
--remote localhost:26657 \
demo

gnokey maketx \
call \
--pkgpath "gno.land/r/waymobetta/gor" \
--func "NewPR" \
--args 2 \
--args 100 \
--args 100 \
--args 1 \
--args 50 \
--args 0 \
--args 100 \
--args 50 \
--args "waymobetta" \
--args "core" \
--args "bug" \
--gas-fee 1000000ugnot \
--gas-wanted 2000000 \
--broadcast \
--chainid dev \
--remote localhost:26657 \
demo
