package home

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/ui"
)

// XXX: p/demo/ui API is crappy, we need to make it more idiomatic
// XXX: use an updatable block system to update content from a DAO
// XXX: var blocks avl.Tree

func Render(_ string) string {
	dom := ui.DOM{Prefix: "r/gnoland/home:"}
	dom.Title = "Welcome to Gno.land"

	// body
	dom.Body.Append(introSection()...)
	dom.Body.Append(ui.Jumbotron(worxDAO()...))
	dom.Body.Append(packageStaffPicks()...)
	dom.Body.Append(ui.Jumbotron(discoverLinks()...))

	// footer
	dom.Footer.Append(
		ui.Columns{2, []ui.Element{
			socialLinks(),
			quoteOfTheBlock(),
		}},
	)

	// Testnet disclaimer
	dom.Footer.Append(
		ui.HR{},
		ui.Bold("This is a testnet."),
		ui.Text("Package names are not guaranteed to be available for production."),
	)

	return dom.String()
}

func introSection() ui.Element {
	return ui.Element{
		ui.H3("An interpretation of the Golang (Go) programming language for advanced developers and intrepid pioneers to build succinct, composable smart contracts for social coordination."),
		ui.Paragraph("If you’re concerned about information censorship and want to contribute to the #GnoWorldOrder, follow our socials to find out how."),
		ui.Paragraph("Gno.land is in building mode. If you want to help lay the foundations of a fairer and freer world through innovative ideas and exceptional code, join us today."),
	}
}

func worxDAO() ui.Element {
	// WorxDAO
	// XXX(manfred): please, let me finish a v0, then we can iterate
	// highest level == highest responsibility
	// teams are responsible for components they don't owne
	// flag : realm maintainers VS facilitators
	// teams
	// committee of trustees to create the directory
	// each directory is a name, has a parent and have groups
	// homepage team - blocks aggregating events
	// XXX: TODO
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
	return ui.Element{
		ui.H3("WorxDAO (WIP)"),
		ui.Text(`- A
  - A1
    - A1A
    - A1B
  - A2
  - A3
    - A3A
      - A3A1
- B
- C`),
	}
}

func quoteOfTheBlock() ui.Element {
	quotes := []string{
		"Gno is for Truth.",
		"Gno is for Social Coordination.",
		"Gno is _not only_ for DeFi.",
		"Now, you Gno.",
		"Come for the Go, Stay for the Gno.",
	}
	height := std.GetHeight()
	idx := int(height) % len(quotes)
	qotb := quotes[idx]

	return ui.Element{
		ui.H3(ufmt.Sprintf("Quote of the ~Day~Block#%d", height)),
		ui.Quote(qotb),
	}
}

func socialLinks() ui.Element {
	return ui.Element{
		ui.H3("Socials"),
		ui.BulletList{
			// XXX: improve UI to support a nice GO api for such links
			ui.Text("Check out our [community projects](https://github.com/gnolang/awesome-gno).)"),
			ui.Text("![Discord](static/img/ico-discord.svg) [Discord](https://discord.gg/S8nKUqwkPn)"),
			ui.Text("![Twitter](static/img/ico-twitter.svg) [Twitter](https://twitter.com/_gnoland)"),
			ui.Text("![Youtube](static/img/ico-youtube.svg) [Youtube](https://www.youtube.com/@_gnoland)"),
			ui.Text("![Telegram](static/img/ico-telegram.svg) [Telegram](https://t.me/gnoland)"),
		},
	}
}

func packageStaffPicks() ui.Element {
	// XXX: make it modifiable from a DAO
	return ui.Element{
		ui.H3("Explore New Packages and Realms"),
		ui.Columns{
			3,
			[]ui.Element{
				{
					ui.H4("r/gnoland"),
					ui.BulletList{
						ui.Link{URL: "r/gnoland/blog"},
						ui.Link{URL: "r/gnoland/dao"},
						ui.Link{URL: "r/gnoland/faucet"},
						ui.Link{URL: "r/gnoland/home"},
						ui.Link{URL: "r/gnoland/pages"},
					},
					ui.H4("r/system"),
					ui.BulletList{
						ui.Link{URL: "r/system/names"},
						ui.Link{URL: "r/system/rewards"},
						ui.Link{URL: "r/system/validators"},
					},
				}, {
					ui.H4("r/demo"),
					ui.BulletList{
						ui.Link{URL: "r/demo/boards"},
						ui.Link{URL: "r/demo/users"},
						ui.Link{URL: "r/demo/banktest"},
						ui.Link{URL: "r/demo/foo20"},
						ui.Link{URL: "r/demo/foo721"},
						ui.Link{URL: "r/demo/microblog"},
						ui.Link{URL: "r/demo/nft"},
						ui.Link{URL: "r/demo/types"},
						ui.Link{URL: "r/demo/art"},
						ui.Link{URL: "r/demo/groups"},
						ui.Text("..."),
					},
				}, {
					ui.H4("p/demo"),
					ui.BulletList{
						ui.Link{URL: "p/demo/avl"},
						ui.Link{URL: "p/demo/blog"},
						ui.Link{URL: "p/demo/ui"},
						ui.Link{URL: "p/demo/ufmt"},
						ui.Link{URL: "p/demo/merkle"},
						ui.Link{URL: "p/demo/bf"},
						ui.Link{URL: "p/demo/flow"},
						ui.Link{URL: "p/demo/gnode"},
						ui.Link{URL: "p/demo/grc/grc20"},
						ui.Link{URL: "p/demo/grc/grc721"},
						ui.Text("..."),
					},
				},
			},
		},
	}
}

func discoverLinks() ui.Element {
	return ui.Element{
		ui.Text(`<div class="columns-3">
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
</div><!-- end columns-3-->`),
	}
}
