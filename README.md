# Gno

[![License](https://img.shields.io/badge/license-GNO%20GPL%20v5-blue.svg)](LICENSE.md)
[![Go Reference](https://pkg.go.dev/badge/hey/google)](https://gnolang.github.io/gno/github.com/gnolang/gno.html)

> At first, there was Bitcoin, out of entropy soup of the greater All.
> Then, there was Ethereum, which was created in the likeness of Bitcoin,
> but made Turing complete.
>
> Among these were Tendermint and Cosmos to engineer robust PoS and IBC.
> Then came Gno upon Cosmos and there spring forth Gnoland,
> simulated by the Gnomes of the Greater Resistance.

**Gno is a blockchain platform that interprets a deterministic variant of Go 
for writing smart contracts, built on Tendermint consensus.**

Gno is an interpreted and fully-deterministic implementation of the Go
programming language, designed to build succinct and composable smart contracts.
The first blockchain to use it is gno.land, a contribution-based chain, backed
by a variation of the [Tendermint](./tm2) consensus engine.

PLEASE NOTE: This is NOT the same project or token as the excellent Gnosis.io project.

## Getting Started

Explore Gno through our comprehensive documentation:

- **[For Users](./docs#use-gnoland)** - Learn how to use Gno applications, 
  manage accounts, and interact with the blockchain
- **[For Builders](./docs#build-on-gnoland)** - Start writing smart contracts, 
  understand the Gno language, and deploy your applications  
- **[Resources](./docs#resources)** - Technical specifications, best 
  practices, and advanced topics

Visit [gno.land](https://gno.land) to see live smart contracts in action.

## Key Features

- **Go Syntax**: If you know Go, you know Gno
- **Deterministic Execution**: Fully predictable contract behavior
- **Composable Packages**: Import and reuse code like regular Go
- **Auto-Persisted State**: Global variables automatically saved between calls
- **Contribution System**: Rewarding open source contributors
- **Developer Experience**: Comprehensive tooling including testing, 
  debugging, and hot-reload development

## Documentation

- [Documentation](./docs/) - Complete documentation portal
- [Examples](./examples) - Sample contracts and patterns
- [Go API Reference](https://gnolang.github.io/gno/github.com/gnolang/gno.html) - 
  Go package documentation

## Gno Playground
<a href="https://play.gno.land/p/VxDC6AmKmK6?run.expr=println(Render(%22%22))">
  <img alt="play.gno.land" src="https://img.shields.io/badge/Play-Hello_World-691a00.svg?logo=data:image/svg%2bxml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIGZpbGw9Im5vbmUiIHZpZXdCb3g9IjAgMCAxNTggMTU4Ij48cGF0aCBkPSJtMTU2IDItNiA1LTIgMmE1NCA1NCAwIDAgMS0yNCAxMmwtMyAxaC00Yy02IDAtOS0xLTE1LTNhNjIgNjIgMCAwIDAtMzUtNmgtMWwtNCAxYTYzIDYzIDAgMCAwLTUwIDY4bDEgMmEyNyAyNyAwIDAgMCAwIDRsNiAxNWEzMyAzMyAwIDAgMSAyIDIxIDYyIDYyIDAgMCAxLTMgOCA2MSA2MSAwIDAgMS0xMyAyMGwtMyA0LTEgMWgxbDEtMmEyMDYgMjA2IDAgMCAxIDYtNWwyLTJhODggODggMCAwIDAgOC03bDEtMSAyLTIgMy0zYTY2MSA2NjEgMCAwIDEgMjYtMjRsNjItNjIgNC00IDMtMyAyLTMgOS05IDMtMyAyLTMgNi02IDExLTEzIDMtMyAxLTEtMSAxWk03OSAyNWM5IDEgMTcgNCAyMyA4IDQgMiA1IDQgNSA3bC0xIDMtMSAydjFsLTItMmE0MSA0MSAwIDAgMC0zNC0xMCA0MyA0MyAwIDAgMC0zMyAyNyA0OSA0OSAwIDAgMC0zIDIwIDMxIDMxIDAgMCAwIDEgNWwyIDRjMSA1IDQgOSA3IDEzbDIgMnYxbC01IDEtNC0xYy0yLTEtNS01LTctMTBsLTEtMmE2MSA2MSAwIDAgMS0zLTEzIDY2IDY2IDAgMCAxIDAtMTRjMS01IDMtMTEgNi0xNmE1MyA1MyAwIDAgMSA0MC0yNmg4Wm0yIDEzaDFsMTEgMyAxMCA4djRMNzkgNzhsLTI2IDI1Yy0yIDEtNCAxLTYtMWwtNS03YTQwIDQwIDAgMCAxLTUtMTljMC04IDMtMTYgOC0yM2w5LThhNDEgNDEgMCAwIDEgMTktOGw4IDFabTQ0IDEzLTMgNnY3YTIyMSAyMjEgMCAwIDEgMiAxNGwtMSAxMGE0NiA0NiAwIDAgMS01NyAzNWgtOWMtMyAwLTcgMi03IDNsNCAyYTU3IDU3IDAgMCAwIDMyIDVoM2wzLTFhNTUgNTUgMCAwIDAgMzQtODJsLTEgMVoiIGZpbGw9IiNmZmYiLz48cGF0aCBkPSJNMTEzIDU5Yy0xIDItMiA0LTEgNmwxIDNhMzggMzggMCAwIDEtNDggNDRoLTRsLTMgMmE0NCA0NCAwIDAgMCAxMCAzaDE1YTQyIDQyIDAgMCAwIDMxLTU4aC0xWiIgZmlsbD0iI2ZmZiIvPjwvc3ZnPg==" />
</a>
</br></br>

[Gno Playground](https://play.gno.land), available at 
[play.gno.land](https://play.gno.land), is a web app that allows users to 
write, share, and deploy Gno code. Developers can seamlessly test, debug, and 
deploy realms and packages on gno.land, while being able to collaborate with 
peers to work on projects together and seek assistance. A key feature of Gno 
Playground is the ability to get started without the need to install any tools 
or manage any services, offering immediate access and convenience for users.

**Note:** The playground may not always reflect the latest changes in this 
repository.

## Repository Structure

* [docs](./docs) - Official documentation
* [examples](./examples) - Smart contract examples and guides
* [gnovm](./gnovm) - GnoVM and the Gno language
* [gno.land](./gno.land) - Blockchain node and tools
* [tm2](./tm2) - Tendermint2 consensus engine
* [contribs](./contribs) - Additional tools and utilities
* [misc](./misc) - Various utilities and scripts

## Community

**Explore the Ecosystem:**
- [Gnoverse](https://github.com/gnoverse) - Community projects and initiatives
- [Awesome Gno](https://github.com/gnoverse/awesome-gno) - Curated list of 
  resources
- [Gnoscan](https://gnoscan.io) - Blockchain explorer
- [Gno Studio](https://gno.studio) - Web IDE for Gno development
- [Become a Gnome](./docs/builders/become-a-gnome.md) - Join the contributor 
  community

**Connect & Get Help:**
- [Discord](https://discord.gg/YFtMjWwUN7) - Real-time support and development 
  discussions
- [Twitter](https://twitter.com/_gnoland) - Official announcements and updates
- [YouTube](https://www.youtube.com/@_gnoland) - Tutorials, workshops, and 
  development calls
- [Workshops](https://github.com/gnolang/workshops) - Educational materials
- [Reddit](https://www.reddit.com/r/gnoland) - Forum-style discussions
- [Telegram](https://t.me/gnoland) - Community-run group

**Contribute:**
- [GitHub Issues](https://github.com/gnolang/gno/issues) - Report bugs and 
  request features
- [Contributing Guide](./CONTRIBUTING.md) - Guidelines for contributors


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

  * [![Go Reference](https://pkg.go.dev/badge/hey/google)](https://gnolang.github.io/gno/github.com/gnolang/gno.html) \
    (pkg.go.dev will not show our repository as it has a license it doesn't recognise)
</details>

