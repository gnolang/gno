haiku is a implementation of a smart contract for creating and transfering text-based NFTs that conform to haiku poetry standards. 

The contract integrates a 240 kB wordlist into the contract that is used to check syllable counts and whether words are valid English, so that only valid Haikus can be added. Haikus are given a "rarity" score that can be used as a indicator of artificial scarcity. 

Add this realm to gno.land:

    gnokey maketx addpkg --pkgpath "gno.land/r/demo/art/haikus" --pkgdir "examples/gno.land/r/demo/art/haiku/app" --deposit 1000000ugnot --gas-fee 200000ugnot --gas-wanted 10000000 --broadcast --chainid dev --remote localhost:26657 <YOURKEY>

Mint a haiku:

     gnokey maketx call --pkgpath "gno.land/r/demo/art/haikus" --func "Mint" --args "Knock over a plant,\ncat's innocent eyes proclaim,\n'Nature needed that!'" --gas-fee "1000000ugnot" --gas-wanted "8000000" --broadcast --chainid dev --remote localhost:26657  <YOURKEY>

Transfer a haiku:

    gnokey maketx call --pkgpath "gno.land/r/demo/art/haikus" --func "Transfer" --args "g1kn4yg8cxc65e6zgzwykwmng2wczkk2mwu5xsgv" --args "be95708bce28ee9eea54a3ab6a719e24b9408aa753c3583ad8a2336b87ec3ca9" --gas-fee "1000000ugnot" --gas-wanted "8000000" --broadcast --chainid dev --remote localhost:26657 <OWNERKEY>

In this case the `g1kn4yg8cxc65e6zgzwykwmng2wczkk2mwu5xsgv` is the recipient address and the `be95708bce28ee9eea54a3ab6a719e24b9408aa753c3583ad8a2336b87ec3ca9` is the token ID of the haiku to transfer (available from gno.land). Only owners can transfer.

Register a user:

    gnokey maketx call --pkgpath "gno.land/r/demo/users" --func "Register" --args "" --args "schollz" --args "https://schollz.com" --gas-fee "1000000ugnot" --gas-wanted "2000000" --broadcast --chainid dev --remote localhost:26657 --send "200000000ugnot" <YOURKEY>

If you register a user, then your username will show up on the haiku pages instead of the address, using the `users` realm.



