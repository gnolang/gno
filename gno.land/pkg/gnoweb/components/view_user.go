package components

import (
	"net/url"
	"time"
	"unicode/utf8"
)

const UserViewType ViewType = "user-view"

type UserLinkType string

const (
	UserLinkTypeGithub   UserLinkType = "github"
	UserLinkTypeTwitter  UserLinkType = "twitter"
	UserLinkTypeDiscord  UserLinkType = "discord"
	UserLinkTypeTelegram UserLinkType = "telegram"
	UserLinkTypeLink     UserLinkType = "link"
)

type UserContributionType int

const (
	UserContributionTypeRealm = iota
	UserContributionTypePackage
)

func (typ UserContributionType) String() string {
	switch typ {
	case UserContributionTypeRealm:
		return "realm"
	case UserContributionTypePackage:
		return "pure"
	}
	return ""
}

type UserLink struct {
	Type  UserLinkType
	URL   string
	Title string
}

type UserContribution struct {
	Title       string
	URL         string
	Type        UserContributionType
	Description string
	Size        int
	Date        *time.Time
}

// UserRegistration is what the user registry answered about the page's
// address. Only an answer says the address is unregistered: a lookup that
// failed, or a chain with no registry to ask, knows nothing either way.
type UserRegistration int

const (
	// UserRegistrationUnknown is the zero value: the registry gave no answer.
	UserRegistrationUnknown UserRegistration = iota
	// UserRegistrationNone is the registry answering that nobody holds it.
	UserRegistrationNone
	// UserRegistrationRegistered is the registry naming the user behind it.
	UserRegistrationRegistered
)

// UserData contains data for the user view
type UserData struct {
	// Username is the name the page is titled by, empty when the namespace is
	// an address with no name behind it: the page is then titled by Address.
	Username string
	// Namespace is the namespace whose home realm the page links to: the
	// name's, the address's when only it has one, or the full address when no
	// name resolves. Links use this, never Username, which may be elided.
	Namespace string
	// HomeLabel is Namespace as the home button prints it, shortened when it
	// is an address, so it differs from Username whenever Namespace is one.
	HomeLabel string
	// Address is the full bech32 address, empty when none could be resolved.
	Address string
	// CurrentName is the name Username now resolves to, set only when it is
	// another one: Username is then an old name, and Address its new owner's.
	CurrentName string
	// Registration is what the registry answered, which an empty Username
	// cannot tell apart: no user, or no answer.
	Registration  UserRegistration
	Bio           string
	Teams         []struct{}
	Links         []UserLink
	Contributions []UserContribution
	PackageCount  int
	RealmCount    int
	PureCount     int
	Content       Component
}

// longNameRunes is the length past which the title steps down a size, so a
// long name wraps onto at most two lines instead of being cut: impersonating
// names tend to differ from the real one at the end.
const longNameRunes = 16

// Unregistered reports whether the registry answered that nobody holds
// Address, the only case the page may say so.
func (d UserData) Unregistered() bool {
	return d.Registration == UserRegistrationNone
}

// LongName reports whether Username is long enough to be titled smaller.
func (d UserData) LongName() bool {
	return utf8.RuneCountInString(d.Username) > longNameRunes
}

// enrichLinks sets the Title of link-type entries to their hostname.
func enrichUserLinks(links []UserLink) {
	for i := range links {
		if links[i].Type == UserLinkTypeLink {
			if u, err := url.Parse(links[i].URL); err == nil {
				links[i].Title = u.Host
			} else {
				links[i].Title = links[i].URL
			}
		}
	}
}

// UserView creates a new user view component
func UserView(data UserData) *View {
	enrichUserLinks(data.Links)

	return NewTemplateView(
		UserViewType,
		"renderUser",
		data,
	)
}
