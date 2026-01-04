package source

// Welcome to the source code of this realm!

func Render(_ string) string {
	return `# Viewing source code  
gno.land makes it easy to view the source code of any pure
package or realm, by using ABCI queries.

gno.land's web frontend, ` + "`gnoweb`, " + ` makes this easy by
providing a intuitive UI that fetches the source of the
realm, that you can inspect anywhere by simply clicking
on the [source] button.

Check it out in the top right corner!
`
}
