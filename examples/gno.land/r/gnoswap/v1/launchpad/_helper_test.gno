package launchpad

import (
	"std"
	"testing"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/testutils"
	"gno.land/p/gnoswap/consts"
)

var (
	admin     = consts.ADMIN
	adminAddr = consts.ADMIN
	adminUser = adminAddr

	fooPath string = "gno.land/r/onbloc/foo"
	barPath string = "gno.land/r/onbloc/bar"
	bazPath string = "gno.land/r/onbloc/baz"
	quxPath string = "gno.land/r/onbloc/qux"
	oblPath string = "gno.land/r/onbloc/obl"

	fee100  uint32 = 100
	fee500  uint32 = 500
	fee3000 uint32 = 3000

	maxApprove uint64 = 18446744073709551615
)

// Realms to mock frames
var (
	adminRealm = std.NewUserRealm(admin)
	posRealm   = std.NewCodeRealm(consts.POSITION_PATH)
	rouRealm   = std.NewCodeRealm(consts.ROUTER_PATH)
)

// MockXGNSToken implements basic functionality for testing xgns token
type MockXGNSToken struct {
	token  *grc20.Token
	ledger *grc20.PrivateLedger
}

func NewMockXGNSToken() *MockXGNSToken {
	token, ledger := grc20.NewToken("XGNS", "xGNS", 6)
	return &MockXGNSToken{
		token:  token,
		ledger: ledger,
	}
}

func (m *MockXGNSToken) TotalSupply() uint64 {
	return m.token.TotalSupply()
}

func (m *MockXGNSToken) VotingSupply() uint64 {
	total := m.token.TotalSupply()
	launchpad := m.token.BalanceOf(consts.LAUNCHPAD_ADDR)
	return total - launchpad
}

func (m *MockXGNSToken) BalanceOf(owner std.Address) uint64 {
	ownerAddr := owner
	return m.token.BalanceOf(ownerAddr)
}

func (m *MockXGNSToken) Mint(to std.Address, amount uint64) error {
	toAddr := to
	return m.ledger.Mint(toAddr, amount)
}

func (m *MockXGNSToken) Burn(from std.Address, amount uint64) error {
	fromAddr := from
	return m.ledger.Burn(fromAddr, amount)
}

// Helper functions for tests
func setupXGNSTest(t *testing.T) (*MockXGNSToken, std.Address) {
	t.Helper()
	mockToken := NewMockXGNSToken()
	testAddr := testutils.TestAddress("test")

	// Set up mock token with initial state
	return mockToken, testAddr
}

func xgnsMint(t *testing.T, token *MockXGNSToken, to std.Address, amount uint64) {
	t.Helper()
	// Set realm to gov/staker or launchpad to have permission
	std.TestSetRealm(std.NewCodeRealm(consts.LAUNCHPAD_PATH))
	err := token.Mint(to, amount)
	if err != nil {
		t.Fatalf("Failed to mint XGNS: %v", err)
	}
}

func xgnsBurn(t *testing.T, token *MockXGNSToken, from std.Address, amount uint64) {
	t.Helper()
	// Set realm to gov/staker or launchpad to have permission
	std.TestSetRealm(std.NewCodeRealm(consts.LAUNCHPAD_PATH))
	err := token.Burn(from, amount)
	if err != nil {
		t.Fatalf("Failed to burn XGNS: %v", err)
	}
}

func xgnsCheckBalance(t *testing.T, token *MockXGNSToken, owner std.Address, expected uint64) {
	t.Helper()
	balance := token.BalanceOf(owner)
	if balance != expected {
		t.Errorf("XGNS balance mismatch for %s: got %d, want %d", owner, balance, expected)
	}
}

func xgnsCheckTotalSupply(t *testing.T, token *MockXGNSToken, expected uint64) {
	t.Helper()
	supply := token.TotalSupply()
	if supply != expected {
		t.Errorf("XGNS total supply mismatch: got %d, want %d", supply, expected)
	}
}

func xgnsCheckVotingSupply(t *testing.T, token *MockXGNSToken, expected uint64) {
	t.Helper()
	supply := token.VotingSupply()
	if supply != expected {
		t.Errorf("XGNS voting supply mismatch: got %d, want %d", supply, expected)
	}
}

// Example test using helper functions
func TestXGNSHelpers(t *testing.T) {
	token, testAddr := setupXGNSTest(t)
	testUser := testAddr

	// Test minting
	xgnsMint(t, token, testUser, 1000)
	xgnsCheckBalance(t, token, testUser, 1000)
	xgnsCheckTotalSupply(t, token, 1000)

	// Test burning
	xgnsBurn(t, token, testUser, 500)
	xgnsCheckBalance(t, token, testUser, 500)
	xgnsCheckTotalSupply(t, token, 500)

	// Test Voting Supply
	launchpadAddr := consts.LAUNCHPAD_ADDR
	xgnsMint(t, token, launchpadAddr, 200)
	xgnsCheckVotingSupply(t, token, 500) // Only testUser's balance counts
}

// Mock data structs
type MockTokenTeller struct {
	balance uint64
	addr    std.Address
}

func (m *MockTokenTeller) TransferFrom(from std.Address, to std.Address, amount uint64) {
	m.balance = amount
	m.addr = to
}

func (m *MockTokenTeller) BalanceOf(addr std.Address) uint64 {
	return m.balance
}

// Additional helper that combines with existing TokenTeller mock
func createMockXGNSTeller() *MockTokenTeller {
	return &MockTokenTeller{
		balance: 1000,
		addr:    testutils.TestAddress("mock_xgns"),
	}
}

/* HELPER */
func ugnotBalanceOf(addr std.Address) uint64 {
	testBanker := std.GetBanker(std.BankerTypeRealmIssue)

	coins := testBanker.GetCoins(addr)
	if len(coins) == 0 {
		return 0
	}

	return uint64(coins.AmountOf("ugnot"))
}
