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
	campaigns      *avl.Tree
	users          *avl.Tree
	campaignPrefix = "/r/demo/campaign:"
)

func init() {
	campaigns = avl.NewTree()
	users = avl.NewTree()
}

func NewUser(name string, email string, address std.Address) string {
	// create the user
	user := pkg.NewUser(name, email, address)

	// the avl.tree key will be address to simplify the fetching
	users.Set(address.String(), user)

	return user.GetName() + " added successfully"
}

func NewCampaign(title string, description string, goal uint, beginT int64, deadlineT int64, owner string, beneficiary string) (int, string) {
	// verify that the user exists
	o, ok := users.Get(owner)
	if !ok {
		panic("owner not found")
	}

	// verify that the benefeciary exists
	b, ok := users.Get(beneficiary)
	if !ok {
		panic("beneficiary not found")
	}

	// cast the owner and beneficiary to *pkg.User
	ownerU := o.(*pkg.User)
	beneficiaryU := b.(*pkg.User)

	// create the time objects
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
	// get the campaign if it exists
	campaign, ok := campaigns.Get(campaignId)
	if !ok {
		return "campaign not found"
	}

	// get the user address
	userAddress := std.GetOrigCaller()

	// get the user if he exists
	user, ok := users.Get(userAddress.String())
	if !ok {
		panic("user is not registered")
	}

	// add the contributor to the campaign
	campaign.(*pkg.Campaign).AddContributor(user.(*pkg.User), amount, display)

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
	// users counter
	userc := 1

	b.WriteString("# Campaigns list:\n\n")
	campaigns.Iterate("", "", func(key string, value interface{}) bool {
		c := value.(*pkg.Campaign)
		campaignURL := ufmt.Sprintf("%s%s:%s", campaignPrefix, key, c.GetTitle())
		b.WriteString(
			ufmt.Sprintf("## [Campaign #%s](%s)\n", key, campaignURL),
		)
		b.WriteString(
			ufmt.Sprintf(
				"### Description: %s\n", c.GetDescription(),
			),
		)
		// b.WriteString(
		// 	ufmt.Sprintf(
		// 		"#### Goal: %d ", c.GetGoal(),
		// 	),
		// )
		// b.WriteString(
		// 	ufmt.Sprintf(
		// 		"********** Current: %d\n", c.GetCurrent(),
		// 	),
		// )
		// b.WriteString(
		// 	ufmt.Sprintf(
		// 		"#### Begin: %s", c.GetBegin().String(),
		// 	),
		// )
		// b.WriteString(
		// 	ufmt.Sprintf(
		// 		"  ------------->  Deadline: %s\n", c.GetDeadline().String(),
		// 	),
		// )
		// b.WriteString(
		// 	ufmt.Sprintf(
		// 		"#### Owner: %s\n", c.GetOwner().GetName(),
		// 	),
		// )
		// b.WriteString(
		// 	ufmt.Sprintf(
		// 		"#### Beneficiary: %s\n", c.GetBeneficiary().GetName(),
		// 	),
		// )
		// b.WriteString(
		// 	ufmt.Sprintf(
		// 		"#### Contributors: %d\n", len(c.GetContributors()),
		// 	),
		// )
		return false
	})
	b.WriteString("# Users:\n")

	users.Iterate("", "", func(key string, value interface{}) bool {
		u := value.(*pkg.User)
		b.WriteString(
			ufmt.Sprintf(
				"## User #%d: %s\n", userc, u.GetName(),
			),
		)
		b.WriteString(
			ufmt.Sprintf(
				"- email: %s\n - address: %s\n", u.GetEmail(), u.GetAddress(),
			),
		)
		userc++
		return false
	})
	return b.String()
}
