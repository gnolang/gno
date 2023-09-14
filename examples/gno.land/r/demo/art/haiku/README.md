haiku is a implementation of a smart contract for creating and transfering text-based NFTs that conform to haiku poetry standards. 

The contract integrates a 240 kB wordlist into the contract that is used to check syllable counts and whether words are valid English, so that only valid Haikus can be added. Haikus are given a "rarity" score that can be used as a indicator of artificial scarcity. 

Add this realm to gno.land:

    gnokey maketx addpkg --pkgpath "gno.land/r/demo/art/haiku" --pkgdir "examples/gno.land/r/demo/art/haiku" --deposit 100000000ugnot --gas-fee 2000000000ugnot --gas-wanted 10000000000 --broadcast --chainid dev --remote localhost:26657 <YOURKEY>

Note: because of the word-list (240 kb) this realm takes a much higher gas than other realms. The gnovm actually had to be edited to be able to do this kind of transaction because it also takes some time and will timeout with the current defaults. Specifically, I had to increase the following parameters:

- `MaxTxBytes`, `MaxDataBytes` increased 10x (consensus parameters)
- `MaxGas` increased 1000x (consensus parameters)
- `TimeoutBroadcastTxCommit` increased 12x (RPC config)
- `maxAllocTx` increased 50x (keeper limits)
- `MaxCycles` increased 100 times (`AddPackage` config)


Mint a haiku:

     gnokey maketx call --pkgpath "gno.land/r/demo/art/haiku" --func "Mint" --args "Knock over a plant,\ncat's innocent eyes proclaim,\n'Nature needed that!'" --gas-fee "1000000ugnot" --gas-wanted "8000000" --broadcast --chainid dev --remote localhost:26657  <YOURKEY>

Transfer a haiku:

    gnokey maketx call --pkgpath "gno.land/r/demo/art/haiku" --func "Transfer" --args "g1k673c6704gzv9qyadjxv045etrysmk60ukug59" --args "be95708bce28ee9eea54a3ab6a719e24b9408aa753c3583ad8a2336b87ec3ca9" --gas-fee "1000000ugnot" --gas-wanted "8000000" --broadcast --chainid dev --remote localhost:26657 <OWNERKEY>

In this case the `g1k673c6704gzv9qyadjxv045etrysmk60ukug59` is the recipient address and the `be95708bce28ee9eea54a3ab6a719e24b9408aa753c3583ad8a2336b87ec3ca9` is the token ID of the haiku to transfer (available from gno.land). Only owners can transfer.

Register a user:

    gnokey maketx call --pkgpath "gno.land/r/demo/users" --func "Register" --args "" --args "schollz" --args "https://schollz.com" --gas-fee "1000000ugnot" --gas-wanted "2000000" --broadcast --chainid dev --remote localhost:26657 --send "200000000ugnot" <YOURKEY>

If you register a user, then your username will show up on the haiku pages instead of the address, using the `users` realm.



