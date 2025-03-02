# 4.3 Creating a key pair

## Prerequisites

- `gnokey` set up. See [Installation](installation.md).

## Overview

In this tutorial, you will learn how to create your Gno key pair using 
[gnokey](../../dev-guides/gnokey/overview.md). A key pair is required to send
transactions to the blockchain, including deploying code, interacting with 
existing applications, and more.

## A word about key pairs

Key pairs are the foundation of how users interact with blockchains; and Gno is 
no exception. By using a 12-word or 24-word [mnemonic phrase](https://www.zimperium.com/glossary/mnemonic-seed/)
as a source of randomness, users can derive a private and a public key.
These two keys can then be used further; a public key derives an address which is
a unique identifier of a user on the blockchain, while a private key is used for
signing messages and transactions for the aforementioned address, proving a user 
has ownership over it. 

Let's see how we can use `gnokey` to generate a Gno key pair locally.

## Generating a key pair

The `gnokey add` command allows you to generate a new key pair locally. Simply 
run the command, while adding a name for your key pair:

```bash
gnokey add MyKey
```

After running the command, `gnokey` will ask you to enter a password that will be
used to encrypt your key pair to the disk. Then, it will show you the following
information:
- Your public key, as well as the Gno address derived from it, starting with `g1`,
- Your randomly generated 12-word mnemonic phrase which was used to derive the key pair.

:::warning Safeguard your mnemonic phrase!

A **mnemonic phrase** is like your master password; you can use it over and over
to derive the same key pairs. This is why it is crucial to store it in a safe,
offline place - writing the phrase on a piece of paper and hiding it is highly
recommended. **If it gets lost, it is unrecoverable.**

::: 

`gnokey` will generate a keybase in which it will store information about your
key pairs. The keybase directory path is stored under the `-home` flag in `gnokey`.

### Gno addresses

Your **Gno address** is like your unique identifier on the network; an address
is visible in the caller stack of an application, it is included in each
transaction you create with your key pair, and anyone who knows your address can
send you [coins](../../concepts/stdlibs/coin.md), etc.

## Conclusion

That's it 🎉

You've successfully created your first Gno key pair. Next, we will learn how to 
spin up a local develoment node and interact with it using your fresh key pair.  

If you wish to learn more about `gnokey` specifically, check out the 
[gnokey developer guides](../../dev-guides/gnokey/overview.md).









