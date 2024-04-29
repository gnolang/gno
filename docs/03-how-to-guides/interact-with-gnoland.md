---
id: interact-with-gnoland
---

# Interact with Gno.land

This tutorial will teach you how to interact with the gno.land blockchain by creating an account and calling various realms to send transactions on the network.

## Prerequisites

- [Installation](../getting-started/local-setup/local-setup.md)

## Create an Account

In order to interact with Gnoland, you need an account that you will use to sign and send transactions. You may create a new account with `gnokey generate` or recover an existing one with `gnokey add`. Confirm that your account was successfully added with `gnokey list` to display all accounts registered in the key base of your device.

```bash
gnokey generate # create a new seed phrase (mnemonic)

gnokey add -recover {your_account_name} # registers a key with the name set as the value you put in {your_account_name} with a seed phrase

gnokey list # check the list of keys
```

## Register As a User

```bash
gnokey maketx call \
    -gas-fee="1ugnot" \
    -gas-wanted="5000000" \
    -broadcast="true" \
    -remote="staging.gno.land:36657" \
    -chainid="test3" \
    -pkgpath="gno.land/r/demo/users" \
    -func="Register" \
    -args="" \
    -args="my_account" \ # (must be at least 6 characters, lowercase alphanumeric with underscore)
    -args="" \
    -send="200000000ugnot" \
    my-account

# username: must be at least 6 characters, lowercase alphanumeric with underscore
```

> **Note:** With a user registration fee of 200 GNOT and a gas fee that ranges up to 2 GNOT, you must have around 202 GNOT to complete this transaction. After registering as a user, you may replace your address with your `username` when developing or publishing a realm package.

## Get Account Information

```bash
# Get account information
gnokey query -remote="staging.gno.land:36657" "auth/accounts/{address}"

# Get account balance
gnokey query -remote="staging.gno.land:36657" "bank/balances/{address}"

# Get /r/demo/boards user information
gnokey query -remote="staging.gno.land:36657" -data "gno.land/r/demo/users
my_account" "vm/qrender"
```

## Send Tokens

The following command will send 1,000,000 ugnot (= 1 GNOT) to the address specified in the `to` argument.

```bash
# Creates and broadcast a token transfer transaction
gnokey maketx send \
    -gas-fee="1ugnot" \
    -gas-wanted="5000000" \
    -broadcast="true" \
    -remote="staging.gno.land:36657" \
    -chainid="test3" \
    -to="{address}" \ # g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
    -send="{amount}{denom}" \ # 1234ugnot
    my-account
```

## Create a Board

Try creating a board called `my_board` on the `gno.land/r/demo/boards` realm with the following command:

```bash
# Calls the CreateBoard function of gno.land/r/demo/boards
gnokey maketx call \
    -gas-fee="1ugnot" \
    -gas-wanted="5000000" \
    -broadcast="true" \
    -remote "staging.gno.land:36657" \
    -chainid="test3" \
    -pkgpath="gno.land/r/demo/boards" \
    -func="CreateBoard" \
    -args="my_board" \
    my-account
```
