# Using the `gnokey` wallet

`gnokey` is the official command-line wallet for Gno.land. This page covers
everyday wallet use: creating and managing keys, checking balances, and sending
coins.

For deploying code, calling realms, scripting, signing workflows, and the full
command and query reference, see the
[gnokey command reference](../resources/gnokey-reference.md). If you'd prefer a graphical
wallet, see [Third-party wallets](./third-party-wallets.md).

## Installing gnokey

See [Installation](../builders/install.md) for prerequisites and install methods.
Once it's installed, check that it's on your `$PATH`:

```sh
gnokey version
```

## Managing key pairs

Every transaction you send is signed by a key pair. `gnokey` derives one from a 12
or 24-word [mnemonic phrase](https://www.zimperium.com/glossary/mnemonic-seed/):
the private key signs your transactions, and the public key derives your `g1...`
address. That address is your identity on chain. It is included in every
transaction you create, it appears in the caller stack of any realm you call, and
anyone who knows it can send you [coins](../resources/gno-stdlibs.md#coin).

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

Keys live in a keybase on disk. The `-home` flag points `gnokey` at a specific
keybase; omit it to use the default. List the keys in one with:

```bash
gnokey list
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
gnokey query bank/balances/g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 -remote https://rpc.gno.land:443
```

```bash
height: 0
data: "227984898927ugnot"
```

You can also see balances visually in [gnoweb](./explore-with-gnoweb.md) or a block explorer.

## Sending coins

Transfer coins with `gnokey maketx send`. Amounts are written as `<amount><denom>`,
for example `100ugnot`. Set `-to` (the recipient) and `-send` (the amount), and
sign with your key:

```bash
gnokey maketx send \
  -to g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5 \
  -send 100ugnot \
  -gas-fee 10000000ugnot \
  -gas-wanted 2000000 \
  -chainid staging \
  -remote "https://rpc.staging.gno.land:443" \
  mykey
```

Sending costs gas, paid in GNOT. On a testnet, get some from the
[Faucet Hub](https://faucet.gno.land) first. The `-chainid` and `-remote` flags
pick the network and must match; find their values in
[Network configuration](../resources/gnoland-networks.md). `-gas-wanted` and
`-gas-fee` cap what you pay; see [Gas fees](../resources/gas-fees.md) to tune them.

For the full base configuration, the transaction output format, and the other
message types, see the
[gnokey command reference](../resources/gnokey-reference.md#making-transactions).

## Next steps

- Deploy code, call realms, script transactions, sign offline, or read chain
  state: the [gnokey command reference](../resources/gnokey-reference.md).
- Write and ship your first realm end to end:
  [Getting started](../builders/getting-started.md).
- Use a graphical wallet instead: [Third-party wallets](./third-party-wallets.md).
