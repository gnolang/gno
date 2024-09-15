package gnofundme // Package gnofundme provides a simple crowdfunding system.

import (
	"strconv"
	"strings"
	"time"
)

// ctype: Campaign type, will be an int
// campaign types :
// 0: Personal, for a person (yourself or another person): can be for humanitarian reason, emergency needs, or other
// 1: Cause, for a cause: you will manage the funds (based on trust)
// 2: Organism, for an organism, charity...
// 3: Donation  Donation Crowdfunding
// 4: RewardC,  Reward Crowdfunding
// 5: EquityC, Equity Crowdfunding

type Campaign struct {
	// Campaign type
	Ctype int
	// Campaign title
	Title string
	// Campaign description
	Description string
	// Campaign goal
	Goal int64
	// Current amount of money raised
	Current int64
	// Campaign begin date
	Begin time.Time
	// Campaign deadline
	Deadline time.Time
	// Campaign owner
	Owner *User
	// Campaign beneficiary (can be different from the owner)
	Beneficiary *User
	// Campaign contributors
	contributors []*Contributor
}

func NewCampaign(ctype int, title string, description string, goal int64, begin time.Time, deadline time.Time, owner *User, beneficiary *User) *Campaign {
	return &Campaign{
		Ctype:        ctype,
		Title:        title,
		Description:  description,
		Goal:         goal,
		Current:      0,
		Begin:        begin,
		Deadline:     deadline,
		Owner:        owner,
		Beneficiary:  beneficiary,
		contributors: make([]*Contributor, 0),
	}
}

func (c Campaign) GetContributors() []*Contributor {
	return c.contributors
}

func (c *Campaign) AddContributor(user *User, date time.Time, amount int64, display bool) *Contributor {
	c.Current += amount
	contributor := NewContributor(user, date, amount, display)
	c.contributors = append(c.contributors, contributor)
	return contributor
}

// CampaignsToJSONString converts a slice of Campaigns structs into a JSON string
func CampaignsToJSONString(campaigns []*Campaign) string {
	var sb strings.Builder
	sb.WriteString("[")

	for i, campaign := range campaigns {
		if i > 0 {
			sb.WriteString(",")
		}

		sb.WriteString(CampaignToJSONString(campaign))
	}
	sb.WriteString("]")

	return sb.String()
}

// CampaignToJSONString returns a Campaign formatted as a JSON string
func CampaignToJSONString(campaign *Campaign) string {
	var sb strings.Builder

	sb.WriteString("{")
	sb.WriteString(`"ctype":"` + strconv.Itoa(campaign.Ctype) + `",`)
	sb.WriteString(`"title":"` + campaign.Title + `",`)
	sb.WriteString(`"description":"` + escapeString(campaign.Description) + `",`)
	sb.WriteString(`"goal":"` + strconv.Itoa(int(campaign.Goal)) + `",`)
	sb.WriteString(`"current":"` + strconv.Itoa(int(campaign.Current)) + `",`)
	sb.WriteString(`"begin":"` + strconv.Itoa(int((campaign.Begin).Unix())) + `",`)
	sb.WriteString(`"deadline":"` + strconv.Itoa(int(campaign.Deadline.Unix())) + `",`)
	sb.WriteString(`"owner":"` + campaign.Owner.GetAddress().String() + `",`)
	sb.WriteString(`"beneficiary":"` + campaign.Beneficiary.GetAddress().String() + `",`)
	sb.WriteString(`"contributors":"` + strconv.Itoa(len(campaign.contributors)) + `"`)
	sb.WriteString("}")

	return sb.String()
}

// escapeString escapes quotes in a string for JSON compatibility.
func escapeString(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`)
}
