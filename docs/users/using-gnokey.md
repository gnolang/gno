# Using the `gnokey` wallet

`gnokey` is the official command-line wallet for Gno.land. It covers
everyday wallet use: creating and managing keys, checking balances, sending
coins, and calling realm functions.

For deploying code, scripting, and the full command and query
reference, see the
[gnokey command reference](../resources/gnokey.md). If you'd prefer a graphical
wallet, see [Third-party wallets](./third-party-wallets.md).

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

# 6. Call a realm function: wrap 1000ugnot into wugnot
gnokey maketx call -pkgpath gno.land/r/gnoland/wugnot -func Deposit -send 1000ugnot \
  -gas-fee 8000ugnot -gas-wanted 8000000 \
  -chainid gnoland-1 -remote https://rpc.gno.land:443 mykey
```

## Installing gnokey

`gnokey` ships with the Gno toolchain. See [Installation](../builders/install.md)
for install methods.

## Managing key pairs

Every transaction you send is signed by a key pair, which `gnokey` derives from a
[mnemonic phrase](https://en.wikipedia.org/wiki/Cryptocurrency_wallet#Seed_phrases).
The private key signs your transactions, and the public key gives your `g1...`
address, the account that holds your coins.

### Generating a key

`gnokey add` creates a new key pair and stores it under a name:

```bash
gnokey add mykey
```

It prompts for a password to encrypt the key on disk, then prints the public key,
the derived `g1...` address, and the generated mnemonic.

:::warning Safeguard your mnemonic phrase!

A mnemonic phrase is your master password: anyone holding it can re-derive your
keys, and if you lose it the key is gone for good. Write it down and keep it
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
mnemonic and a new encryption password, then re-derives the same address:

```bash
gnokey add --recover mykey
```

## Checking your balance

Reading the chain doesn't cost gas. To see what an address holds, query its
balance, pointing `-remote` at the network you care about:

```bash
gnokey query bank/balances/g15vj5q08amlvyd0nx6zjgcvwq2d0gt9fcchrvum -remote https://rpc.gno.land:443
```

```bash
height: 0
data: "5147968835830ugnot"
```

Balances are denominated in [`ugnot`](../resources/glossary.md#ugnot), the
smallest unit of GNOT, Gno.land's native coin: 1 GNOT = 1,000,000 ugnot.

For a visual view of a balance, use a block explorer such as
[GnoScan](https://gnoscan.io/).

## Anatomy of a gnokey transaction

Every `gnokey maketx` command shares the same base flags. A `send` shows the
ones every transaction needs:

```bash
gnokey maketx send \
  -to g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 \
  -send 100ugnot \
  -gas-fee 2000ugnot \
  -gas-wanted 2000000 \
  -chainid staging \
  -remote https://rpc.staging.gno.land:443 \
  mykey
```

`-to` and `-send` belong to `send`. The rest are the base flags every
transaction needs:

| Flag | What it is | Where to get it |
|------|-----------|-----------------|
| `-chainid` | the network's identifier | [Network configuration](../resources/gnoland-networks.md) |
| `-remote` | that network's node RPC address, must match `-chainid` | [Network configuration](../resources/gnoland-networks.md) |
| `-gas-wanted` | max gas units the transaction may use | gas note below |
| `-gas-fee` | the total fee paid for the transaction, in `ugnot` | [Gas fees](../resources/gas-fees.md) |
| `mykey` | the key that signs, the final argument | `gnokey list` |

Transactions cost gas, paid in GNOT. On a testnet, get some from the
[Faucet Hub](https://faucet.gno.land) first. To pay the right amount, let `gnokey`
estimate the gas: `-simulate only` runs the transaction as a dry run without
broadcasting and prints the gas used, a suggested `-gas-wanted` and a fee.
[Gas estimation](../resources/gas-fees.md#gas-estimation) explains which figures
to pass.

For the full base configuration, the output format, and every message type, see
the [gnokey command reference](../resources/gnokey.md#making-transactions).

## Sending coins

`gnokey maketx send` transfers coins. Amounts are written as `<amount><denom>`,
for example `100ugnot`. Set `-to` (the recipient) and `-send` (the amount), then
add the base flags. The command shown in
[Anatomy of a gnokey transaction](#anatomy-of-a-gnokey-transaction) is a complete
send; adjust `-to` and `-send` for your transfer.

## Calling a realm

[Realms](../resources/realms.md), Gno.land's smart contracts, expose functions
you invoke with `gnokey maketx call`. Set `-pkgpath` (the realm's on-chain path)
and `-func` (the function), pass any arguments with `-args`, and add the base
flags. This calls `Deposit` on the `wugnot` realm to wrap `1000ugnot`. `Deposit`
takes no `-args`. The `-send` flag attaches the coins the call deposits:

```bash
gnokey maketx call \
  -pkgpath gno.land/r/gnoland/wugnot \
  -func Deposit \
  -send 1000ugnot \
  -gas-fee 8000ugnot -gas-wanted 8000000 \
  -chainid staging -remote https://rpc.staging.gno.land:443 \
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
