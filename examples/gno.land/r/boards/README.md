This is a demo of Gno smart contract programming.  This document was
constructed by Gno onto a smart contract hosted on the data Realm
name ["gno.land/r/boards"](https://gno.land/r/boards/)
([github](https://github.com/gnolang/gno/tree/master/examples/gno.land/r/boards)).



## Build `gnokey`, create your account, and interact with Gno.

NOTE: Where you see `--remote gno.land:36657` here, that flag can be replaced
with `--remote localhost:26657` for local testnets.

### Build `gnokey`.

```bash
git clone git@github.com:gnolang/gno.git
cd ./gno
make
```

### Generate a seed/mnemonic code.

```bash
./build/gnokey generate
```

NOTE: You can generate 24 words with any good bip39 generator.

### Create a new account using your mnemonic.

```bash
./build/gnokey add KEYNAME --recover
```

NOTE: `KEYNAME` is your key identifier, and should be changed.

### Verify that you can see your account locally.

```bash
./build/gnokey list
```

## Interact with the blockchain:

### Get your current balance, account number, and sequence number.

```bash
./build/gnokey query auth/accounts/ACCOUNT_ADDR --remote gno.land:36657
```

NOTE: you can retrieve your `ACCOUNT_ADDR` with `./build/gnokey list`.

### Acquire testnet tokens using the official faucet.

Go to https://gno.land/faucet

### Create a board with a smart contract call.

```bash
./build/gnokey maketx call KEYNAME --pkgpath "gno.land/r/boards" --func CreateBoard --args "BOARDNAME" --gas-fee 1gnot --gas-wanted 2000000 > createboard.unsigned.txt
./build/gnokey sign KEYNAME --txpath createboard.unsigned.txt --chainid "testchain" --number ACCOUNT_NUMBER --sequence SEQUENCE_NUMBER > createboard.signed.txt
./build/gnokey broadcast createboard.signed.txt --remote gno.land:36657
```

Next, query for the permanent board ID by querying (you need this to create a new post):

```bash
./build/gnokey query "vm/qeval" --data "gno.land/r/boards
GetBoardIDFromName(\"BOARDNAME\")"
```

### Create a post of a board with a smart contract call.

```bash
./build/gnokey maketx call KEYNAME --pkgpath "gno.land/r/boards" --func CreatePost --args 1 --args "Hello World" --args#file "./examples/gno.land/r/boards/README.md" --gas-fee 1gnot --gas-wanted 2000000 > createpost.unsigned.txt
./build/gnokey sign KEYNAME --txpath createpost.unsigned.txt --chainid "testchain" --number ACCOUNT_NUMBER --sequence SEQUENCE_NUMBER > createpost.signed.txt
./build/gnokey broadcast createpost.signed.txt --remote gno.land:36657
```

### Create a comment to a post.

```bash
./build/gnokey maketx call KEYNAME --pkgpath "gno.land/r/boards" --func CreateReply --args 1 --args 1 --args "A comment" --gas-fee 1gnot --gas-wanted 2000000 > createcomment.unsigned.txt
./build/gnokey sign KEYNAME --txpath createcomment.unsigned.txt --chainid "testchain" --number ACCOUNT_NUMBER --sequence SEQUENCE_NUMBER > createcomment.signed.txt
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

## Starting a local `gnoland` node:

### Add test account.

```bash
./build/gnokey add test1 --recover
```

Use this mneonic:
> source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast

### Start `gnoland` node.

```bash
./build/gnoland
```

NOTE: This can be reset with `make reset`

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
