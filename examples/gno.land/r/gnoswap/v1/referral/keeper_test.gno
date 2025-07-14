package referral

import (
	"std"
	"testing"
	"time"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"

	"gno.land/r/gnoswap/v1/access"
)

const STRESS_TEST_NUM = 10000 // arbitrary number

var (
	validAddr1  = testutils.TestAddress("valid1")
	validAddr2  = testutils.TestAddress("valid2")
	validAddr3  = testutils.TestAddress("valid3")
	validAddr4  = testutils.TestAddress("valid4")
	validAddr5  = testutils.TestAddress("valid5")
	validAddr6  = testutils.TestAddress("valid6")
	validAddr7  = testutils.TestAddress("valid7")
	validAddr8  = testutils.TestAddress("valid8")
	invalidAddr = testutils.TestAddress("invalid")
)

// time mocking
var currentTime int64 = time.Now().Unix()

func mockTimeNow() time.Time {
	return time.Unix(currentTime, 0)
}

func setupKeeper() *keeper { return NewKeeper().(*keeper) }

func mockValidCaller() func() {
	origCaller := std.OriginCaller()
	routerAddr, _ := access.GetAddress(access.ROLE_ROUTER)
	testing.SetOriginCaller(routerAddr)
	return func() {
		testing.SetOriginCaller(origCaller)
	}
}

func TestRegister(t *testing.T) {
	tests := []struct {
		name        string
		addr        std.Address
		refAddr     std.Address
		setupCaller func() func()
		wantErr     bool
	}{
		{
			name:        "valid registration",
			addr:        validAddr1,
			refAddr:     validAddr2,
			setupCaller: mockValidCaller,
			wantErr:     false,
		},
		{
			name:    "unauthorized caller",
			addr:    validAddr1,
			refAddr: validAddr2,
			setupCaller: func() func() {
				origCaller := std.OriginCaller()
				testing.SetOriginCaller(std.Address("unauthorized"))
				return func() {
					testing.SetOriginCaller(origCaller)
				}
			},
			wantErr: true,
		},
		{
			name:        "self referral",
			addr:        validAddr1,
			refAddr:     validAddr1,
			setupCaller: mockValidCaller,
			wantErr:     true,
		},
		{
			name:        "zero address referral",
			addr:        validAddr1,
			refAddr:     zeroAddress,
			setupCaller: mockValidCaller,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := setupKeeper()
			cleanup := tt.setupCaller()
			defer cleanup()

			err := k.register(tt.addr, tt.refAddr)

			if tt.wantErr {
				if err == nil {
					t.Errorf("register() error = nil, wantErr %v", tt.wantErr)
					return
				}
			} else if err != nil {
				t.Errorf("register() unexpected error = %v", err)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		name        string
		addr        std.Address
		refAddr     std.Address
		setupState  func(*keeper)
		setupCaller func() func()
		wantErr     bool
	}{
		{
			name:    "valid update",
			addr:    validAddr1,
			refAddr: validAddr2,
			setupState: func(k *keeper) {
				k.store.Set(validAddr1.String(), "old_ref_addr")
			},
			setupCaller: mockValidCaller,
			wantErr:     false,
		},
		{
			name:        "update non-existent referral",
			addr:        validAddr1,
			refAddr:     validAddr2,
			setupState:  func(k *keeper) {},
			setupCaller: mockValidCaller,
			wantErr:     false,
		},
		{
			name:    "update to self referral",
			addr:    validAddr1,
			refAddr: validAddr1,
			setupState: func(k *keeper) {
				k.store.Set(validAddr1.String(), validAddr2.String())
			},
			setupCaller: mockValidCaller,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := setupKeeper()
			tt.setupState(k)
			cleanup := tt.setupCaller()
			defer cleanup()

			err := k.update(tt.addr, tt.refAddr)

			if tt.wantErr {
				if err == nil {
					t.Errorf("update() error = nil, wantErr %v", tt.wantErr)
					return
				}
			} else if err != nil {
				t.Errorf("update() unexpected error = %v", err)
			}
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name       string
		addr       std.Address
		setupState func(*keeper)
		wantAddr   std.Address
		wantErr    bool
		errCode    int
	}{
		{
			name: "get existing referral",
			addr: validAddr1,
			setupState: func(k *keeper) {
				k.store.Set(validAddr1.String(), validAddr2.String())
			},
			wantAddr: validAddr2,
			wantErr:  false,
		},
		{
			name:       "get non-existent referral",
			addr:       validAddr1,
			setupState: func(k *keeper) {},
			wantAddr:   zeroAddress,
			wantErr:    true,
		},
		{
			name:       "get with invalid address",
			addr:       invalidAddr,
			setupState: func(k *keeper) {},
			wantAddr:   zeroAddress,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := setupKeeper()
			tt.setupState(k)

			gotAddr, err := k.get(tt.addr)

			if tt.wantErr {
				if err == nil {
					t.Errorf("get() error = nil, wantErr %v", tt.wantErr)
					return
				}
			} else {
				if err != nil {
					t.Errorf("get() unexpected error")
					return
				}
				if gotAddr != tt.wantAddr {
					t.Errorf("get() gotAddr = %v, want %v", gotAddr, tt.wantAddr)
				}
			}
		})
	}
}

func TestHas(t *testing.T) {
	tests := []struct {
		name       string
		addr       std.Address
		setupState func(*keeper)
		want       bool
	}{
		{
			name: "has existing referral",
			addr: validAddr1,
			setupState: func(k *keeper) {
				k.store.Set(validAddr1.String(), validAddr2.String())
			},
			want: true,
		},
		{
			name:       "does not have referral",
			addr:       validAddr1,
			setupState: func(k *keeper) {},
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := setupKeeper()
			tt.setupState(k)

			if got := k.has(tt.addr); got != tt.want {
				t.Errorf("has() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemove(t *testing.T) {
	tests := []struct {
		name        string
		addr        std.Address
		setupState  func(*keeper)
		setupCaller func() func()
		wantErr     bool
	}{
		{
			name: "remove existing referral",
			addr: validAddr1,
			setupState: func(k *keeper) {
				k.store.Set(validAddr1.String(), validAddr2.String())
			},
			setupCaller: mockValidCaller,
			wantErr:     false,
		},
		{
			name:        "remove non-existent referral",
			addr:        validAddr1,
			setupState:  func(k *keeper) {},
			setupCaller: mockValidCaller,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := setupKeeper()
			tt.setupState(k)
			cleanup := tt.setupCaller()
			defer cleanup()

			println("addr: ", tt.addr.String())

			err := k.remove(tt.addr)

			if tt.wantErr {
				if err == nil {
					t.Errorf("remove() error = nil, wantErr %v", tt.wantErr)
					return
				}
			} else if err != nil {
				t.Errorf("remove() unexpected error = %v", err)
			}

			if k.has(tt.addr) {
				t.Errorf("remove() referral still exists after removal")
			}
		})
	}
}

func TestUpdateNonExistentReferral(t *testing.T) {
	k := setupKeeper()
	cleanup := mockValidCaller()
	defer cleanup()

	err := k.update(validAddr1, validAddr2)
	if err != nil {
		t.Errorf("update() for non-existent referral failed: %v", err)
	}

	refAddr, err := k.get(validAddr1)
	if err != nil {
		t.Errorf("get() after update failed: %v", err)
	}
	if refAddr != validAddr2 {
		t.Errorf("got refAddr = %v, want %v", refAddr, validAddr2)
	}
}

func TestReferralCycles(t *testing.T) {
	k := setupKeeper()
	cleanup := mockValidCaller()
	defer cleanup()

	addr1 := testutils.TestAddress("cycle1")
	addr2 := testutils.TestAddress("cycle2")
	addr3 := testutils.TestAddress("cycle3")

	// A -> B -> C
	err := k.register(addr1, addr2)
	if err != nil {
		t.Fatalf("Failed to register addr1->addr2: %v", err)
	}

	err = k.register(addr2, addr3)
	if err != nil {
		t.Fatalf("Failed to register addr2->addr3: %v", err)
	}

	// reference cycle: C -> A
	err = k.register(addr3, addr1)
	if err != nil {
		t.Fatalf("Failed to register addr3->addr1: %v", err)
	}

	refAddr, _ := k.get(addr1)
	if refAddr != addr2 {
		t.Error("addr1's referral should be addr2")
	}

	refAddr, _ = k.get(addr2)
	if refAddr != addr3 {
		t.Error("addr2's referral should be addr3")
	}

	refAddr, _ = k.get(addr3)
	if refAddr != addr1 {
		t.Error("addr3's referral should be addr1")
	}
}

func TestStress(t *testing.T) {
	t.Skip("Skipping stress test")

	k := setupKeeper()
	cleanup := mockValidCaller()
	defer cleanup()

	addresses := make([]std.Address, STRESS_TEST_NUM)

	for i := 0; i < STRESS_TEST_NUM; i++ {
		addresses[i] = testutils.TestAddress(ufmt.Sprintf("addr%d", i))
	}

	for i := 0; i < STRESS_TEST_NUM; i++ {
		err := k.register(addresses[i], addresses[(i+1)%STRESS_TEST_NUM])
		if err != nil {
			t.Fatalf("Registration failed at index %d: %v", i, err)
		}

		err = k.update(addresses[i], addresses[(i+2)%STRESS_TEST_NUM])
		if err != nil {
			t.Fatalf("Update failed at index %d: %v", i, err)
		}

		// remove some addresses
		if i%3 == 0 {
			err = k.remove(addresses[i])
			if err != nil {
				t.Fatalf("Remove failed at index %d: %v", i, err)
			}
		}

		// check data consistency
		if i%1000 == 0 {
			for j := 0; j <= i; j++ {
				if j%3 == 0 {
					// check removed address
					if k.has(addresses[j]) {
						t.Errorf("Removed address still exists at index %d", j)
					}
				} else {
					// check registered address
					refAddr, err := k.get(addresses[j])
					if err != nil {
						t.Errorf("Failed to get referral at index %d: %v", j, err)
					}
					expectedAddr := addresses[(j+2)%STRESS_TEST_NUM]
					if refAddr != expectedAddr {
						t.Errorf("Incorrect referral at index %d", j)
					}
				}
			}
		}
	}
}

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name       string
		setupState func(*keeper)
		want       bool
	}{
		{
			name:       "new keeper must empty",
			setupState: func(k *keeper) {},
			want:       true,
		},
		{
			name: "keeper with data must not be empty",
			setupState: func(k *keeper) {
				k.store.Set(validAddr1.String(), validAddr2.String())
			},
			want: false,
		},
		{
			name: "keeper with all data removed must be empty",
			setupState: func(k *keeper) {
				k.store.Set(validAddr1.String(), validAddr2.String())
				k.store.Remove(validAddr1.String())
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := setupKeeper()
			tt.setupState(k)

			if got := k.isEmpty(); got != tt.want {
				t.Errorf("isEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateAddresses(t *testing.T) {
	invalid1 := std.Address("000000000000")
	invalidRef := std.Address("000000000001")

	tests := []struct {
		name    string
		addr    std.Address
		refAddr std.Address
		wantErr error
	}{
		{
			name:    "valid addresses",
			addr:    validAddr1,
			refAddr: validAddr2,
			wantErr: nil,
		},
		{
			name:    "invalid main address",
			addr:    invalid1,
			refAddr: validAddr2,
			wantErr: ErrInvalidAddress,
		},
		{
			name:    "invalid referral address",
			addr:    validAddr1,
			refAddr: invalid1,
			wantErr: ErrInvalidAddress,
		},
		{
			name:    "self referral",
			addr:    validAddr1,
			refAddr: validAddr1,
			wantErr: ErrSelfReferral,
		},
		{
			name:    "zero address referral",
			addr:    validAddr1,
			refAddr: zeroAddress,
			wantErr: nil,
		},
		{
			name:    "both addresses invalid",
			addr:    invalid1,
			refAddr: invalidRef,
			wantErr: ErrInvalidAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := setupKeeper()
			err := k.validateAddresses(tt.addr, tt.refAddr)

			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("validateAddresses() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else {
				if err == nil {
					t.Errorf("validateAddresses() error = nil, wantErr %v", tt.wantErr)
				} else if err != tt.wantErr {
					t.Errorf("validateAddresses() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

func TestSetReferral(t *testing.T) {
	tests := []struct {
		name        string
		addr        std.Address
		refAddr     std.Address
		eventType   string
		setupState  func(*keeper)
		setupCaller func() func()
		wantErr     bool
		checkStore  func(*keeper) bool
	}{
		{
			name:        "valid registration",
			addr:        validAddr1,
			refAddr:     validAddr2,
			eventType:   EventTypeRegister,
			setupState:  func(k *keeper) {},
			setupCaller: mockValidCaller,
			wantErr:     false,
			checkStore: func(k *keeper) bool {
				val, exists := k.store.Get(validAddr1.String())
				return exists && val == validAddr2.String()
			},
		},
		{
			name:      "valid update",
			addr:      validAddr1,
			refAddr:   validAddr2,
			eventType: EventTypeUpdate,
			setupState: func(k *keeper) {
				k.store.Set(validAddr1.String(), "old_ref_addr")
			},
			setupCaller: mockValidCaller,
			wantErr:     false,
			checkStore: func(k *keeper) bool {
				val, exists := k.store.Get(validAddr1.String())
				return exists && val == validAddr2.String()
			},
		},
		{
			name:      "remove referral with zero address",
			addr:      validAddr1,
			refAddr:   zeroAddress,
			eventType: EventTypeRemove,
			setupState: func(k *keeper) {
				k.store.Set(validAddr1.String(), validAddr2.String())
			},
			setupCaller: mockValidCaller,
			wantErr:     false,
			checkStore: func(k *keeper) bool {
				// GSW-2258:
				// check the address is removed from the store
				_, exists := k.store.Get(validAddr1.String())
				if exists {
					return false
				}

				// double checks
				if GetReferral(validAddr1.String()) != "" {
					return false
				}

				if HasReferral(validAddr1.String()) {
					return false
				}

				return true
			},
		},
		{
			name:       "unauthorized caller",
			addr:       validAddr1,
			refAddr:    validAddr2,
			eventType:  EventTypeRegister,
			setupState: func(k *keeper) {},
			setupCaller: func() func() {
				origCaller := std.OriginCaller()
				testing.SetOriginCaller(std.Address("unauthorized"))
				return func() {
					testing.SetOriginCaller(origCaller)
				}
			},
			wantErr: true,
			checkStore: func(k *keeper) bool {
				_, exists := k.store.Get(validAddr1.String())
				return !exists
			},
		},
		{
			name:        "self referral",
			addr:        validAddr1,
			refAddr:     validAddr1,
			eventType:   EventTypeRegister,
			setupState:  func(k *keeper) {},
			setupCaller: mockValidCaller,
			wantErr:     true,
			checkStore: func(k *keeper) bool {
				_, exists := k.store.Get(validAddr1.String())
				return !exists
			},
		},
		{
			name:        "invalid address",
			addr:        std.Address("000000000000"),
			refAddr:     validAddr2,
			eventType:   EventTypeRegister,
			setupState:  func(k *keeper) {},
			setupCaller: mockValidCaller,
			wantErr:     true,
			checkStore: func(k *keeper) bool {
				_, exists := k.store.Get("000000000000")
				return !exists
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := setupKeeper()
			tt.setupState(k)
			cleanup := tt.setupCaller()
			defer cleanup()

			err := k.setReferral(tt.addr, tt.refAddr, tt.eventType)

			if tt.wantErr {
				if err == nil {
					t.Errorf("setReferral() error = nil, wantErr %v", tt.wantErr)
				}
			} else {
				if err != nil {
					t.Errorf("setReferral() unexpected error = %v", err)
				}
			}

			if !tt.checkStore(k) {
				t.Errorf("setReferral() store state is incorrect")
			}
		})
	}
}

/*             Getter Tests               */

func TestGetReferral(t *testing.T) {
	tests := []struct {
		name       string
		addr       string
		setupState func()
		want       string
	}{
		{
			name: "retrieve existing referral",
			addr: validAddr1.String(),
			setupState: func() {
				gReferralKeeper = NewKeeper()
				cleanup := mockValidCaller()
				defer cleanup()
				gReferralKeeper.register(validAddr1, validAddr2)
			},
			want: validAddr2.String(),
		},
		{
			name: "retrieve non-existent referral",
			addr: validAddr1.String(),
			setupState: func() {
				gReferralKeeper = NewKeeper()
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupState()

			if GetReferral(tt.addr) != tt.want {
				t.Errorf("GetReferral() = %v, want %v", GetReferral(tt.addr), tt.want)
			}
		})
	}
}

func TestHasReferral(t *testing.T) {
	tests := []struct {
		name       string
		addr       string
		setupState func()
		want       bool
	}{
		{
			name: "referral exists",
			addr: validAddr1.String(),
			setupState: func() {
				gReferralKeeper = NewKeeper()
				cleanup := mockValidCaller()
				defer cleanup()
				gReferralKeeper.register(validAddr1, validAddr2)
			},
			want: true,
		},
		{
			name: "referral does not exist",
			addr: validAddr1.String(),
			setupState: func() {
				gReferralKeeper = NewKeeper()
			},
			want: false,
		},
		{
			name: "zero address referral",
			addr: validAddr1.String(),
			setupState: func() {
				gReferralKeeper = NewKeeper()
				cleanup := mockValidCaller()
				defer cleanup()
				gReferralKeeper.register(validAddr1, zeroAddress)
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupState()
			if got := HasReferral(tt.addr); got != tt.want {
				t.Errorf("HasReferral() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGlobalIsEmpty(t *testing.T) {
	tests := []struct {
		name       string
		setupState func()
		want       bool
	}{
		{
			name: "new global keeper must be empty",
			setupState: func() {
				gReferralKeeper = NewKeeper()
			},
			want: true,
		},
		{
			name: "keeper with data must not be empty",
			setupState: func() {
				gReferralKeeper = NewKeeper()
				cleanup := mockValidCaller()
				defer cleanup()
				gReferralKeeper.register(validAddr1, validAddr2)
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupState()
			if got := IsEmpty(); got != tt.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTryRegister(t *testing.T) {
	tests := []struct {
		name                         string
		inputUserAddr                std.Address
		inputFirstRefAddr            std.Address
		inputSecondRefAddr           std.Address
		inputSkipBlockHeight         int64
		expectedFirstIsSuccess       bool
		expectedFirstRegisteredAddr  std.Address
		expectedSecondIsSuccess      bool
		expectedSecondRegisteredAddr std.Address
	}{
		{
			name:                         "try register with invalid referral address then valid referral address",
			inputUserAddr:                validAddr3,
			inputFirstRefAddr:            std.Address("000000000001"),
			inputSecondRefAddr:           validAddr4,
			inputSkipBlockHeight:         0,
			expectedFirstIsSuccess:       false,
			expectedFirstRegisteredAddr:  zeroAddress,
			expectedSecondIsSuccess:      true,
			expectedSecondRegisteredAddr: validAddr4,
		},
		{
			name:                         "try register with valid referral address then failed by rate limit",
			inputUserAddr:                validAddr3,
			inputFirstRefAddr:            validAddr4,
			inputSecondRefAddr:           validAddr5,
			inputSkipBlockHeight:         0,
			expectedFirstIsSuccess:       true,
			expectedFirstRegisteredAddr:  validAddr4,
			expectedSecondIsSuccess:      false,
			expectedSecondRegisteredAddr: validAddr4,
		},
		{
			name:                         "try register with valid referral address then success",
			inputUserAddr:                validAddr3,
			inputFirstRefAddr:            validAddr4,
			inputSecondRefAddr:           validAddr5,
			inputSkipBlockHeight:         (60 * 60 * 24 / 5), // 5 seconds per block (60 * 60 * 24 / 5 = 1day)
			expectedFirstIsSuccess:       true,
			expectedFirstRegisteredAddr:  validAddr4,
			expectedSecondIsSuccess:      true,
			expectedSecondRegisteredAddr: validAddr5,
		},
		{
			name:                         "try register with empty referral address then valid referral address",
			inputUserAddr:                validAddr3,
			inputFirstRefAddr:            "",
			inputSecondRefAddr:           validAddr5,
			inputSkipBlockHeight:         (60 * 60 * 24 / 5), // 5 seconds per block (60 * 60 * 24 / 5 = 1day)
			expectedFirstIsSuccess:       true,
			expectedFirstRegisteredAddr:  zeroAddress,
			expectedSecondIsSuccess:      true,
			expectedSecondRegisteredAddr: validAddr5,
		},
		{
			name:                         "try register with valid referral address then zero address",
			inputUserAddr:                validAddr3,
			inputFirstRefAddr:            validAddr4,
			inputSecondRefAddr:           zeroAddress,
			inputSkipBlockHeight:         (60 * 60 * 24 / 5), // 5 seconds per block (60 * 60 * 24 / 5 = 1day)
			expectedFirstIsSuccess:       true,
			expectedFirstRegisteredAddr:  validAddr4,
			expectedSecondIsSuccess:      true,
			expectedSecondRegisteredAddr: zeroAddress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			cleanup := mockValidCaller()
			defer cleanup()

			gReferralKeeper = NewKeeper()

			routerAddr, _ := access.GetAddress(access.ROLE_ROUTER)
			testing.SetOriginCaller(routerAddr)

			// when: first register
			success := TryRegister(cross, tt.inputUserAddr, tt.inputFirstRefAddr.String())
			uassert.Equal(t, tt.expectedFirstIsSuccess, success)

			// then: first registered address should be the first referral address
			firstRegisteredAddr := GetReferral(tt.inputUserAddr.String())
			uassert.Equal(t, tt.expectedFirstRegisteredAddr.String(), firstRegisteredAddr)

			// skip block height
			testing.SkipHeights(tt.inputSkipBlockHeight)

			// when: second register
			success = TryRegister(cross, tt.inputUserAddr, tt.inputSecondRefAddr.String())
			uassert.Equal(t, tt.expectedSecondIsSuccess, success)

			// then: second registered address should be the second referral address
			secondRegisteredAddr := GetReferral(tt.inputUserAddr.String())
			uassert.Equal(t, tt.expectedSecondRegisteredAddr.String(), secondRegisteredAddr)
		})
	}
}
