// Pacakge web25 provides an opinionated way to register an external web2 frontend to provide a "better" web2.5 experience.
package web25

import "strings"

type Config struct {
	CID string
	URL string
}

func (c *Config) SetRemoteFrontendByURL(url string) {
	c.CID = ""
	c.URL = url
}

func (c *Config) SetRemoteFrontendByCID(cid string) {
	c.CID = cid
	c.URL = ""
}

func (c Config) GetLink() string {
	if c.CID != "" {
		return "https://ipfs.io/ipfs/" + c.CID
	}
	return c.URL
}

// RenderFrontendLink displays a frontend link at the top of your realm's Render function in a concistent way to help gno visitors to have a consistent experience.
//
// if query is not nil, then it will check if it's not disable by ?no-web25, so that you can call the render function from an external point of view.
func RenderFrontendLink(path string) string {
	if strings.Contains(path, "?no-web25") {
		return ""
	}
	return "Discover this Realm from our better [web2.5 frontend](" + c.GetLink() + ")"
}
