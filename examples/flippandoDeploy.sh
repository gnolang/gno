

#/bin/bash

echo "Deploying flippandoserver..."

gnokey maketx addpkg  \
-deposit="1ugnot" \
-gas-fee="1ugnot" \
-gas-wanted="6000000" \
-broadcast="true" \
-remote="localhost:26657" \
-chainid="dev" \
-pkgdir="gno.land/p/demo/flippandoserver" \
-pkgpath="gno.land/p/demo/flippandoserver" \
test

echo "Deploying flippando..."

gnokey maketx addpkg  \
-deposit="1ugnot" \
-gas-fee="1ugnot" \
-gas-wanted="8000000" \
-broadcast="true" \
-remote="localhost:26657" \
-chainid="dev" \
-pkgdir="gno.land/r/demo/flippando" \
-pkgpath="gno.land/r/demo/flippando" \
test
