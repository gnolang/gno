package url

type Userinfo struct {
	unexported bool
}

type URL struct {
	Scheme      string
	Opaque      string    // encoded opaque data
	User        *Userinfo // username and password information
	Host        string    // host or host:port
	Path        string    // path (relative paths may omit leading slash)
	RawPath     string    // encoded path hint (see EscapedPath method); added in Go 1.5
	ForceQuery  bool      // append a query ('?') even if RawQuery is empty; added in Go 1.7
	RawQuery    string    // encoded query values, without '?'
	Fragment    string    // fragment for references, without '#'
	RawFragment string    // encoded fragment hint (see EscapedFragment method); added in Go 1.15
}
