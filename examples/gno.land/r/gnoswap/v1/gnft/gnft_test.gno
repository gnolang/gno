package gnft

import (
	"std"
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"

	"gno.land/p/gnoswap/consts"
)

const (
	errInvalidTokenId = "invalid token id"
)

var (
	positionAddr  = consts.POSITION_ADDR
	positionRealm = std.NewCodeRealm(consts.POSITION_PATH)

	addr01      = testutils.TestAddress("addr01")
	addr01Realm = std.NewUserRealm(addr01)

	addr02      = testutils.TestAddress("addr02")
	addr02Realm = std.NewUserRealm(addr02)
)

func TestMetadata(t *testing.T) {
	tests := []struct {
		name     string
		fn       func() string
		expected string
	}{
		{"Name()", Name, "GNOSWAP NFT"},
		{"Symbol()", Symbol, "GNFT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uassert.Equal(t, tt.expected, tt.fn())
		})
	}
}

func TestTotalSupply(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		expected uint64
	}{
		{
			name:     "initial total supply",
			expected: uint64(0),
		},
		{
			name: "total supply after minting",
			setup: func() {
				std.TestSetRealm(positionRealm)
				Mint(addr01, tid(1))
				Mint(addr01, tid(2))
			},
			expected: uint64(2),
		},
		{
			name: "total supply after burning",
			setup: func() {
				std.TestSetRealm(positionRealm)
				Burn(tid(2))
			},
			expected: uint64(1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			uassert.Equal(t, tt.expected, TotalSupply())
		})
	}
}

func TestBalanceOf(t *testing.T) {
	tests := []struct {
		name     string
		addr     std.Address
		expected uint64
	}{
		{"BalanceOf(addr01)", addr01, uint64(1)},
		{"BalanceOf(addr02)", addr02, uint64(0)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			balance, _ := BalanceOf(tt.addr)
			uassert.Equal(t, tt.expected, balance)
		})
	}
}

func TestOwnerOf(t *testing.T) {
	tests := []struct {
		name        string
		tokenId     uint64
		shouldPanic bool
		panicMsg    string
		expected    std.Address
	}{
		{"OwnerOf(1)", 1, false, "", addr01},
		{"OwnerOf(500)", 500, false, errInvalidTokenId, addr01},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				uassert.PanicsWithMessage(t, tt.panicMsg, func() {
					OwnerOf(tid(tt.tokenId))
				})
			} else {
				ownerAddr, err := OwnerOf(tid(tt.tokenId))
				if err != nil {
					uassert.Equal(t, tt.panicMsg, err.Error())
				} else {
					uassert.Equal(t, tt.expected, ownerAddr)
				}
			}
		})
	}
}

func TestIsApprovedForAll(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		expected bool
	}{
		{
			name:     "IsApprovedForAll(addr01, addr02)",
			expected: false,
		},
		{
			name: "IsApprovedForAll(addr01, addr02) after setting approval",
			setup: func() {
				std.TestSetRealm(addr01Realm)
				SetApprovalForAll((addr02), true)
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			uassert.Equal(t, tt.expected, IsApprovedForAll((addr01), (addr02)))
		})
	}
}

func TestGetApproved(t *testing.T) {
	tests := []struct {
		name         string
		setup        func()
		expectedAddr std.Address
	}{
		{
			name:         "GetApproved(1)",
			expectedAddr: std.Address(""),
		},
		{
			name: "GetApproved(1) after approving",
			setup: func() {
				std.TestSetRealm(addr01Realm)
				Approve(addr02, tid(1))
			},
			expectedAddr: addr02,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			addr, _ := GetApproved(tid(1))
			uassert.Equal(t, tt.expectedAddr, addr)
		})
	}
}

func TestTransferFrom(t *testing.T) {
	resetObject(t)
	std.TestSetRealm(positionRealm)
	Mint(addr01, tid(1))

	tests := []struct {
		name              string
		setup             func()
		callerRealm       std.Realm
		fromAddr          std.Address
		toAddr            std.Address
		tokenIdToTransfer uint64
		shouldPanic       bool
		panicMsg          string
		expected          std.Address
		verifyTokenList   func()
	}{
		{
			name:              "transfer non-existent token id",
			callerRealm:       std.NewUserRealm(addr01),
			fromAddr:          addr01,
			toAddr:            addr02,
			tokenIdToTransfer: 99,
			shouldPanic:       true,
			panicMsg:          "[GNOSWAP-GNFT-001] caller has no permission || caller (g1q646ctzhvn60v492x8ucvyqnrj2w30cwh6efk5) is not the owner or operator of token (99)",
		},
		{
			name:              "transfer token owned by other user without approval",
			callerRealm:       std.NewUserRealm(addr02),
			fromAddr:          addr01,
			toAddr:            addr02,
			tokenIdToTransfer: 1,
			shouldPanic:       true,
			panicMsg:          "[GNOSWAP-GNFT-001] caller has no permission || caller (g1q646ctzhvn60v492x8ucvyqnrj2w30cwh6efk5) is not the owner or operator of token (1)",
		},
		{
			name: "transfer token owned by other user with approval",
			setup: func() {
				std.TestSetRealm(addr01Realm)
				Approve((addr02), tid(1))
			},
			callerRealm:       std.NewUserRealm(addr02),
			fromAddr:          addr01,
			toAddr:            addr02,
			tokenIdToTransfer: 1,
			verifyTokenList: func() {
				uassert.Equal(t, 0, len(mustGetTokenList(addr01)))
				uassert.Equal(t, 1, len(mustGetTokenList(addr02)))
			},
		},
		{
			name:              "transfer token owned by caller",
			callerRealm:       std.NewUserRealm(addr02),
			fromAddr:          addr02,
			toAddr:            addr01,
			tokenIdToTransfer: 1,
			verifyTokenList: func() {
				uassert.Equal(t, 1, len(mustGetTokenList(addr01)))
				uassert.Equal(t, 0, len(mustGetTokenList(addr02)))
			},
		},
		{
			name:              "transfer from is invalid address",
			callerRealm:       std.NewUserRealm(addr01),
			fromAddr:          std.Address(""),
			toAddr:            addr02,
			tokenIdToTransfer: 1,
			shouldPanic:       true,
			panicMsg:          "[GNOSWAP-GNFT-004] invalid addresss || ",
		},
		{
			name:              "transfer to is invalid address",
			callerRealm:       std.NewUserRealm(addr01),
			fromAddr:          addr01,
			toAddr:            std.Address("this_is_invalid_address"),
			tokenIdToTransfer: 1,
			shouldPanic:       true,
			panicMsg:          "[GNOSWAP-GNFT-004] invalid addresss || this_is_invalid_address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			if tt.shouldPanic {
				uassert.PanicsWithMessage(t, tt.panicMsg, func() {
					TransferFrom((tt.fromAddr), (tt.toAddr), tid(tt.tokenIdToTransfer))
				})
			} else {
				std.TestSetRealm(tt.callerRealm)
				TransferFrom((tt.fromAddr), (tt.toAddr), tid(tt.tokenIdToTransfer))
				tt.verifyTokenList()
			}
		})
	}
}

func TestMint(t *testing.T) {
	resetObject(t)

	tests := []struct {
		name            string
		callerRealm     std.Realm
		tokenIdToMint   uint64
		addressToMint   std.Address
		shouldPanic     bool
		panicMsg        string
		expected        string
		verifyTokenList func()
	}{
		{
			name:        "mint without permission",
			shouldPanic: true,
			panicMsg:    "ownable: caller is not owner",
		},
		{
			name:          "mint first nft to addr01",
			callerRealm:   std.NewCodeRealm(consts.POSITION_PATH),
			tokenIdToMint: 1,
			addressToMint: addr01,
			expected:      "1",
			verifyTokenList: func() {
				uassert.Equal(t, 1, len(mustGetTokenList(addr01)))
			},
		},
		{
			name:          "mint second nft to addr02",
			callerRealm:   std.NewCodeRealm(consts.POSITION_PATH),
			tokenIdToMint: 2,
			addressToMint: addr02,
			expected:      "2",
			verifyTokenList: func() {
				uassert.Equal(t, 1, len(mustGetTokenList(addr02)))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				uassert.PanicsWithMessage(t, tt.panicMsg, func() {
					Mint((tt.addressToMint), tid(tt.tokenIdToMint))
				})

			} else {
				std.TestSetRealm(tt.callerRealm)
				mintedTokenId := Mint((tt.addressToMint), tid(tt.tokenIdToMint))
				uassert.Equal(t, tt.expected, string(mintedTokenId))
				tt.verifyTokenList()
			}
		})
	}
}

func TestBurn(t *testing.T) {
	tests := []struct {
		name            string
		callerRealm     std.Realm
		tokenIdToBurn   uint64
		shouldPanic     bool
		panicMsg        string
		verifyTokenList func()
	}{
		{
			name:          "burn without permission",
			tokenIdToBurn: 1,
			shouldPanic:   true,
			panicMsg:      "ownable: caller is not owner",
		},
		{
			name:          "burn non-existent token id",
			callerRealm:   std.NewCodeRealm(consts.POSITION_PATH),
			tokenIdToBurn: 99,
			shouldPanic:   true,
			panicMsg:      errInvalidTokenId,
		},
		{
			name:          "burn token id(2)",
			callerRealm:   std.NewCodeRealm(consts.POSITION_PATH),
			tokenIdToBurn: 2,
			shouldPanic:   false,
			verifyTokenList: func() {
				uassert.Equal(t, 0, len(mustGetTokenList(addr02)))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			std.TestSetRealm(tt.callerRealm)

			if tt.shouldPanic {
				uassert.PanicsWithMessage(t, tt.panicMsg, func() {
					Burn(tid(tt.tokenIdToBurn))
				})
			} else {
				uassert.NotPanics(t, func() {
					Burn(tid(tt.tokenIdToBurn))
				})
				tt.verifyTokenList()
			}
		})
	}
}

func TestSetTokenURI(t *testing.T) {
	tests := []struct {
		name        string
		callerRealm std.Realm
		tokenId     uint64
		shouldPanic bool
		panicMsg    string
	}{
		{
			name:        "set token uri without permission",
			tokenId:     1,
			shouldPanic: true,
			panicMsg:    `[GNOSWAP-GNFT-001] caller has no permission || caller () is not the owner of token (1)`,
		},
		{
			name:        "set token uri of non-minted token id",
			tokenId:     99,
			shouldPanic: true,
			panicMsg:    `[GNOSWAP-GNFT-002] cannot set URI || invalid token id (99)`,
		},
		{
			name:        "set token uri of token id(1)",
			callerRealm: addr01Realm,
			tokenId:     1,
		},
		{
			name:        "set token uri of token id(1) - twice",
			callerRealm: addr01Realm,
			tokenId:     1,
			shouldPanic: true,
			panicMsg:    "[GNOSWAP-GNFT-002] cannot set URI || token id (1) has already set URI",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			std.TestSetRealm(tt.callerRealm)

			if tt.shouldPanic {
				uassert.PanicsWithMessage(t, tt.panicMsg, func() {
					SetTokenURIByImageURI(tid(tt.tokenId))
				})
			} else {
				uassert.NotPanics(t, func() {
					SetTokenURIByImageURI(tid(tt.tokenId))
				})
			}
		})
	}
}

func TestTokenURI(t *testing.T) {
	resetObject(t)

	tests := []struct {
		name        string
		setup       func()
		tokenId     uint64
		shouldPanic bool
		panicMsg    string
	}{
		{
			name:        "get token uri of non-minted token id",
			tokenId:     99,
			shouldPanic: true,
			panicMsg:    errInvalidTokenId,
		},
		{
			name: "get token uri of minted token but not set token uri",
			setup: func() {
				std.TestSetRealm(positionRealm)
				Mint((addr01), tid(1))
			},
			tokenId:     1,
			shouldPanic: true,
			panicMsg:    errInvalidTokenId,
		},
		{
			name: "get token uri of minted token after setting token uri",
			setup: func() {
				std.TestSetRealm(addr01Realm)
				SetTokenURIByImageURI(tid(1))
			},
			tokenId: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			if tt.shouldPanic {
				uassert.PanicsWithMessage(t, tt.panicMsg, func() {
					TokenURI(tid(tt.tokenId))
				})
			} else {
				uassert.NotEmpty(t, TokenURI(tid(tt.tokenId)))
			}
		})
	}
}

func TestSetTokenURILast(t *testing.T) {
	resetObject(t)
	std.TestSetRealm(positionRealm)
	Mint(addr01, tid(1))
	Mint(addr01, tid(2)) // last minted

	t.Run("set token uri last", func(t *testing.T) {
		std.TestSetRealm(addr01Realm)
		SetTokenURILast()
	})

	t.Run("token uri(2)", func(t *testing.T) {
		uassert.NotEmpty(t, TokenURI(tid(2)))
	})
}

func resetObject(t *testing.T) {
	t.Helper()

	gnft = grc721.NewBasicNFT("GNOSWAP NFT", "GNFT")
	tokenList = avl.NewTree()
}
