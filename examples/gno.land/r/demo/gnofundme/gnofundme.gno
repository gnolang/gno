package gnofundme

import (
	"bytes"
	"std"
	"strconv"
	"time"

	"gno.land/p/demo/avl"
	pkg "gno.land/p/demo/gnofundme"
	"gno.land/p/demo/ufmt"
)

var (
	campaigns *avl.Tree
	users     *avl.Tree
)

func init() {
	campaigns = avl.NewTree()
	users = avl.NewTree()
}

func NewUser(name string, email string) (int, string) {
	// create the user
	user := pkg.NewUser(name, email, std.GetOrigCaller())
	// the key will be the size of the tree
	key := users.Size()
	users.Set(strconv.Itoa(key), user)
	return key, "user added successfully"
}

func NewCampaign(title string, description string, goal uint, begin time.Time, deadline time.Time, owner *pkg.User, beneficiary *pkg.User) (int, string) {
	// create the campaign
	campaign := pkg.NewCampaign(title, description, goal, begin, deadline, owner, beneficiary)
	// the key will be the size of the tree
	key := campaigns.Size()
	campaigns.Set(strconv.Itoa(key), campaign)
	return key, "campaign created successfully"
}

func AddContributor(campaignId string, amount uint, display bool) string {
	var user *pkg.User
	campaign, ok := campaigns.Get(campaignId)
	if !ok {
		return "campaign not found"
	}
	userAddress := std.GetOrigCaller()
	users.Iterate("", "", func(key string, value interface{}) bool {
		u := value.(*pkg.User)
		if u.GetAddress() == userAddress {
			user = u
			break
		}
		return false
	})
	// add the contributor to the campaign
	campaign.(*pkg.Campaign).AddContributor(user, amount, display)
	return "contributor added successfully"
}

func Render(path string) string {
	if path == "" {
		return renderHomePage()
	}
	return "unknown page"
}

func renderHomePage() string {
	// define empty buffer
	var b bytes.Buffer

	b.WriteString("# Campaigns list:\n\n")
	campaigns.Iterate("", "", func(key string, value interface{}) bool {
		c := value.(*pkg.Campaign)
		b.WriteString(
			ufmt.Sprintf(
				"## Campaign #%s: %s\n", key, c.GetTitle(),
			),
		)
		b.WriteString(
			ufmt.Sprintf(
				"### Description: %s\n", c.GetDescription(),
			),
		)
		return false
	})
	return b.String()
}
