package gnofundme

import (
	"bytes"
	"std"
	"strconv"
	"strings"
	"time"

	"gno.land/p/demo/avl"
	GFMpkg "gno.land/p/demo/gnofundme" // GnoFundMe package
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
	user := GFMpkg.NewUser(name, email, address)

	// the avl.tree key will be the address to simplify the fetching
	users.Set(address.String(), user)
	std.Emit("User Creation", "user_created", user.GetAddress().String())
	return user.GetName() + " added successfully"
}

func NewCampaign(ctype int, title string, description string, goal int64, beginT int64, deadlineT int64, beneficiary std.Address) string {
	if ctype < 0 || ctype > 5 {
		panic("invalid campaign type")
	}

	var owner, beneficiaryU *GFMpkg.User
	// gets the caller address (owner of the campaign)
	caller := std.GetOrigCaller()

	// verify that the user owner exists
	o, ok := users.Get(caller.String())
	if !ok {
		// panic("owner not found")
		// XXX TO BE MODIFIED: temporarily creates the user if he doesn't exist. In the final version,
		// the user will have to be registered using the user struct of this project, or maybe I'll use
		// the profile pkg that is still a "work in progress"
		user := GFMpkg.NewUser("user1", "email1", caller)
		// the avl.tree key will be address to simplify the fetching
		users.Set(caller.String(), user)
		owner = user
	} else {
		// cast the owner and beneficiary to *GFMpkg.User
		owner = o.(*GFMpkg.User)
	}

	// verify that the benefeciary exists if it's different from caller
	if beneficiary != caller {
		b, ok := users.Get(beneficiary.String())
		if !ok {
			// panic("beneficiary not found")
			user := GFMpkg.NewUser("user2", "email2", beneficiary)
			users.Set(beneficiary.String(), user)
			beneficiaryU = user
		} else {
			beneficiaryU = b.(*GFMpkg.User)
		}
	} else {
		beneficiaryU = owner
	}

	// Verfiy that the dates are correct
	if beginT < time.Now().Unix() {
		panic("begin date must be in the future")
	}
	if deadlineT < time.Now().Unix() {
		panic("deadline date must be in the future")
	}
	if deadlineT < beginT {
		panic("deadline date must be after begin date")
	}

	// creates the time objects
	begin := time.Unix(beginT, 0)
	deadline := time.Unix(deadlineT, 0)
	// create the campaign
	campaign := GFMpkg.NewCampaign(ctype, title, description, goal, begin, deadline, owner, beneficiaryU)

	// the key will be the size of the tree
	key := campaigns.Size()
	campaigns.Set(strconv.Itoa(key), campaign)
	std.Emit("Campaign Creation", "campaign_created", GFMpkg.CampaignToJSONString(campaign))
	return "Campaign N°" + strconv.Itoa(key) + " created successfully"
}

func AddContributor(campaignId string, display bool) string {
	// get the campaign if it exists
	camp, ok := campaigns.Get(campaignId)
	if !ok {
		panic("campaign not found")
	}

	campaign := camp.(*GFMpkg.Campaign)
	// Verfiy that the dates are correct
	if time.Now().Unix() < campaign.Begin.Unix() {
		panic(ufmt.Sprintf("%d, %d,the campaign didn't start yet", time.Now().Unix(), campaign.Begin.Unix())) // added the execution date and the beginning date in unix format to the error message to do verfications, sometimes the execution time is the same even when testing several minutes after!!!
	}
	if time.Now().Unix() > campaign.Deadline.Unix() {
		panic("the campaign is over")
	}

	// get the user address
	userAddress := std.GetOrigCaller()

	// get the user if he exists
	user, ok := users.Get(userAddress.String())
	// XXX Maybe this condition will be deleted in the final version, as anyone can contribute to a campaign without hvaving to be registered
	if !ok {
		panic("user is not registered")
	}

	sentCoins := std.GetOrigSend()
	amount := sentCoins.AmountOf("ugnot")

	// add the contributor to the campaign
	contributor := campaign.AddContributor(user.(*GFMpkg.User), time.Unix(time.Now().Unix(), 0), amount, display)
	std.Emit("Add Contributor", "contributor_added", contributor.String())
	return "contributor added successfully"
}

func GetCampaigns() string {
	if campaigns.Size() == 0 {
		return "[]"
	}
	campaignsArray := make([]*GFMpkg.Campaign, 0, campaigns.Size())
	campaigns.Iterate("", "", func(key string, value interface{}) bool {
		c := value.(*GFMpkg.Campaign)
		campaignsArray = append(campaignsArray, c)
		return false
	})
	return GFMpkg.CampaignsToJSONString(campaignsArray)
}

func IsUserRegistered(userAddress string) bool {
	// get the user address
	// userAddress := std.GetOrigCaller()
	// verify if the user exists
	_, ok := users.Get(userAddress)
	return ok
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
	banker := std.GetBanker(std.BankerTypeReadonly)
	// banker.IssueCoin(std.CurrentRealm().Addr(), "ugnot", 350000000)
	balance := banker.GetCoins(std.CurrentRealm().Addr())
	// define empty buffer
	var b bytes.Buffer

	// users counter
	userc := 1
	b.WriteString("# Welcome to GnoFundMe\n\n")
	b.WriteString(ufmt.Sprintf("Balance: %s.\n", balance.String()))

	b.WriteString("# Campaigns :\n\n")
	campaigns.Iterate("", "", func(key string, value interface{}) bool {
		c := value.(*GFMpkg.Campaign)
		campaignURL := ufmt.Sprintf("%scampaign/%s", campaignPrefix, key)
		b.WriteString(
			ufmt.Sprintf("## [Campaign %s](%s)\n", key, campaignURL),
		)
		b.WriteString(
			ufmt.Sprintf(
				"### Description: %s\n", c.Description,
			),
		)

		return false
	})
	b.WriteString("# Users:\n")

	users.Iterate("", "", func(key string, value interface{}) bool {
		u := value.(*GFMpkg.User)
		b.WriteString(
			ufmt.Sprintf(
				"## User #%d: %s, Key: %s\n", userc, u.GetName(), key,
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

	// define an empty buffer
	var b bytes.Buffer
	c := campaign.(*GFMpkg.Campaign)
	b.WriteString(
		ufmt.Sprintf("# Campaign %s\n", id),
	)
	b.WriteString(
		ufmt.Sprintf("## Title: %s\n", c.Title),
	)
	b.WriteString(
		ufmt.Sprintf("## Description: %s\n", c.Description),
	)
	b.WriteString(
		ufmt.Sprintf("### Goal: %d ", c.Goal),
	)
	b.WriteString(
		ufmt.Sprintf("********** Current: %d\n", c.Current),
	)
	b.WriteString(
		ufmt.Sprintf("### Begin: %s", c.Begin.String()),
	)
	b.WriteString(
		ufmt.Sprintf("  ------------->  Deadline: %s\n", c.Deadline.String()),
	)
	b.WriteString(
		ufmt.Sprintf("### Owner: %s, %s\n", c.Owner.GetName(), c.Owner.GetAddress()),
	)
	b.WriteString(
		ufmt.Sprintf("### Beneficiary: %s, %s\n", c.Beneficiary.GetName(), c.Beneficiary.GetAddress()),
	)
	b.WriteString(
		ufmt.Sprintf("## Contributors: %d\n", len(c.GetContributors())),
	)
	if len(c.GetContributors()) != 0 {
		contributors := c.GetContributors()
		for _, contributor := range contributors {
			if contributor.GetDisplay() {
				b.WriteString(
					ufmt.Sprintf(
						"### %s: %dugnot, %s\n", contributor.GetUser().GetName(), contributor.Amount, contributor.Date.String(),
					), // added the date of contribution in the display, but adds "m=+1726419880.000000001" after the date =>> the problems comes from time.now(), had to modify and use time.Unix(time.Now().Unix(), 0) to solve the problem.
				)
			} else {
				b.WriteString(
					ufmt.Sprintf(
						"### Anonymous: %dugnot\n", contributor.Amount,
					),
				)
			}
		}
	}
	return b.String()
}
