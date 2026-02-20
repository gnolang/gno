package home

import (
	"chain/runtime"
	"strconv"

	"gno.land/p/moul/dynreplacer"
	"gno.land/p/nt/ownable"
	blog "gno.land/r/gnoland/blog"

	"gno.land/p/leon/svgbtn"
	"gno.land/r/devrels/events"
	users "gno.land/r/gnoland/users/v1"
	"gno.land/r/leon/hor"
)

var (
	override string
	Admin    = ownable.NewWithAddressByPrevious("g1manfred47kzduec920z88wfr64ylksmdcedlf5") // @moul
)

func Render(_ string) string {
	r := dynreplacer.New()
	r.RegisterCallback(":latest-blogposts:", func() string {
		return blog.RenderLastPostsWidget(4)
	})
	r.RegisterCallback(":upcoming-events:", func() string {
		out, _ := events.RenderEventWidget(events.MaxWidgetSize)
		return out
	})
	r.RegisterCallback(":latest-hor:", func() string {
		return hor.RenderExhibWidget(5)
	})
	r.RegisterCallback(":qotb:", quoteOfTheBlock)
	r.RegisterCallback(":newsletter-button:", newsletterButton)
	r.RegisterCallback(":latest-gnomes:", func() string {
		return users.RenderLatestUsersWidget(5)
	})
	r.RegisterCallback(":chain-height:", func() string {
		return strconv.Itoa(int(runtime.ChainHeight()))
	})

	template := `# Welcome to Gno.land

We’re building Gno.land, set to become the leading open-source smart contract
platform, using Gno, an interpreted and fully deterministic variation of the
Go programming language for succinct and composable smart contracts.

With transparent and timeless code, Gno.land is the next generation of smart
contract platforms, serving as the “GitHub” of the ecosystem, with realms built
using fully transparent, auditable code that anyone can inspect and reuse.

Intuitive and easy to use, Gno.land lowers the barrier to web3 and makes
censorship-resistant platforms accessible to everyone. If you want to help lay
the foundations of a fairer and freer world, join us today.


<gno-columns>
## Learn about Gno.land

- [About](/about)
- [GitHub](https://github.com/gnolang)
- [Blog](/blog)
- [Events](/events)
- [Partners, Fund, Grants](/partners)
- [Explore the Ecosystem](/ecosystem)
- [Careers](https://jobs.ashbyhq.com/allinbits)

|||

## Build with Gno

- [Write Gno in the browser](https://play.gno.land)
- [Read about the Gno Language](/gnolang)
- [Visit the official documentation](https://docs.gno.land)
- [Gno by Example](/r/docs/home)
- [Efficient local development for Gno](https://docs.gno.land/builders/local-dev-with-gnodev)
- [Get testnet GNOTs](https://faucet.gno.land)

|||

## Explore the universe

- [Discover demo packages](https://github.com/gnolang/gno/tree/master/examples)
- [Gnoscan](https://gnoscan.io)
- [Staging chain](https://docs.gno.land/resources/gnoland-networks/#staging-environments-portal-loops)
- [Testnet 11](https://test11.testnets.gno.land/)
- [Faucet Hub](https://faucet.gno.land)
- [Boards](https://gno.land/r/gnoland/boards2/v1:OpenDiscussions): community forum

</gno-columns>

<gno-columns>

## [Latest Blogposts](/r/gnoland/blog)

:latest-blogposts:

|||

## [Latest Events](/r/devrels/events)

:upcoming-events:

|||

## [Hall of Realms](/r/leon/hor)

:latest-hor:

</gno-columns>

---

## [Gno Playground](https://play.gno.land)

Gno Playground is a web application designed for building, running, testing, and
interacting with your Gno code, enhancing your understanding of the Gno
language. With Gno Playground, you can share your code, execute tests, deploy
your realms and packages to Gno.land, and explore a multitude of other features.

Experience the convenience of code sharing and rapid experimentation with
[Gno Playground](https://play.gno.land).

---

## Explore New Packages and Realms

All code in Gno.land is organized in packages, and each package lives at a unique package path like
"r/gnoland/home". You can browse packages, inspect their source, and use them in your own libraries and realms.

<gno-columns>

### r/gnoland

Official realm packages developed by the Gno.land core team.

[Browse](/r/gnoland)

|||

### r/sys

System-level realm packages used by the chain.

[Browse](/r/sys)

|||

### r/demo

Demo realm packages showcasing what’s possible.

[Browse](/r/demo)

|||

### p/demo

Pure packages for demo purposes.

[Browse](/p/demo)

</gno-columns>

---

<gno-columns>
## Latest Gnomes

Latest Gnomes in the [registry](/r/gnoland/users/v1).

:latest-gnomes:

|||

## Socials

- Check out our [community projects](https://github.com/gnolang/awesome-gno)
- [Discord](https://discord.gg/S8nKUqwkPn)
- [Twitter](https://twitter.com/_gnoland)
- [Youtube](https://www.youtube.com/@_gnoland)
- [Telegram](https://t.me/gnoland)

|||

## Quote of the ~Day~ Block #:chain-height:

> :qotb:

</gno-columns>

---

## Sign up for our newsletter

Stay in the Gno by signing up for our newsletter. You'll get the scoop on dev updates, fresh content, and community news.

:newsletter-button:

---

**This is a testnet.** Package names are not guaranteed to be available for production.`

	if override != "" {
		template = override
	}
	result := r.Replace(template)
	return result
}

func newsletterButton() string {
	return svgbtn.Button(
		256,
		44,
		"#226c57",
		"#ffffff",
		"Subscribe to stay in the Gno",
		"https://land.us18.list-manage.com/subscribe?u=8befe3303cf82796d2c1a1aff&id=271812000b",
	)
}

func quoteOfTheBlock() string {
	quotes := []string{
		"Gno is for Truth.",
		"Gno is for Social Coordination.",
		"Gno is _not only_ for DeFi.",
		"Now, you Gno.",
		"Come for the Go, Stay for the Gno.",
	}
	height := runtime.ChainHeight()
	idx := int(height) % len(quotes)
	qotb := quotes[idx]
	return qotb
}

func AdminSetOverride(cur realm, content string) {
	Admin.AssertOwnedByPrevious()
	override = content
}

func AdminTransferOwnership(cur realm, newOwner address) {
	if err := Admin.TransferOwnershipByPrevious(newOwner); err != nil {
		panic(err)
	}
}
