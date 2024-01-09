#/bin/bash

echo "GetBalances"

 gnokey maketx call -broadcast -pkgpath gno.land/r/demo/flippando -gas-wanted 3000000 -gas-fee 1000000ugnot -func GetFLIPBalance -args g1j8kjkw4afsmzzs4xp9lz8yrglf6ly2apchzpyx test

echo "Mint a basic NFT, token id = 2

gnokey maketx call -broadcast -pkgpath gno.land/r/demo/flippando -gas-wanted 3000000 -gas-fee 1000000ugnot -func MintLocked -args g1j8kjkw4afsmzzs4xp9lz8yrglf6ly2apchzpyx -args 2 -args 1 test

echo "GetBalances"

 gnokey maketx call -broadcast -pkgpath gno.land/r/demo/flippando -gas-wanted 3000000 -gas-fee 1000000ugnot -func GetFLIPBalance -args g1j8kjkw4afsmzzs4xp9lz8yrglf6ly2apchzpyx test


echo "UnlockAndTransfer basic NFT, token id = 3, should fail"

gnokey maketx call -broadcast -pkgpath gno.land/r/demo/flippando -gas-wanted 3000000 -gas-fee 1000000ugnot -func UnlockAndTransferFLIP -args g1j8kjkw4afsmzzs4xp9lz8yrglf6ly2apchzpyx -args 3 test

echo "UnlockAndTransfer basic NFT, token id = 2"

gnokey maketx call -broadcast -pkgpath gno.land/r/demo/flippando -gas-wanted 3000000 -gas-fee 1000000ugnot -func UnlockAndTransferFLIP -args g1j8kjkw4afsmzzs4xp9lz8yrglf6ly2apchzpyx -args 2 test

echo "GetBalances"

 gnokey maketx call -broadcast -pkgpath gno.land/r/demo/flippando -gas-wanted 3000000 -gas-fee 1000000ugnot -func GetFLIPBalance -args g1j8kjkw4afsmzzs4xp9lz8yrglf6ly2apchzpyx test

echo "Transfer 2 FLIP fungible tokens, should fail"

gnokey maketx call -broadcast -pkgpath gno.land/r/demo/flippando -gas-wanted 3000000 -gas-fee 1000000ugnot -func Transfer -args g1ym25u5nasres8qyew5cr8ftcx69xww0gsxn22m -args 2 test

echo "Transfer 1 FLIP fungible token, should work"

gnokey maketx call -broadcast -pkgpath gno.land/r/demo/flippando -gas-wanted 3000000 -gas-fee 1000000ugnot -func Transfer -args g1ym25u5nasres8qyew5cr8ftcx69xww0gsxn22m -args 1 test
