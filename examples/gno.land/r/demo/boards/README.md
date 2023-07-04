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
./build/gnokey add --recover KEYNAME
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

Go to https://test3.gno.land/faucet

### Create a board with a smart contract call.

NOTE: `BOARDNAME` will be the slug of the board, and should be changed.

```bash
./build/gnokey maketx call -pkgpath "gno.land/r/boards" -func "CreateBoard" -args "BOARDNAME" -gas-fee "1000000ugnot" -gas-wanted "2000000" -broadcast -chainid testchain -remote gno.land:36657 KEYNAME
```

Interactive documentation: https://gno.land/r/boards?help&__func=CreateBoard

Next, query for the permanent board ID by querying (you need this to create a new post):

```bash
./build/gnokey query "vm/qeval" -data "gno.land/r/boards
GetBoardIDFromName(\"BOARDNAME\")" -remote gno.land:36657
```

### Create a post of a board with a smart contract call.

NOTE: If a board was created successfully, your SEQUENCE_NUMBER would have increased.

```bash
./build/gnokey maketx call -pkgpath "gno.land/r/boards" -func "CreateThread" -args BOARD_ID -args "Hello gno.land" -args\#file "./examples/gno.land/r/boards/example_post.md" -gas-fee 1000000ugnot -gas-wanted 2000000 -broadcast -chainid testchain -remote gno.land:36657 KEYNAME
```

Interactive documentation: https://gno.land/r/boards?help&__func=CreateThread

### Create a comment to a post.

```bash
./build/gnokey maketx call -pkgpath "gno.land/r/boards" -func "CreateReply" -args "BOARD_ID" -args "1" -args "1" -args "Nice to meet you too." -gas-fee 1000000ugnot -gas-wanted 2000000 -broadcast -chainid testchain -remote gno.land:36657 KEYNAME
```

Interactive documentation: https://gno.land/r/boards?help&__func=CreateReply

```bash
./build/gnokey query "vm/qrender" -data "gno.land/r/boards
BOARDNAME/1" -remote gno.land:36657
```

### Render page with optional path expression.

The contents of `https://gno.land/r/boards:` and `https://gno.land/r/boards:gnolang` are rendered by calling
the `Render(path string)` function like so:

```bash
./build/gnokey query "vm/qrender" -data "gno.land/r/boards
gnolang"
```

## Starting a local `gnoland` node:

### Add test account.

```bash
./build/gnokey add -recover test1
```

Use this mneonic:
> source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast

### Start `gnoland` node.

```bash
./build/gnoland
```

NOTE: This can be reset with `make reset`

### Publish the "gno.land/p/demo/avl" package.

```bash
./build/gnokey maketx addpkg -pkgpath "gno.land/p/demo/avl" -pkgdir "examples/gno.land/p/demo/avl" -deposit 100000000ugnot -gas-fee 1000000ugnot -gas-wanted 2000000 -broadcast -chainid dev -remote localhost:26657 test1
```

### Publish the "gno.land/r/boards" realm package.

```bash
./build/gnokey maketx addpkg -pkgpath "gno.land/r/boards" -pkgdir "examples/gno.land/r/boards" -deposit 100000000ugnot -gas-fee 1000000ugnot -gas-wanted 300000000 -broadcast -chainid dev -remote localhost:26657 test1
```
