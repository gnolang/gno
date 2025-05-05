package home

import (
	"std"
	"strconv"

	"gno.land/p/demo/ownable"
	"gno.land/p/demo/ui"
	"gno.land/p/moul/dynreplacer"
	blog "gno.land/r/gnoland/blog"
	"gno.land/r/gnoland/events"
	"gno.land/r/leon/hor"
)

var (
	override string
	Admin    = ownable.NewWithAddress("g1manfred47kzduec920z88wfr64ylksmdcedlf5") // @moul
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
	r.RegisterCallback(":chain-height:", func() string {
		return strconv.Itoa(int(std.ChainHeight()))
	})

	template := `# Welcome to gno.land

We’re building gno.land, set to become the leading open-source smart contract
platform, using Gno, an interpreted and fully deterministic variation of the
Go programming language for succinct and composable smart contracts.

With transparent and timeless code, gno.land is the next generation of smart
contract platforms, serving as the “GitHub” of the ecosystem, with realms built
using fully transparent, auditable code that anyone can inspect and reuse.

Intuitive and easy to use, gno.land lowers the barrier to web3 and makes
censorship-resistant platforms accessible to everyone. If you want to help lay
the foundations of a fairer and freer world, join us today. 

## Learn about gno.land

- [About](/about)
- [GitHub](https://github.com/gnolang)
- [Blog](/blog)
- [Events](/events)
- Tokenomics (soon)
- [Partners, Fund, Grants](/partners)
- [Explore the Ecosystem](/ecosystem)
- [Careers](https://jobs.ashbyhq.com/allinbits)

## Build with Gno

- [Write Gno in the browser](https://play.gno.land)
- [Read about the Gno Language](/gnolang)
- [Visit the official documentation](https://docs.gno.land)
- [Gno by Example](https://gno-by-example.com/)
- [Efficient local development for Gno](https://docs.gno.land/gno-tooling/cli/gno-tooling-gnodev)
- [Get testnet GNOTs](https://faucet.gno.land)

## Explore the universe

- [Discover demo packages](https://github.com/gnolang/gno/tree/master/examples)
- [Gnoscan](https://gnoscan.io)
- [Staging chain](https://docs.gno.land/concepts/portal-loop)
- [Testnet 6](https://test6.gno.land/)
- [Testnet 5](https://test5.gno.land/)
- [Faucet Hub](https://faucet.gno.land)

## [Latest Blogposts](/r/gnoland/blog)

:latest-blogposts:

## [Latest Events](/r/gnoland/events)

:upcoming-events:

## [Hall of Realms](/r/leon/hor)

:latest-hor:

---

## [Gno Playground](https://play.gno.land)


Gno Playground is a web application designed for building, running, testing, and
interacting with your Gno code, enhancing your understanding of the Gno
language. With Gno Playground, you can share your code, execute tests, deploy
your realms and packages to gno.land, and explore a multitude of other features.

Experience the convenience of code sharing and rapid experimentation with
[Gno Playground](https://play.gno.land).

## Explore New Packages and Realms

### [r/gnoland](https://github.com/gnolang/gno/tree/master/examples/gno.land/r/gnoland)

- [r/gnoland/blog](/r/gnoland/blog)
- [r/gnoland/users](/r/gnoland/users)
- [r/gnoland/home](/r/gnoland/home)
- [r/gnoland/pages](/r/gnoland/pages)

### [r/sys](https://github.com/gnolang/gno/tree/master/examples/gno.land/r/sys)

- [r/sys/names](/r/sys/names)
- [r/sys/users](/r/sys/users)
- [r/sys/rewards](/r/sys/rewards)
- [/r/sys/validators/v2](/r/sys/validators/v2)

### [r/demo](https://github.com/gnolang/gno/tree/master/examples/gno.land/r/demo)

- [r/demo/boards](/r/demo/boards)
- [r/demo/banktest](/r/demo/banktest)
- [r/demo/foo20](/r/demo/foo20)
- [r/demo/foo721](/r/demo/foo721)
- [r/demo/microblog](/r/demo/microblog)
- [r/demo/nft](/r/demo/nft)
- [r/demo/types](/r/demo/types)
- [r/demo/art/gnoface](/r/demo/art/gnoface)
- [r/demo/art/millipede](/r/demo/art/millipede)
- [r/demo/groups](/r/demo/groups)
- ...

### [p/demo](https://github.com/gnolang/gno/tree/master/examples/gno.land/p/demo)

- [p/demo/avl](/p/demo/avl)
- [p/demo/blog](/p/demo/blog)
- [p/demo/ui](/p/demo/ui)
- [p/demo/ufmt](/p/demo/ufmt)
- [p/demo/merkle](/p/demo/merkle)
- [p/demo/bf](/p/demo/bf)
- [p/demo/flow](/p/demo/flow)
- [p/demo/gnode](/p/demo/gnode)
- [p/demo/grc/grc20](/p/demo/grc/grc20)
- [p/demo/grc/grc721](/p/demo/grc/grc721)
- ...

---

## Socials

- Check out our [community projects](https://github.com/gnolang/awesome-gno)
- [Discord](https://discord.gg/S8nKUqwkPn)
- [Twitter](https://twitter.com/_gnoland)
- [Youtube](https://www.youtube.com/@_gnoland)
- [Telegram](https://t.me/gnoland)

## Quote of the ~Day~ Block#:chain-height:

> :qotb:

---

**This is a testnet.**
Package names are not guaranteed to be available for production.
`

	if override != "" {
		template = override
	}
	result := r.Replace(template)
	return result
}

func latestHOFItems(num int) ui.Element {
	submissions := hor.RenderExhibWidget(num)

	return ui.Element{
		ui.H2("[Hall of Realms](/r/leon/hor)"),
		ui.Text(submissions),
	}
}

func quoteOfTheBlock() string {
	quotes := []string{
		"Gno is for Truth.",
		"Gno is for Social Coordination.",
		"Gno is _not only_ for DeFi.",
		"Now, you Gno.",
		"Come for the Go, Stay for the Gno.",
	}
	height := std.ChainHeight()
	idx := int(height) % len(quotes)
	qotb := quotes[idx]
	return qotb
}

func AdminSetOverride(content string) {
	crossing()
	Admin.AssertOwnedByPrevious()
	override = content
}
