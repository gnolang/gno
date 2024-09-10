package home

import (
	"std"

	"gno.land/p/demo/ownable"
	"gno.land/p/demo/page_builder"
	blog "gno.land/r/gnoland/blog"
	events "gno.land/r/gnoland/events"
)

type Link struct {
	Name string
	URL  string
}

var (
	override string
	admin    = ownable.NewWithAddress("g1u7y667z64x2h7vc6fmpcprgey4ck233jaww9zq") // @manfred by default

	aboutLinks = []Link{
		{"About", "about"},
		{"GitHub", "https://github.com/gnolang"},
		{"Blog", "blog"},
		{"Events", "events"},
		{"Tokenomics (soon)", ""},
		{"Partners, Fund, Grants", "partners"},
		{"Explore the Ecosystem", "ecosystem"},
		{"Careers", "https://jobs.lever.co/allinbits?department=Gno.land"},
	}

	buildLinks = []Link{
		{"Write Gno in the browser", "https://play.gno.land"},
		{"Read about the Gno Language", "gnolang"},
		{"Visit the official documentation", "https://docs.gno.land"},
		{"Gno by Example", "https://gno-by-example.com/"},
		{"Efficient local development for Gno", "https://docs.gno.land/gno-tooling/cli/gno-tooling-gnodev"},
		{"Get testnet GNOTs", "https://faucet.gno.land"},
	}

	exploreLinks = []Link{
		{"Discover demo packages", "https://github.com/gnolang/gno/tree/master/examples"},
		{"Gnoscan", "https://gnoscan.io"},
		{"Portal Loop", "https://docs.gno.land/concepts/portal-loop"},
		{"Testnet 4 (Launched July 2024!)", "https://test4.gno.land/"},
		{"Testnet 3 (archive)", "https://test3.gno.land/"},
		{"Testnet 2 (archive)", "https://test2.gno.land/"},
		{"Testnet Faucet Hub (soon)", ""},
	}
)

func Render(_ string) string {
	body := page_builder.New(1)
	body.Append(renderIntro())
	body.Append(renderJumbotron())
	body.Append(renderEvents())
	body.HR()
	body.Append(renderPlayground())
	body.HR()
	body.H1("To be continued")

	return body.Render()
}

func renderIntro() string {
	container := page_builder.New(1)
	container.H1("Welcome to gno.land")
	container.H3("We’re building gno.land, set to become the leading open-source smart contract platform, using Gno, an interpreted and fully deterministic variation of the Go programming language for succinct and composable smart contracts.")
	container.P("With transparent and timeless code, gno.land is the next generation of smart contract platforms, serving as the “GitHub” of the ecosystem, with realms built using fully transparent, auditable code that anyone can inspect and reuse.")
	container.P("Intuitive and easy to use, gno.land lowers the barrier to web3 and makes censorship-resistant platforms accessible to everyone. If you want to help lay the foundations of a fairer and freer world, join us today.")
	return container.Render()
}

func renderJumbotron() string {
    jumbotron := page_builder.New(3)

    jumbotron.Append(renderLinkDiv("Learn about gno.land", aboutLinks))
    jumbotron.Append(renderLinkDiv("Build with Gno", buildLinks))
    jumbotron.Append(renderLinkDiv("Explore the universe", exploreLinks))

    return jumbotron.Render()
}

func renderLinkDiv(header string, links []Link) string {
    linkDiv := page_builder.New(1)
    linkDiv.H3(header)
    
    var listItems []string
    for _, link := range links {
        listItems = append(listItems, linkDiv.Link(link.URL, link.Name))
    }

    linkDiv.UL(listItems)
    return linkDiv.Render()
}

func renderEvents() string {
    container := page_builder.New(3)

	blogSection := page_builder.New(1)
	blogSection.H3(blogSection.Link("/r/gnoland/blog", "Latest Blogposts"))
	blogSection.Append(blog.RenderLastPostsWidget(4))
	container.Append(blogSection.Render())
	
	eventSection := page_builder.New(1)
	eventSection.H3(eventSection.Link("/r/gnoland/events", "Latest Events"))
	eventList, _ := events.RenderEventWidget(events.MaxWidgetSize)
	eventSection.Append(eventList)
	container.Append(eventSection.Render())

	contributionSection := page_builder.New(1)
	contributionSection.H3("Latest Contributions")
	contributionSection.P(contributionSection.Link("https://github.com/gnolang/gno/pulls", "View latest contributions"))
	container.Append(contributionSection.Render())

    return container.Render()
}

func renderPlayground() string {
	container := page_builder.New(1)
	container.H3(container.Link("https://play.gno.land/", "Gno Playground"))
	container.P("Gno Playground is a web application designed for building, running, testing, and interacting with your Gno code, enhancing your understanding of the Gno language. With Gno Playground, you can share your code, execute tests, deploy your realms and packages to gno.land, and explore a multitude of other features.")
	container.P("Experience the convenience of code sharing and rapid experimentation with " + container.Link("https://play.gno.land", "Gno Playground."))
	return container.Render()
}

func AdminSetOverride(content string) {
	admin.AssertCallerIsOwner()
	override = content
}

func AdminTransferOwnership(newAdmin std.Address) {
	admin.AssertCallerIsOwner()
	admin.TransferOwnership(newAdmin)
}
