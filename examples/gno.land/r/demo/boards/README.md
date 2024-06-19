This is a demo of Gno smart contract programming.  This document was
constructed by Gno onto a smart contract hosted on the data Realm
name ["gno.land/r/demo/boards"](https://gno.land/r/demo/boards/)
([github](https://github.com/gnolang/gno/tree/master/examples/gno.land/r/demo/boards)).



## Build `gnokey`, create your account, and interact with Gno.

NOTE: Where you see `-remote localhost:26657` here, that flag can be replaced
with `-remote test3.gno.land:26657` if you have $GNOT on the testnet.
(To use the testnet, also replace `-chainid dev` with `-chainid test3` .)

### Build `gnokey` (and other tools).

```bash
git clone git@github.com:gnolang/gno.git
cd gno/gno.land
make build
```

### Generate a seed/mnemonic code.

```bash
./build/gnokey generate
```

NOTE: You can generate 24 words with any good bip39 generator.

### Create a new account using your mnemonic.

```bash
./build/gnokey add -recover KEYNAME
```

NOTE: `KEYNAME` is your key identifier, and should be changed.

### Verify that you can see your account locally.

```bash
./build/gnokey list
```

Take note of your `addr` which looks something like `g17sphqax3kasjptdkmuqvn740u8dhtx4kxl6ljf` .
You will use this as your `ACCOUNT_ADDR`.

## Interact with the blockchain.

### Add $GNOT for your account.

Before starting the `gnoland` node for the first time, your new account can be given $GNOT in the node genesis.
Edit the file `gno.land/genesis/genesis_balances.txt` and add the following line (simlar to the others), using
your `ACCOUNT_ADDR` and `KEYNAME`

`ACCOUNT_ADDR=10000000000ugnot # @KEYNAME`

### Alternative: Run a faucet to add $GNOT.

Instead of editing `gno.land/genesis/genesis_balances.txt`, a more general solution (with more steps)
is to run a local "faucet" and use the web browser to add $GNOT. (This can be done at any time.)
See this page: https://github.com/gnolang/gno/blob/master/gno.land/cmd/gnofaucet/README.md 

### Start the `gnoland` node.

```bash
./build/gnoland start
```

NOTE: The node already has the "boards" realm.

Leave this running in the terminal. In a new terminal, cd to the same folder `gno/gno.land` .

### Get your current balance, account number, and sequence number.

```bash
./build/gnokey query auth/accounts/ACCOUNT_ADDR -remote localhost:26657
```

### Register a board username with a smart contract call.

The `USERNAME` for posting can different than your `KEYNAME`. It is internally linked to your `ACCOUNT_ADDR`. It must be at least 6 characters, lowercase alphanumeric with underscore.

```bash
./build/gnokey maketx call -pkgpath "gno.land/r/demo/users" -func "Register" -args "" -args "USERNAME" -args "Profile description" -gas-fee "10000000ugnot" -gas-wanted "2000000" -send "200000000ugnot" -broadcast -chainid dev -remote 127.0.0.1:26657 KEYNAME
```

Interactive documentation: https://test3.gno.land/r/demo/users?help&__func=Register

### Create a board with a smart contract call.

```bash
./build/gnokey maketx call -pkgpath "gno.land/r/demo/boards" -func "CreateBoard" -args "BOARDNAME" -gas-fee "1000000ugnot" -gas-wanted "10000000" -broadcast -chainid dev -remote localhost:26657 KEYNAME
```

Interactive documentation: https://test3.gno.land/r/demo/boards?help&__func=CreateBoard

Next, query for the permanent board ID by querying (you need this to create a new post):

```bash
./build/gnokey query "vm/qeval" -data "gno.land/r/demo/boards
GetBoardIDFromName(\"BOARDNAME\")" -remote localhost:26657
```

### Create a post of a board with a smart contract call.

NOTE: If a board was created successfully, your SEQUENCE_NUMBER would have increased.

```bash
./build/gnokey maketx call -pkgpath "gno.land/r/demo/boards" -func "CreateThread" -args BOARD_ID -args "Hello gno.land" -args "Text of the post" -gas-fee 1000000ugnot -gas-wanted 2000000 -broadcast -chainid dev -remote localhost:26657 KEYNAME
```

Interactive documentation: https://test3.gno.land/r/demo/boards?help&__func=CreateThread

### Create a comment to a post.

```bash
./build/gnokey maketx call -pkgpath "gno.land/r/demo/boards" -func "CreateReply" -args BOARD_ID -args "1" -args "1" -args "Nice to meet you too." -gas-fee 1000000ugnot -gas-wanted 2000000 -broadcast -chainid dev -remote localhost:26657 KEYNAME
```

Interactive documentation: https://test3.gno.land/r/demo/boards?help&__func=CreateReply

```bash
./build/gnokey query "vm/qrender" -data "gno.land/r/demo/boards
BOARDNAME/1" -remote localhost:26657
```

### Render page with optional path expression.

The contents of `https://gno.land/r/demo/boards:` and `https://gno.land/r/demo/boards:gnolang` are rendered by calling
the `Render(path string)` function like so:

```bash
./build/gnokey query "vm/qrender" -data "gno.land/r/demo/boards
gnolang"
```
## View the board in the browser.

### Start the web server.

```bash
./build/gnoweb
```

This should print something like `Running on http://127.0.0.1:8888` . Leave this running in the terminal.

### View in the browser

In your browser, navigate to the printed address http://127.0.0.1:8888 .
To see you post, click on the package `/r/demo/boards` .
