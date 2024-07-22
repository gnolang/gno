---
id: creating-a-keypair
---

# Creating a Keypair

## Overview

In this tutorial, you will learn how to create your Gno keypair using 
[`gnokey`](../../gno-tooling/cli/gnokey/gnokey.md). 

Keypairs are the foundation of how users interact with blockchains; and Gno is 
no exception. By using a 12-word or 24-word [mnemonic phrase](https://www.zimperium.com/glossary/mnemonic-seed/#:~:text=A%20mnemonic%20seed%2C%20also%20known,wallet%20software%20or%20hardware%20device.) 
as a source of randomness, users can derive a private and a public key.
These two keys can then be used further; a public key derives an address which is
a unique identifier of a user on the blockchain, while a private key is used for
signing messages and transactions for the aforementioned address, proving a user 
has ownership over it. 

Let's see how we can use `gnokey` to generate a keypair.

## Generating a keypair

The `gnokey add` command allows you to generate a new keypair. Simply run the 
command, while adding a name for your keypair:

```bash
gnokey add MyKey
```

![gnokey-add-random](../../assets/getting-started/local-setup/creating-a-key-pair/gnokey-add-random.gif)

After running the command, `gnokey` will ask you to enter a password that will be
used to encrypt your keypair to the disk. Then, it will show you the following
information:
- Your public key, as well as the Gno address derived from it, starting with `g1...`,
- Your randomly generated 12-word mnemonic phrase which was used to derive the keypair.

:::warning Safeguard your mnemonic phrase!

A **mnemonic phrase** is like your master password; you can use it over and over
to derive the same keypair. This is why it is crucial to store it in a safe,
offline place - writing the phrase on a piece of paper and hiding it is highly
recommended. **If it gets lost, it is unrecoverable.**

::: 

`gnokey` will generate a keybase in which it will store information about your
keypairs. The keybase directory is stored under the `-home` flag in `gnokey`.

### Gno addresses

Your **Gno address** is like your username on the network; an address is visible
in the caller stack of an application, it is included in each transaction you create
with your keypair, and anyone who knows your address can send you [coins](../../concepts/stdlibs/coin.md),
etc.

## Conclusion

That's it 🎉

You've successfully created your first Gno keypair. Let's see how we can use it.










