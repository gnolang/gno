package sponsor

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

var sponsors *avl.Tree

func init() {
	sponsors = avl.NewTree() // maping address -> Sponsor
}

type Sponsor struct {
	Name        string
	ApiVersion  string
	DiscoverURL string
	Description string
	ContactInfo string
}

// RegisterSponsor a new sponsor
func RegisterSponsor(name, apiVersion, url, description, contactInfo string) error {
	caller := std.GetOrigCaller()

	if found := sponsors.Has(caller.String()); found {
		return ufmt.Errorf("sponsor %s already existed", name)
	}

	sponsors.Set(caller.String(), Sponsor{
		Name:        name,
		ApiVersion:  apiVersion,
		DiscoverURL: url,
		Description: description,
		ContactInfo: contactInfo,
	})

	return nil
}

// Lookup a sponsor by name
func GetByAddress(addr std.Address) Sponsor {
	value, found := sponsors.Get(addr.String())
	if !found {
		panic(ufmt.Sprintf("sponsor %s not found", addr.String()))
	}

	return value.(Sponsor)
}
