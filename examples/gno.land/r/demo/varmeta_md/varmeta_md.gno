package varmeta_md

import (
	"gno.land/p/demo/varmeta_md"
)

func runMarkdown(input string) string {
	return string(varmeta_md.MarkdownCommon([]byte(input)))
}

func introSection() string {
	markdown := "# Welcome to gno.land\n" +
		"### We’re building gno.land, set to become the leading open-source smart contract platform, using Gno, an interpreted and fully deterministic variation of the Go programming language for succinct and composable smart contracts.\n" +
		"With transparent and timeless code, gno.land is the next generation of smart contract platforms, serving as the “GitHub” of the ecosystem, with realms built using fully transparent, auditable code that anyone can inspect and reuse.\n\n" +
		"Intuitive and easy to use, gno.land lowers the barrier to web3 and makes censorship-resistant platforms accessible to everyone. If you want to help lay the foundations of a fairer and freer world, join us today.\n"
	return runMarkdown(markdown)
}

func discoverLinks() string {
	return runMarkdown(`
| Learn about gno.land &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| Build with Gno | Explore the universe                         |
|:---------------------------------------------|:---------------------------------------------|:---------------------------------------------|
| [About](/about)                              | [Write Gno in the browser](https://play.gno.land) | [Discover demo packages](https://github.com/gnolang/gno/tree/master/examples) |
| [GitHub](https://github.com/gnolang)         | [Read about the Gno Language](/gnolang)      | [Gnoscan](https://gnoscan.io)               |
| [Blog](/blog)                                | [Visit the official documentation](https://docs.gno.land) | [Portal Loop](https://docs.gno.land/concepts/portal-loop) |
| [Events](/events)                            | [Gno by Example](https://gno-by-example.com/) | [Testnet 4](https://test4.gno.land/) (Launched July 2024!) |
| Tokenomics (soon)                            | [Efficient local development for Gno](https://docs.gno.land/gno-tooling/cli/gno-tooling-gnodev) | [Testnet 3](https://test3.gno.land/) (archive) |
| [Partners, Fund, Grants](/partners)          | [Get testnet GNOTs](https://faucet.gno.land) | [Testnet 2](https://test2.gno.land/) (archive) |
| [Explore the Ecosystem](/ecosystem)          |                                              | Testnet Faucet Hub (soon)                   |
| [Careers](https://jobs.lever.co/allinbits?department=Gno.land) |                                              |                                             |

`)
}

func playgroundSection() string {
	markdown := "### [Gno Playground](https://play.gno.land)\n" +
		"Gno Playground is a web application designed for building, running, testing, and interacting with your Gno code, enhancing your understanding of the Gno language. With Gno Playground, you can share your code, execute tests, deploy your realms and packages to gno.land, and explore a multitude of other features.\n\n" +
		"Experience the convenience of code sharing and rapid experimentation with [Gno Playground](https://play.gno.land)."
	return runMarkdown(markdown)
}

func packageStaffPicks() string {
	markdown := `### Explore New Packages and Realms

| [r/gnoland](https://github.com/gnolang/gno/tree/master/examples/gno.land/r/gnoland) &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; | [r/demo](https://github.com/gnolang/gno/tree/master/examples/gno.land/r/demo) | [p/demo](https://github.com/gnolang/gno/tree/master/examples/gno.land/p/demo) |
|:--------------------------------------------|:--------------------------------------------|:--------------------------------------------|
| [r/gnoland/blog](/r/gnoland/blog)           | [r/demo/boards](/r/demo/boards)             | [p/demo/avl](/p/demo/avl)                   |
| [r/gnoland/dao](/r/gnoland/dao)             | [r/demo/users](/r/demo/users)               | [p/demo/blog](/p/demo/blog)                 |
| [r/gnoland/faucet](/r/gnoland/faucet)       | [r/demo/banktest](/r/demo/banktest)         | [p/demo/ui](/p/demo/ui)                     |
| [r/gnoland/home](/r/gnoland/home)           | [r/demo/foo20](/r/demo/foo20)               | [p/demo/ufmt](/p/demo/ufmt)                 |
| [r/gnoland/pages](/r/gnoland/pages)         | [r/demo/foo721](/r/demo/foo721)             | [p/demo/merkle](/p/demo/merkle)             |
|                                             | [r/demo/microblog](/r/demo/microblog)       | [p/demo/bf](/p/demo/bf)                     |
| [r/sys](https://github.com/gnolang/gno/tree/master/examples/gno.land/r/sys)         | [r/demo/nft](/r/demo/nft)                   | [p/demo/flow](/p/demo/flow)                 |
| [r/sys/names](/r/sys/names)                 | [r/demo/types](/r/demo/types)               | [p/demo/gnode](/p/demo/gnode)               |
| [r/sys/rewards](/r/sys/rewards)             | [r/demo/art/gnoface](/r/demo/art/gnoface)   | [p/demo/grc/grc20](/p/demo/grc/grc20)       |
| [r/sys/validators](/r/sys/validators)       | [r/demo/art/millipede](/r/demo/art/millipede) | [p/demo/grc/grc721](/p/demo/grc/grc721)     |
|                                             | [r/demo/groups](/r/demo/groups)             |                                             |
|                                             |                                             |                                             |
`
	return runMarkdown(markdown)
}

func SocialsAndQuotes() string {
	markdown := `| Socials &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; | Quote of the Day                            |
|:---------------------------------------------|:---------------------------------------------|
| [Check out our community projects](https://github.com/gnolang/awesome-gno) | Come for the Go, Stay for the Gno.           |
| [Discord](https://discord.gg/S8nKUqwkPn) |                                              |
| [Twitter](https://twitter.com/_gnoland)  |                                              |
| [Youtube](https://www.youtube.com/@_gnoland) |                                              |
| [Telegram](https://t.me/gnoland)       |                                              |
`
	return runMarkdown(markdown)
}

func worxDAO() string {
	markdown := "### Contributions (WorxDAO &amp; GoR)\n" +
		"coming soon"
	return runMarkdown(markdown)
}

func RenderHome() string {
	out := introSection()
	out += discoverLinks()
	out += playgroundSection()
	out += "\n---\n"
	out += packageStaffPicks()
	out += "\n---\n"
	out += worxDAO()
	out += "\n---\n"
	out += SocialsAndQuotes()
	out += "\n---\n"
	out += runMarkdown("**This is a testnet.** Package names are not guaranteed to be available for production.\n")
	out += runMarkdown("**This page was built using Markdown. Check more examples [here](/r/demo/varmeta_md:example)**")
	return out
}

func RenderExampleLists() string {
	input := `
# Unordered List

* Item 1
    * Subitem 1.1
    	* Subitem 1.1.1
* Item 2
    * Subitem 2.1
    	* Subitem 2.1.1


# Ordered List

1. First item
   1. Subitem 1.1
      1. Subitem 1.1.1
2. Second item
   1. Subitem 2.1
      1. Subitem 2.1.1
`

	out := runMarkdown(input)

	return out
}

func RenderExampleHeader() string {
	input := `
# Heading Level 1
## Heading Level 2
### Heading Level 3
#### Heading Level 4
##### Heading Level 5
###### Heading Level 6
`
	out := runMarkdown(input)
	return out
}

func RenderExampleCode() string {
	input := `# Code Block
Here is an inline code example: ` + "`" + `fmt.Println("Hello Gno!")` + "`" + `.

` + "``` go\n" + `func foo() bool {
	return true;
	}
` + "```\n"

	out := runMarkdown(input)

	return out
}

func RenderExampleBlockquote() string {
	input := `# Blockquote #
> This is a blockquote.

> Here is another blockquote with **bold** and _italic_ text.
`

	out := runMarkdown(input)

	return out
}

func RenderExampleLinkAndImage() string {
	input := `
# Link And Image
Here is a [link to VarMeta](https://www.var-meta.com/).
![Gno logo](https://www.var-meta.com/favicon-16x16.png)
`

	out := runMarkdown(input)

	return out
}

func RenderExampleEmphasis() string {
	input := `# Emphasis
This is **bold** text, and this is _italic_ text.
`
	out := runMarkdown(input)
	return out
}

func RenderExampleTable() string {
	// Markdown string containing a table
	input := `# Table
| Header 1 | Header 2 |
| -------- | -------- |
| Cell 1   | Cell 2   |
| Cell 3   | Cell 4   |
`
	// Call runMarkdown to convert the input Markdown to HTML
	out := runMarkdown(input)

	// Return the rendered HTML
	return out
}

func RenderExample() string {
	out := RenderExampleHeader()
	out += "\n---\n"
	out += RenderExampleLists()
	out += "\n---\n"
	out += RenderExampleLinkAndImage()
	out += "\n---\n"
	out += RenderExampleCode()
	out += "\n---\n"
	out += RenderExampleEmphasis()
	out += "\n---\n"
	out += RenderExampleBlockquote()
	out += "\n---\n"
	out += RenderExampleTable()
	return out
}

func Render(path string) string {
	switch {
	case path == "":
		return RenderHome()
	case path == "example":
		return RenderExample()
	default:
		return "404\n"
	}
}
