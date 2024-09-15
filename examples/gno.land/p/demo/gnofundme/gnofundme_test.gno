package gnofundme

import (
	"strconv"
	"testing"
	"time"

	"gno.land/p/demo/testutils"
)

var (
	ctype              = 0
	title              = "Campaign 1"
	description        = "Description 1"
	goal               = 1000
	ownerAddress       = testutils.TestAddress("owner")
	beneficiaryAddress = testutils.TestAddress("beneficiary")
	owner              = NewUser("Owner", "", ownerAddress)
	beneficiary        = NewUser("Beneficiary", "", beneficiaryAddress)
	// Test case 1 Valid campaign
	begin    = time.Now()
	deadline = begin.Add(time.Hour * 24)
	campaign = NewCampaign(ctype, title, description, 1000, begin, deadline, owner, beneficiary)
)

func TestNewCampaign(t *testing.T) {
	if campaign.GetCtype() != ctype {
		t.Errorf("Expected campaign type to be Personal, got %s", campaign.GetCtype())
	}

	if campaign.GetTitle() != title {
		t.Errorf("Expected campaign name to be 'Campaign 1', got '%s'", campaign.GetTitle())
	}
	if campaign.GetDescription() != description {
		t.Errorf("Expected campaign description to be 'Description 1', got '%s'", campaign.GetDescription())
	}
	if campaign.GetGoal() != 1000 {
		t.Errorf("Expected campaign target amount to be 1000, got %d", campaign.GetGoal())
	}
	if campaign.GetCurrent() != 0 {
		t.Errorf("Expected campaign current amount to be 0, got %d", campaign.GetCurrent())
	}
	if campaign.GetBegin() != begin {
		t.Errorf("Expected campaign begin to be %s, got %s", begin, campaign.GetBegin())
	}
	if campaign.GetDeadline() != deadline {
		t.Errorf("Expected campaign deadline to be %s, got %s", deadline, campaign.GetDeadline())
	}
	if campaign.GetOwner() != owner {
		t.Errorf("Expected campaign owner to be %s, got %s", owner, campaign.GetOwner())
	}
	if campaign.GetBeneficiary() != beneficiary {
		t.Errorf("Expected campaign beneficiary to be %s, got %s", beneficiary, campaign.GetBeneficiary())
	}
	if len(campaign.GetContributors()) != 0 {
		t.Errorf("Expected 0 contributors, got %d", len(campaign.GetContributors()))
	}
}

func TestAddContributor(t *testing.T) {
	contributorAddress := testutils.TestAddress("contributor")
	user := NewUser("Contributor", "", contributorAddress)
	var amount uint = 500
	display := true
	campaign.AddContributor(user, amount, display)

	contributors := campaign.GetContributors()
	if campaign.current != 500 {
		t.Errorf("Expected campaign current amount to be 500, got %d", campaign.current)
	}

	if len(contributors) != 1 {
		t.Errorf("Expected 1 contributor, got %d", len(contributors))
	}

	if contributors[0].user != user {
		t.Errorf("Wrong user")
	}

	if contributors[0].amount != 500 {
		t.Errorf("Expected contributor amount to be 500, got %d", contributors[0].amount)
	}
}

func TestCampaignsToJSONString(t *testing.T) {
	campaigns := []*Campaign{
		NewCampaign(0, "Campaign 1", "Description 1", 1000, time.Now(), time.Now(), NewUser("Owner", "", ownerAddress), NewUser("Beneficiary", "", beneficiaryAddress)),
		NewCampaign(1, "Campaign 2", "Description 2", 2000, time.Now(), time.Now(), NewUser("Owner", "", ownerAddress), NewUser("Beneficiary", "", beneficiaryAddress)),
		NewCampaign(2, "Campaign 3", "Description 3", 3000, time.Now(), time.Now(), NewUser("Owner", "", ownerAddress), NewUser("Beneficiary", "", beneficiaryAddress)),
	}

	expected := `[{"ctype":"0","title":"Campaign 1","description":"Description 1","goal":"1000","current":"0","begin":"` + strconv.Itoa(int(time.Now().Unix())) + `","deadline":"` + strconv.Itoa(int(time.Now().Unix())) + `","owner":"` + ownerAddress.String() + `","beneficiary":"` + beneficiaryAddress.String() + `","contributors":"0"},{"ctype":"1","title":"Campaign 2","description":"Description 2","goal":"2000","current":"0","begin":"` + strconv.Itoa(int(time.Now().Unix())) + `","deadline":"` + strconv.Itoa(int(time.Now().Unix())) + `","owner":"` + ownerAddress.String() + `","beneficiary":"` + beneficiaryAddress.String() + `","contributors":"0"},{"ctype":"2","title":"Campaign 3","description":"Description 3","goal":"3000","current":"0","begin":"` + strconv.Itoa(int(time.Now().Unix())) + `","deadline":"` + strconv.Itoa(int(time.Now().Unix())) + `","owner":"` + ownerAddress.String() + `","beneficiary":"` + beneficiaryAddress.String() + `","contributors":"0"}]`

	result := CampaignsToJSONString(campaigns)
	if result != expected {
		t.Errorf("Expected JSON string: %s,\n*****got JSON string: %s", expected, result)
	}
}

func TestCampaignToJSONString(t *testing.T) {
	campaign := NewCampaign(0, "Campaign 1", "Description 1", 1000, time.Now(), time.Now(), NewUser("Owner", "", ownerAddress), NewUser("Beneficiary", "", beneficiaryAddress))

	expected := `{"ctype":"0","title":"Campaign 1","description":"Description 1","goal":"1000","current":"0","begin":"` + strconv.Itoa(int(time.Now().Unix())) + `","deadline":"` + strconv.Itoa(int(time.Now().Unix())) + `","owner":"` + ownerAddress.String() + `","beneficiary":"` + beneficiaryAddress.String() + `","contributors":"0"}`

	result := CampaignToJSONString(campaign)
	if result != expected {
		t.Errorf("Expected JSON string: %s, \ngot: %s", expected, result)
	}
}
