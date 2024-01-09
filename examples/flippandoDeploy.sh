

#/bin/bash

# echo "Deploying math_rand to math_rand..."

# gnokey maketx addpkg  \
# -deposit="1ugnot" \
# -gas-fee="1ugnot" \
# -gas-wanted="6000000" \
# -broadcast="true" \
# -remote="localhost:26657" \
# -chainid="dev" \
# -pkgdir="gno.land/p/demo/math_rand" \
# -pkgpath="gno.land/p/demo/math_rand" \
# test

echo "Deploying grc721f to grc721f..."

gnokey maketx addpkg  \
-deposit="1ugnot" \
-gas-fee="1ugnot" \
-gas-wanted="6000000" \
-broadcast="true" \
-remote="localhost:26657" \
-chainid="dev" \
-pkgdir="gno.land/p/demo/grc/grc721f" \
-pkgpath="gno.land/p/demo/grc/grc721f" \
test

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
