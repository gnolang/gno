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

Building from source requires **Go 1.25+**, `git`, and `make` — see
[install.md](../../../docs/builders/install.md) for details and troubleshooting.

```bash
git clone git@github.com:gnolang/gno.git
cd gno/gno.land
make install.gnoland
```

Or, with the one-line installer, which downloads a prebuilt binary and needs no
Go toolchain (`--full` adds `gnoland`):

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
gnoland config init                       # writes a default config.toml
gnoland secrets init                      # generates validator + node keys
gnoland config set <key> <value>          # edit config.toml
gnoland config get <key>                  # read a value back
gnoland secrets get node_id.p2p_address   # your node's dialable p2p address
gnoland start -chainid <id> -genesis genesis.json -log-level info
```

To build a genesis file yourself — adding validators, balances, or transactions
— use [`gnogenesis`](../../../contribs/gnogenesis), a dedicated CLI for
generating and manipulating `genesis.json`.

> ⚠️ **Set `-log-level info` on any node you leave running.** The default is
> `debug`, which writes enough to fill a disk: one operator reported 1.3 TB of
> logs on a testnet node
> ([#5903](https://github.com/gnolang/gno/issues/5903)). Rotate your logs too.

Flags worth knowing on `gnoland start`:

| Flag | What it does |
|------|--------------|
| `-lazy` | Generate secrets, config, and genesis if missing (local dev only) |
| `-genesis` | Path to `genesis.json` (default `genesis.json`) |
| `-chainid` | Chain ID; must match the network's exactly (default `dev`) |
| `-data-dir` | Node data directory (default `gnoland-data`) |
| `-log-level` | `debug`, `info`, `warn`, `error` — **default `debug`**, see above |
| `-skip-genesis-sig-verification` | Don't panic on genesis txs with invalid signatures |
| `-skip-failing-genesis-txs` | Don't panic when a genesis tx fails to replay |

`gnoland start -h` lists the rest. `gnoland config init -h` and
`gnoland secrets init -h` do the same for those subcommands.

Released networks generally require `-skip-genesis-sig-verification`: some
genesis transactions carry placeholder or intentionally-invalidated signatures,
and the node panics on startup without it.

### Config keys worth setting

`config.toml` is large; these are the ones that matter for a node that joins a
network. Set them with `gnoland config set <key> <value>`.

| Key | What it does |
|-----|--------------|
| `moniker` | Human-readable node name, shown to peers |
| `p2p.laddr` | P2P listen address (default `tcp://0.0.0.0:26656`) |
| `p2p.external_address` | Your public `host:26656` — without it, peers cannot dial you back |
| `p2p.persistent_peers` | Comma-separated `<node-id>@<host>:26656` list to stay connected to |
| `p2p.pex` | Peer exchange: `true` to discover peers, `false` on a validator behind sentries |
| `p2p.private_peer_ids` | Peer IDs never gossiped to others — a sentry's validator |
| `p2p.max_num_outbound_peers` | Outbound peer cap, excluding persistent peers |
| `rpc.laddr` | RPC listen address (default `tcp://127.0.0.1:26657`) — keep off the public internet |
| `mempool.size` | Max transactions held in the mempool |
| `application.prune_strategy` | `everything`, `nothing`, or `syncable` (default). `nothing` keeps all history — needed for historical queries |
| `consensus.timeout_commit` | Chain-wide; must match the network |

Networks pin several of these; always start from the `config.toml` in the
network's deployment directory rather than from the defaults.

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

**RAM: 16 GB minimum.** Node startup temporarily exceeds 8 GB while executing
genesis, so a box sized for steady state alone will OOM on first boot.

**Storage: NVMe, not spinning disk or network storage.** The node's bottleneck
is disk I/O, and consensus is latency-sensitive — a slow disk shows up as missed
blocks, not as a slow node. Size it for growth: the data directory grows with
chain history, and `application.prune_strategy = nothing` (needed for historical
queries) means it never shrinks. Budget for logs separately, and set
`-log-level info` — see the warning above.

Per-network requirements, when they differ, are in that network's deployment
directory.

## Become a validator

Running the node is the easy half. Joining a validator set is a *process*, and
it ends in a governance vote rather than a config change.

Get the node right **first** — the review team checks that your node is synced
on the network before approving anything, so starting the paperwork early just
means waiting twice.

1. **Run the node properly.** Sentries, remote signing, monitoring, log rotation
   — see the sections above.
2. **Sync a full node** on the target network, following its
   [deployment directory](../../../misc/deployments). Wait until
   `/status` reports `catching_up: false`.
3. **Start onboarding on [Discord](https://discord.gg/YFtMjWwUN7).** Run
   `/candidate-testnet` in `#general-chat`. The bot assigns you the *Testnet
   Validator Candidate* role and opens access to `#testnet-onboarding`.
4. **Follow the pinned instructions** in `#testnet-onboarding`, which include
   registering your valoper profile on-chain through the
   [`r/gnops/valopers`](../../../examples/gno.land/r/gnops/valopers) realm.
5. **Run `/submit-request`** in `#testnet-onboarding` with your operator
   address. The team reviews it and either grants the *Testnet Validator* role
   or tells you what is still missing.
6. **Get voted in.** Entering the set is a GovDAO decision. Registering does not
   entitle you to a slot, on a testnet or on mainnet.

The `#testnet-*` channels are only visible once you hold the candidate or
validator role, so step 3 is what unlocks the rest. Current instructions are
announced in `#announcements` and `#general-chat`.

Mainnet onboarding will differ; ask on Discord rather than assuming this flow
carries over.

## Getting help

Ask on [Discord](https://discord.gg/YFtMjWwUN7) — `#general-chat` if you don't
hold a testnet role yet, the `#testnet-*` channels once you do. Say which
network you're on and paste your `config.toml` diff from that network's default;
that is usually enough to answer in one round-trip.
