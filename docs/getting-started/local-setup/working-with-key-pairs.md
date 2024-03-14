---
id: working-with-key-pairs
---

# Working with Key Pairs

## Overview

In this tutorial, you will learn how to manage private user keys, which are required for interacting with the Gno.land
blockchain. You will understand what mnemonics are, how they are used, and how you can make interaction seamless with
Gno.

## Prerequisites

- **`gnokey` set up. Reference the [Local Setup](local-setup.md#3-installing-other-gno-tools) guide for steps**

## Listing available keys

`gnokey` works by creating a local directory in the filesystem for storing (encrypted!) user private keys.

You can find this repository by checking the value of the `--home` flag when running the following command:

```bash
gnokey --help
```

Example output:

```bash
USAGE
  <subcommand> [flags] [<arg>...]

Manages private keys for the node

SUBCOMMANDS
  add        Adds key to the keybase
  delete     Deletes a key from the keybase
  generate   Generates a bip39 mnemonic
  export     Exports private key armor
  import     Imports encrypted private key armor
  list       Lists all keys in the keybase
  sign       Signs the document
  verify     Verifies the document signature
  query      Makes an ABCI query
  broadcast  Broadcasts a signed document
  maketx     Composes a tx document to sign

FLAGS
  -config ...                                          config file (optional)
  -home $XDG_CONFIG/gno  home directory
  -insecure-password-stdin=false                       WARNING! take password from stdin
  -quiet=false                                         suppress output during execution
  -remote 127.0.0.1:26657                              remote node URL
```

In this example, the directory where `gnokey` will store working data
is `/Users/zmilos/Library/Application Support/gno`.

Keep note of this directory, in case you need to reset the keystore, or migrate it for some reason.
You can provide a specific `gnokey` working directory using the `--home` flag.

To list keys currently present in the keystore, we can run:

```bash
gnokey list
```

In case there are no keys present in the keystore, the command will simply return an empty response.
Otherwise, it will return the list of keys and their accompanying metadata as a list, for example:

```bash
0. Manfred (local) - addr: g15uk9d6feap7z078ttcnwc94k60ullrvhmynxjt pub: gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pqvn87u43scec4zfgn4la3nt237nehzydzayqxe43fx63lq6rty9c5almet4, path: <nil>
1. Milos (local) - addr: g15lppu0tuxets0c0t80tncs4enqzgxt7v4eftcj pub: gpub1pgfj7ard9eg82cjtv4u4xetrwqer2dntxyfzxz3pqw2kkzujprgrfg7vumg85mccsf790n5ep6htpygkuwedwuumf2g7ydm4vqf, path: <nil>
```

The key response consists of a few pieces of information:

- The name of the private key
- The derived address (`addr`)
- The public key (`pub`)

Using these pieces of information, we can interact with Gno.land tools and write blockchain applications.

## Generating a BIP39 mnemonic

Using `gnokey`, we can generate a [mnemonic phrase](https://en.bitcoin.it/wiki/Seed_phrase) based on
the [BIP39 standard](https://github.com/bitcoin/bips/blob/master/bip-0039.mediawiki).

To generate the mnemonic phrase in the console, you can run:

```bash
gnokey generate
```

![gnokey generate](../../assets/getting-started/local-setup/creating-a-key-pair/gnokey-generate.gif)

## Adding a random private key

If we wanted to add a new private key to the keystore, we can run the following command:

```bash
gnokey add MyKey
```

Of course, you can replace `MyKey` with whatever name you want for your key.

The `gnokey` tool will prompt you to enter a password to encrypt the key on disk (don't forget this!).
After you enter the password, the `gnokey` tool will add the key to the keystore, and return the accompanying [mnemonic
phrase](https://en.bitcoin.it/wiki/Seed_phrase), which you should remember somewhere if you want to recover the key at a
future point in time.

![gnokey add random](../../assets/getting-started/local-setup/creating-a-key-pair/gnokey-add-random.gif)

You can check that the key was indeed added to the keystore, by listing available keys:

```bash
gnokey list
```

![gnokey list](../../assets/getting-started/local-setup/creating-a-key-pair/gnokey-list.gif)

## Adding a private key using a mnemonic

To add a private key to the `gnokey` keystore [using an existing mnemonic](#generating-a-bip39-mnemonic), we can run the
following command with the
`--recover` flag:

```bash
gnokey add --recover MyKey
```

Of course, you can replace `MyKey` with whatever name you want for your key.

By following the prompts to encrypt the key on disk, and providing a BIP39 mnemonic, we can successfully add
the key to the keystore.

![gnokey add mnemonic](../../assets/getting-started/local-setup/creating-a-key-pair/gnokey-add-mnemonic.gif)

## Deleting a private key

To delete a private key from the `gnokey` keystore, we need to know the name or address of the key to remove.
After we have this information, we can run the following command:

```bash
gnokey delete MyKey
```

After entering the key decryption password, the key will be deleted from the keystore.

:::caution Recovering a private key
In case you delete or lose access to your private key in the `gnokey` keystore, you
can recover it using the key's mnemonic, or by importing it if it was exported at a previous point in time.
:::

## Exporting a private key

Private keys stored in the `gnokey` keystore can be exported to a desired place
on the user's filesystem.

Keys are exported in their original armor, encrypted or unencrypted.

To export a key from the keystore, you can run:

```bash
gnokey export -key MyKey -output-path ~/Work/gno-key.asc
```

Follow the prompts presented in the terminal. Namely, you will be asked to decrypt the key in the keystore,
and later to encrypt the armor file on disk. It is worth noting that you can also export unencrypted key armor, using
the `--unsafe` flag.

![gnokey export](../../assets/getting-started/local-setup/creating-a-key-pair/gnokey-export.gif)

## Importing a private key

If you have an exported private key file, you can import it into `gnokey` fairly easily.

For example, if the key is exported at `~/Work/gno-key.asc`, you can run the following command:

```bash
gnokey import -armor-path ~/Work/gno-key.asc -name ImportedKey
```

You will be asked to decrypt the encrypted private key armor on disk (if it is encrypted, if not, use the `--unsafe`
flag), and then to provide an encryption password for storing the key in the keystore.

After executing the previous command, the `gnokey` keystore will have imported `ImportedKey`.

![gnokey import](../../assets/getting-started/local-setup/creating-a-key-pair/gnokey-import.gif)
