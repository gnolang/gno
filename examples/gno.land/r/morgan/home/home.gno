package home

import "gno.land/r/demo/hof"

const staticHome = `# morgan's (gn)home

- [📝 sign my guestbook](/r/morgan/guestbook)
`

func init() { hof.Register() }

func Render(path string) string {
	return staticHome
}
