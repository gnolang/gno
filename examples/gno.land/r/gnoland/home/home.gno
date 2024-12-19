package home

import "github.com/gnolang/gno/examples/gno.land/p/moul/md"

var sections []string

func Render(_ string) string {
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
	out += md.H2("We're building gno.land, xyz")
	out += md.Paragraph(`gno.land is a blockchain platform that allows users
to have full transparency over the applications they are using`)

	text := `
# Welcome to gno.land

	gno.land is xyz bla bla



	## What is this website?
	You're looking at the universal web frontend for gno.land - ` + "`gnoweb`" + `. Gnoweb queries the gno.land blockchain in 
	order to get data from smart contracts called realms. This text, and everything on this page is a smart contract 
	that represents gno.land's home page.`

	return out
}
