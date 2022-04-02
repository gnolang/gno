This is a demo of Gno smart contract programming.  This document was
constructed by Gno onto a smart contract hosted on the data Realm
name ["gno.land/r/boards"](https://gno.land/r/boards/)
([github](https://github.com/gnolang/gno/tree/master/examples/gno.land/r/boards)).

## Starting the `gnoland` node node/validator:

NOTE: Where you see `--remote gno.land:36657` here, that flag can be replaced
with `--remote localhost:26657` for local testnets.

### Build gnoland.

```bash
git clone git@github.com:gnolang/gno.git
cd ./gno
make
```

### Add test account.

```bash
./build/gnokey add test1 --recover
```

Use this mnemonic:
> source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast

### Start gnoland validator node.

```bash
./build/gnoland
```

(This can be reset with `make reset`).

### Start gnoland web server (optional).

```bash
cd ./gnoland/website; go run \*.go
```

## Signing and broadcasting transactions:

### Publish the "gno.land/p/avl" package.

```bash
./build/gnokey maketx addpkg test1 --pkgpath "gno.land/p/avl" --pkgdir "examples/gno.land/p/avl" --deposit 100gnot --gas-fee 1gnot --gas-wanted 2000000 > addpkg.avl.unsigned.txt
./build/gnokey query "auth/accounts/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"
./build/gnokey sign test1 --txpath addpkg.avl.unsigned.txt --chainid "testchain" --number 0 --sequence 0 > addpkg.avl.signed.txt
./build/gnokey broadcast addpkg.avl.signed.txt --remote gno.land:36657
```

### Publish the "gno.land/r/boards" realm package.

```bash
./build/gnokey maketx addpkg test1 --pkgpath "gno.land/r/boards" --pkgdir "examples/gno.land/r/boards" --deposit 100gnot --gas-fee 1gnot --gas-wanted 300000000 > addpkg.boards.unsigned.txt
./build/gnokey sign test1 --txpath addpkg.boards.unsigned.txt --chainid "testchain" --number 0 --sequence 1 > addpkg.boards.signed.txt
./build/gnokey broadcast addpkg.boards.signed.txt --remote gno.land:36657
```

### Create a board with a smart contract call.

```bash
./build/gnokey maketx call test1 --pkgpath "gno.land/r/boards" --func CreateBoard --args "gnolang" --gas-fee 1gnot --gas-wanted 2000000 > createboard.unsigned.txt
./build/gnokey sign test1 --txpath createboard.unsigned.txt --chainid "testchain" --number 0 --sequence 2 > createboard.signed.txt
./build/gnokey broadcast createboard.signed.txt --remote gno.land:36657
```
Next, query for the permanent board ID by querying (you need this to create a new post):

```bash
./build/gnokey query "vm/qeval" --data "gno.land/r/boards
GetBoardIDFromName(\"gnolang\")"
```

### Create a post of a board with a smart contract call.

```bash
./build/gnokey maketx call test1 --pkgpath "gno.land/r/boards" --func CreatePost --args 1 --args "Hello World" --args#file "./examples/gno.land/r/boards/README.md" --gas-fee 1gnot --gas-wanted 2000000 > createpost.unsigned.txt
./build/gnokey sign test1 --txpath createpost.unsigned.txt --chainid "testchain" --number 0 --sequence 3 > createpost.signed.txt
./build/gnokey broadcast createpost.signed.txt --remote gno.land:36657
```

### Create a comment to a post.

```bash
./build/gnokey maketx call test1 --pkgpath "gno.land/r/boards" --func CreateReply --args 1 --args 1 --args "A comment" --gas-fee 1gnot --gas-wanted 2000000 > createcomment.unsigned.txt
./build/gnokey sign test1 --txpath createcomment.unsigned.txt --chainid "testchain" --number 0 --sequence 4 > createcomment.signed.txt
./build/gnokey broadcast createcomment.signed.txt --remote gno.land:36657
```

```bash
./build/gnokey query "vm/qrender" --data "gno.land/r/boards
gnolang/1"
```

### Render page with optional path expression.

The contents of `https://gno.land/r/boards:` and `https://gno.land/r/boards:gnolang` are rendered by calling
the `Render(path string)` function like so:

```bash
./build/gnokey query "vm/qrender" --data "gno.land/r/boards
gnolang"
```
