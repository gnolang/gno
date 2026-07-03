# Using the `gnokey` wallet

`gnokey` is the official command-line wallet for Gno.land. It covers
everyday wallet use: creating and managing keys, checking balances, sending
coins, and calling realm functions.

For deploying code, scripting, signing workflows, and the full
command and query reference, see the
[gnokey command reference](../resources/gnokey-reference.md). If you'd prefer a graphical
wallet, see [Third-party wallets](./third-party-wallets.md).

## Installing gnokey

`gnokey` ships with the Gno toolchain. See [Installation](../builders/install.md)
for install methods and [verifying the binary](../builders/install.md#verify-installation).

## Managing key pairs

Every transaction you send is signed by a key pair. `gnokey` derives one from a 12
or 24-word [mnemonic phrase](https://www.zimperium.com/glossary/mnemonic-seed/):
the private key signs your transactions, and the public key derives your `g1...`
address. That address is your on-chain identity: every transaction you send
carries it, and it owns your [coins](../resources/gno-stdlibs.md#coin).

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

The `-home` flag selects which keybase to use; omit it for the default. Point it
at different directories to keep separate keybases for separate contexts, for
example testnet keys apart from mainnet keys, so each `gnokey list` shows only that
context's keys:

```bash
gnokey list -home ~/.gnokey-testnet
```

Every `gnokey` command takes `-home`, so the same flag keeps later transactions on
the right keybase.

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
gnokey query bank/balances/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 -remote https://rpc.gno.land:443
```

```bash
height: 0
data: "227984898927ugnot"
```

Balances are denominated in [`ugnot`](../resources/glossary.md#ugnot), the
smallest unit of GNOT, Gno.land's native coin: 1 GNOT = 1,000,000 ugnot.

For a visual view of a balance, use a block explorer such as
[GnoScan](https://gnoscan.io/).

## Anatomy of a gnokey transaction

Every state-changing command (`maketx send`, `maketx call`, and the rest) shares
the same base flags. A `send` shows them all:

```bash
gnokey maketx send \
  -to g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 \
  -send 100ugnot \
  -gas-fee 1000000ugnot \
  -gas-wanted 2000000 \
  -chainid staging \
  -remote https://rpc.staging.gno.land:443 \
  mykey
```

The `-to` and `-send` flags are specific to `send`; each message type has its own
(see [Sending coins](#sending-coins) and [Calling a realm](#calling-a-realm)
below). The rest are the base flags present on every transaction:

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
broadcasting and reports the real gas used, which you then pass to `-gas-wanted`.
See [Gas estimation](../resources/gas-fees.md#gas-estimation).

For the full base configuration, the output format, and every message type, see
the [gnokey command reference](../resources/gnokey-reference.md#making-transactions).

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
flags. Calling `Deposit` on the `wugnot` realm to wrap `1000ugnot`. `Deposit`
takes no `-args`. The `-send` flag attaches the coins the call deposits:

```bash
gnokey maketx call \
  -pkgpath gno.land/r/gnoland/wugnot \
  -func Deposit \
  -send 1000ugnot \
  -gas-fee 1000000ugnot -gas-wanted 2000000 \
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

For arguments, variadic functions, return values, and the `Run` scripting message,
see the [gnokey command reference: `Call`](../resources/gnokey-reference.md#call).

## Next steps

- Deploy code, script transactions, sign offline, or read chain state: the
  [gnokey command reference](../resources/gnokey-reference.md).
- Write and ship your first realm end to end:
  [Getting started](../builders/getting-started.md).
- Use a graphical wallet instead: [Third-party wallets](./third-party-wallets.md).
