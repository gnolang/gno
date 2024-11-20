package source

// Welcome to the source code of this realm!

func Render(_ string) string {
	return `# Viewing source code  
gno.land makes it easy to view the source code of any pure
package or realm, by using ABCI queries.

Specifically for gno.land's universal frontend, ` + "`gnoweb`, " + ` makes this easy by
providing a intuitive UI that fetches the source of the
realm and providing a "Source" tab.

Check it out in the top right corner!
`
}
