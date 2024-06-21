package sponsors

import (
	"std"
	"strconv"
)

type Sponsor struct {
	Name        string
	Address     std.Address
	ApiVersion  string
	DiscoverURL string
	Description string
	ContactInfo string
	Height      int64
}

// Render function to display sponsor information
func (s *Sponsor) Render() string {
	str := "## Sponsor: " + s.Name + "\n" +
		"\n" +
		" * Address: " + s.Address.String() + "\n" +
		" * API Version: " + s.ApiVersion + "\n" +
		" * Discover URL: " + s.DiscoverURL + "\n" +
		" * Description: " + s.Description + "\n" +
		" * Contact Info: " + s.ContactInfo + "\n" +
		" * Height: " + strconv.Itoa(int(s.Height)) + "\n"
	return str
}
