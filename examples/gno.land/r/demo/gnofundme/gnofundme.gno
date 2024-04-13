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

func NewUser(name string, email string, address std.Address) (int, string) {
	// create the user
	user := pkg.NewUser(name, email, address)
	// the key will be the size of the tree
	key := users.Size()
	users.Set(strconv.Itoa(key), user)
	return key, "user added successfully"
}

func NewCampaign(title string, description string, goal uint, beginT int64, deadlineT int64, owner string, beneficiary string) (int, string) {
	o, ok := users.Get(owner)
	if !ok {
		panic("owner not found")
	}
	ownerU := o.(*pkg.User)
	b, ok := users.Get(beneficiary)
	if !ok {
		panic("owner not found")
	}
	beneficiaryU := b.(*pkg.User)
	begin := time.Unix(beginT, 0)
	deadline := time.Unix(deadlineT, 0)
	// create the campaign
	campaign := pkg.NewCampaign(title, description, goal, begin, deadline, ownerU, beneficiaryU)
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
		}
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
		b.WriteString(
			ufmt.Sprintf(
				"#### Goal: %d ", c.GetGoal(),
			),
		)
		b.WriteString(
			ufmt.Sprintf(
				"********** Current: %d\n", c.GetCurrent(),
			),
		)
		b.WriteString(
			ufmt.Sprintf(
				"#### Begin: %s", c.GetBegin().String(),
			),
		)
		b.WriteString(
			ufmt.Sprintf(
				"  ------------->  Deadline: %s\n", c.GetDeadline().String(),
			),
		)
		b.WriteString(
			ufmt.Sprintf(
				"#### Owner: %s\n", c.GetOwner().GetName(),
			),
		)
		b.WriteString(
			ufmt.Sprintf(
				"#### Beneficiary: %s\n", c.GetBeneficiary().GetName(),
			),
		)
		b.WriteString(
			ufmt.Sprintf(
				"#### Contributors: %d\n", len(c.GetContributors()),
			),
		)
		return false
	})
	b.WriteString("# Users:\n")

	users.Iterate("", "", func(key string, value interface{}) bool {
		u := value.(*pkg.User)
		b.WriteString(
			ufmt.Sprintf(
				"## User #%s: %s\n", key, u.GetName(),
			),
		)
		b.WriteString(
			ufmt.Sprintf(
				"- email: %s\n - address: %s\n", u.GetEmail(), u.GetAddress(),
			),
		)
		return false
	})
	return b.String()
}
