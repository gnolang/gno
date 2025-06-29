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
	UserContributionTypePure
)

func (typ UserContributionType) String() string {
	switch typ {
	case UserContributionTypeRealm:
		return "realm"
	case UserContributionTypePure:
		return "pure"
	}
	return ""
}

type UserCardList struct {
	Title             string
	Items             []UserContribution
	Categories        []UserCardListCategory
	TotalCount        int
	SearchPlaceholder string
}

type UserCardListCategory struct {
	ID    string
	Name  string
	Icon  string
	Count int
	Items []UserContribution
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
	Username       string
	Handlename     string
	Bio            string
	Teams          []struct{}
	Links          []UserLink
	Contributions  []UserContribution
	PackageCount   int
	RealmCount     int
	PureCount      int
	Content        Component
	CardsListTitle string
	CardsList      *UserCardList
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

// enrichUserCardList enriches the user card list with the data from the user data
// NOTE: Categories ID must be same as the Type of the UserContribution
func enrichUserCardList(data *UserData) {
	data.CardsListTitle = "Contributions"
	data.CardsList = &UserCardList{
		Title: "Packages",
		Items: data.Contributions,
		Categories: []UserCardListCategory{
			{ID: "realm", Name: "Realm", Icon: "ico-realm", Count: data.RealmCount},
			{ID: "pure", Name: "Pure", Icon: "ico-pure", Count: data.PureCount},
		},
		TotalCount:        data.PackageCount,
		SearchPlaceholder: "Search contributions",
	}
}

// UserView creates a new user view component
func UserView(data UserData) *View {
	enrichUserLinks(data.Links)
	enrichUserCardList(&data)

	return NewTemplateView(
		UserViewType,
		"renderUser",
		data,
	)
}
