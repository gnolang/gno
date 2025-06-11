package components

import (
	"net/url"
	"time"
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

// UserData contains data for the user view
type UserData struct {
	Username      string
	Handlename    string
	Bio           string
	Teams         []struct{}
	Links         []UserLink
	Contributions []UserContribution
	PackageCount  int
	RealmCount    int
	PureCount     int
	Content       Component
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
