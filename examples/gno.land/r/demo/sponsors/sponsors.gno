package sponsors

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/sponsors"
)

var (
	name2Sponsor  *avl.Tree
	addr2Sponsor  *avl.Tree
	totalSponsors int
)

func init() {
	name2Sponsor = avl.NewTree() // maping name -> Sponsor
	addr2Sponsor = avl.NewTree() // maping address -> Sponsor
	totalSponsors = 0
}

// RegisterSponsor registers a new sponsor
func RegisterSponsor(name, apiVersion, url, description, contactInfo string) {
	caller := std.PrevRealm().Addr()
	// assert not already registered.
	if found := addr2Sponsor.Has(caller.String()); found {
		panic("address already registered")
	}

	if found := name2Sponsor.Has(name); found {
		panic("sponsorName already registered")
	}

	sponsor := &sponsors.Sponsor{
		Address:     caller,
		Name:        name,
		ApiVersion:  apiVersion,
		DiscoverURL: url,
		Description: description,
		ContactInfo: contactInfo,
		Height:      std.GetHeight(),
	}

	name2Sponsor.Set(name, sponsor)
	addr2Sponsor.Set(caller.String(), sponsor)

	totalSponsors += 1
}

// GetSponsorByAddress retrieves the sponsor by their address
func GetSponsorByAddress(addr std.Address) *sponsors.Sponsor {
	value, found := addr2Sponsor.Get(addr.String())
	if !found {
		return nil
	}

	return value.(*sponsors.Sponsor)
}

// GetSponsorByName retrieves the sponsor by their name
func GetSponsorByName(name string) *sponsors.Sponsor {
	value, found := name2Sponsor.Get(name)
	if !found {
		return nil
	}

	return value.(*sponsors.Sponsor)
}

// GetURLByAddress retrieves the sponsor's URL by their address
func GetURLByAddress(addr std.Address) string {
	value, found := addr2Sponsor.Get(addr.String())
	if !found {
		return ""
	}

	return value.(*sponsors.Sponsor).DiscoverURL
}

// GetURLByName retrieves the sponsor's URL by their name
func GetURLByName(name string) string {
	value, found := name2Sponsor.Get(name)
	if !found {
		return ""
	}

	return value.(*sponsors.Sponsor).DiscoverURL
}

// GetList retrieves pairs of (name, address) with each sponsor
func GetList() []string {
	var sponsorList []string

	name2Sponsor.Iterate("", "", func(key string, value interface{}) bool {
		sponsor := value.(*sponsors.Sponsor)

		sponsorList = append(sponsorList, key, sponsor.Address.String())

		return false
	})

	return sponsorList
}

// GetTotalSponsors returns the total number of registered sponsors
func GetTotalSponsors() int {
	return totalSponsors
}

// Render displays information based on the given path
func Render(path string) string {
	if path == "" {
		return renderHome()
	} else if len(path) >= 38 { // 39? 40?
		if path[:2] != "g1" {
			return "invalid address " + path
		}
		sponsor := GetSponsorByAddress(std.Address(path))
		if sponsor == nil {
			// TODO: display basic information about account.
			return "unknown address " + path
		}
		return sponsor.Render()
	} else {
		sponsor := GetSponsorByName(path)
		if sponsor == nil {
			return "unknown sponsorname " + path
		}
		return sponsor.Render()
	}
}

// renderHome renders the home page with a list of sponsors
func renderHome() string {
	doc := ""
	name2Sponsor.Iterate("", "", func(key string, value interface{}) bool {
		sponsor := value.(*sponsors.Sponsor)
		doc += " * [" + sponsor.Name + "](/r/demo/sponsors:" + sponsor.Name + ")\n"
		return false
	})
	return doc
}
