package launchpad

import (
	"std"
	"testing"
	"time"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"
	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/gns"
	"gno.land/r/gnoswap/v1/gov/xgns"
)

func setupTestDeposit(t *testing.T) (*MockTokenTeller, std.Address) {
	t.Helper()
	projects = make(map[string]Project)
	deposits = make(map[string]Deposit)
	depositsByProject = make(map[string]map[string][]string)
	depositsByUser = make(map[std.Address][]string)
	depositsByUserByProject = make(map[std.Address]map[string][]string)

	mockTeller := &MockTokenTeller{balance: 1000}
	testAddr := testutils.TestAddress("test")

	return mockTeller, testAddr
}

func createTestProject(t *testing.T) Project {
	now := uint64(time.Now().Unix())
	height := uint64(std.ChainHeight())

	tiers := make(map[uint64]Tier)
	tiersRatios := make(map[uint64]uint64)

	tiers[30] = Tier{
		id:                  ufmt.Sprintf("gno.land/r/gnoswap/v1/gns:%d:30", height),
		collectWaitDuration: TIMESTAMP_3DAYS * 1000 / 1000,
		tierAmount:          300,
		started: TimeInfo{
			height: height,
			time:   now,
		},
		ended: TimeInfo{
			height: height + (30 * 24 * 60 * 60),
			time:   now + (30 * 24 * 60 * 60),
		},
		reward: *NewReward(u256.Zero(), 0, 0),
	}
	tiersRatios[30] = 30

	tiers[90] = Tier{
		id:                  ufmt.Sprintf("gno.land/r/gnoswap/v1/gns:%d:90", height),
		collectWaitDuration: TIMESTAMP_7DAYS * 1000 / 1000,
		tierAmount:          300,
		started: TimeInfo{
			height: height,
			time:   now,
		},
		ended: TimeInfo{
			height: height + (90 * 24 * 60 * 60),
			time:   now + (90 * 24 * 60 * 60),
		},
		reward: *NewReward(u256.Zero(), 0, 0),
	}
	tiersRatios[90] = 30

	tiers[180] = Tier{
		id:                  ufmt.Sprintf("gno.land/r/gnoswap/v1/gns:%d:180", height),
		collectWaitDuration: TIMESTAMP_14DAYS * 1000 / 1000,
		tierAmount:          400,
		started: TimeInfo{
			height: height,
			time:   now,
		},
		ended: TimeInfo{
			height: height + (180 * 24 * 60 * 60),
			time:   now + (180 * 24 * 60 * 60),
		},
		reward: *NewReward(u256.Zero(), 0, 0),
	}
	tiersRatios[180] = 40

	return Project{
		id:            ufmt.Sprintf("gno.land/r/gnoswap/v1/gns:%d", height),
		name:          "Test Project",
		tokenPath:     "gno.land/r/gnoswap/v1/gns",
		depositAmount: 1000,
		recipient:     testutils.TestAddress("recipient"),
		conditions:    make(map[string]Condition),
		tiers:         tiers,
		tiersRatios:   tiersRatios,
		created: TimeInfo{
			height: height,
			time:   now,
		},
		started: TimeInfo{
			height: height,
			time:   now,
		},
		ended: TimeInfo{
			height: height + (180 * 24 * 60 * 60),
			time:   now + (180 * 24 * 60 * 60),
		},
		stats: ProjectStats{
			totalDeposit:      0,
			actualDeposit:     0,
			totalParticipant:  0,
			actualParticipant: 0,
			totalCollected:    0,
		},
		refund: RefundInfo{
			amount: 0,
			height: 0,
			time:   0,
		},
	}
}

func TestValidateProjectTier(t *testing.T) {
	mockTeller, _ := setupTestDeposit(t)
	project := createTestProject(t)
	projects[project.id] = project

	tests := []struct {
		name        string
		projectId   string
		tierStr     string
		shouldError bool
	}{
		{
			name:        "Valid project and tier",
			projectId:   project.id,
			tierStr:     "30",
			shouldError: false,
		},
		{
			name:        "Invalid project id",
			projectId:   "nonexistent",
			tierStr:     "30",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := validateProjectTier(tt.projectId, tt.tierStr)
			switch tt.shouldError {
			case true:
				uassert.Error(t, err)
			case false:
				uassert.NoError(t, err)
			}
		})
	}
}

func TestCreateDeposit(t *testing.T) {
	_, testAddr := setupTestDeposit(t)
	project := createTestProject(t)
	now := uint64(time.Now().Unix())
	height := uint64(std.ChainHeight())

	info := ProjectTierInfo{
		Project:       project,
		Tier:          project.tiers[30],
		TierType:      "30",
		CurrentHeight: height,
		CurrentTime:   now,
	}

	deposit, err := createDeposit(info, 100)
	if err != nil {
		t.Fatalf("Failed to create deposit: %v", err)
	}

	uassert.Equal(t, deposit.amount, uint64(100))
	uassert.Equal(t, deposit.depositHeight, height)
	uassert.Equal(t, deposit.depositTime, now)
	uassert.Equal(t, deposit.projectId, project.id)
}

func TestCalculateClaimableTimes(t *testing.T) {
	tests := []struct {
		name       string
		info       ProjectTierInfo
		wantHeight uint64
		wantTime   uint64
	}{
		{
			name: "Tier 30 - Normal Case",
			info: ProjectTierInfo{
				Project: Project{},
				Tier: Tier{
					collectWaitDuration: 100,
					ended: TimeInfo{
						height: 1000,
						time:   1000000,
					},
				},
				TierType:      "30",
				CurrentHeight: 500,
				CurrentTime:   500000,
			},
			wantHeight: 600,    // CurrentHeight(500) + collectWaitDuration(100)
			wantTime:   759200, // CurrentTime(500000) + TIMESTAMP_3DAYS
		},
		{
			name: "Tier 90 - Exceeds End CurrentHeight/Time",
			info: ProjectTierInfo{
				Project: Project{},
				Tier: Tier{
					collectWaitDuration: 200,
					ended: TimeInfo{
						height: 550,
						time:   600000,
					},
				},
				TierType:      "90",
				CurrentHeight: 400,
				CurrentTime:   400000,
			},
			wantHeight: 550,    // min(400+200, endHeight(550))
			wantTime:   600000, // min(400000+TIMESTAMP_7DAYS, endTime(600000))
		},
		{
			name: "Tier 180 - Early Stage",
			info: ProjectTierInfo{
				Project: Project{},
				Tier: Tier{
					collectWaitDuration: 300,
					ended: TimeInfo{
						height: 2000,
						time:   2000000,
					},
				},
				TierType:      "180",
				CurrentHeight: 100,
				CurrentTime:   100000,
			},
			wantHeight: 400,     // CurrentHeight(100) + collectWaitDuration(300)
			wantTime:   1309600, // CurrentTime(100000) + TIMESTAMP_14DAYS
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHeight, gotTime := calculateClaimableTimes(tt.info)
			uassert.Equal(t, gotHeight, tt.wantHeight)
			uassert.Equal(t, gotTime, tt.wantTime)
		})
	}
}

func TestCollectDeposit(t *testing.T) {
	t.Skip("fix this test")
	_, testAddr := setupTestDeposit(t)
	project := createTestProject(t)
	projects[project.id] = project
	currentTime := uint64(time.Now().Unix())
	height := uint64(std.ChainHeight())

	// Create test deposit
	info := ProjectTierInfo{
		Project:       project,
		Tier:          project.tiers[30],
		TierType:      "30",
		CurrentHeight: height,
		CurrentTime:   currentTime,
	}

	deposit, err := createDeposit(info, 100)
	if err != nil {
		t.Fatalf("Failed to create deposit: %v", err)
	}

	// Add deposit to indices
	deposits[deposit.id] = deposit
	updateDepositIndices(deposit, &DepositState{
		Deposits:              make(map[string]Deposit),
		DepositsByProject:     depositsByProject,
		DepositsByUser:        depositsByUser,
		DepositsByUserProject: depositsByUserByProject,
	})

	t.Run("Cannot collect before claimable height", func(t *testing.T) {
		std.TestSetRealm(std.NewUserRealm(testAddr))
		amount, _ := processDepositCollection([]string{deposit.id}, "")
		uassert.Equal(t, uint64(0), amount)
	})

	t.Run("Can collect after claimable height", func(t *testing.T) {
		// Skip to after claimable height
		std.TestSkipHeights(int64(deposit.claimableHeight - height + 1))

		std.TestSetRealm(std.NewUserRealm(testAddr))
		amount, errCode := processDepositCollection([]string{deposit.id}, "")
		uassert.NoError(t, errCode)
		uassert.Equal(t, uint64(100), amount)
	})
}

func TestDepositGns(t *testing.T) {
	_, testAddr := setupTestDeposit(t)
	project := createTestProject(t)
	projects[project.id] = project

	t.Run("Fail with insufficient balance", func(t *testing.T) {
		std.TestSetRealm(std.NewUserRealm(testAddr))

		defer func() {
			r := recover()
			if r != nil {
				uassert.Equal(t, "[GNOSWAP-LAUNCHPAD-017] invalid amount || amount(1000) should greater than minimum deposit amount(1000000)", r.(string))
			}
		}()

		DepositGns(project.id+":30", 1000, "")
		std.TestSkipHeights(1)
	})

	t.Run("Success with sufficient balance", func(t *testing.T) {
		// transfer enough gns
		std.TestSetRealm(adminRealm)
		gns.Transfer(testAddr, 1000000)

		std.TestSetRealm(std.NewUserRealm(testAddr))
		gns.Approve(consts.LAUNCHPAD_ADDR, 1000000)

		std.TestSkipHeights(10)
		depositId := DepositGns(project.id+":30", 1000000, "")
		deposit := deposits[depositId]

		uassert.Equal(t, deposit.amount, uint64(1000000))
		uassert.Equal(t, deposit.depositor, testAddr)
	})
}

func TestUpdateDepositIndices(t *testing.T) {
	_, testAddr := setupTestDeposit(t)
	project := createTestProject(t)

	deposit := Deposit{
		id:        "test_deposit_1",
		projectId: project.id,
		tier:      "30",
		depositor: testAddr,
		amount:    100,
	}

	state := &DepositState{
		Deposits:              make(map[string]Deposit),
		DepositsByProject:     make(map[string]map[string][]string),
		DepositsByUser:        make(map[std.Address][]string),
		DepositsByUserProject: make(map[std.Address]map[string][]string),
	}

	updateDepositIndices(deposit, state)

	// Check depositsByUser
	userDeposits := state.DepositsByUser[testAddr]
	uassert.Equal(t, 1, len(userDeposits))
	uassert.Equal(t, deposit.id, userDeposits[0])

	// Check depositsByProject
	projectDeposits := state.DepositsByProject[project.id][deposit.tier]
	uassert.Equal(t, 1, len(projectDeposits))
	uassert.Equal(t, deposit.id, projectDeposits[0])

	// Check depositsByUserProject
	userProjectDeposits := state.DepositsByUserProject[testAddr][project.id]
	uassert.Equal(t, 1, len(userProjectDeposits))
	uassert.Equal(t, deposit.id, userProjectDeposits[0])
}

func TestMultipleDeposits(t *testing.T) {
	_, testAddr := setupTestDeposit(t)
	project := createTestProject(t)
	projects[project.id] = project

	state := &DepositState{
		Deposits:              make(map[string]Deposit),
		DepositsByProject:     depositsByProject,
		DepositsByUser:        depositsByUser,
		DepositsByUserProject: depositsByUserByProject,
	}

	// Create multiple deposits
	for i := 0; i < 3; i++ {
		info := ProjectTierInfo{
			Project:       project,
			Tier:          project.tiers[30],
			TierType:      "30",
			CurrentHeight: uint64(std.ChainHeight()),
			CurrentTime:   uint64(time.Now().Unix()),
		}

		std.TestSetRealm(std.NewUserRealm(testAddr))
		deposit, _ := createDeposit(info, 100)
		deposits[deposit.id] = deposit
		updateDepositIndices(deposit, state)

		std.TestSkipHeights(100) // Skip some blocks between deposits
	}

	// Verify indices are correct
	userDeposits := state.DepositsByUser[testAddr]
	uassert.Equal(t, 3, len(userDeposits))

	projectDeposits := state.DepositsByProject[project.id]["30"]
	uassert.Equal(t, 3, len(projectDeposits))
}

func TestUpdateDepositIndices_EmptyState(t *testing.T) {
	_, testAddr := setupTestDeposit(t)
	deposit := Deposit{
		id:        "test_deposit",
		projectId: "test_project",
		tier:      "30",
		depositor: testAddr,
		amount:    100,
	}

	state := &DepositState{
		Deposits:              make(map[string]Deposit),
		DepositsByProject:     make(map[string]map[string][]string),
		DepositsByUser:        make(map[std.Address][]string),
		DepositsByUserProject: make(map[std.Address]map[string][]string),
	}

	updateDepositIndices(deposit, state)

	uassert.NotEqual(t, nil, state.DepositsByUser)
	uassert.NotEqual(t, nil, state.DepositsByProject)
	uassert.NotEqual(t, nil, state.DepositsByUserProject)

	uassert.Equal(t, 1, len(state.DepositsByUser[testAddr]))
	uassert.Equal(t, deposit.id, state.DepositsByUser[testAddr][0])
}

func TestIsProjectActive(t *testing.T) {
	project := Project{
		started: TimeInfo{height: 100},
		ended:   TimeInfo{height: 200},
	}

	tests := []struct {
		name           string
		height         uint64
		shouldBeActive bool
	}{
		{
			name:           "Before start",
			height:         50,
			shouldBeActive: false,
		},
		{
			name:           "At start",
			height:         100,
			shouldBeActive: true,
		},
		{
			name:           "During active period",
			height:         150,
			shouldBeActive: true,
		},
		{
			name:           "At end",
			height:         200,
			shouldBeActive: true,
		},
		{
			name:           "After end",
			height:         250,
			shouldBeActive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			active := isProjectActive(project, tt.height)
			uassert.Equal(t, tt.shouldBeActive, active)
		})
	}
}

func TestIsTierActive(t *testing.T) {
	tier := Tier{
		ended: TimeInfo{height: 200},
	}

	tests := []struct {
		name           string
		height         uint64
		shouldBeActive bool
	}{
		{
			name:           "Before end",
			height:         150,
			shouldBeActive: true,
		},
		{
			name:           "At end",
			height:         200,
			shouldBeActive: true,
		},
		{
			name:           "After end",
			height:         250,
			shouldBeActive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			active := isTierActive(tier, tt.height)
			uassert.Equal(t, tt.shouldBeActive, active)
		})
	}
}

// Mock for GRC20 Token and Ledger
type MockPrivateLedger struct {
	burnCalled bool
	burnAmount uint64
	burnFrom   std.Address
	mintCalled bool
	mintAmount uint64
	mintTo     std.Address
}

func (m *MockPrivateLedger) Burn(from std.Address, amount uint64) error {
	m.burnCalled = true
	m.burnFrom = from
	m.burnAmount = amount
	return nil
}

func (m *MockPrivateLedger) Mint(to std.Address, amount uint64) error {
	m.mintCalled = true
	m.mintTo = to
	m.mintAmount = amount
	return nil
}

type MockToken struct {
	balances map[string]uint64
}

func NewMockToken() *MockToken {
	return &MockToken{
		balances: make(map[string]uint64),
	}
}

func (m *MockToken) BalanceOf(addr std.Address) uint64 {
	return m.balances[addr.String()]
}

// Mock for emission
type MockEmission struct {
	mintCalled bool
}

func (m *MockEmission) MintAndDistributeGns() {
	m.mintCalled = true
}

type MockCommon struct {
	haltCheckCalled bool
}

func (m *MockCommon) IsHalted() {
	m.haltCheckCalled = true
}

func TestProcessCollectedDeposits_TransactionFail(t *testing.T) {
	t.Skip("fix this test")
	mockXGNS, _ := setupXGNSTest(t)
	project := createTestProject(t)
	projects[project.id] = project

	info := ProjectTierInfo{
		Project:       project,
		Tier:          project.tiers[30],
		TierType:      "30",
		CurrentHeight: uint64(std.ChainHeight()),
		CurrentTime:   uint64(time.Now().Unix()),
	}

	deposit, _ := createDeposit(info, 1000)
	deposits[deposit.id] = deposit

	xgnsMint(t, mockXGNS, consts.LAUNCHPAD_ADDR, 500)

	std.TestSetRealm(std.NewCodeRealm(consts.LAUNCHPAD_PATH))

	processCollectedDeposits([]string{deposit.id}, "")
}

func TestProcessCollectedDeposits_BalanceChanges(t *testing.T) {
	t.Skip("fix this test")
	mockXGNS, _ := setupXGNSTest(t)
	project := createTestProject(t)
	projects[project.id] = project

	depositAmount := uint64(1000)
	xgnsMint(t, mockXGNS, consts.LAUNCHPAD_ADDR, depositAmount)

	info := ProjectTierInfo{
		Project:       project,
		Tier:          project.tiers[30],
		TierType:      "30",
		CurrentHeight: uint64(std.ChainHeight()),
		CurrentTime:   uint64(time.Now().Unix()),
	}

	deposit, _ := createDeposit(info, depositAmount)
	deposits[deposit.id] = deposit

	std.TestSkipHeights(int64(deposit.claimableHeight - uint64(std.ChainHeight()) + 1))

	std.TestSetRealm(std.NewCodeRealm(consts.LAUNCHPAD_PATH))
	amount := processCollectedDeposits([]string{deposit.id}, "")

	xgnsCheckTotalSupply(t, mockXGNS, 1000)
}

func TestCollectDepositGnsByDepositId_ClaimableCheck(t *testing.T) {
	_, depositor := setupTestDeposit(t)
	std.TestSkipHeights(10)
	project := createTestProject(t)
	projects[project.id] = project

	// TierInfo
	now := uint64(time.Now().Unix())
	height := uint64(std.ChainHeight())
	info := ProjectTierInfo{
		Project:       project,
		Tier:          project.tiers[30],
		TierType:      "30",
		CurrentHeight: height,
		CurrentTime:   now,
	}

	std.TestSkipHeights(10)
	deposit, err := createDeposit(info, 100)
	if err != nil {
		t.Fatalf("failed create deposit: %v", err)
	}

	updateDepositIndices(deposit, &DepositState{
		Deposits:              deposits,
		DepositsByProject:     depositsByProject,
		DepositsByUser:        depositsByUser,
		DepositsByUserProject: depositsByUserByProject,
	})

	t.Run("Collect early => Should panic", func(t *testing.T) {
		std.TestSetRealm(std.NewUserRealm(depositor))

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic but did not occur")
			}
		}()
		CollectDepositGnsByDepositId(deposit.id)
	})
}

func TestCollectDepositGnsByDepositId_OtherOwner(t *testing.T) {
	_, userA := setupTestDeposit(t)
	userB := testutils.TestAddress("userB")

	project := createTestProject(t)
	projects[project.id] = project

	// create deposit (owner = userA)
	now := uint64(time.Now().Unix())
	height := uint64(std.ChainHeight())
	info := ProjectTierInfo{
		Project:       project,
		Tier:          project.tiers[30],
		TierType:      "30",
		CurrentHeight: height,
		CurrentTime:   now,
	}

	deposit, err := createDeposit(info, 100)
	if err != nil {
		t.Fatalf("failed to create deposit: %v", err)
	}

	updateDepositIndices(deposit, &DepositState{
		Deposits:              deposits,
		DepositsByProject:     depositsByProject,
		DepositsByUser:        depositsByUser,
		DepositsByUserProject: depositsByUserByProject,
	})

	t.Run("UserB tries to collect userA's deposit => panic expected", func(t *testing.T) {
		std.TestSetRealm(std.NewUserRealm(userB))
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic but did not occur")
			} else {
				uassert.Equal(t, "[GNOSWAP-LAUNCHPAD-007] invalid input data || user has no deposits in any project", r.(string))
			}
		}()
		CollectDepositGnsByDepositId(deposit.id)
	})
}

func TestCollectDepositGnsByDepositId_ExactClaimable(t *testing.T) {
	_, userA := setupTestDeposit(t)
	project := createTestProject(t)
	projects[project.id] = project

	now := uint64(time.Now().Unix())
	initialHeight := uint64(std.ChainHeight())

	info := ProjectTierInfo{
		Project:       project,
		Tier:          project.tiers[30],
		TierType:      "30",
		CurrentHeight: initialHeight,
		CurrentTime:   now,
	}

	std.TestSetRealm(std.NewUserRealm(userA))
	// deposit
	deposit, err := createDeposit(info, 200)
	if err != nil {
		t.Fatalf("failed create deposit: %v", err)
	}
	deposit.claimableHeight = initialHeight + 10
	deposit.claimableTime = now + 60

	updateDepositIndices(deposit, &DepositState{
		Deposits:              deposits,
		DepositsByProject:     depositsByProject,
		DepositsByUser:        depositsByUser,
		DepositsByUserProject: depositsByUserByProject,
	})

	std.TestSetRealm(std.NewCodeRealm(consts.EMISSION_PATH))
	std.TestSkipHeights(int64(10))
	gns.MintGns(consts.LAUNCHPAD_ADDR)
	std.TestSkipHeights(int64(10))

	std.TestSetRealm(std.NewUserRealm(userA))
	gns.Approve(consts.LAUNCHPAD_ADDR, 1000)
	token, _ := setupXGNSTest(t)

	std.TestSetRealm(std.NewCodeRealm(consts.LAUNCHPAD_PATH))
	std.TestSetOriginCaller(std.Address(consts.LAUNCHPAD_PATH))
	xgnsMint(t, token, consts.LAUNCHPAD_ADDR, 1000)

	std.TestSetRealm(std.NewUserRealm(consts.LAUNCHPAD_ADDR))
	xgns.Mint(
		consts.LAUNCHPAD_ADDR,
		1000,
	)

	skipCount := deposit.claimableHeight - uint64(std.ChainHeight())
	if skipCount > 0 {
		std.TestSkipHeights(int64(skipCount))
	}

	std.TestSetRealm(std.NewUserRealm(userA))
	collected := CollectDepositGnsByDepositId(deposit.id)
	uassert.Equal(t, uint64(200), collected, "should collect the deposit fully after claimableHeight/time")
}
