package gnofundme

import (
	"bytes"
	"std"
	"strconv"
	"strings"
	"time"

	"gno.land/p/demo/avl"
	gfm "gno.land/p/demo/gnofundme" // GnoFundMe package
	"gno.land/p/demo/ufmt"
)

var (
	campaigns      *avl.Tree
	users          *avl.Tree
	banker         std.Banker
	balance        int64
	campaignPrefix = "/r/demo/gnofundme:"
	campaignsTotal = make(map[string]int64)
)

func init() {
	campaigns = avl.NewTree()
	users = avl.NewTree()
	banker = std.GetBanker(std.BankerTypeRealmIssue)
	// banker.IssueCoin(std.CurrentRealm().Addr(), "ugnot", 350000000)
}

func NewUser(name string, email string) string {
	// get the user address
	address := std.PrevRealm().Addr()
	// create the user
	user := gfm.NewUser(name, email, address)

	// the avl.tree key will be the address to simplify the fetching
	users.Set(address.String(), user)
	std.Emit("User Creation", "user_created", user.GetAddress().String())
	return user.GetName() + " added successfully"
}

func NewCampaign(ctype int, title string, description string, goal int64, beginT int64, deadlineT int64, beneficiary std.Address) string {
	if ctype < 0 || ctype > 5 {
		panic("invalid campaign type")
	}

	var owner, beneficiaryU *gfm.User
	// gets the caller address (owner of the campaign)
	caller := std.PrevRealm().Addr()

	// verify that the user owner exists
	o, ok := users.Get(caller.String())
	if !ok {
		// panic("owner not found")
		// XXX TO BE MODIFIED: temporarily creates the user if he doesn't exist. In the final version,
		// the user will have to be registered using the user struct of this project, or maybe I'll use
		// the profile pkg that is still a "work in progress"
		user := gfm.NewUser("user1", "email1", caller)
		// the avl.tree key will be address to simplify the fetching
		users.Set(caller.String(), user)
		owner = user
	} else {
		// cast the owner and beneficiary to *gfm.User
		owner = o.(*gfm.User)
	}

	// verify that the benefeciary exists if it's different from caller
	if beneficiary != caller {
		b, ok := users.Get(beneficiary.String())
		if !ok {
			// panic("beneficiary not found")
			user := gfm.NewUser("user2", "email2", beneficiary)
			users.Set(beneficiary.String(), user)
			beneficiaryU = user
		} else {
			beneficiaryU = b.(*gfm.User)
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
	campaign := gfm.NewCampaign(ctype, title, description, goal, begin, deadline, owner, beneficiaryU)

	// the key will be the size of the tree
	key := campaigns.Size()
	campaigns.Set(strconv.Itoa(key), campaign)
	std.Emit("Campaign Creation", "campaign_created", gfm.CampaignToJSONString(campaign))
	return "Campaign N°" + strconv.Itoa(key) + " created successfully"
}

func AddContributor(campaignId string, display bool) string {
	// get the campaign if it exists
	camp, ok := campaigns.Get(campaignId)
	if !ok {
		panic("campaign not found")
	}

	campaign := camp.(*gfm.Campaign)
	// Verfiy that the dates are correct
	if time.Now().Unix() < campaign.Begin.Unix() {
		panic(ufmt.Sprintf("Now: %d, Begin: %d,the campaign didn't start yet", time.Now().Unix(), campaign.Begin.Unix())) // added the execution date and the beginning date in unix format to the error message to do verfications, sometimes the execution time is the same even when testing several minutes after!!!
	}
	if time.Now().Unix() > campaign.Deadline.Unix() {
		panic("the campaign is over")
	}

	// get the user address
	userAddress := std.PrevRealm().Addr()

	// get the user if he exists
	user, ok := users.Get(userAddress.String())
	// XXX Maybe this condition will be deleted in the final version, as anyone can contribute to a campaign without hvaving to be registered
	if !ok {
		panic("user is not registered")
	}

	sentCoins := std.GetOrigSend()
	amount := sentCoins.AmountOf("ugnot")
	campaignsTotal[campaignId] += amount

	// add the contributor to the campaign
	contributor := campaign.AddContributor(user.(*gfm.User), time.Unix(time.Now().Unix(), 0), amount, display)
	std.Emit("Add Contributor", "contributor_added", contributor.String())
	return "contributor added successfully"
}

func GetCampaigns() string {
	if campaigns.Size() == 0 {
		return "[]"
	}
	campaignsArray := make([]*gfm.Campaign, 0, campaigns.Size())
	campaigns.Iterate("", "", func(key string, value interface{}) bool {
		c := value.(*gfm.Campaign)
		campaignsArray = append(campaignsArray, c)
		return false
	})
	return gfm.CampaignsToJSONString(campaignsArray)
}

func IsUserRegistered(userAddress string) bool {
	// verify if the user exists
	_, ok := users.Get(userAddress)
	return ok
}

func Withdraw(campaignId string) {
	// get the campaign if it exists
	camp, ok := campaigns.Get(campaignId)
	if !ok {
		panic("campaign not found")
	}

	campaign := camp.(*gfm.Campaign)
	// get the user address
	callerAddress := std.PrevRealm().Addr()

	// verify that the user is the owner of the campaign
	if callerAddress != campaign.Owner.GetAddress() {
		panic("you are not the owner of the campaign")
	}

	// verify that the campaign is over
	if time.Now().Unix() < campaign.Deadline.Unix() {
		panic("the campaign is not over yet")
	}

	// XXX TO BE DISCUSSED: the funds can be withdrawn even if the goal is not reached or not?
	// verify that the campaign is successful
	// if campaign.Current < campaign.Goal {
	// 	panic("the campaign is not successful")
	// }

	// get the beneficiary address
	beneficiaryAddress := campaign.Beneficiary.GetAddress()
	// get the amount to withdraw
	amount := campaignsTotal[campaignId]
	campaignsTotal[campaignId] = 0
	// transfer the amount to the beneficiary
	banker.SendCoins(std.CurrentRealm().Addr(), beneficiaryAddress, std.Coins{std.NewCoin("ugnot", amount)})
	std.Emit("Withdraw", "withdraw", "withdrawn successfully")
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
	balance = banker.GetCoins(std.CurrentRealm().Addr()).AmountOf("ugnot")
	// define empty buffer
	var b bytes.Buffer

	// users counter
	userc := 1
	b.WriteString("# Welcome to GnoFundMe\n\n")
	b.WriteString(ufmt.Sprintf("Balance: %d ugnot.\n", balance))

	b.WriteString("# Campaigns :\n\n")
	campaigns.Iterate("", "", func(key string, value interface{}) bool {
		c := value.(*gfm.Campaign)
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
		u := value.(*gfm.User)
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
	c := campaign.(*gfm.Campaign)

	return c.Render(id)
}
