package gnofundme

import (
	"strconv"
	"testing"
	"time"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/r/demo/users"
)

const (
	ct                 = ctype(0)
	title              = "Campaign 1"
	description        = "Description 1"
	goal               = uint64(1000)
	ownerAddress       = testutils.TestAddress("owner")
	beneficiaryAddress = testutils.TestAddress("beneficiary")
	owner              = users.GetUserByName("demo")
	beneficiary        = users.GetUserByName("gno")
	// Test case 1 Valid campaign
	begin    = time.Now()
	deadline = begin.Add(time.Hour * 24)
)

var campaign = NewCampaign(0, title, description, 1000, begin, deadline, owner, owner)

func TestNewCampaign(t *testing.T) {
	uassert.Equal(t, campaign.Ctype.String(), ct.String())

	uassert.Equal(t, campaign.Title, title)
	uassert.Equal(t, campaign.Description, description)
	uassert.Equal(t, campaign.Goal, goal, "goal")
	uassert.Equal(t, campaign.Current, uint64(0))
	uassert.Equal(t, strconv.Itoa(int(campaign.Begin.Unix())), strconv.Itoa(int(begin.Unix())))
	uassert.Equal(t, strconv.Itoa(int(campaign.Deadline.Unix())), strconv.Itoa(int(deadline.Unix())))
	uassert.Equal(t, campaign.Owner.Address.String(), owner.Address.String())
	uassert.Equal(t, campaign.Beneficiary.Address.String(), owner.Address.String())
	uassert.Equal(t, len(campaign.GetContributors()), 0)
}

func TestAddContributor(t *testing.T) {
	contributorAddress := testutils.TestAddress("contributor")
	user := users.GetUserByName("demo")
	var amount uint64 = 500
	display := true
	date := time.Unix(time.Now().Unix(), 0)
	campaign.AddContributor(user, date, amount, display)

	contributors := campaign.GetContributors()
	uassert.Equal(t, campaign.Current, uint64(500))

	uassert.Equal(t, len(contributors), 1)

	uassert.Equal(t, contributors[0].user.Address.String(), user.Address.String())

	uassert.Equal(t, contributors[0].Amount, uint64(500))
}

func TestCampaignsToJSONString(t *testing.T) {
	campaigns := []*Campaign{
		NewCampaign(0, "Campaign 1", "Description 1", 1000, time.Now(), time.Now(), users.GetUserByName("demo"), users.GetUserByName("demo")),
		NewCampaign(1, "Campaign 2", "Description 2", 2000, time.Now(), time.Now(), users.GetUserByName("demo"), users.GetUserByName("demo")),
		NewCampaign(2, "Campaign 3", "Description 3", 3000, time.Now(), time.Now(), users.GetUserByName("demo"), users.GetUserByName("demo")),
	}

	expected := `[{"Ctype":"Personal","Title":"Campaign 1","Description":"Description 1","Goal":1000,"Current":0,"Begin":"` + strconv.Itoa(int(time.Now().Unix())) + `","Deadline":"` + strconv.Itoa(int(time.Now().Unix())) + `","Owner":"` + owner.Address.String() + `","Beneficiary":"` + owner.Address.String() + `","Contributors":0},{"Ctype":"Cause","Title":"Campaign 2","Description":"Description 2","Goal":2000,"Current":0,"Begin":"` + strconv.Itoa(int(time.Now().Unix())) + `","Deadline":"` + strconv.Itoa(int(time.Now().Unix())) + `","Owner":"` + owner.Address.String() + `","Beneficiary":"` + owner.Address.String() + `","Contributors":0},{"Ctype":"Organism","Title":"Campaign 3","Description":"Description 3","Goal":3000,"Current":0,"Begin":"` + strconv.Itoa(int(time.Now().Unix())) + `","Deadline":"` + strconv.Itoa(int(time.Now().Unix())) + `","Owner":"` + owner.Address.String() + `","Beneficiary":"` + owner.Address.String() + `","Contributors":0}]`

	result := CampaignsToJSONString(campaigns)
	uassert.Equal(t, result, expected)
}

func TestCampaignToJSONString(t *testing.T) {
	campaign := NewCampaign(0, "Campaign 1", "Description 1", 1000, time.Now(), time.Now(), owner, owner)

	expected := `{"Ctype":"Personal","Title":"Campaign 1","Description":"Description 1","Goal":1000,"Current":0,"Begin":"` + strconv.Itoa(int(time.Now().Unix())) + `","Deadline":"` + strconv.Itoa(int(time.Now().Unix())) + `","Owner":"` + owner.Address.String() + `","Beneficiary":"` + owner.Address.String() + `","Contributors":0}`

	result := campaign.CampaignToJSONString()
	uassert.Equal(t, result, expected)
}
