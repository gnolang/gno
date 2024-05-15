package home

import (
	"std"

	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/ui"
	blog "gno.land/r/gnoland/blog"
)

// XXX: p/demo/ui API is crappy, we need to make it more idiomatic
// XXX: use an updatable block system to update content from a DAO
// XXX: var blocks avl.Tree

func Render(_ string) string {
	dom := ui.DOM{Prefix: "r/gnoland/home:"}
	dom.Title = "Welcome to Gno.land"

	// body
	dom.Body.Append(introSection()...)

	dom.Body.Append(ui.Jumbotron(discoverLinks()))

	dom.Body.Append(
		ui.Columns{3, []ui.Element{
			lastBlogposts(4),
			upcomingEvents(4),
			lastContributions(4),
		}},
	)

	dom.Body.Append(ui.HR{})
	dom.Body.Append(playgroundSection()...)
	dom.Body.Append(ui.HR{})
	dom.Body.Append(packageStaffPicks()...)
	dom.Body.Append(ui.HR{})
	dom.Body.Append(worxDAO()...)
	dom.Body.Append(ui.HR{})
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

func lastBlogposts(limit int) ui.Element {
	posts := blog.RenderLastPostsWidget(limit)
	return ui.Element{
		ui.H3("Latest Blogposts"),
		ui.Text(posts),
	}
}

func lastContributions(limit int) ui.Element {
	return ui.Element{
		ui.H3("Latest Contributions"),
		// TODO: import r/gh to
		ui.Link{Text: "View latest contributions", URL: "https://github.com/gnolang/gno/pulls"},
	}
}

func upcomingEvents(limit int) ui.Element {
	return ui.Element{
		ui.H3("Upcoming Events"),
		// TODO: replace with r/gnoland/events
		ui.Text("[View upcoming events](/events)"),
	}
}

func introSection() ui.Element {
	return ui.Element{
		ui.H3("We’re building Gno.land, the first open-source smart contract system, using Gno, an interpreted and fully deterministic variation of the Go programming language for succinct and composable smart contracts."),
		ui.Paragraph("With transparent and timeless code, Gno.land is the next generation of smart contract platforms, serving as the “GitHub” of the ecosystem, with realms built using fully transparent, auditable code that anyone can inspect and reuse."),
		ui.Paragraph("Intuitive and easy to use, Gno.land lowers the barrier to web3 and makes censorship-resistant platforms accessible to everyone. If you want to help lay the foundations of a fairer and freer world, join us today."),
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
		ui.H3("Contributions (WorxDAO & GoR)"),
		// TODO: GoR dashboard + WorxDAO topics
		ui.Text(`coming soon`),
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
		ui.H3(ufmt.Sprintf("Quote of the ~Day~ Block#%d", height)),
		ui.Quote(qotb),
	}
}

func socialLinks() ui.Element {
	return ui.Element{
		ui.H3("Socials"),
		ui.BulletList{
			// XXX: improve UI to support a nice GO api for such links
			ui.Text("Check out our [community projects](https://github.com/gnolang/awesome-gno)"),
			ui.Text("![Discord](static/img/ico-discord.svg) [Discord](https://discord.gg/S8nKUqwkPn)"),
			ui.Text("![Twitter](static/img/ico-twitter.svg) [Twitter](https://twitter.com/_gnoland)"),
			ui.Text("![Youtube](static/img/ico-youtube.svg) [Youtube](https://www.youtube.com/@_gnoland)"),
			ui.Text("![Telegram](static/img/ico-telegram.svg) [Telegram](https://t.me/gnoland)"),
		},
	}
}

func playgroundSection() ui.Element {
	return ui.Element{
		ui.H3("[Gno Playground](https://play.gno.land)"),
		ui.Paragraph(`Gno Playground is a web application designed for building, running, testing, and interacting
with your Gno code, enhancing your understanding of the Gno language. With Gno Playground, you can share your code,
execute tests, deploy your realms and packages to Gno.land, and explore a multitude of other features.`),
		ui.Paragraph("Experience the convenience of code sharing and rapid experimentation with [Gno Playground](https://play.gno.land)."),
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
					ui.H4("[r/gnoland](https://github.com/gnolang/gno/tree/master/examples/gno.land/r/gnoland)"),
					ui.BulletList{
						ui.Link{URL: "r/gnoland/blog"},
						ui.Link{URL: "r/gnoland/dao"},
						ui.Link{URL: "r/gnoland/faucet"},
						ui.Link{URL: "r/gnoland/home"},
						ui.Link{URL: "r/gnoland/pages"},
					},
					ui.H4("[r/sys](https://github.com/gnolang/gno/tree/master/examples/gno.land/r/sys)"),
					ui.BulletList{
						ui.Link{URL: "r/sys/names"},
						ui.Link{URL: "r/sys/rewards"},
						ui.Link{URL: "r/sys/validators"},
					},
				}, {
					ui.H4("[r/demo](https://github.com/gnolang/gno/tree/master/examples/gno.land/r/demo)"),
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
					ui.H4("[p/demo](https://github.com/gnolang/gno/tree/master/examples/gno.land/p/demo)"),
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
- [Blog](/blog)
- [Events](/events)
- Tokenomics (soon)
- [Partners, Fund, Grants](/partners)
- [Explore the Ecosystem](/ecosystem)

</div><!-- end column-->

<div class="column">

### Build with Gno

- [Write Gno in the browser](https://play.gno.land)
- [Read about the Gno Language](/gnolang)
- [Visit the official documentation](https://docs.gno.land)
- [Gno by Example](https://gno-by-example.com/)
- [Efficient local development for Gno](https://docs.gno.land/gno-tooling/cli/gno-tooling-gnodev)

</div><!-- end column-->
<div class="column">

### Explore the universe

- [Discover demo packages](https://github.com/gnolang/gno/tree/master/examples)
- [Gnoscan](https://gnoscan.io)
- [Portal Loop](https://docs.gno.land/concepts/portal-loop) 
- Testnet 4 (upcoming)
- [Testnet 3](https://test3.gno.land/) (archive)
- [Testnet 2](https://test2.gno.land/) (archive)
- Testnet Faucet Hub (soon)

</div><!-- end column-->
</div><!-- end columns-3-->`),
	}
}
