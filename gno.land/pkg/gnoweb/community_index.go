package gnoweb

import (
	"fmt"
	"net/url"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/weburl"
)

// CommunityIndex says which community pages search engines may index, as set
// by -index-community. Official pages are always indexable.
type CommunityIndex int

const (
	// IndexNoCommunity keeps every community page out of search results. It
	// is the zero value, so a handler configured without a choice fails
	// closed.
	IndexNoCommunity CommunityIndex = iota
	// IndexRegisteredCommunity indexes the bare page of a package or user
	// under a registered name, and no other community page.
	IndexRegisteredCommunity
	// IndexAllCommunity indexes community pages like official ones.
	IndexAllCommunity
)

// String is the -index-community value of c, the one table of names.
func (c CommunityIndex) String() string {
	switch c {
	case IndexNoCommunity:
		return "none"
	case IndexRegisteredCommunity:
		return "registered"
	case IndexAllCommunity:
		return "all"
	default:
		return fmt.Sprintf("CommunityIndex(%d)", int(c))
	}
}

// MarshalText and UnmarshalText let the value be a flag.TextVar.
func (c CommunityIndex) MarshalText() ([]byte, error) { return []byte(c.String()), nil }

func (c *CommunityIndex) UnmarshalText(text []byte) error {
	for v := IndexNoCommunity; v <= IndexAllCommunity; v++ {
		if string(text) == v.String() {
			*c = v
			return nil
		}
	}
	return fmt.Errorf("unknown community index %q: want none, registered or all", text)
}

// robots is what a page tells search engines, in its meta and in the
// X-Robots-Tag header. The values order from strictest to loosest, and the
// zero value is the strictest, so an unset policy fails closed.
type robots int

const (
	noIndexNoFollow robots = iota
	noIndexFollow
	indexFollow
)

func (r robots) String() string {
	switch r {
	case indexFollow:
		return "index, follow"
	case noIndexFollow:
		return "noindex, follow"
	default:
		return "noindex, nofollow"
	}
}

// canonicalViews are the $ keys an indexed page may carry: gnoweb links to a
// package's source and to each file in it. Any other key ($help, $state, a
// key nobody links to) is a view of a page, not a page of its own.
func canonicalViews(q url.Values) bool {
	for key := range q {
		if key != "source" && key != "file" {
			return false
		}
	}
	return true
}

// robots decides what search engines may do with u, a page of kind k.
//
// A community page under "registered" is indexed only as the bare page of a
// package or user under a registered name, not as one of its files or its
// listing: once r/sys/names is enabled the chain lets nobody deploy under a
// name they do not hold, while anyone may deploy under their own address,
// which makes address namespaces free to throw away. Under "none" no
// community page is indexed, and under "all" community pages follow the rule
// of official ones.
//
// An official page is indexed, but not under a query or a $ view other than
// its source: both let any link multiply one page into as many URLs as it
// cares to write, and a query reaches Render. Its links stay followed.
func (p pagePolicy) robots(k pageKind, u *weburl.GnoURL) robots {
	if k == pageCommunity {
		switch p.index {
		case IndexAllCommunity:
		case IndexRegisteredCommunity:
			if u.Args != "" || len(u.Query) > 0 || len(u.WebQuery) > 0 || u.IsFile() || u.IsDir() ||
				isGnoAddress(u.Namespace()) {
				return noIndexNoFollow
			}
			return indexFollow
		default:
			return noIndexNoFollow
		}
	}
	if len(u.Query) > 0 || !canonicalViews(u.WebQuery) {
		return noIndexFollow
	}
	return indexFollow
}
