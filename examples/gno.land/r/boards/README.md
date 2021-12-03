This is a demo of Gno smart contract programming.  This document was
constructed by Gno. To see how it was done, follow the steps below.

The smart contract files that were uploaded to make this
possible can be found here:
https://github.com/gnolang/gno/tree/master/examples/gno.land

## add test account

> make
> ./build/gnokey add test1 --recover

Use this mnemonic:
> source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast

## start gnoland validator.

> ./build/gnoland

(This can be reset with `make reset`).

## sign an addpkg (add avl package) transaction.

```bash
./build/gnokey maketx addpkg test1 --pkgpath "gno.land/p/avl" --pkgdir "examples/gno.land/p/avl" --deposit 100gnot --gas-fee 1gnot --gas-wanted 2000000 > addpkg.avl.unsigned.txt
./build/gnokey query "auth/accounts/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
./build/gnokey sign test1 --txpath addpkg.avl.unsigned.txt --chainid "testchain" --number 0 --sequence 0 > addpkg.avl.signed.txt
./build/gnokey broadcast addpkg.avl.signed.txt
```

## sign an addpkg (add "gno.land/r/boards" package) transaction.

```bash
./build/gnokey maketx addpkg test1 --pkgpath "gno.land/r/boards" --pkgdir "examples/gno.land/r/boards" --deposit 100gnot --gas-fee 1gnot --gas-wanted 4000000 > addpkg.boards.unsigned.txt
./build/gnokey sign test1 --txpath addpkg.boards.unsigned.txt --chainid "testchain" --number 0 --sequence 1 > addpkg.boards.signed.txt
./build/gnokey broadcast addpkg.boards.signed.txt
```

## sign a (contract) function call transaction -- create board.

```bash
./build/gnokey maketx call test1 --pkgpath "gno.land/r/boards" --func CreateBoard --args "Gno.land" --gas-fee 1gnot --gas-wanted 2000000 > createboard.unsigned.txt
./build/gnokey sign test1 --txpath createboard.unsigned.txt --chainid "testchain" --number 0 --sequence 2 > createboard.signed.txt
./build/gnokey broadcast createboard.signed.txt
```
The boardcast of the createboard transaction should return the resulting board's BoardID (e.g. 1).
You can also look this up by querying:

```bash
./build/gnokey query "vm/qeval" --data "gno.land/r/boards
GetBoardIDFromName(\"Gno.land\")"
```

## sign a (contract) function call transaction -- create post.

```bash
./build/gnokey maketx call test1 --pkgpath "gno.land/r/boards" --func CreatePost --args 1 --args "Hello World" --args#file "./examples/gno.land/r/boards/README.md" --gas-fee 1gnot --gas-wanted 2000000 > createpost.unsigned.txt
./build/gnokey sign test1 --txpath createpost.unsigned.txt --chainid "testchain" --number 0 --sequence 3 > createpost.signed.txt
./build/gnokey broadcast createpost.signed.txt
```

## render page with ABCI query (evalquery).

```bash
./build/gnokey query "vm/qeval" --data "gno.land/r/boards
RenderBoard(1)"
```
