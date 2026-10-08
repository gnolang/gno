# Using the `gnokey` wallet

`gnokey` is the official command-line wallet for Gno.land. It covers
everyday wallet use: creating and managing keys, checking balances, sending
coins, and calling realm functions.

If you'd prefer a graphical wallet, see
[Third-party wallets](./third-party-wallets.md).

## TL;DR

```sh
# 1. Create a key and write down its mnemonic
gnokey add mykey

# 2. List the keys in your keybase
gnokey list

# 3. Restore a key from its mnemonic
gnokey add --recover mykey

# 4. Check an address's balance
gnokey query bank/balances/<g1-address> -remote https://rpc.gno.land:443

# 5. Send coins
gnokey maketx send -to <g1-address> -send 1000000ugnot \
  -gas-fee 2000ugnot -gas-wanted 2000000 \
  -chainid gnoland-1 -remote https://rpc.gno.land:443 mykey

# 6. Call a realm function: turn 1000ugnot into wugnot tokens
gnokey maketx call -pkgpath gno.land/r/gnoland/wugnot -func Deposit -send 1000ugnot \
  -gas-fee 8000ugnot -gas-wanted 8000000 \
  -chainid gnoland-1 -remote https://rpc.gno.land:443 mykey
```

## Installing gnokey

`gnokey` ships with the Gno toolchain. See [Installation](../builders/install.md)
for install methods.

## Managing key pairs

Every transaction you send is signed by a key pair, which `gnokey` makes from a
[mnemonic phrase](https://en.wikipedia.org/wiki/Cryptocurrency_wallet#Seed_phrases).
The private key signs your transactions, and the public key gives your `g1...`
address, the account that holds your coins.

### Generating a key

`gnokey add` creates a new key pair and stores it under a name:

```bash
gnokey add mykey
```

It prompts for a password to encrypt the key on disk, then prints the public key,
your `g1...` address, and the generated mnemonic.

:::warning Safeguard your mnemonic phrase!

A mnemonic phrase is your master password: anyone holding it controls your
coins, and if you lose it the key is gone for good. Write it down and keep it
somewhere safe and offline.

:::

Keys live in a keybase on disk. List the keys in one with:

```bash
gnokey list
```

`-home` picks the keybase directory, so mainnet keys can live apart from test
keys. Every `gnokey` command takes it:

```bash
gnokey list -home ~/.gnokey-mainnet
```

### Importing an existing key

To restore a key from its mnemonic, add it with `--recover`. `gnokey` asks for the
mnemonic and a new encryption password, and you get back the same address:

```bash
gnokey add --recover mykey
```

## Checking your balance

Reading the chain doesn't cost gas. To see what an address holds, query its
balance, pointing `-remote` at the network you care about:

```bash
gnokey query bank/balances/<your-g1-address> -remote https://rpc.gno.land:443
```

An address holding 1 GNOT prints:

```console
height: 0
data: "1000000ugnot"
```

Balances are counted in [`ugnot`](../resources/glossary.md#ugnot), the smallest
unit of GNOT, Gno.land's native coin: 1 GNOT = 1,000,000 ugnot.

For a visual view of a balance, use a block explorer such as
[GnoScan](https://gnoscan.io/).

## Sending coins

`gnokey maketx send` moves coins from your key to another address. `-to` takes
the recipient's `g1...` address and `-send` the amount in `ugnot`, so 1 GNOT is
`1000000ugnot`. This sends 1 GNOT on mainnet:

```bash
gnokey maketx send \
  -to <recipient-g1-address> \
  -send 1000000ugnot \
  -gas-fee 2000ugnot \
  -gas-wanted 2000000 \
  -chainid gnoland-1 \
  -remote https://rpc.gno.land:443 \
  mykey
```

Check the recipient address before you send: a transfer cannot be taken back.

## Anatomy of a gnokey transaction

Every `gnokey maketx` command takes the same base flags as `send`:

| Flag | What it is | Where to get it |
|------|-----------|-----------------|
| `-chainid` | the network's identifier | [Network configuration](../resources/gnoland-networks.md) |
| `-remote` | that network's node RPC address, must match `-chainid` | [Network configuration](../resources/gnoland-networks.md) |
| `-gas-wanted` | max gas units the transaction may use | [Gas fees](../resources/gas-fees.md) |
| `-gas-fee` | the total fee paid for the transaction, in `ugnot` | [Gas fees](../resources/gas-fees.md) |
| `mykey` | the key that signs, the final argument | `gnokey list` |

Transactions cost gas, paid in GNOT. On a testnet, get some from the
[Faucet Hub](https://faucet.gno.land) first.

## Calling a realm

[Realms](../resources/realms.md), Gno.land's smart contracts, expose functions
you invoke with `gnokey maketx call`. Set `-pkgpath` (the realm's on-chain path)
and `-func` (the function), pass any arguments with `-args`, and add the base
flags. This calls `Deposit` on the `wugnot` realm, which turns the `1000ugnot`
sent with `-send` into 1000 `wugnot` tokens. `Deposit` takes no `-args`:

```bash
gnokey maketx call \
  -pkgpath gno.land/r/gnoland/wugnot \
  -func Deposit \
  -send 1000ugnot \
  -gas-fee 8000ugnot -gas-wanted 8000000 \
  -chainid gnoland-1 -remote https://rpc.gno.land:443 \
  mykey
```

:::tip Let gnoweb write the command for you

Every realm page in [gnoweb](./explore-with-gnoweb.md) has an **Actions** tab
listing the realm's callable functions. Fill in the arguments and it generates the
exact `gnokey maketx call` command, ready to copy and run. The
[Getting started](../builders/getting-started.md) walkthrough shows this end to
end.

:::

For more on arguments and return values, see the
[gnokey command reference](../resources/gnokey.md#call).

## Next steps

- Deploy code, script transactions, or read chain state: the
  [gnokey command reference](../resources/gnokey.md).
- Sign offline, use a multisig, or set up session accounts: the
  [`gnokey` README](../../gno.land/cmd/gnokey/README.md).
- Write and ship your first realm end to end:
  [Getting started](../builders/getting-started.md).
- Use a graphical wallet instead: [Third-party wallets](./third-party-wallets.md).
