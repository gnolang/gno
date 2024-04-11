package gnofundme

import (
	"testing"
	"time"

	"gno.land/p/demo/testutils"
)

func TestNewCampaign(t *testing.T) {
	var (
		title       = "Campaign 1"
		description = "Description 1"
		goal        = 1000
	)
	ownerAddress := testutils.TestAddress("owner")
	beneficiaryAddress := testutils.TestAddress("beneficiary")
	owner := NewUser("Owner", "", ownerAddress)
	beneficiary := NewUser("Beneficiary", "", beneficiaryAddress)
	// Test case 1: Valid campaign
	begin := time.Now()
	deadline := begin.Add(time.Hour * 24)
	campaign := NewCampaign(title, description, 1000, begin, deadline, owner, beneficiary)

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
	// Test case 2: Empty campaign name
	// campaign = NewCampaign("", "Description 2", 2000)
	// if campaign.Name != "" {
	// 	t.Errorf("Expected campaign name to be '', got '%s'", campaign.Name)
	// }
	// if campaign.Description != "Description 2" {
	// 	t.Errorf("Expected campaign description to be 'Description 2', got '%s'", campaign.Description)
	// }
	// if campaign.TargetAmount != 2000 {
	// 	t.Errorf("Expected campaign target amount to be 2000, got %d", campaign.TargetAmount)
	// }
}

// func TestAddContributor(t *testing.T) {
// 	campaign := NewCampaign("Campaign 1", "Description 1", 1000)

// 	// Test case 1: Valid contributor
// 	contributor := Contributor{
// 		Name:   "Contributor 1",
// 		Email:  "contributor1@example.com",
// 		Amount: 500,
// 	}
// 	err := campaign.AddContributor(contributor)
// 	if err != nil {
// 		t.Errorf("Expected no error, got '%s'", err.Error())
// 	}
// 	if len(campaign.Contributors) != 1 {
// 		t.Errorf("Expected 1 contributor, got %d", len(campaign.Contributors))
// 	}
// 	if campaign.Contributors[0].Name != "Contributor 1" {
// 		t.Errorf("Expected contributor name to be 'Contributor 1', got '%s'", campaign.Contributors[0].Name)
// 	}
// 	if campaign.Contributors[0].Email != "contributor1@example.com" {
// 		t.Errorf("Expected contributor email to be 'contributor1@example.com', got '%s'", campaign.Contributors[0].Email)
// 	}
// 	if campaign.Contributors[0].Amount != 500 {
// 		t.Errorf("Expected contributor amount to be 500, got %d", campaign.Contributors[0].Amount)
// 	}

// 	// Test case 2: Invalid contributor (negative amount)
// 	contributor = Contributor{
// 		Name:   "Contributor 2",
// 		Email:  "contributor2@example.com",
// 		Amount: -200,
// 	}
// 	err = campaign.AddContributor(contributor)
// 	if err == nil {
// 		t.Errorf("Expected an error, got nil")
// 	}
// 	if len(campaign.Contributors) != 1 {
// 		t.Errorf("Expected 1 contributor, got %d", len(campaign.Contributors))
// 	}

// 	// Test case 3: Invalid contributor (empty name)
// 	contributor = Contributor{
// 		Name:   "",
// 		Email:  "contributor3@example.com",
// 		Amount: 1000,
// 	}
// 	err = campaign.AddContributor(contributor)
// 	if err == nil {
// 		t.Errorf("Expected an error, got nil")
// 	}
// 	if len(campaign.Contributors) != 1 {
// 		t.Errorf("Expected 1 contributor, got %d", len(campaign.Contributors))
// 	}
// }
