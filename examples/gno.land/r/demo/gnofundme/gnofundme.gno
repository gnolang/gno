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
	pusers "gno.land/p/demo/users"
	"gno.land/r/demo/users"
)

var (
	campaigns      *avl.Tree
	banker         std.Banker
	balance        int64
	campaignPrefix = "/r/demo/gnofundme:"
	campaignsTotal = make(map[string]int64)
)

func init() {
	campaigns = avl.NewTree()
	banker = std.GetBanker(std.BankerTypeRealmIssue)
}

func NewCampaign(ctype int, title string, description string, goal int64, beginT int64, deadlineT int64, beneficiary std.Address) string {
	if ctype < 0 || ctype > 5 {
		panic("invalid campaign type")
	}

	var beneficiaryU *pusers.User
	// gets the caller address (owner of the campaign)
	caller := std.PrevRealm().Addr()

	// verify that the user owner exists
	owner := users.GetUserByAddress(caller)
	if owner == nil {
		panic("owner not registered in 'gno.land/r/demo/users'")
	}

	// verify that the benefeciary exists if it's different from caller
	if beneficiary != caller {
		beneficiaryU := users.GetUserByAddress(beneficiary)
		if beneficiaryU == nil {
			panic("beneficiary not registered in 'gno.land/r/demo/users'")
		}
	} else {
		beneficiaryU = owner
	}

	// Verfiy that the dates are correct
	if beginT < time.Now().Unix() {
		panic("begin date must be in the future")
	}
	if deadlineT <= beginT {
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
		panic(ufmt.Sprintf("The campaign didn't start yet. Begin: %d,", campaign.Begin.Unix()))
	}
	if time.Now().Unix() > campaign.Deadline.Unix() {
		panic("the campaign is over")
	}

	// get the user address
	userAddress := std.PrevRealm().Addr()

	// get the user if he exists
	user := users.GetUserByAddress(userAddress)
	// XXX Maybe this condition will be deleted in the final version, as anyone can contribute to a campaign without hvaving to be registered
	if user == nil {
		panic("user is not registered in 'gno.land/r/demo/users")
	}

	sentCoins := std.GetOrigSend()
	amount := sentCoins.AmountOf("ugnot")
	campaignsTotal[campaignId] += amount

	// add the contributor to the campaign
	contributor := campaign.AddContributor(user, time.Unix(time.Now().Unix(), 0), amount, display)
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

func IsUserRegistered(userAddress std.Address) bool {
	// verify if the user exists
	ok := users.GetUserByAddress(userAddress)
	return ok != nil
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
	if callerAddress != campaign.Owner.Address {
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
	beneficiaryAddress := campaign.Beneficiary.Address
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
	b.WriteString("\n")
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
