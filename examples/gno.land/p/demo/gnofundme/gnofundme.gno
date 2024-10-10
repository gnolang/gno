package gnofundme // Package gnofundme provides a simple crowdfunding system.

import (
	"bytes"
	"strconv"
	"strings"
	"time"

	"gno.land/p/demo/json"
	"gno.land/p/demo/ufmt"
	pusers "gno.land/p/demo/users"
)

type ctype int

const (
	Personal ctype = iota
	// for a person (yourself or another person): can be for humanitarian reason, emergency needs, or other (beneficiary = owner)
	Cause
	// for a cause: you will manage the funds (based on trust) (beneficiary = owner)
	Organism
	// for an organism, charity... (beneficiary != owner)
	// the next types are for companies(startup projects...) crowdfunding
	Donation
	Reward
	Equity
)

func (c ctype) String() string {
	return [...]string{"Personal", "Cause", "Organism", "Donation", "Reward", "Equity"}[c]
}

type Campaign struct {
	// Campaign type
	Ctype       ctype
	Title       string
	Description string
	Goal        uint64
	Current     uint64
	Begin       time.Time
	Deadline    time.Time
	Owner       *pusers.User
	// Campaign beneficiary (can be the owner or another address)
	Beneficiary  *pusers.User
	contributors []*Contributor
}

func NewCampaign(ct int, title string, description string, goal uint64, begin time.Time, deadline time.Time, owner *pusers.User, beneficiary *pusers.User) *Campaign {
	campaign := &Campaign{
		Ctype:        ctype(ct),
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

	// owner.campaigns = append(owner.campaigns, campaign)
	return campaign
}

func (c Campaign) GetContributors() []*Contributor {
	return c.contributors
}

func (c *Campaign) AddContributor(user *pusers.User, date time.Time, amount uint64, display bool) *Contributor {
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

		sb.WriteString(campaign.CampaignToJSONString())
	}
	sb.WriteString("]")

	return sb.String()
}

// CampaignToJSONString returns a Campaign formatted as a JSON string
func (c *Campaign) CampaignToJSONString() string {
	node := json.ObjectNode("", map[string]*json.Node{
		"Ctype":        json.StringNode("Ctype", c.Ctype.String()),
		"Title":        json.StringNode("Title", c.Title),
		"Description":  json.StringNode("Description", c.Description),
		"Goal":         json.NumberNode("Goal", float64(c.Goal)),
		"Current":      json.NumberNode("Current", float64(c.Current)),
		"Begin":        json.StringNode("Begin", strconv.Itoa(int(c.Begin.Unix()))),
		"Deadline":     json.StringNode("Deadline", strconv.Itoa(int(c.Deadline.Unix()))),
		"Owner":        json.StringNode("Owner", c.Owner.Address.String()),
		"Beneficiary":  json.StringNode("Beneficiary", c.Beneficiary.Address.String()),
		"Contributors": json.NumberNode("Contributors", float64(len(c.contributors))),
	})
	value, err := json.Marshal(node)
	if err != nil {
		panic(ufmt.Sprintf("unexpected error: %s", err))
	}

	return string(value)
}

func (c Campaign) Render(id string) string {
	var b bytes.Buffer
	b.WriteString(ufmt.Sprintf("# Campaign %s\n", id))
	b.WriteString(ufmt.Sprintf("## Title: %s\n", c.Title))
	b.WriteString(ufmt.Sprintf("## Description: %s\n", c.Description))
	b.WriteString(ufmt.Sprintf("### Goal: %d GNOT", c.Goal))
	b.WriteString(ufmt.Sprintf("********** Current: %d GNOT\n", c.Current/1_000_000))
	b.WriteString(ufmt.Sprintf("### Begin: %s", c.Begin.String()))
	b.WriteString(ufmt.Sprintf("  ------------->  Deadline: %s\n", c.Deadline.String()))
	b.WriteString(ufmt.Sprintf("### Owner: %s, %s\n", c.Owner.Name, c.Owner.Address.String()))
	b.WriteString(ufmt.Sprintf("### Beneficiary: %s, %s\n", c.Beneficiary.Name, c.Beneficiary.Address.String()))
	b.WriteString(ufmt.Sprintf("## Last contributors: %d\n", len(c.GetContributors())))
	if contributors := c.GetContributors(); len(c.GetContributors()) != 0 {
		length := len(c.GetContributors())
		if length > 10 {
			length = 10
		}

		for i := length - 1; i >= 0; i-- {
			if contributors[i].GetDisplay() {
				b.WriteString(
					ufmt.Sprintf(
						"### %s: %dugnot, %s\n", contributors[i].user.Name, contributors[i].Amount, contributors[i].Date.String(),
					), // added the date of contribution in the display, but adds "m=+1726419880.000000001" after the date =>> the problems comes from time.now(), had to modify and use time.Unix(time.Now().Unix(), 0) to solve the problem.
				)
			} else {
				b.WriteString(
					ufmt.Sprintf(
						"### Anonymous: %dugnot\n", contributors[i].Amount,
					),
				)
			}
		}
	}
	return b.String()
}
