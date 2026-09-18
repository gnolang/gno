package home

import "gno.land/r/leon/hor"

const staticHome = `# morgan's (gn)home

- [📝 sign my guestbook](/r/morgan/guestbook)
`

func init(cur realm) { hor.Register(cross(cur), "Morgan's Home Realm", "") }

func Render(path string) string {
	return staticHome
}
