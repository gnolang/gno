---
id: full-security-tx
---

# Making an airgapped transaction

## Prerequisites

- **`gnokey` installed.** Reference the
  [Local Setup](../../../getting-started/local-setup/installation.md#2-installing-the-required-tools) guide for steps

## Overview

`gnokey` provides a way to create a transaction, sign it, and later
broadcast it to a chain in the most secure fashion. This approach, while more 
complicated, grants full control and provides airgap support.

The indented purpose of this functionality is to provide maximum security when 
signing and broadcasting a transaction. In practice, this procedure should take
place on two separate machines, one with access to the internet (`Machine 1`), 
and the other one without (`Machine 2`), with the separation of steps as follows:
1. `Machine 1`: Fetch account information from the chain
2. `Machine 2`: Create an unsigned transaction locally
3. `Machine 2`: Sign the transaction
4. `Machine 1`: Broadcast the transaction

For the sake of simplicity, in this example, we will assume that the procedure 
is happening on two machines, and we will again use the Userbook 
realm on the Portal Loop testnet.

## Fetching account information from the chain

First, we need to fetch data for the account we are using to sign the transaction,
using the [auth/accounts](./querying-a-network.md#authaccounts) query:

```bash
gnokey query auth/accounts/<your_address> -remote "https://rpc.gno.land:443"
```

We need to extract the account number and sequence from the output:

```bash
height: 0
data: {
  "BaseAccount": {
    "address": "g1zzqd6phlfx0a809vhmykg5c6m44ap9756s7cjj",
    "coins": "10000000ugnot",
    "public_key": null,
    "account_number": "468",
    "sequence": "0"
  }
}
```

In this case, the account number is `468`, and the sequence (nonce) is `0`. We
will need these values to sign the transaction later.

## Creating an unsigned transaction locally

To create the transaction you want, you can use the aforementioned `call` API,
without the `-broadcast` flag, while redirecting the output to a local file:

```bash
gnokey maketx call \
-pkgpath "gno.land/r/demo/userbook" \
-func "SignUp" \
-gas-fee 1000000ugnot \
-gas-wanted 2000000 \
mykey > userbook.tx
```

This will create a `userbook.tx` file with a null `signature` field.
Now we are ready to sign the transaction.

## Signing the transaction

To add a signature to the transaction, we can use the `gnokey sign` subcommand.
To sign, we must set the correct flags for the subcommand:
- `-tx-path` - path to the transaction file to sign, in our case, `userbook.tx`
- `-chainid` - id of the chain to sign for
- `-account-number` - number of the account fetched previously
- `-account-sequence` - sequence of the account fetched previously

```bash
gnokey sign \
-tx-path userbook.tx \
-chainid "portal-loop" \
-account-number 468 \
-account-sequence 0 \
mykey
```

After inputting the correct values, `gnokey` will ask for the password to decrypt
the keypair. Once we input the password, we should receive the message that the
signing was completed. If we open the `userbook.tx` file, we will be able to see
that the signature field has been populated.

We are now ready to broadcast this transaction to the chain.

## Broadcasting the transaction

To broadcast the signed transaction to the chain, we can use the `gnokey broadcast`
subcommand, giving it the path to the signed transaction:

```bash
gnokey broadcast -remote "https://rpc.gno.land:443" userbook.tx
```

In this case, we do not need to specify a keypair, as the transaction has already
been signed in a previous step and `gnokey` is only sending it to the RPC endpoint.

## Verifying a transaction's signature

To verify a transaction's signature is correct, you can use the `gnokey verify`
subcommand. We can provide the path to the transaction document using the `-docpath`
flag, provide the key we signed the transaction with, and the signature itself.
Make sure the signature is in the `hex` format.

```bash
gnokey verify -docpath userbook.tx mykey <signature>
```

## Conclusion

That's it! 🎉

In this tutorial, you've learned to use `gnokey` for creating maximum-security
transactions in an airgapped manner.