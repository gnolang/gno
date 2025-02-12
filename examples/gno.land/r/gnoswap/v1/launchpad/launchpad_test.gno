package launchpad

import (
	"std"
	"testing"
	"time"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/gns"
	"gno.land/r/onbloc/obl"

	u256 "gno.land/p/gnoswap/uint256"
)

func setupTest(t *testing.T) {
	projects = make(map[string]Project)
	projectTiersWithoutDeposit = make(map[string]bool)
}

func createTestAddress(t *testing.T) std.Address {
	addr := testutils.TestAddress("test")
	return addr
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name      string
		inputName string
		wantErr   bool
	}{
		{"ValidName", "MyProject", false},
		{"EmptyName", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := ProjectInput{Name: tt.inputName}
			err := in.ValidateName()
			if tt.wantErr {
				uassert.Error(t, err)
			} else {
				uassert.NoError(t, err)
			}
		})
	}
}

func TestValidateTokenPath(t *testing.T) {
	tests := []struct {
		name      string
		tokenPath string
		wantErr   bool
	}{
		{"ValidToken", "gno.land/r/gnoswap/v1/gns", false},
		{"EmptyToken", "", true},
		{"UnregisteredToken", "invalid/token", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := ProjectInput{TokenPath: tt.tokenPath}
			err := in.ValidateTokenPath()
			if tt.wantErr {
				uassert.Error(t, err)
			} else {
				uassert.NoError(t, err)
			}
		})
	}
}

func TestValidateRecipient(t *testing.T) {
	tests := []struct {
		name      string
		recipient std.Address
		wantErr   bool
	}{
		{"ValidRecipient", std.Address(consts.ADMIN), false},
		{"InvalidRecipient", std.Address("xyz..."), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := ProjectInput{Recipient: tt.recipient}
			err := in.ValidateRecipient()
			if tt.wantErr {
				uassert.Error(t, err)
			} else {
				uassert.NoError(t, err)
			}
		})
	}
}

func TestValidateDepositAmount(t *testing.T) {
	tests := []struct {
		name          string
		depositAmount uint64
		wantErr       bool
	}{
		{"Valid", 100, false},
		{"ZeroDeposit", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := ProjectInput{DepositAmount: tt.depositAmount}
			err := in.ValidateDepositAmount()
			if tt.wantErr {
				uassert.Error(t, err)
			} else {
				uassert.NoError(t, err)
			}
		})
	}
}

func TestValidateRatio(t *testing.T) {
	tests := []struct {
		name           string
		t30, t90, t180 uint64
		wantErr        bool
	}{
		{"ValidSum100", 30, 50, 20, false},
		{"Not100", 10, 10, 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := ProjectInput{
				Tier30Ratio:  tt.t30,
				Tier90Ratio:  tt.t90,
				Tier180Ratio: tt.t180,
			}
			err := in.ValidateRatio()
			if tt.wantErr {
				uassert.Error(t, err)
			} else {
				uassert.NoError(t, err)
			}
		})
	}
}

func TestValidateStartTime(t *testing.T) {
	now := uint64(1000)
	tests := []struct {
		name      string
		startTime uint64
		wantErr   bool
	}{
		{"ValidFuture", 2000, false},
		{"Now", 1000, true},
		{"Past", 999, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := ProjectInput{StartTime: tt.startTime}
			err := in.ValidateStartTime(now)
			if tt.wantErr {
				uassert.Error(t, err)
			} else {
				uassert.NoError(t, err)
			}
		})
	}
}

func TestParseConditions(t *testing.T) {
	tests := []struct {
		name        string
		condToken   string
		condAmount  string
		wantErr     bool
		expectedLen int
	}{
		{
			name:        "ValidMultiple",
			condToken:   "gno.land/r/demo/wugnot*PAD*gno.land/r/gnoswap/v1/gns",
			condAmount:  "100*PAD*200",
			wantErr:     false,
			expectedLen: 2,
		},
		{
			name:        "EmptyToken",
			condToken:   "",
			condAmount:  "100",
			wantErr:     true,
			expectedLen: 0,
		},
		{
			name:        "LenMismatch",
			condToken:   "valid/T0*PAD*valid/T1*PAD*valid/T2",
			condAmount:  "100*PAD*200",
			wantErr:     true,
			expectedLen: 0,
		},
		{
			name:        "UnregisteredToken",
			condToken:   "invalid/token",
			condAmount:  "100",
			wantErr:     true,
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := ProjectInput{
				ConditionsToken:  tt.condToken,
				ConditionsAmount: tt.condAmount,
			}
			conds, err := in.ParseConditions()
			if tt.wantErr {
				uassert.Error(t, err)
			} else {
				uassert.NoError(t, err)
				uassert.Equal(t, tt.expectedLen, len(conds))
				if tt.expectedLen > 0 {
					uassert.Equal(t, uint64(100), conds["gno.land/r/demo/wugnot"].minAmount)
					uassert.Equal(t, uint64(200), conds["gno.land/r/gnoswap/v1/gns"].minAmount)
				}
			}
		})
	}
}

func TestValidateAll(t *testing.T) {
	now := uint64(1000)
	tests := []struct {
		name    string
		input   ProjectInput
		wantErr bool
	}{
		{
			name: "AllValid",
			input: ProjectInput{
				Name:             "ProjectName",
				TokenPath:        "gno.land/r/gnoswap/v1/gns",
				Recipient:        std.Address(consts.ADMIN),
				DepositAmount:    1000,
				ConditionsToken:  "gno.land/r/demo/wugnot",
				ConditionsAmount: "10",
				Tier30Ratio:      50,
				Tier90Ratio:      30,
				Tier180Ratio:     20,
				StartTime:        2000,
			},
			wantErr: false,
		},
		{
			name: "InvalidName",
			input: ProjectInput{
				Name:             "",
				TokenPath:        "valid/token",
				Recipient:        std.Address("g1valid"),
				DepositAmount:    10,
				ConditionsToken:  "valid/T0",
				ConditionsAmount: "1",
				Tier30Ratio:      80,
				Tier90Ratio:      10,
				Tier180Ratio:     10,
				StartTime:        2000,
			},
			wantErr: true,
		},
		{
			name: "BadTokenPath",
			input: ProjectInput{
				Name:             "Proj",
				TokenPath:        "invalid/token",
				Recipient:        std.Address("g1valid"),
				DepositAmount:    10,
				ConditionsToken:  "valid/T0",
				ConditionsAmount: "1",
				Tier30Ratio:      80,
				Tier90Ratio:      10,
				Tier180Ratio:     10,
				StartTime:        2000,
			},
			wantErr: true,
		},
		{
			name: "BadRecipient",
			input: ProjectInput{
				Name:             "Proj",
				TokenPath:        "valid/token",
				Recipient:        std.Address("xinvalid"),
				DepositAmount:    10,
				ConditionsToken:  "valid/T0",
				ConditionsAmount: "1",
				Tier30Ratio:      80,
				Tier90Ratio:      10,
				Tier180Ratio:     10,
				StartTime:        2000,
			},
			wantErr: true,
		},
		{
			name: "BadDepositAmount",
			input: ProjectInput{
				Name:             "Proj",
				TokenPath:        "valid/token",
				Recipient:        std.Address("g1abc"),
				DepositAmount:    0,
				ConditionsToken:  "valid/T0",
				ConditionsAmount: "1",
				Tier30Ratio:      80,
				Tier90Ratio:      10,
				Tier180Ratio:     10,
				StartTime:        2000,
			},
			wantErr: true,
		},
		{
			name: "BadRatio",
			input: ProjectInput{
				Name:             "Proj",
				TokenPath:        "valid/token",
				Recipient:        std.Address("g1abc"),
				DepositAmount:    100,
				ConditionsToken:  "valid/T0",
				ConditionsAmount: "1",
				Tier30Ratio:      80,
				Tier90Ratio:      10,
				Tier180Ratio:     30, // sum=120 !=100
				StartTime:        2000,
			},
			wantErr: true,
		},
		{
			name: "StartTimePast",
			input: ProjectInput{
				Name:             "Proj",
				TokenPath:        "valid/token",
				Recipient:        std.Address("g1abc"),
				DepositAmount:    100,
				ConditionsToken:  "valid/T0",
				ConditionsAmount: "1",
				Tier30Ratio:      80,
				Tier90Ratio:      10,
				Tier180Ratio:     10,
				StartTime:        999, // <=now
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.input.ValidateAll(now)
			if tt.wantErr {
				uassert.Error(t, err)
			} else {
				uassert.NoError(t, err)
			}
		})
	}
}

func TestValidateProjectInput(t *testing.T) {
	now := uint64(time.Now().Unix())
	testCases := []struct {
		name        string
		input       ProjectInput
		shouldError bool
	}{
		{
			name: "Valid input",
			input: ProjectInput{
				Name:             "Test Project",
				TokenPath:        "gno.land/r/gnoswap/v1/gns",
				Recipient:        createTestAddress(t),
				DepositAmount:    1000,
				ConditionsToken:  "gno.land/r/demo/wugnot",
				ConditionsAmount: "10",
				Tier30Ratio:      30,
				Tier90Ratio:      30,
				Tier180Ratio:     40,
				StartTime:        now + 3600, // 1 hour in future
			},
			shouldError: false,
		},
		{
			name: "Invalid tier ratios",
			input: ProjectInput{
				Name:          "Test Project",
				TokenPath:     "test/token",
				Recipient:     createTestAddress(t),
				DepositAmount: 1000,
				Tier30Ratio:   30,
				Tier90Ratio:   30,
				Tier180Ratio:  30, // Sum != 100
				StartTime:     now + 3600,
			},
			shouldError: true,
		},
		{
			name: "Start time in past",
			input: ProjectInput{
				Name:          "Test Project",
				TokenPath:     "test/token",
				Recipient:     createTestAddress(t),
				DepositAmount: 1000,
				Tier30Ratio:   30,
				Tier90Ratio:   30,
				Tier180Ratio:  40,
				StartTime:     now - 3600, // 1 hour in past
			},
			shouldError: true,
		},
		{
			name: "Zero deposit amount",
			input: ProjectInput{
				Name:          "Test Project",
				TokenPath:     "test/token",
				Recipient:     createTestAddress(t),
				DepositAmount: 0,
				Tier30Ratio:   30,
				Tier90Ratio:   30,
				Tier180Ratio:  40,
				StartTime:     now + 3600,
			},
			shouldError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.input.ValidateAll(now)
			if tc.shouldError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tc.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err.Error())
			}
		})
	}
}

func TestProject_GetSet(t *testing.T) {
	conditions := make(map[string]Condition)
	tiers := make(map[uint64]Tier)
	tiersRatios := make(map[uint64]uint64)

	p := NewProject(
		"pid",
		"MyProject",
		"gno.land/r/gnoswap/v1/gns",
		1000,
		"g1qxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		conditions,
		tiers,
		tiersRatios,
		TimeInfo{height: 1, time: 1001},
		TimeInfo{height: 2, time: 1002},
		TimeInfo{height: 3, time: 1003},
		*NewProjectStats(10, 20, 30, 40, 50),
		RefundInfo{amount: 99, height: 10, time: 999},
	)

	if p.ID() != "pid" {
		t.Errorf("NewProject id mismatch, got=%s", p.ID())
	}
	if p.Name() != "MyProject" {
		t.Errorf("NewProject name mismatch, got=%s", p.Name())
	}
	if p.TokenPath() != "gno.land/r/gnoswap/v1/gns" {
		t.Errorf("NewProject tokenPath mismatch, got=%s", p.TokenPath())
	}
	if p.DepositAmount() != 1000 {
		t.Errorf("NewProject depositAmount mismatch, got=%d", p.DepositAmount())
	}
	if p.Recipient().String() != "g1qxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" {
		t.Errorf("NewProject recipient mismatch, got=%s", p.Recipient())
	}
	created := p.Created()
	if created.Height() != 1 {
		t.Errorf("NewProject Created Height mismatch, got=%d", created.Height())
	}
	ended := p.Ended()
	if ended.Time() != 1003 {
		t.Errorf("NewProject Ended Time mismatch, got=%d", ended.Time())
	}
	stats := p.Stats()
	if stats.TotalCollected() != 50 {
		t.Errorf("NewProject Stats mismatch, got=%d", stats.TotalCollected())
	}
	refund := p.Refund()
	if refund.Amount() != 99 {
		t.Errorf("NewProject Refund mismatch, got=%d", refund.Amount())
	}

	p.SetID("pid2")
	if p.ID() != "pid2" {
		t.Errorf("SetID failed, got=%s", p.ID())
	}
	p.SetName("MyProject2")
	if p.Name() != "MyProject2" {
		t.Errorf("SetName failed, got=%s", p.Name())
	}
	p.SetTokenPath("new/tokenPath")
	if p.TokenPath() != "new/tokenPath" {
		t.Errorf("SetTokenPath failed, got=%s", p.TokenPath())
	}
	p.SetDepositAmount(2000)
	if p.DepositAmount() != 2000 {
		t.Errorf("SetDepositAmount failed, got=%d", p.DepositAmount())
	}
	t.Run("TestProject_isActivated", func(t *testing.T) {
		now := uint64(1)
		if p.isActivated(now) {
			t.Errorf("isActivated(%d) returned true, want false (before started)", now)
		}

		now = uint64(2)
		if !p.isActivated(now) {
			t.Errorf("isActivated(%d) returned false, want true (exactly started)", now)
		}

		now = uint64(2)
		if !p.isActivated(now) {
			t.Errorf("isActivated(%d) returned false, want true (between start and end)", now)
		}

		now = uint64(3)
		if p.isActivated(now) {
			t.Errorf("isActivated(%d) returned true, want false (exactly ended)", now)
		}

		now = uint64(4)
		if p.isActivated(now) {
			t.Errorf("isActivated(%d) returned true, want false (after ended)", now)
		}
	})
}

func TestTier_GetSet(t *testing.T) {
	emptyUint := u256.Zero()

	tr := NewTier(
		"tierId", 10, 100,
		emptyUint.Clone(),
		TimeInfo{height: 0, time: 0},
		TimeInfo{height: 100, time: 9999},
		1000, 800, 50, 45, 12, 200,
		*NewReward(u256.NewUint(1234), 10, 9999),
	)

	if tr.ID() != "tierId" {
		t.Errorf("Tier ID mismatch, got=%s", tr.ID())
	}
	if tr.CollectWaitDuration() != 10 {
		t.Errorf("Tier CollectWaitDuration mismatch, got=%d", tr.CollectWaitDuration())
	}
	if tr.TierAmount() != 100 {
		t.Errorf("Tier TierAmount mismatch, got=%d", tr.TierAmount())
	}
	if tr.TierAmountPerBlockX128().Cmp(u256.Zero()) != 0 {
		t.Errorf("Tier TierAmountPerBlockX128 mismatch, expect=0")
	}
	if tr.TotalDepositAmount() != 1000 {
		t.Errorf("Tier TotalDepositAmount mismatch, got=%d", tr.TotalDepositAmount())
	}
	if tr.ActualDepositAmount() != 800 {
		t.Errorf("Tier ActualDepositAmount mismatch, got=%d", tr.ActualDepositAmount())
	}
	if tr.TotalParticipant() != 50 {
		t.Errorf("Tier TotalParticipant mismatch, got=%d", tr.TotalParticipant())
	}
	if tr.ActualParticipant() != 45 {
		t.Errorf("Tier ActualParticipant mismatch, got=%d", tr.ActualParticipant())
	}
	if tr.UserCollectedAmount() != 12 {
		t.Errorf("Tier UserCollectedAmount mismatch, got=%d", tr.UserCollectedAmount())
	}
	if tr.CalculatedAmount() != 200 {
		t.Errorf("Tier CalculatedAmount mismatch, got=%d", tr.CalculatedAmount())
	}

	tr.SetID("tierIdNew")
	if tr.ID() != "tierIdNew" {
		t.Errorf("SetID failed, got=%s", tr.ID())
	}
	tr.SetTierAmount(999)
	if tr.TierAmount() != 999 {
		t.Errorf("SetTierAmount failed, got=%d", tr.TierAmount())
	}
	tr.SetActualDepositAmount(777)
	if tr.ActualDepositAmount() != 777 {
		t.Errorf("SetActualDepositAmount failed, got=%d", tr.ActualDepositAmount())
	}
}

func TestRewardInfo_GetSet(t *testing.T) {
	ri := NewRewardInfo(
		u256.NewUint(111),
		222, 333, 444, 555, 666, 777,
	)
	if ri.PriceDebt().ToString() != "111" {
		t.Errorf("PriceDebt mismatch, got=%s", ri.PriceDebt().ToString())
	}
	if ri.DepositAmount() != 222 {
		t.Errorf("DepositAmount mismatch, got=%d", ri.DepositAmount())
	}
	if ri.RewardAmount() != 333 {
		t.Errorf("RewardAmount mismatch, got=%d", ri.RewardAmount())
	}
	if ri.Claimed() != 444 {
		t.Errorf("Claimed mismatch, got=%d", ri.Claimed())
	}
	if ri.StartHeight() != 555 {
		t.Errorf("StartHeight mismatch, got=%d", ri.StartHeight())
	}
	if ri.GetEndHeight() != 666 {
		t.Errorf("EndHeight mismatch, got=%d", ri.GetEndHeight())
	}
	if ri.GetLastHeight() != 777 {
		t.Errorf("LastHeight mismatch, got=%d", ri.GetLastHeight())
	}
	ri.SetPriceDebt(u256.NewUint(999))
	if ri.PriceDebt().ToString() != "999" {
		t.Errorf("SetPriceDebt failed, got=%s", ri.PriceDebt().ToString())
	}
	ri.SetRewardAmount(1234)
	if ri.RewardAmount() != 1234 {
		t.Errorf("SetRewardAmount failed, got=%d", ri.RewardAmount())
	}
}

func TestTimeInfo_GetSet(t *testing.T) {
	ti := NewTimeInfo(50, 5000)

	if ti.Height() != 50 {
		t.Errorf("CurrentHeight mismatch, got=%d", ti.Height())
	}
	if ti.Time() != 5000 {
		t.Errorf("Time mismatch, got=%d", ti.Time())
	}

	ti.SetHeight(100)
	ti.SetTime(9999)
	if ti.Height() != 100 {
		t.Errorf("SetHeight failed, got=%d", ti.Height())
	}
	if ti.Time() != 9999 {
		t.Errorf("SetTime failed, got=%d", ti.Time())
	}
}

func TestRefundInfo_GetSet(t *testing.T) {
	ri := NewRefundInfo(11, 22, 33)
	if ri.Amount() != 11 {
		t.Errorf("Amount mismatch, got=%d", ri.Amount())
	}
	if ri.Height() != 22 {
		t.Errorf("CurrentHeight mismatch, got=%d", ri.Height())
	}
	if ri.Time() != 33 {
		t.Errorf("Time mismatch, got=%d", ri.Time())
	}

	ri.SetAmount(111)
	ri.SetHeight(222)
	ri.SetTime(333)
	if ri.Amount() != 111 {
		t.Errorf("SetAmount failed, got=%d", ri.Amount())
	}
	if ri.Height() != 222 {
		t.Errorf("SetHeight failed, got=%d", ri.Height())
	}
	if ri.Time() != 333 {
		t.Errorf("SetTime failed, got=%d", ri.Time())
	}
}

func TestCondition_GetSet(t *testing.T) {
	c := NewCondition("my/token", 1000)
	if c.TokenPath() != "my/token" {
		t.Errorf("Condition TokenPath mismatch, got=%s", c.TokenPath())
	}
	if c.MinAmount() != 1000 {
		t.Errorf("Condition MinAmount mismatch, got=%d", c.MinAmount())
	}

	c.SetTokenPath("new/token")
	if c.TokenPath() != "new/token" {
		t.Errorf("SetTokenPath failed, got=%s", c.TokenPath())
	}
	c.SetMinAmount(2000)
	if c.MinAmount() != 2000 {
		t.Errorf("SetMinAmount failed, got=%d", c.MinAmount())
	}
}

func TestReward_GetSet(t *testing.T) {
	acc := u256.NewUint(1000)
	r := NewReward(acc, 10, 1000)

	if r.AccumRewardPerDeposit().Cmp(u256.NewUint(1000)) != 0 {
		t.Errorf("AccumRewardPerDeposit mismatch, got=%s", r.AccumRewardPerDeposit().ToString())
	}
	if r.GetLastHeight() != 10 {
		t.Errorf("LastHeight mismatch, got=%d", r.GetLastHeight())
	}
	if r.GetEndHeight() != 1000 {
		t.Errorf("EndHeight mismatch, got=%d", r.GetEndHeight())
	}
	if r.Info() == nil {
		t.Errorf("Info mismatch, got=%v, want=not nil", r.Info())
	}
	r.SetAccumRewardPerDeposit(u256.NewUint(9999))
	if r.AccumRewardPerDeposit().ToString() != "9999" {
		t.Errorf("SetAccumRewardPerDeposit failed, got=%s", r.AccumRewardPerDeposit().ToString())
	}
	r.SetLastHeight(111)
	if r.GetLastHeight() != 111 {
		t.Errorf("SetLastHeight failed, got=%d", r.GetLastHeight())
	}
}

func TestCreateProject(t *testing.T) {
	t.Skip("fix this test")
	projectAddr := testutils.TestAddress("projectAddr")
	fooBarWithpad := "gno.land/r/onbloc/foo*PAD*gno.land/r/onbloc/bar"
	tests := []struct {
		name          string
		projectName   string
		tokenPath     string
		projectOwner  std.Address
		totalAmount   uint64
		conditions    string
		ratios        string
		tier30Ratio   uint64
		tier90Ratio   uint64
		tier180Ratio  uint64
		startTime     uint64
		expectedId    string
		expectedError string
		duplicateTest bool
	}{
		{
			name:          "normal project creation",
			projectName:   "Obl Protocol",
			tokenPath:     oblPath,
			projectOwner:  projectAddr,
			totalAmount:   uint64(1_000_000_000),
			conditions:    fooBarWithpad,
			ratios:        "1*PAD*2",
			tier30Ratio:   uint64(10),
			tier90Ratio:   uint64(20),
			tier180Ratio:  uint64(70),
			startTime:     uint64(time.Now().Unix() + 10),
			expectedId:    "gno.land/r/onbloc/obl:124",
			expectedError: "",
		},
		{
			name:          "invalid ratio sum (less than 100%)",
			projectName:   "Invalid Ratio",
			tokenPath:     oblPath,
			projectOwner:  projectAddr,
			totalAmount:   uint64(1_000_000_000),
			conditions:    fooBarWithpad,
			ratios:        "1*PAD*2",
			tier30Ratio:   uint64(10),
			tier90Ratio:   uint64(20),
			tier180Ratio:  uint64(60),
			startTime:     uint64(time.Now().Unix() + 10),
			expectedError: "[GNOSWAP-LAUNCHPAD-007] invalid input data || invalid ratio, sum of all tiers(30:10, 90:20, 180:60) should be 100",
		},
		{
			name:          "project name is empty",
			projectName:   "",
			tokenPath:     oblPath,
			projectOwner:  projectAddr,
			totalAmount:   uint64(1_000_000_000),
			conditions:    fooBarWithpad,
			ratios:        "1*PAD*2",
			tier30Ratio:   uint64(10),
			tier90Ratio:   uint64(20),
			tier180Ratio:  uint64(70),
			startTime:     uint64(time.Now().Unix() + 10),
			expectedError: "[GNOSWAP-LAUNCHPAD-007] invalid input data || project name cannot be empty",
		},
		{
			name:          "project owner address is invalid",
			projectName:   "TTT",
			tokenPath:     oblPath,
			projectOwner:  "abcdef",
			totalAmount:   uint64(1_000_000_000),
			conditions:    fooBarWithpad,
			ratios:        "1*PAD*2",
			tier30Ratio:   uint64(10),
			tier90Ratio:   uint64(20),
			tier180Ratio:  uint64(70),
			startTime:     uint64(time.Now().Unix() + 10),
			expectedError: "[GNOSWAP-LAUNCHPAD-002] invalid address || recipient address(abcdef)",
		},
		{
			name:          "invalid start time",
			projectName:   "Obl Protocol",
			tokenPath:     oblPath,
			projectOwner:  projectAddr,
			totalAmount:   uint64(1_000_000_000),
			conditions:    fooBarWithpad,
			ratios:        "1*PAD*2",
			tier30Ratio:   uint64(10),
			tier90Ratio:   uint64(20),
			tier180Ratio:  uint64(70),
			startTime:     uint64(time.Now().Unix()),
			expectedId:    "gno.land/r/onbloc/obl:124",
			expectedError: "[GNOSWAP-LAUNCHPAD-007] invalid input data || start time(1234567890) must be greater than now(1234567902)",
		},
		{
			name:          "invalid token path",
			projectName:   "Obl Protocol",
			tokenPath:     "gno.land/r/test/token",
			projectOwner:  projectAddr,
			totalAmount:   uint64(1_000_000_000),
			conditions:    fooBarWithpad,
			ratios:        "1*PAD*2",
			tier30Ratio:   uint64(10),
			tier90Ratio:   uint64(20),
			tier180Ratio:  uint64(70),
			startTime:     uint64(time.Now().Unix() + 50),
			expectedId:    "gno.land/r/onbloc/obl:124",
			expectedError: "[GNOSWAP-LAUNCHPAD-007] invalid input data || tokenPath(gno.land/r/test/token) not registered",
		},
		{
			name:          "duplicate project test - origin",
			projectName:   "Obl Protocol",
			tokenPath:     oblPath,
			projectOwner:  projectAddr,
			totalAmount:   uint64(1_000_000_000),
			conditions:    fooBarWithpad,
			ratios:        "1*PAD*2",
			tier30Ratio:   uint64(10),
			tier90Ratio:   uint64(20),
			tier180Ratio:  uint64(70),
			startTime:     uint64(time.Now().Unix() + 50),
			expectedId:    "gno.land/r/onbloc/obl:130",
			expectedError: "",
			duplicateTest: true,
		},
		{
			name:          "duplicate project test - duplicate",
			projectName:   "Obl Protocol",
			tokenPath:     oblPath,
			projectOwner:  projectAddr,
			totalAmount:   uint64(1_000_000_000),
			conditions:    fooBarWithpad,
			ratios:        "1*PAD*2",
			tier30Ratio:   uint64(10),
			tier90Ratio:   uint64(20),
			tier180Ratio:  uint64(70),
			startTime:     uint64(time.Now().Unix() + 50),
			expectedId:    "gno.land/r/onbloc/obl:130",
			expectedError: "[GNOSWAP-LAUNCHPAD-008] can not create same project in same block || project(gno.land/r/onbloc/obl:130) already exists",
			duplicateTest: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			std.TestSetRealm(adminRealm)
			obl.Approve(consts.LAUNCHPAD_ADDR, tt.totalAmount)
			if !tt.duplicateTest {
				std.TestSkipHeights(1)
			}

			defer func() {
				if r := recover(); r != nil {
					if tt.expectedError == "" {
						t.Errorf("unexpected error: %v", r)
					} else {
						uassert.Equal(t, r.(string), tt.expectedError)
					}
				}
			}()

			projectId := CreateProject(
				tt.projectName,
				tt.tokenPath,
				tt.projectOwner,
				tt.totalAmount,
				tt.conditions,
				tt.ratios,
				tt.tier30Ratio,
				tt.tier90Ratio,
				tt.tier180Ratio,
				tt.startTime,
			)

			if tt.expectedError == "" {
				uassert.Equal(t, projectId, tt.expectedId)
			}
			if !tt.duplicateTest {
				std.TestSkipHeights(1)
			}
		})
	}
}

func TestCreateProject_TierAndBalance(t *testing.T) {
	t.Skip("fix this test")
	projectAddr := testutils.TestAddress("projectAddr")
	fooBarWithpad := "gno.land/r/onbloc/foo*PAD*gno.land/r/onbloc/bar"
	tests := []struct {
		name          string
		projectName   string
		tokenPath     string
		projectOwner  std.Address
		totalAmount   uint64
		conditions    string
		ratios        string
		tier30Ratio   uint64
		tier90Ratio   uint64
		tier180Ratio  uint64
		startTime     uint64
		expectedId    string
		expectedError string
	}{
		{
			name:          "normal project creation",
			projectName:   "Obl Protocol",
			tokenPath:     oblPath,
			projectOwner:  projectAddr,
			totalAmount:   uint64(1_000_000_000),
			conditions:    fooBarWithpad,
			ratios:        "1*PAD*2",
			tier30Ratio:   uint64(10),
			tier90Ratio:   uint64(20),
			tier180Ratio:  uint64(70),
			startTime:     uint64(time.Now().Unix() + 10),
			expectedId:    "gno.land/r/onbloc/obl:124",
			expectedError: "[GNOSWAP-LAUNCHPAD-008] can not create same project in same block || project(gno.land/r/onbloc/obl:124) already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			std.TestSetRealm(adminRealm)
			obl.Approve(consts.LAUNCHPAD_ADDR, tt.totalAmount)
			std.TestSkipHeights(1)

			defer func() {
				if r := recover(); r != nil {
					if tt.expectedError == "" {
						t.Errorf("unexpected error: %v", r)
					} else {
						uassert.Equal(t, r.(string), tt.expectedError)
					}
				}
			}()

			projectId := CreateProject(
				tt.projectName,
				tt.tokenPath,
				tt.projectOwner,
				tt.totalAmount,
				tt.conditions,
				tt.ratios,
				tt.tier30Ratio,
				tt.tier90Ratio,
				tt.tier180Ratio,
				tt.startTime,
			)

			if tt.expectedError == "" {
				uassert.Equal(t, projectId, tt.expectedId)
			}
			std.TestSkipHeights(1)
		})
	}
}

func TestMakeConditions(t *testing.T) {
	t.Skip("fix this test ")
	tests := []struct {
		name             string
		conditionsToken  string
		conditionsAmount string
		wantLen          int
		wantPanic        bool
		wantAmounts      map[string]uint64
	}{
		{
			name:             "normal multiple input",
			conditionsToken:  "gno.land/r/gnoswap/v1/gns*PAD*gno.land/r/onbloc/foo*PAD*gno.land/r/onbloc/bar",
			conditionsAmount: "100*PAD*200*PAD*300",
			wantLen:          3,
			wantPanic:        false,
			wantAmounts: map[string]uint64{
				"gno.land/r/gnoswap/v1/gns": 100,
				"gno.land/r/onbloc/foo":     200,
				"gno.land/r/onbloc/bar":     300,
			},
		},
		{
			name:             "empty input",
			conditionsToken:  "",
			conditionsAmount: "",
			wantLen:          0,
			wantPanic:        true,
			wantAmounts:      nil,
		},
		{
			name:             "single condition",
			conditionsToken:  "gno.land/r/gnoswap/v1/gns",
			conditionsAmount: "100",
			wantLen:          1,
			wantPanic:        false,
			wantAmounts: map[string]uint64{
				"gno.land/r/gnoswap/v1/gns": 100,
			},
		},
		{
			name:             "invalid amount format",
			conditionsToken:  "token1",
			conditionsAmount: "invalid",
			wantLen:          0,
			wantPanic:        true,
			wantAmounts:      nil,
		},
		{
			name:             "token and amount count mismatch",
			conditionsToken:  "token1*PAD*token2",
			conditionsAmount: "100",
			wantLen:          0,
			wantPanic:        true,
			wantAmounts:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := ProjectInput{
				Name:             tt.name,
				TokenPath:        "gno.land/r/gnoswap/v1/gns",
				Recipient:        std.Address(consts.ADMIN),
				DepositAmount:    1000000,
				ConditionsToken:  tt.conditionsToken,
				ConditionsAmount: tt.conditionsAmount,
				Tier30Ratio:      20,
				Tier90Ratio:      30,
				Tier180Ratio:     50,
				StartTime:        uint64(time.Now().Unix() + 100),
			}
			got, err := input.ParseConditions()
			if !tt.wantPanic {
				if tt.wantAmounts == nil {
					if got != nil {
						t.Errorf("makeConditions() = %v, want nil", got)
					}
					return
				}

				count := 0
				for range got {
					count++
				}
				uassert.Equal(t, count, tt.wantLen)

				for token, wantAmount := range tt.wantAmounts {
					condition, exists := got[token]
					uassert.True(t, exists)
					uassert.Equal(t, condition.minAmount, wantAmount)
				}
			} else {
				uassert.Error(t, err)
			}
		})
	}
}

func TestCalculateLeftReward(t *testing.T) {
	tests := []struct {
		name        string
		project     Project
		want        uint64
		description string
	}{
		{
			name: "all tiers have remaining rewards",
			project: Project{
				tiers: map[uint64]Tier{
					30: {
						tierAmount:       1000,
						calculatedAmount: 400,
					},
					90: {
						tierAmount:       2000,
						calculatedAmount: 800,
					},
					180: {
						tierAmount:       3000,
						calculatedAmount: 1200,
					},
				},
			},
			want:        3600, // (1000-400) + (2000-800) + (3000-1200)
			description: "Should sum up all remaining rewards from each tier",
		},
		{
			name: "some tiers fully claimed",
			project: Project{
				tiers: map[uint64]Tier{
					30: {
						tierAmount:       1000,
						calculatedAmount: 1000,
					},
					90: {
						tierAmount:       2000,
						calculatedAmount: 800,
					},
					180: {
						tierAmount:       3000,
						calculatedAmount: 3000,
					},
				},
			},
			want:        1200, // (1000-1000) + (2000-800) + (3000-3000)
			description: "Should handle fully claimed tiers correctly",
		},
		{
			name: "all tiers fully claimed",
			project: Project{
				tiers: map[uint64]Tier{
					30: {
						tierAmount:       1000,
						calculatedAmount: 1000,
					},
					90: {
						tierAmount:       2000,
						calculatedAmount: 2000,
					},
					180: {
						tierAmount:       3000,
						calculatedAmount: 3000,
					},
				},
			},
			want:        0,
			description: "Should return 0 when all rewards are claimed",
		},
		{
			name: "no rewards calculated yet",
			project: Project{
				tiers: map[uint64]Tier{
					30: {
						tierAmount:       1000,
						calculatedAmount: 0,
					},
					90: {
						tierAmount:       2000,
						calculatedAmount: 0,
					},
					180: {
						tierAmount:       3000,
						calculatedAmount: 0,
					},
				},
			},
			want:        6000, // 1000 + 2000 + 3000
			description: "Should return total tier amounts when no rewards calculated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateLeftReward(tt.project)
			uassert.Equal(t, got, tt.want)
		})
	}
}

func TestCreateTier(t *testing.T) {
	// Constants for test
	const (
		AVG_BLOCK_TIME_MS = 2000
		TIMESTAMP_DAY     = 86400
	)

	tests := []struct {
		name                string
		projectId           string
		duration            uint64
		amount              uint64
		startHeight         uint64
		startTime           uint64
		collectWaitDuration uint64
		want                Tier
		description         string
	}{
		{
			name:                "30-day tier creation",
			projectId:           "project1",
			duration:            30,
			amount:              1000,
			startHeight:         1000,
			startTime:           1600000000,
			collectWaitDuration: 100,
			want: Tier{
				id:                  "project1:30",
				collectWaitDuration: 100,
				tierAmount:          1000,
				ended: TimeInfo{
					height: 1000 + (30 * TIMESTAMP_DAY * 1000 / AVG_BLOCK_TIME_MS),
					time:   1600000000 + (30 * TIMESTAMP_DAY),
				},
			},
			description: "Should create a 30-day tier with correct end time and height",
		},
		{
			name:                "90-day tier creation",
			projectId:           "project2",
			duration:            90,
			amount:              2000,
			startHeight:         2000,
			startTime:           1700000000,
			collectWaitDuration: 200,
			want: Tier{
				id:                  "project2:90",
				collectWaitDuration: 200,
				tierAmount:          2000,
				ended: TimeInfo{
					height: 2000 + (90 * TIMESTAMP_DAY * 1000 / AVG_BLOCK_TIME_MS),
					time:   1700000000 + (90 * TIMESTAMP_DAY),
				},
			},
			description: "Should create a 90-day tier with correct end time and height",
		},
		{
			name:                "180-day tier creation",
			projectId:           "project3",
			duration:            180,
			amount:              3000,
			startHeight:         3000,
			startTime:           1800000000,
			collectWaitDuration: 300,
			want: Tier{
				id:                  "project3:180",
				collectWaitDuration: 300,
				tierAmount:          3000,
				ended: TimeInfo{
					height: 3000 + (180 * TIMESTAMP_DAY * 1000 / AVG_BLOCK_TIME_MS),
					time:   1800000000 + (180 * TIMESTAMP_DAY),
				},
			},
			description: "Should create a 180-day tier with correct end time and height",
		},
		{
			name:                "zero duration tier",
			projectId:           "project4",
			duration:            0,
			amount:              1000,
			startHeight:         4000,
			startTime:           1900000000,
			collectWaitDuration: 400,
			want: Tier{
				id:                  "project4:0",
				collectWaitDuration: 400,
				tierAmount:          1000,
				ended: TimeInfo{
					height: 4000,
					time:   1900000000,
				},
			},
			description: "Should handle zero duration correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectTiersWithoutDeposit = make(map[string]bool)

			got := createTier(
				tt.projectId,
				tt.duration,
				tt.amount,
				tt.startHeight,
				tt.startTime,
				tt.collectWaitDuration,
			)

			uassert.Equal(t, got.id, tt.want.id)
			uassert.Equal(t, got.collectWaitDuration, tt.want.collectWaitDuration)
			uassert.Equal(t, got.tierAmount, tt.want.tierAmount)
			uassert.Equal(t, got.ended.height, tt.want.ended.height)
			uassert.Equal(t, got.ended.time, tt.want.ended.time)
			uassert.True(t, projectTiersWithoutDeposit[got.id])
		})
	}
}

func TestCalculateProjectParams(t *testing.T) {
	currentTime := uint64(time.Now().Unix())
	baseHeight := uint64(std.GetHeight())

	tests := []struct {
		name    string
		input   ProjectInput
		now     uint64
		want    *ProjectCalculationResult
		wantErr bool
	}{
		{
			name: "valid input with equal ratio distribution",
			input: ProjectInput{
				DepositAmount: 3000,
				Tier30Ratio:   33,
				Tier90Ratio:   33,
				Tier180Ratio:  34,
				StartTime:     currentTime + 1000, // 1000 seconds in future
			},
			now: currentTime,
			want: &ProjectCalculationResult{
				Tier30Amount:  990,              // 3000 * 33 / 100
				Tier90Amount:  990,              // 3000 * 33 / 100
				Tier180Amount: 1020,             // 3000 * 34 / 100
				StartHeight:   baseHeight + 500, // 1000 seconds = 500 blocks
			},
			wantErr: false,
		},
		{
			name: "valid input with uneven ratio distribution",
			input: ProjectInput{
				DepositAmount: 10000,
				Tier30Ratio:   20,
				Tier90Ratio:   30,
				Tier180Ratio:  50,
				StartTime:     currentTime + 2000, // 2000 seconds in future
			},
			now: currentTime,
			want: &ProjectCalculationResult{
				Tier30Amount:  2000,              // 10000 * 20 / 100
				Tier90Amount:  3000,              // 10000 * 30 / 100
				Tier180Amount: 5000,              // 10000 * 50 / 100
				StartHeight:   baseHeight + 1000, // 2000 seconds = 1000 blocks
			},
			wantErr: false,
		},
		{
			name: "start time in the past",
			input: ProjectInput{
				DepositAmount: 1000,
				Tier30Ratio:   30,
				Tier90Ratio:   30,
				Tier180Ratio:  40,
				StartTime:     currentTime - 1000,
			},
			now:     currentTime,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculateProjectParams(tt.input, tt.now)

			if tt.wantErr {
				uassert.Error(t, err)
				return
			}

			uassert.NoError(t, err)
			uassert.Equal(t, got.Tier30Amount, tt.want.Tier30Amount)
			uassert.Equal(t, got.Tier90Amount, tt.want.Tier90Amount)
			uassert.Equal(t, got.Tier180Amount, tt.want.Tier180Amount)
			uassert.Equal(t, got.StartHeight, tt.want.StartHeight)
		})
	}
}

func TestGenerateProjectId(t *testing.T) {
	setupTest(t)

	currentHeight := std.GetHeight()

	tests := []struct {
		name       string
		tokenPath  string
		mockHeight int64
		want       string
	}{
		{
			name:       "normal token path",
			tokenPath:  "gno.land/r/gnoswap/gns",
			mockHeight: currentHeight + 10,
			want:       "",
		},
		{
			name:       "another token path",
			tokenPath:  "gno.land/r/onbloc/obl",
			mockHeight: currentHeight + 5,
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			heightBefore := std.GetHeight()
			skip := tt.mockHeight - heightBefore
			if skip > 0 {
				std.TestSkipHeights(skip)
			}

			got := generateProjectId(tt.tokenPath)

			want := ufmt.Sprintf("%s:%d", tt.tokenPath, std.GetHeight())

			uassert.Equal(t, got, want)
		})
	}
}

func TestGenerateTierId(t *testing.T) {
	setupTest(t)

	tests := []struct {
		name      string
		projectId string
		duration  uint64
		want      string
	}{
		{
			name:      "30-day tier",
			projectId: "gno.land/r/gnoswap/gns:100",
			duration:  30,
			want:      "gno.land/r/gnoswap/gns:100:30",
		},
		{
			name:      "90-day tier",
			projectId: "gno.land/r/onbloc/obl:999",
			duration:  90,
			want:      "gno.land/r/onbloc/obl:999:90",
		},
		{
			name:      "180-day tier",
			projectId: "myProject:1234",
			duration:  180,
			want:      "myProject:1234:180",
		},
		{
			name:      "zero-day tier",
			projectId: "testProj:555",
			duration:  0,
			want:      "testProj:555:0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateTierId(tt.projectId, tt.duration)
			uassert.Equal(t, got, tt.want)
		})
	}
}

func TestCovertTimeToHeight(t *testing.T) {
	setupTest(t)

	tests := []struct {
		name      string
		timestamp uint64
		want      uint64
	}{
		{
			name:      "3600 seconds => 1800 blocks",
			timestamp: 3600, // 1hour
			want:      1800, // 3600 * 1000 / 2000
		},
		{
			name:      "0 seconds => 0 blocks",
			timestamp: 0,
			want:      0,
		},
		{
			name:      "small seconds => 1 block",
			timestamp: 2, // 2 seconds
			want:      1, // 2 * 1000 / 2000 = 1
		},
		{
			name:      "larger value",
			timestamp: 86400, // 1day (seconds)
			want:      43200, // 86400 * 1000 / 2000
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertTimeToHeight(tt.timestamp)
			uassert.Equal(t, got, tt.want)
		})
	}

	// division by zero test
	t.Run("blocktime is zero => panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic but got none")
			} else {
				uassert.Equal(t, r.(string), "division by zero")
			}
		}()

		std.TestSetRealm(adminRealm)
		gns.SetAvgBlockTimeInMsByAdmin(0)
		_ = convertTimeToHeight(1000)
	})
}

func TestTransferLeftFromProjectByAdmin(t *testing.T) {
	setupTest(t)

	validRecipient := testutils.TestAddress("recipient")
	invalidRecipient := std.Address("invalidRecipient")

	projectId := "dummyProject:100"
	dummyProject := Project{
		id:        projectId,
		name:      "Dummy Project",
		tokenPath: oblPath,
		tiers: map[uint64]Tier{
			30:  {tierAmount: 1000, userCollectedAmount: 400},
			90:  {tierAmount: 2000, userCollectedAmount: 2000},
			180: {tierAmount: 3000, userCollectedAmount: 2500},
		},
		ended: TimeInfo{
			height: 120,
			time:   uint64(time.Now().Unix() - 1),
		},
	}

	tests := []struct {
		name          string
		setupFunc     func()
		projectId     string
		recipient     std.Address
		mockHeight    uint64
		expectedPanic string
	}{
		{
			name: "project not found",
			setupFunc: func() {
				projects = make(map[string]Project)
			},
			projectId:     "nonExistentProject",
			recipient:     validRecipient,
			mockHeight:    200,
			expectedPanic: "[GNOSWAP-LAUNCHPAD-003] requested data not found || projectId(nonExistentProject) not found",
		},
		{
			name: "invalid recipient address",
			setupFunc: func() {
				projects = make(map[string]Project)
				projects[projectId] = dummyProject
			},
			projectId:     projectId,
			recipient:     invalidRecipient,
			mockHeight:    200,
			expectedPanic: "[GNOSWAP-LAUNCHPAD-013] invalid transfer condition || invalid recipient address(invalidRecipient)",
		},
		{
			name: "project not ended yet",
			setupFunc: func() {
				projects = make(map[string]Project)
				p := dummyProject
				p.ended.height = 300
				projects[projectId] = p
			},
			projectId:     projectId,
			recipient:     validRecipient,
			mockHeight:    250,
			expectedPanic: "[GNOSWAP-LAUNCHPAD-013] invalid transfer condition || project not ended yet(current:250, endHeight: 300)",
		},
		{
			name: "project already refunded",
			setupFunc: func() {
				projects = make(map[string]Project)
				p := dummyProject
				p.refund = RefundInfo{
					amount: 10,
					height: 110,
					time:   uint64(time.Now().Unix()),
				}
				projects[projectId] = p
			},
			projectId:     projectId,
			recipient:     validRecipient,
			mockHeight:    200,
			expectedPanic: "[GNOSWAP-LAUNCHPAD-013] invalid transfer condition || project already refunded(height:110)",
		},
		{
			name: "success leftover transfer",
			setupFunc: func() {
				projects = make(map[string]Project)
				p := dummyProject
				p.ended.height = 120
				p.depositAmount = 6000
				p.stats.totalCollected = 4900
				p.refund = RefundInfo{}
				projects[projectId] = p
			},
			projectId:     projectId,
			recipient:     validRecipient,
			mockHeight:    250,
			expectedPanic: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			std.TestSetRealm(adminRealm)
			// set up
			tt.setupFunc()

			// control block height
			currHeight := std.GetHeight()
			skip := int64(tt.mockHeight) - int64(currHeight)
			if skip > 0 {
				std.TestSkipHeights(skip)
			}

			defer func() {
				if r := recover(); r != nil {
					if tt.expectedPanic == "" {
						t.Errorf("unexpected panic: %v", r)
					} else {
						uassert.Equal(t, r.(string), tt.expectedPanic)
					}
				} else if tt.expectedPanic != "" {
					t.Errorf("expected panic(%s) but got none", tt.expectedPanic)
				}
			}()

			// run
			got := TransferLeftFromProjectByAdmin(tt.projectId, tt.recipient)

			if tt.expectedPanic == "" {
				project := projects[tt.projectId]
				// leftover = (1000-400) + (2000-2000) + (3000-2500) = 1100
				uassert.Equal(t, got, uint64(1100))
				uassert.Equal(t, project.refund.amount, uint64(1100))
				uassert.Equal(t, project.refund.height, tt.mockHeight)
			}
		})
	}
}
