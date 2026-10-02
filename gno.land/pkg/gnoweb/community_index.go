package gnoweb

import (
	"fmt"

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

// indexable reports whether search engines may index u, a page of kind k.
// Under "registered", only the bare page of a package or user page is, and
// only under a registered name: once r/sys/names is enabled the chain lets
// nobody deploy under a name they do not hold, while anyone may deploy under
// their own address, which makes address namespaces free to throw away.
// Arguments, a query or a $ view multiply one package into as many URLs as
// a link cares to write, so none of them is indexed.
func (p pagePolicy) indexable(k pageKind, u *weburl.GnoURL) bool {
	if k != pageCommunity {
		return true
	}
	switch p.index {
	case IndexAllCommunity:
		return true
	case IndexRegisteredCommunity:
		return u.Args == "" && len(u.Query) == 0 && len(u.WebQuery) == 0 && !isAddress(u.Namespace())
	default:
		return false
	}
}
