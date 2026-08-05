# `gnoland`

`gnoland` is the production binary powering the gno.land chain. This is the
node-operator documentation: running a chain, joining a network, and the
architecture and process expected of a validator.

Not sure you need a node at all? Read
[Running a node](../../../docs/builders/running-a-node.md) first — for writing
and testing contracts you want [gnodev](../../../contribs/gnodev), not this.

> Note: The `gnoland` binary is **specific to the gno.land chain**. Other chains
> in the Gno ecosystem will use different binaries tailored to their own
> configurations and goals.

## Install

```bash
git clone git@github.com:gnolang/gno.git
cd gno/gno.land
make install.gnoland
```

Or, with the one-line installer (`--full` adds `gnoland` to the toolchain):

```bash
curl -fsSL https://raw.githubusercontent.com/gnolang/gno/master/misc/install.sh | sh -s -- --full
```

## Run a local chain from scratch

One command, no genesis to fetch — `-lazy` generates the node secrets, the
config, and a single-validator `genesis.json` on first start:

```bash
gnoland start -lazy -skip-genesis-sig-verification
```

The node listens on `tcp://127.0.0.1:26657` (RPC) and `tcp://0.0.0.0:26656`
(p2p), with its data under `gnoland-data/`. Interact with it using
[gnokey](../gnokey) and [gnoweb](../gnoweb).

To do the same thing step by step — which is what you need for any real
deployment — generate the pieces yourself:

```bash
gnoland config init                      # writes a default config.toml
gnoland secrets init                     # generates validator + node keys
gnoland config set <key> <value>          # edit config.toml
gnoland secrets get node_id.p2p_address   # your node's dialable p2p address
gnoland start -chainid <id> -genesis genesis.json
```

Flags worth knowing on `gnoland start`:

| Flag | What it does |
|------|--------------|
| `-lazy` | Generate secrets, config, and genesis if missing (local dev only) |
| `-genesis` | Path to `genesis.json` (default `genesis.json`) |
| `-chainid` | Chain ID; must match the network's exactly (default `dev`) |
| `-data-dir` | Node data directory |
| `-skip-genesis-sig-verification` | Don't panic on genesis txs with invalid signatures |
| `-skip-failing-genesis-txs` | Don't panic when a genesis tx fails to replay |
| `-log-level` | `debug`, `info`, `warn`, `error` |

`gnoland start -h` lists the rest. `gnoland config init -h` and
`gnoland secrets init -h` do the same for those subcommands.

Released networks generally require `-skip-genesis-sig-verification`: some
genesis transactions carry placeholder or intentionally-invalidated signatures,
and the node panics on startup without it.

## Join an existing network

Every network pins its own binary revision, genesis file, and peer list, so the
authoritative instructions ship with the network rather than with this README.
Look under
[`misc/deployments/<network>/`](../../../misc/deployments)
on that network's `chain/<name>` branch, for its `config.toml`, its genesis (or
the script that regenerates it), and a `README.md` / `VALIDATOR.md`. The current
networks are listed in
[Networks](../../../docs/resources/gnoland-networks.md#deployment-files).

Two things that bite people:

- **Build from the network's `chain/<name>` branch, not `master`.** A `master`
  build will not reach consensus with a chain running a pinned release.
- **Use `p2p.persistent_peers`, not `p2p.seeds`.** The `seeds` key exists in
  `config.toml` but is not consumed by the node; a peer list set there is
  silently ignored.

Also set `p2p.external_address` to your public `host:26656`, or peers cannot
dial you back.

## Sentry-node architecture

A validator that accepts inbound connections from the public network is
reachable, and therefore attackable — a DoS against it costs you blocks. The
standard mitigation is to keep the validator off the public network entirely and
put full nodes ("sentries") in front of it.

```
   public network
         │
   ┌─────┴─────┬───────────┐
   │           │           │
sentry 1    sentry 2    sentry 3     (public, pex on)
   │           │           │
   └─────┬─────┴───────────┘
         │  private links
   ┌─────┴─────┐
   │ validator │                     (no public p2p, pex off)
   └───────────┘
```

The validator only ever talks to its own sentries; the sentries never tell the
rest of the network that the validator exists.

On the **validator**:

| Key | Value |
|-----|-------|
| `p2p.laddr` | bound to the private interface only |
| `p2p.persistent_peers` | every sentry, as `<node-id>@<host>:26656` |
| `p2p.pex` | `false` — don't discover or gossip peers |
| `p2p.external_address` | empty — nothing public should dial it |

On each **sentry**:

| Key | Value |
|-----|-------|
| `p2p.persistent_peers` | the validator, plus the other sentries |
| `p2p.private_peer_ids` | the validator's node ID — keeps it out of gossip |
| `p2p.pex` | `true` |
| `p2p.external_address` | its own public `host:26656` |

Get a node's ID with `gnoland secrets get node_id.id`, or the full dialable
address with `gnoland secrets get node_id.p2p_address`.

Two more habits: firewall the validator's p2p port to the sentry source IPs
(config is not a security boundary), and keep the RPC listener
(`rpc.laddr`) off the public internet on every node that doesn't need to serve
it.

## Keep the consensus key off the node

A sentry setup protects availability, not the key. To have the consensus key
held by a dedicated signer — [tmkms](https://github.com/iqlusioninc/tmkms) with
an HSM backend, or Horcrux for threshold signing — see [TMKMS.md](./TMKMS.md).

This is the expected setup for anything with material slashing risk.

## Hardware

At least **16 GB RAM**. Node startup temporarily exceeds 8 GB while executing
genesis. Prefer NVMe storage: the node's bottleneck is disk.

Per-network requirements, when they differ, are in that network's deployment
directory.

## Become a validator

Running the node is the easy half. Joining a validator set is a *process*, not a
deployment, and it is gated on people rather than on config:

1. **Join [Discord](https://discord.gg/YFtMjWwUN7)** and ask for the validator
   role in `#testnet-general`. This is not only where launches are announced —
   it's where coordination during upgrades, incidents, and valset changes
   happens, and operators are expected to be reachable there.
2. **Complete the onboarding checks** the network requires, including identity
   verification (KYC) where applicable.
3. **Run the node properly** — sentries, tmkms, monitoring. See the sections
   above.
4. **Sync a full node** on the target network, following its
   [deployment directory](../../../misc/deployments).
5. **Register a valoper profile** on-chain, per that network's `VALIDATOR.md`.
6. **Get voted in.** Entering the set is a GovDAO decision. Registering does not
   entitle you to a slot, on a testnet or on mainnet.

Skipping steps 1 and 2 is the usual reason a technically-fine node never makes
it into a set.

## Getting help

Ask on [Discord](https://discord.gg/YFtMjWwUN7). For node operation, the
testnet channels are the right place; say which network and paste your
`config.toml` diff from the network's default.
