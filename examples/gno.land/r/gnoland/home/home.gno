package home

import "std"

var body = `# Welcome to Gno.land

---

## An interpretation of the Golang (Go) programming language for advanced developers and intrepid pioneers to build succinct, composable smart contracts for social coordination.

If you’re concerned about information censorship and want to contribute to the #GnoWorldOrder, follow our socials to find out how.

Gno.land is in building mode. If you want to help lay the foundations of a fairer and freer world through innovative ideas and exceptional code, join us today.

<div class="stack">

[About Gno.land](/about)
[Explore Github repo](https://github.com/gnolang)

</div><!-- end stack-->
<div class="jumbotron">

### Game of Realms

The first high-stakes contest will see participants compete for tiered membership to co-own the Gno.land blockchain. A series of complex technical and non-technical tasks will challenge contributors to create innovative patterns that push the chain to new limits. Start building the foundation for tomorrow through key smart contracts and other contributions that change our understanding of the world.

<div class="stack">

[Join Game Of Realms](/game-of-realms)

</div><!-- end stack-->
</div><!-- end jumbotron-->

### Fund and Grants Program

Are you a builder, tinkerer, or researcher? If you’re looking to create awesome dApps, tooling, infrastructure, or smart contract libraries on Gno.land, you can apply for a grant. The Gno.land Ecosystem Fund and Grants program provides financial contributions for individuals and teams to innovate on the platform.

<div class="stack">

[More information here](https://github.com/gnolang/ecosystem-fund-grants)

</div><!-- end stack-->

<div class="columns-3">
<div class="column">

### Learn about Gno.land

- [About](/about)
- [Subscribe](#subscribe)
- [Tokenomics (soon)](#)
- [Blog](/blog)
- [Events](/events)
- [Partners (soon)](#)

</div><!-- end column-->

<div class="column">

### Build with Gnolang

- [Gno dev with CLI (soon)](#)
- [Explore the Universe](/ecosystem)
- [Test in the browser (soon)](#)
- [About the Gnolang Language](/gnolang)
- [Docs/ Tutorials](https://github.com/gnolang)
- [Gno by example](https://gno-by-example.com/)
- [Getting started video (soon)](#)

</div><!-- end column-->
<div class="column">

### Explore the universe

- [Discover demo packages](https://github.com/gnolang/gno/tree/master/examples)
- [Install Gno Key instructions](/r/demo/boards:testboard/5)
- [Testnets 3](https://test3.gno.land/)
- [Testnets 2](https://test2.gno.land/)
- [Explorer links(soon)](#)
- [Testnet Tokens (faucet)](https://test3.gno.land/faucet)

</div><!-- end column-->
</div><!-- end columns-3-->

---

### Socials

- ![Discord](static/img/ico-discord.svg) [Discord](https://discord.gg/S8nKUqwkPn)
- ![Twitter](static/img/ico-twitter.svg) [Twitter](https://twitter.com/_gnoland)
- ![Youtube](static/img/ico-youtube.svg) [Youtube](https://www.youtube.com/@_gnoland)
- ![Telegram](static/img/ico-telegram.svg) [Telegram](https://t.me/gnoland)
`

var body2 = `# Welcome to **Gno.land**

- [About Gno.land](/r/gnoland/pages:about)
- [Blogs](/r/gnoland/blog)
- [Install ` + "`" + `gnokey` + "`" + `](https://github.com/gnolang/gno/tree/master/gno.land/cmd/gnokey)
- [Acquire testnet tokens](/faucet)
- [Game of Realms](/game-of-realms) - An open worldwide competition for developers to build the best Gnolang smart-contracts.

# Explore new packages.

- r/gnoland
  - [/r/gnoland/blog](/r/gnoland/blog)
  - [/r/gnoland/faucet](/r/gnoland/faucet)
- r/system
  - [/r/system/names](/r/system/names)
  - [/r/system/rewards](/r/system/rewards)
  - [/r/system/validators](/r/system/validators)
- r/demo
  - [/r/demo/banktest](/r/demo/banktest)
  - [/r/demo/boards](/r/demo/boards)
  - [/r/demo/foo20](/r/demo/foo20)
  - [/r/demo/nft](/r/demo/nft)
  - [/r/demo/types](/r/demo/types)
  - [/r/demo/users](/r/demo/users)
  - [/r/demo/groups](/r/demo/groups)
- p/demo
  - [/p/demo/avl](/p/demo/avl)
  - [/p/demo/blog](/p/demo/blog)
  - [/p/demo/flow](/p/demo/flow)
  - [/p/demo/gnode](/p/demo/gnode)
  - [/p/demo/grc/exts](/p/demo/grc/exts)
  - [/p/demo/grc/grc20](/p/demo/grc/grc20)
  - [/p/demo/grc/grc721](/p/demo/grc/grc721)

# Other Testnets

- **[staging.gno.land](https://staging.gno.land) (wiped every commit to master)**
- _[test3.gno.land](https://test3.gno.land) (latest)_
- _[test2.gno.land](https://test2.gno.land) (archive)_
- _[test1.gno.land](https://test1.gno.land) (archive)_

**This is a testnet.**
Package names are not guaranteed to be available for production.

# Social

Check out our [community projects](https://github.com/gnolang/awesome-gno).

Official channel: [Discord](https://discord.gg/S8nKUqwkPn)<br />
Other channels: [Telegram](https://t.me/gnoland) [Twitter](https://twitter.com/_gnoland) [Youtube](https://www.youtube.com/@_gnoland)
`

var adminAddr std.Address = std.Address("")

func Render(_ string) string {
	// highest level == highest responsibility
	// teams are responsible for components they don't owne
	// flag : realm maintainers VS facilitators
	// teams
	// committee of trustees to create the directory
	// each directory is a name, has a parent and have groups
	// homepage team - blocks aggregating events

	/*`
	# Directory

	* gno.land (owned by group)
	  *
	* gnovm
	  * gnolang (language)
	  * gnovm
	    - current challenges / concerns / issues
	* tm2
	  * amino
	  *

	## Contributors
	``*/

	// XXX: dynamic content: contributions
	// XXX:
	// XXX: community curated picks
	// XXX:
	body += randomBanner()
	return body
}

func randomBanner() string {
	// not sure to keep this...
	banners := []string{
		"Gno is for Truth.",
		"Gno is for Social Coordination.",
		"Gno is _not only_ for DeFi.",
	}
	height := std.GetHeight()
	idx := int(height) % len(banners)
	return banners[idx]
}
