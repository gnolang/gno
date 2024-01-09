#/bin/bash
gnokey maketx addpkg  \
-deposit="1ugnot" \
-gas-fee="1ugnot" \
-gas-wanted="5000000" \
-broadcast="true" \
-remote="localhost:26657" \
-chainid="dev" \
-pkgdir="gno.land/p/demo/flippando" \
-pkgpath="gno.land/p/demo/flippando" \
test

gnokey maketx addpkg  \
-deposit="1ugnot" \
-gas-fee="1ugnot" \
-gas-wanted="5000000" \
-broadcast="true" \
-remote="localhost:26657" \
-chainid="dev" \
-pkgdir="gno.land/r/demo/uflippando" \
-pkgpath="gno.land/r/demo/uflippando" \
test
