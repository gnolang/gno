package home

import (
	"github.com/gnolang/gno/examples/gno.land/p/demo/ownable"
	"std"
	"strings"

	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/md"
)

var (
	sections []Section
	Ownable  = ownable.NewWithAddress("g1manfred47kzduec920z88wfr64ylksmdcedlf5") // @moul
)

type Section struct {
	text    func() string
	ordinal uint8
	enabled bool
}

func init() {
	sections = []Section{
		{
			text:    intro,
			ordinal: 0,
			enabled: true,
		},
		{
			text:    networkBanner,
			ordinal: 1,
			enabled: true,
		},
	}
}

func UpdateSection(ord1, ord2 uint8) error {
	if !Ownable.CallerIsOwner() {
		return ownable.ErrUnauthorized
	}

	if ord1 != ord2 {
		sections[ord1], sections[ord2] = sections[ord2], sections[ord1]
	}
}

func AddSection(text string, ordinal uint8, enabled bool) error {
	if !Ownable.CallerIsOwner() {
		return ownable.ErrUnauthorized
	}

}

func Render(_ string) string {

	var out string
	for _, section := range sections {
		if section.enabled {
			out += section.text()
		}
	}

	return out

	// what is gnoland & gno
	// Latest in gno.land:
	// registered users, blogs, hof items, new pkgs (maybe registry pattern)
	// learn about gno:pages, github
	//
	// User? view r/
	// Dev? View r/docs
	// view: examples  (p/ & r/),
	// footer:
	// looking at X chain's home realm render
	// links, other stats
}

func intro() string {
	out := md.H1("Welcome to gno.land")
	out += md.Paragraph("gno.land is xyz bla bla")
	out += md.H2("What is this website?")
	out += md.Paragraph(`
You're looking at the universal web frontend for gno.land - ` + "`gnoweb`" + `. Gnoweb queries the gno.land blockchain in 
	order to get data from smart contracts called realms. This text, and everything on this page is a smart contract 
	that represents gno.land's home page.`,
	)

	return out
}

func networkBanner() string {
	chain := "Local"
	chainid := std.GetChainID()

	switch chainid {
	case "portal-loop":
		chain = "the Portal Loop"
	case "dev":
		chain = "Local"
	default:
		if strings.HasPrefix(chainid, "test") {
			chain = "Test" + strings.TrimPrefix(chainid, "test")
		}
	}

	return md.H2(ufmt.Sprintf("You're on network `%s`!", chain))
}
