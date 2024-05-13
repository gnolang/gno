package gnofundme

import (
	"bytes"
	"std"
	"strconv"
	"strings"
	"time"

	"gno.land/p/demo/avl"
	pkg "gno.land/p/demo/gnofundme"
	"gno.land/p/demo/ufmt"
)

var (
	campaigns      *avl.Tree
	users          *avl.Tree
	campaignPrefix = "/r/demo/gnofundme:"
)

func init() {
	campaigns = avl.NewTree()
	users = avl.NewTree()
}

func NewUser(name string, email string) string {
	// get the user address
	address := std.GetOrigCaller()
	// create the user
	user := pkg.NewUser(name, email, address)

	// the avl.tree key will be address to simplify the fetching
	users.Set(address.String(), user)

	return user.GetName() + " added successfully"
}

func NewCampaign(title string, description string, goal uint, beginT int64, deadlineT int64, beneficiary std.Address) (int, string) {
	// get the caller as the owner address
	caller := std.GetOrigCaller()
	user := pkg.NewUser("user1", "email1", caller)

	// the avl.tree key will be address to simplify the fetching
	users.Set(caller.String(), user)

	user = pkg.NewUser("user2", "email2", beneficiary)

	// the avl.tree key will be address to simplify the fetching
	users.Set(beneficiary.String(), user)
	// verify that the user exists
	o, ok := users.Get(caller.String())
	if !ok {
		panic("owner not found")
	}

	// verify that the benefeciary exists
	b, ok := users.Get(beneficiary.String())
	if !ok {
		panic("beneficiary not found")
	}

	// cast the owner and beneficiary to *pkg.User
	owner := o.(*pkg.User)
	beneficiaryU := b.(*pkg.User)

	// create the time objects
	begin := time.Unix(beginT, 0)
	deadline := time.Unix(deadlineT, 0)

	// create the campaign
	campaign := pkg.NewCampaign(title, description, goal, begin, deadline, owner, beneficiaryU)

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

	sentCoins := std.GetOrigSend()
	// amount := sentCoins.AmountOf("ugnot")

	// add the contributor to the campaign
	campaign.(*pkg.Campaign).AddContributor(user.(*pkg.User), amount, display)

	return "contributor added successfully"
}

func GetCampaigns() string {
	if campaigns.Size() == 0 {
		return "[]"
	}
	campaignsArray := make([]*pkg.Campaign, 0, campaigns.Size())
	campaigns.Iterate("", "", func(key string, value interface{}) bool {
		c := value.(*pkg.Campaign)
		campaignsArray = append(campaignsArray, c)
		return false
	})
	return pkg.CampaignsToJSONString(campaignsArray)
}

func Render(path string) string {
	if path == "" {
		return renderHomePage()
	}

	// split the path
	parts := strings.Split(path, "/")
	switch {
	case len(parts) == 2 && parts[0] == "campaign":
		return renderCampaign(parts[1])
	}
	return "unknown page"
}

func renderHomePage() string {
	banker := std.GetBanker(std.BankerTypeRealmIssue)
	banker.IssueCoin(std.CurrentRealm().Addr(), "ugnot", 350000000)
	balance := banker.GetCoins(std.CurrentRealm().Addr())
	// define empty buffer
	var b bytes.Buffer

	// users counter
	userc := 1
	b.WriteString("# Welcome to GnoFundMe\n\n")
	b.WriteString(ufmt.Sprintf("Balance: %s.\n", balance.String()))

	b.WriteString("# Campaigns :\n\n")
	campaigns.Iterate("", "", func(key string, value interface{}) bool {
		c := value.(*pkg.Campaign)
		campaignURL := ufmt.Sprintf("%scampaign/%s", campaignPrefix, key)
		b.WriteString(
			ufmt.Sprintf("## [Campaign %s](%s)\n", key, campaignURL),
		)
		b.WriteString(
			ufmt.Sprintf(
				"### Description: %s\n", c.GetDescription(),
			),
		)

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

func renderCampaign(id string) string {
	// get the campaign if it exists
	campaign, ok := campaigns.Get(id)
	if !ok {
		return "campaign not found"
	}

	// define empty buffer
	var b bytes.Buffer
	c := campaign.(*pkg.Campaign)
	b.WriteString(
		ufmt.Sprintf(
			"# Campaign %s\n", id,
		),
	)
	b.WriteString(
		ufmt.Sprintf(
			"## Description: %s\n", c.GetDescription(),
		),
	)
	b.WriteString(
		ufmt.Sprintf(
			"### Goal: %d ", c.GetGoal(),
		),
	)
	b.WriteString(
		ufmt.Sprintf(
			"********** Current: %d\n", c.GetCurrent(),
		),
	)
	b.WriteString(
		ufmt.Sprintf(
			"### Begin: %s", c.GetBegin().String(),
		),
	)
	b.WriteString(
		ufmt.Sprintf(
			"  ------------->  Deadline: %s\n", c.GetDeadline().String(),
		),
	)
	b.WriteString(
		ufmt.Sprintf(
			"### Owner: %s, %s\n", c.GetOwner().GetName(), c.GetOwner().GetAddress(),
		),
	)
	b.WriteString(
		ufmt.Sprintf(
			"### Beneficiary: %s, %s\n", c.GetBeneficiary().GetName(), c.GetBeneficiary().GetAddress(),
		),
	)
	b.WriteString(
		ufmt.Sprintf(
			"## Contributors: %d\n", len(c.GetContributors()),
		),
	)
	if len(c.GetContributors()) != 0 {
		contributors := c.GetContributors()
		for _, contributor := range contributors {
			if contributor.GetDisplay() {
				b.WriteString(
					ufmt.Sprintf(
						"### %s: %d\n", contributor.GetUser().GetName(), contributor.GetAmount(),
					),
				)
			} else {
				b.WriteString(
					ufmt.Sprintf(
						"### Anonymous: %d\n", contributor.GetAmount(),
					),
				)
			}
		}
	}
	return b.String()
}
