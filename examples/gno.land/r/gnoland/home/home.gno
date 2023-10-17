package home

import "std"

var body = `# Welcome to Gno.land

### An interpretation of the Golang (Go) programming language for advanced developers and intrepid pioneers to build succinct, composable smart contracts for social coordination.

If you’re concerned about information censorship and want to contribute to the #GnoWorldOrder, follow our socials to find out how.

Gno.land is in building mode. If you want to help lay the foundations of a fairer and freer world through innovative ideas and exceptional code, join us today.

---

### WorxDAO

TODO

- A
  - A1
    - A1A
    - A1B
  - A2
  - A3
    - A3A
      - A3A1
- B
- C

---

<div class="columns-3">
<div class="column">

### Learn about Gno.land

- [About](/about)
- [GitHub](https://github.com/gnolang)
- [Subscribe](#subscribe)
- [Tokenomics (soon)](#)
- [Blog](/blog)
- [Events](/events)
- [Partners, Fund, Grants](/partners)

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

## Explore New Packages and Realms

<div class="columns-3">
<div class="column">

#### r/gnoland
- [/r/gnoland/blog](/r/gnoland/blog)
- [/r/gnoland/dao](/r/gnoland/dao)
- [/r/gnoland/faucet](/r/gnoland/faucet)
- [/r/gnoland/home](/r/gnoland/home)
- [/r/gnoland/pages](/r/gnoland/pages)

#### r/system
- [/r/system/names](/r/system/names)
- [/r/system/rewards](/r/system/rewards)
- [/r/system/validators](/r/system/validators)

</div><!-- end column-->
<div class="column">

#### r/demo
- [/r/demo/boards](/r/demo/boards)
- [/r/demo/users](/r/demo/users)
- [/r/demo/banktest](/r/demo/banktest)
- [/r/demo/foo20](/r/demo/foo20)
- [/r/demo/foo721](/r/demo/foo721)
- [/r/demo/microblog](/r/demo/microblog)
- [/r/demo/nft](/r/demo/nft)
- [/r/demo/types](/r/demo/types)
- [/r/demo/art](/r/demo/art)
- [/r/demo/groups](/r/demo/groups)
- ...

</div><!-- end column-->
<div class="column">

#### p/demo
- [/p/demo/avl](/p/demo/avl)
- [/p/demo/blog](/p/demo/blog)
- [/p/demo/ui](/p/demo/ui)
- [/p/demo/ufmt](/p/demo/ufmt)
- [/p/demo/merkle](/p/demo/merkle)
- [/p/demo/bf](/p/demo/bf)
- [/p/demo/flow](/p/demo/flow)
- [/p/demo/gnode](/p/demo/gnode)
- [/p/demo/grc/grc20](/p/demo/grc/grc20)
- [/p/demo/grc/grc721](/p/demo/grc/grc721)
- ...

</div><!-- end column-->
</div><!-- end column-3-->

---

### Socials

- Check out our [community projects](https://github.com/gnolang/awesome-gno).
- ![Discord](static/img/ico-discord.svg) [Discord](https://discord.gg/S8nKUqwkPn)
- ![Twitter](static/img/ico-twitter.svg) [Twitter](https://twitter.com/_gnoland)
- ![Youtube](static/img/ico-youtube.svg) [Youtube](https://www.youtube.com/@_gnoland)
- ![Telegram](static/img/ico-telegram.svg) [Telegram](https://t.me/gnoland)

**This is a testnet.**
Package names are not guaranteed to be available for production.
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
	body += "\n\n" + randomBanner()

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
