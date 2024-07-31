#!/bin/bash

gnokey maketx addpkg --pkgpath "gno.land/p/demo/zentasktic" --pkgdir "./" --gas-fee 10000000ugnot --gas-wanted 20000000 --broadcast --chainid dev --remote localhost:26657 test
