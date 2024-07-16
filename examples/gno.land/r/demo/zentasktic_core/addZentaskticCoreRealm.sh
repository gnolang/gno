#!/bin/bash

gnokey maketx addpkg --pkgpath "gno.land/r/demo/zentasktic_core" --pkgdir "./" --gas-fee 10000000ugnot --gas-wanted 20000000 --broadcast --chainid dev --remote localhost:26657 test
