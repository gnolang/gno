# Running a node

Usually you don't need to. Running a node is an infrastructure job, and
most of what people want one for is already covered by a tool — or a
public endpoint — that needs no node at all.

So before anything else, name what you're actually trying to do:

| I want to…                                          | What you actually need           | Where to go                                                     |
|-----------------------------------------------------|----------------------------------|-----------------------------------------------------------------|
| Write, test, and debug contracts                    | `gnodev`, or a current testnet   | [Develop contracts](#develop-contracts)                         |
| Read chain state, or build an app on top of a chain  | A public RPC endpoint, or an indexer | [Read a chain](#read-a-chain)                               |
| My own synced, low-latency copy of a public network  | A full node                      | [Join a network as a full node](#join-a-network-as-a-full-node)  |
| Sign blocks on a public network                     | A validator node                 | [Become a validator](#become-a-validator)                       |
| Learn how a Gno chain works, or just for fun        | A throwaway local chain          | [Run a chain locally](#run-a-chain-locally)                     |

## Develop contracts

No node needed. [`gnodev`](../resources/gnodev.md) boots a local devnet
with hot reload, pre-funded accounts, and a built-in web UI — a faster
loop than any real chain. `gno test` runs your tests with no chain at
all, and the [Playground](https://play.gno.land) runs Gno in your
browser with nothing installed.

When you want your code on a shared chain, deploy to a current testnet
rather than operating your own — see
[Networks](../resources/gnoland-networks.md).

Start with [Getting started](./getting-started.md), or
[Quick Start](./quickstart.md) if you already know Go.

## Read a chain

No node needed. Every network exposes a public RPC endpoint — see
[Networks](../resources/gnoland-networks.md) — and
[`tx-indexer`](https://github.com/gnolang/tx-indexer) serves indexed
history over GraphQL for the queries an RPC node can't answer
efficiently.

- [RPC clients](./rpc-clients.md) — Go and JavaScript clients.
- [Ways to interact with Gno.land](../resources/comparison-of-ways-to-interact-with-gnoland.md)
  — pick the right client for your language and platform.

Run your own node for this only if you need the state locally: heavy
read traffic, custom indexing, or archival history.

## Join a network as a full node

A full node syncs a public network's blocks without taking part in
consensus. Two things to read, and neither is in this section of the
docs: the operator guide in
[`gno.land/cmd/gnoland`](../../gno.land/cmd/gnoland),
and the target network's own genesis, peer list, and `README.md` under
[deployment files](../resources/gnoland-networks.md#deployment-files) —
each network pins its own.

:::warning
Don't point a `master` build at a released network. Chains run pinned
releases; a node built from `master` will not reach consensus with them.
:::

## Become a validator

This is not "a full node, but voted in". It's a full node behind
sentries, signing through a remote signer (tmkms), monitored, run by an
operator who completed the network's onboarding — identity checks
included — and who stays reachable on Discord for upgrades and
incidents. Then GovDAO votes; registering entitles you to nothing.

Three places, in order:

- [Discord](https://discord.gg/YFtMjWwUN7) — ask for the validator role
  in `#testnet-general`; onboarding and coordination happen there.
- [`gno.land/cmd/gnoland`](../../gno.land/cmd/gnoland)
  — the operator documentation: sentry architecture, remote signing,
  hardware, and the full checklist.
- [`misc/deployments/`](../../misc/deployments)
  — the target network's own genesis, peers, and `VALIDATOR.md`.

## Run a chain locally

Here a node is the point: no network to join, no genesis to fetch — just
`gnoland` running a single-validator chain you can break. See
[`gno.land/cmd/gnoland`](../../gno.land/cmd/gnoland).

Install the binary with the `--full` flag of the
[one-line installer](./install.md).

## Still not sure?

Ask on [Discord](https://discord.gg/YFtMjWwUN7). Say what you're trying
to build; someone will tell you whether a node is part of the answer.
