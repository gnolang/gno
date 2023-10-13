# Gno

> At first, there was Bitcoin, out of entropy soup of the greater All.
> Then, there was Ethereum, which was created in the likeness of Bitcoin,
> but made Turing complete.
>
> Among these were Tendermint and Cosmos to engineer robust PoS and IBC.
> Then came Gno upon Cosmos and there spring forth Gnoland,
> simulated by the Gnomes of the Greater Resistance.

Gno is an interpreted and fully-deterministic implementation of the Go
programming language, designed to build succint and composable smart contracts.
The first blockchain to use it is Gno.land, a
[Proof of Contribution](./docs/proof-of-contribution.md)-based chain, backed by
a variation of the [Tendermint](https://docs.tendermint.com/v0.34/introduction/what-is-tendermint.html)
consensus engine.

## Getting started

If you haven't already, take a moment to check out our [website](https://gno.land/).

> The website is a deployment of our [gnoweb](./gno.land/cmd/gnoweb) frontend; you
> can use it to check out [some](https://test3.gno.land/r/demo/boards) [example](https://test3.gno.land/r/gnoland/blog)
> [contracts](https://test3.gno.land/r/demo/users).
>
> Use the `[source]` button in the header to inspect the program's source; use
> the `[help]` button to view how you can use [`gnokey`](./gno.land/cmd/gnokey)
> to interact with the chain from your command line.

If you have already played around with the website, use our
[Getting Started](https://github.com/gnolang/getting-started) guide to learn how
to write and deploy your first smart contract. No local set-up required!

Once you're done, learn how to set up your local environment with the
[quickstart guide](./examples/gno.land/r/demo/boards/README.md) and the
[contributing guide](./CONTRIBUTING.md).

You can find out more existing tools & documentation for Gno on our
[awesome-gno](https://github.com/gnolang/awesome-gno) repository.
We look forward to seeing your first PR!

## Repository structure

* [examples](./examples) - smart-contract examples and guides for new Gno developers.
* [gnovm](./gnovm) - GnoVM and Gnolang.
* [gno.land](./gno.land) - Gno.land blockchain and tools.
* [tm2](./tm2) - Tendermint2.

## Socials & Contact

* [**Discord**](https://discord.gg/YFtMjWwUN7): good for general chat-based
  conversations, as well as for asking support on developing with Gno.
* [**Reddit**](https://www.reddit.com/r/gnoland): more "permanent" and
  forum-style discussions. Feel free to post anything Gno-related, as well as
  any question related to Gno programming!
* [**Telegram**](https://t.me/gnoland): unofficial Telegram group.
* [**Twitter**](https://twitter.com/_gnoland): official Twitter account. Follow
   us to know about new developments, events & official announcements about Gno!
* [**YouTube**](https://www.youtube.com/@_gnoland): here we post all of our
  video content, like workshops, talks and public development calls. Follow
  along on our development journey!

<details><summary>Short doc about all the commands</summary>

  User commands:

  * [gnokey](./gno.land/cmd/gnokey) - key manipulation, also general interaction with gnoland
  * [gnoland](./gno.land/cmd/gnoland) - runs the blockchain node
  * [gnoweb](./gno.land/cmd/gnoweb) - serves gno website, along with user-defined content
  * [logos](./misc/logos) - intended to be used as a browser

  Developer commands:

  * [gno](./gnovm/cmd/gno) - handy tool for developing gno packages & realms
  * [gnotxsync](./gno.land/cmd/gnotxsync) - importing/exporting transactions from local blockchain node storage
  * [goscan](./misc/goscan) - dumps imports from specified file’s AST
  * [genproto](./misc/genproto) - helper for generating .proto implementations
  * [gnofaucet](./gno.land/cmd/gnofaucet) - serves GNOT faucet
</details>

<details><summary>CI/CD/Tools badges and links</summary>

  GitHub Actions:

  * [![gno.land](https://github.com/gnolang/gno/actions/workflows/gnoland.yml/badge.svg)](https://github.com/gnolang/gno/actions/workflows/gnoland.yml)
  * [![gnovm](https://github.com/gnolang/gno/actions/workflows/gnovm.yml/badge.svg)](https://github.com/gnolang/gno/actions/workflows/gnovm.yml)
  * [![tm2](https://github.com/gnolang/gno/actions/workflows/tm2.yml/badge.svg)](https://github.com/gnolang/gno/actions/workflows/tm2.yml)
  * [![examples](https://github.com/gnolang/gno/actions/workflows/examples.yml/badge.svg)](https://github.com/gnolang/gno/actions/workflows/examples.yml)
  * [![docker](https://github.com/gnolang/gno/actions/workflows/docker.yml/badge.svg)](https://github.com/gnolang/gno/actions/workflows/docker.yml)

  Codecov:

  * General: [![codecov](https://codecov.io/gh/gnolang/gno/branch/master/graph/badge.svg?token=HPP82HR1P4)](https://codecov.io/gh/gnolang/gno)
  * tm2: [![codecov](https://codecov.io/gh/gnolang/gno/branch/master/graph/badge.svg?token=HPP82HR1P4&flag=tm2)](https://codecov.io/gh/gnolang/gno/tree/master/tm2)
  * gnovm: [![codecov](https://codecov.io/gh/gnolang/gno/branch/master/graph/badge.svg?token=HPP82HR1P4&flag=gnovm)](https://codecov.io/gh/gnolang/gno/tree/master/gnovm)
  * gno.land: [![codecov](https://codecov.io/gh/gnolang/gno/branch/master/graph/badge.svg?token=HPP82HR1P4&flag=gno.land)](https://codecov.io/gh/gnolang/gno/tree/master/gno.land)
  * examples: TODO

  Go Report Card:

  * [![Go Report Card](https://goreportcard.com/badge/github.com/gnolang/gno)](https://goreportcard.com/report/github.com/gnolang/gno)
  * tm2, gnovm, gno.land: TODO (blocked by tm2 split, because we need go mod workspaces)

  Pkg.go.dev

  * [![Go Reference](https://pkg.go.dev/badge/github.com/gnolang/gno.svg)](https://pkg.go.dev/github.com/gnolang/gno)
  * TODO: host custom docs on gh-pages, to bypass license limitation
</details>
