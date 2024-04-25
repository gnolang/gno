---
id: gno-tooling-gnokey
---

# gnokey

Used for account & key management and general interactions with the Gnoland blockchain.

## Generate a New Seed Phrase

Generate a new seed phrase and add it to your keybase with the following command.

```bash
gnokey generate
```

## Add a New Key

You can add a new private key to the keybase using the following command.

```bash
gnokey add {KEY_NAME}
```

#### **Options**

| Name        | Type       | Description                                                                            |
|-------------|------------|----------------------------------------------------------------------------------------|
| `account`   | UInt       | Account number for HD derivation.                                                      |
| `dryrun`    | Boolean    | Performs action, but doesn't add key to local keystore.                                |
| `index`     | UInt       | Address index number for HD derivation.                                                |
| `ledger`    | Boolean    | Stores a local reference to a private key on a Ledger device.                          |
| `multisig`  | String \[] | Constructs and stores a multisig public key (implies `--pubkey`).                      |
| `nobackup`  | Boolean    | Doesn't print out seed phrase (if others are watching the terminal).                   |
| `nosort`    | Boolean    | Keys passed to `--multisig` are taken in the order they're supplied.                   |
| `pubkey`    | String     | Parses a public key in bech32 format and save it to disk.                              |
| `recover`   | Boolean    | Provides seed phrase to recover existing key instead of creating.                      |
| `threshold` | Int        | K out of N required signatures. For use in conjunction with --multisig (default: `1`). |

> **Test Seed Phrase:** source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast

### Using a ledger device

You can add a ledger device using the following command

> [!NOTE]
> Before running this command make sure your ledger device is connected, with the cosmos app installed and open in it.

```bash
gnokey add {LEDGER_KEY_NAME} --ledger
```

## List all Known Keys

List all keys stored in your keybase with the following command.

```bash
gnokey list
```

## Delete a Key

Delete a key from your keybase with the following command.

```bash
gnokey delete {KEY_NAME}
```

#### **Options**

| Name    | Type    | Description                  |
|---------|---------|------------------------------|
| `yes`   | Boolean | Skips confirmation prompt.   |
| `force` | Boolean | Removes key unconditionally. |


## Export a Private Key (Encrypted & Unencrypted)

Export a private key's (encrypted or unencrypted) armor using the following command.

```bash
gnokey export
```

#### **Options**

| Name          | Type   | Description                                 |
|---------------|--------|---------------------------------------------|
| `key`         | String | Name or Bech32 address of the private key   |
| `output-path` | String | The desired output path for the armor file  |
| `unsafe`      | Bool   | Export the private key armor as unencrypted |


## Import a Private Key (Encrypted & Unencrypted)

Import a private key's (encrypted or unencrypted) armor with the following command.

```bash
gnokey import
```

#### **Options**

| Name         | Type   | Description                                 |
|--------------|--------|---------------------------------------------|
| `armor-path` | String | The path to the encrypted armor file.       |
| `name`       | String | The name of the private key.                |
| `unsafe`     | Bool   | Import the private key armor as unencrypted |


## Make an ABCI Query

Make an ABCI Query with the following command.

```bash
gnokey query {QUERY_PATH}
```

#### **Query**

| Query Path                | Description                                                        | Example                                                                                |
|---------------------------|--------------------------------------------------------------------|----------------------------------------------------------------------------------------|
| `auth/accounts/{ADDRESS}` | Returns information about an account.                              | `gnokey query auth/accounts/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5`                  |
| `bank/balances/{ADDRESS}` | Returns balances of an account.                                    | `gnokey query bank/balances/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5`                  |
| `vm/qfuncs`               | Returns public facing function signatures as JSON.                 | `gnokey query vm/qfuncs --data "gno.land/r/demo/boards"`                               |
| `vm/qfile`                | Returns the file bytes, or list of files if directory.             | `gnokey query vm/qfile --data "gno.land/r/demo/boards"`                                |
| `vm/qrender`              | Calls .Render(path) in readonly mode.                              | `gnokey query vm/qrender --data "gno.land/r/demo/boards"`                              |
| `vm/qeval`                | Evaluates any expression in readonly mode and returns the results. | `gnokey query vm/qeval --data "gno.land/r/demo/boards GetBoardIDFromName("my_board")"` |
| `vm/store`                | (not yet supported) Fetches items from the store.                  | -                                                                                      |
| `vm/package`              | (not yet supported) Fetches a package's files.                     | -                                                                                      |

#### **Options**

| Name     | Type      | Description                              |
|----------|-----------|------------------------------------------|
| `data`   | UInt8 \[] | Queries data bytes.                      |
| `height` | Int64     | (not yet supported) Queries height.      |
| `prove`  | Boolean   | (not yet supported) Proves query result. |


## Sign and Broadcast a Transaction

You can sign and broadcast a transaction with the following command.

```bash
gnokey maketx {SUB_COMMAND} {ADDRESS or KeyName}
```

#### **Subcommands**

| Name     | Description                  |
|----------|------------------------------|
| `addpkg` | Uploads a new package.       |
| `call`   | Calls a public function.     |
| `send`   | The amount of coins to send. |

### `addpkg`

This subcommand lets you upload a new package.

```bash
gnokey maketx addpkg \
    -deposit="1ugnot" \
    -gas-fee="1ugnot" \
    -gas-wanted="5000000" \
    -pkgpath={Registered Realm path} \
    -pkgdir={Package folder path} \
    {ADDRESS} \
    > unsigned.tx
```

#### **SignBroadcast Options**

| Name         | Type    | Description                                                              |
|--------------|---------|--------------------------------------------------------------------------|
| `gas-wanted` | Int64   | The maximum amount of gas to use for the transaction.                    |
| `gas-fee`    | String  | The gas fee to pay for the transaction.                                  |
| `memo`       | String  | Any descriptive text.                                                    |
| `broadcast`  | Boolean | Broadcasts the transaction.                                              |
| `chainid`    | String  | Defines the chainid to sign for (should only be used with `--broadcast`) |

#### **makeTx AddPackage Options**

| Name      | Type   | Description                           |
|-----------|--------|---------------------------------------|
| `pkgpath` | String | The package path (required).          |
| `pkgdir`  | String | The path to package files (required). |
| `deposit` | String | The amount of coins to send.          |

### `call`

This subcommand lets you call a public function.

```bash
# Register
gnokey maketx call \
    -gas-fee="1ugnot" \
    -gas-wanted="5000000" \
    -pkgpath="gno.land/r/demo/users" \
    -send="200000000ugnot" \
    -func="Register" \
    -args="" \
    -args={NAME} \
    -args="" \
    {ADDRESS} \
    > unsigned.tx
```

#### **SignBroadcast Options**

| Name         | Type    | Description                                                      |
|--------------|---------|------------------------------------------------------------------|
| `gas-wanted` | Int64   | The maximum amount of gas to use for the transaction.            |
| `gas-fee`    | String  | The gas fee to pay for the transaction.                          |
| `memo`       | String  | Any descriptive text.                                            |
| `broadcast`  | Boolean | Broadcasts the transaction.                                      |
| `chainid`    | String  | The chainid to sign for (should only be used with `--broadcast`) |

#### **makeTx Call Options**

| Name      | Type   | Description                                                                                                                                          |
|-----------|--------|------------------------------------------------------------------------------------------------------------------------------------------------------|
| `send`    | String | The amount of coins to send.                                                                                                                         |
| `pkgpath` | String | The package path (required).                                                                                                                         |
| `func`    | String | The contract to call (required).                                                                                                                     |
| `args`    | String | An argument of the function being called. Can be used multiple times in a single `call` command to accommodate possible multiple function arguments. |

:::info
Currently, only primitive types are supported as `-args` parameters. This limitation will be addressed in the future.
Alternatively, see how `maketx run` works.
:::

### `send`

This subcommand lets you send a native currency to an address.

```bash
gnokey maketx send \
    -gas-fee="1ugnot" \
    -gas-wanted="5000000" \
    -send={SEND_AMOUNT} \
    -to={TO_ADDRESS} \
    {ADDRESS} \
    > unsigned.tx
```

#### **SignBroadcast Options**

| Name         | Type    | Description                                           |
|--------------|---------|-------------------------------------------------------|
| `gas-wanted` | Int64   | The maximum amount of gas to use for the transaction. |
| `gas-fee`    | String  | The gas fee to pay for the transaction.               |
| `memo`       | String  | Any descriptive text.                                 |
| `broadcast`  | Boolean | Broadcasts the transaction.                           |
| `chainid`    | String  | The chainid to sign for (implies `--broadcast`)       |

#### **makeTx Send Options**

| Name   | Type   | Description              |
|--------|--------|--------------------------|
| `send` | String | Amount of coins to send. |
| `to`   | String | The destination address. |


## Sign a Document

Sign a document with the following command.

```bash
gnokey sign
```

#### **Options**

| Name             | Type    | Description                                                |
|------------------|---------|------------------------------------------------------------|
| `txpath`         | String  | The path to file of tx to sign (default: `-`).             |
| `chainid`        | String  | The chainid to sign for (default: `dev`).                  |
| `number`         | UInt    | The account number of the account to sign with (required)  |
| `sequence`       | UInt    | The sequence number of the account to sign with (required) |
| `show-signbytes` | Boolean | Shows signature bytes.                                     |


## Verify a Document Signature

Verify a document signature with the following command.

```bash
gnokey verify
```

#### **Options**

| Name      | Type   | Description                              |
|-----------|--------|------------------------------------------|
| `docpath` | String | The path of the document file to verify. |

## Broadcast a Signed Document

Broadcast a signed document with the following command.

```bash
gnokey broadcast {signed transaction file document}
```
