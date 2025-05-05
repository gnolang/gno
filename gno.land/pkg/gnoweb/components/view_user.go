package components

import (
	"fmt"
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
	UserLinkTypeLinkedin UserLinkType = "linkedin"
	UserLinkTypeLink     UserLinkType = "link"
)

type UserContributionType struct {
	Id   int
	Name string
}

var (
	UserContributionTypeRealm = UserContributionType{
		Id:   1,
		Name: "realm",
	}
	UserContributionTypePackage = UserContributionType{
		Id:   2,
		Name: "package",
	}
)

type UserLink struct {
	Type  UserLinkType
	URL   string
	Title string
}

type UserContribution struct {
	Title       string
	Description string
	URL         string
	Size        int
	Stars       int // TODO: would be great to have this
	Type        UserContributionType
	Date        time.Time
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
}

// FormatRelativeTime formats a time into a relative string (e.g. "1 month ago")
func FormatRelativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	units := []struct {
		unit  time.Duration
		label string
	}{
		{time.Minute, "minute"},
		{time.Hour, "hour"},
		{24 * time.Hour, "day"},
		{7 * 24 * time.Hour, "week"},
		{30 * 24 * time.Hour, "month"},
		{365 * 24 * time.Hour, "year"},
	}

	for i := len(units) - 1; i >= 0; i-- {
		u := units[i]
		if diff >= u.unit {
			value := int(diff / u.unit)
			if value == 1 {
				return fmt.Sprintf("1 %s ago", u.label)
			}
			return fmt.Sprintf("%d %ss ago", value, u.label)
		}
	}

	return "just now"
}

// UserView creates a new user view component
func UserView(data UserData) *View {
	// Set the title of the link to the host of the URL if it's a link
	for i := range data.Links {
		if data.Links[i].Type == UserLinkTypeLink {
			if u, err := url.Parse(data.Links[i].URL); err == nil {
				data.Links[i].Title = u.Host
			} else {
				data.Links[i].Title = data.Links[i].URL
			}
		}
	}

	// Count realms, packages is the rest
	data.RealmCount = 0
	data.PackageCount = len(data.Contributions)
	for _, contribution := range data.Contributions {
		if contribution.Type.Id == UserContributionTypeRealm.Id {
			data.RealmCount++
		}
	}
	data.PureCount = data.PackageCount - data.RealmCount

	return NewTemplateView(
		UserViewType,
		"renderUser",
		data,
	)
}
