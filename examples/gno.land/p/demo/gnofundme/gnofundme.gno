package gnofundme

import (
	"std"
	"strconv"
	"strings"
	"time"

	"gno.land/p/demo/avl"
)

type Campaign struct {
	title        string
	description  string
	goal         uint
	current      uint
	begin        time.Time
	deadline     time.Time
	owner        *User
	beneficiary  *User
	contributors []*Contributor
}

func NewCampaign(title string, description string, goal uint, begin time.Time, deadline time.Time, owner *User, beneficiary *User) *Campaign {
	return &Campaign{
		title:        title,
		description:  description,
		goal:         goal,
		current:      0,
		begin:        begin,
		deadline:     deadline,
		owner:        owner,
		beneficiary:  beneficiary,
		contributors: make([]*Contributor, 0),
	}
}

func (c Campaign) GetTitle() string {
	return c.title
}

func (c Campaign) GetDescription() string {
	return c.description
}

func (c Campaign) GetGoal() uint {
	return c.goal
}

func (c Campaign) GetCurrent() uint {
	return c.current
}

func (c Campaign) GetBegin() time.Time {
	return c.begin
}

func (c Campaign) GetDeadline() time.Time {
	return c.deadline
}

func (c Campaign) GetOwner() *User {
	return c.owner
}

func (c Campaign) GetBeneficiary() *User {
	return c.beneficiary
}

func (c *Campaign) AddContributor(user *User, amount uint, display bool) {
	c.current += amount
	contributor := &Contributor{
		user:    user,
		date:    time.Now(),
		amount:  amount,
		display: display,
	}
	c.contributors = append(c.contributors, contributor)
}

func (c Campaign) GetContributors() []*Contributor {
	return c.contributors
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
	sb.WriteString(`"title":"` + campaign.title + `",`)
	sb.WriteString(`"description":"` + escapeString(campaign.description) + `",`)
	sb.WriteString(`"goal":"` + strconv.Itoa(int(campaign.goal)) + `",`)
	sb.WriteString(`"current":"` + strconv.Itoa(int(campaign.current)) + `",`)
	sb.WriteString(`"begin":"` + strconv.Itoa(int((campaign.begin).Unix())) + `",`)
	sb.WriteString(`"deadline":"` + strconv.Itoa(int(campaign.deadline.Unix())) + `",`)
	sb.WriteString(`"owner":"` + campaign.owner.GetAddress().String() + `",`)
	sb.WriteString(`"beneficiary":"` + campaign.beneficiary.GetAddress().String() + `",`)
	sb.WriteString(`"contributors":"` + strconv.Itoa(len(campaign.contributors)) + `"`)
	sb.WriteString("}")

	return sb.String()
}

// escapeString escapes quotes in a string for JSON compatibility.
func escapeString(s string) string {
	return strings.ReplaceAll(s, `"`, `\"`)
}
