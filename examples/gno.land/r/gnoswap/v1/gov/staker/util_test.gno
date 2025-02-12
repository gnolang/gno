package staker

import (
	b64 "encoding/base64"
	"math"
	"std"
	"testing"
	"time"

	"gno.land/p/demo/json"
	"gno.land/p/demo/uassert"

	"gno.land/p/demo/avl"
	"gno.land/p/gnoswap/consts"
)

func TestMarshal(t *testing.T) {
	fakeNode := json.ObjectNode("", map[string]*json.Node{
		"height":    json.NumberNode("height", float64(std.GetHeight())),
		"timestamp": json.NumberNode("timestamp", float64(time.Now().Unix())),
	})

	got := marshal(fakeNode)
	if !(len(got) > 0) {
		t.Errorf("TestMarshal() failed, got empty string")
	}
	t.Run("TestMarshal", func(t *testing.T) {
		uassert.Equal(t, "{\"height\":123,\"timestamp\":1234567890}", got)
	})
}

func TestB64Encode(t *testing.T) {
	input := "Hello World"
	want := b64.StdEncoding.EncodeToString([]byte(input))
	got := b64Encode(input)
	if got != want {
		t.Errorf("TestB64Encode() = %s; want %s", got, want)
	}
}

func TestGetPrevRealm(t *testing.T) {
	got := getPrevRealm()
	if got.PkgPath() != "" {
		t.Errorf("TestPrevRealm() got package path")
	}
}

func TestIsUserCall(t *testing.T) {
	tests := []struct {
		name     string
		action   func() bool
		expected bool
	}{
		{
			name: "called from user",
			action: func() bool {
				userRealm := std.NewUserRealm(std.Address("user"))
				std.TestSetRealm(userRealm)
				return isUserCall()
			},
			expected: true,
		},
		{
			name: "called from realm",
			action: func() bool {
				fromRealm := std.NewCodeRealm("gno.land/r/realm")
				std.TestSetRealm(fromRealm)
				return isUserCall()
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uassert.Equal(t, tt.expected, tt.action())
		})
	}
}

func TestGetPrev(t *testing.T) {
	tests := []struct {
		name            string
		action          func() (string, string)
		expectedAddr    string
		expectedPkgPath string
	}{
		{
			name: "user call",
			action: func() (string, string) {
				userRealm := std.NewUserRealm(std.Address("user"))
				std.TestSetRealm(userRealm)
				return getPrev()
			},
			expectedAddr:    "user",
			expectedPkgPath: "",
		},
		{
			name: "code call",
			action: func() (string, string) {
				codeRealm := std.NewCodeRealm("gno.land/r/demo/realm")
				std.TestSetRealm(codeRealm)
				return getPrev()
			},
			expectedAddr:    std.DerivePkgAddr("gno.land/r/demo/realm").String(),
			expectedPkgPath: "gno.land/r/demo/realm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, pkgPath := tt.action()
			uassert.Equal(t, tt.expectedAddr, addr)
			uassert.Equal(t, tt.expectedPkgPath, pkgPath)
		})
	}
}

func TestFormatUint(t *testing.T) {
	tests := []struct {
		input    uint64
		expected string
	}{
		{0, "0"},
		{12345, "12345"},
		{math.MaxUint64, "18446744073709551615"},
	}

	for _, tt := range tests {
		result := formatUint(tt.input)
		if result != tt.expected {
			t.Errorf("formatUint(%d) = %s; want %s", tt.input, result, tt.expected)
		}
	}
}

func TestFormatInt(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0"},
		{-12345, "-12345"},
		{math.MaxInt64, "9223372036854775807"},
		{math.MinInt64, "-9223372036854775808"},
	}

	for _, tt := range tests {
		result := formatInt(tt.input)
		if result != tt.expected {
			t.Errorf("formatInt(%d) = %s; want %s", tt.input, result, tt.expected)
		}
	}
}

func TestAssertCallerIsAdmin(t *testing.T) {
	adminAddr := consts.ADMIN
	nonAdminAddr := std.Address("user456")

	std.TestSetOrigCaller(adminAddr)
	assertCallerIsAdmin()

	// Test non-admin
	std.TestSetOrigCaller(nonAdminAddr)
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("assertCallerIsAdmin() did not panic for non-admin")
		}
	}()
	assertCallerIsAdmin()
}

func TestAssertCallerIsLaunchpad(t *testing.T) {
	launchpadAddr := consts.LAUNCHPAD_ADDR
	otherAddr := std.Address("other123")

	// Test valid launchpad
	std.TestSetOrigCaller(launchpadAddr)
	assertCallerIsLaunchpad()

	// Test invalid caller
	std.TestSetOrigCaller(otherAddr)
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("assertCallerIsLaunchpad() did not panic for non-launchpad caller")
		}
	}()
	assertCallerIsLaunchpad()
}

func TestGetUint64FromTree(t *testing.T) {
	tree := avl.NewTree()
	tree.Set("key1", uint64(100))

	if value := getUint64FromTree(tree, "key1"); value != 100 {
		t.Errorf("getUint64FromTree(tree, 'key1') = %d; want 100", value)
	}

	if value := getUint64FromTree(tree, "key2"); value != 0 {
		t.Errorf("getUint64FromTree(tree, 'key2') = %d; want 0", value)
	}
}

func TestUpdateUint64InTree(t *testing.T) {
	tree := avl.NewTree()
	tree.Set("key1", uint64(100))

	// Add value
	if newValue := updateUint64InTree(tree, "key1", 50, true); newValue != 150 {
		t.Errorf("updateUint64InTree add failed; got %d, want 150", newValue)
	}

	// Subtract value
	if newValue := updateUint64InTree(tree, "key1", 50, false); newValue != 100 {
		t.Errorf("updateUint64InTree subtract failed; got %d, want 100", newValue)
	}

	// Attempt to subtract too much
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("updateUint64InTree did not panic on insufficient balance")
		}
	}()
	updateUint64InTree(tree, "key1", 150, false)
}

func TestGetOrCreateInnerTree(t *testing.T) {
	tree := avl.NewTree()

	// Test creation of new inner tree
	innerTree := getOrCreateInnerTree(tree, "key1")
	if innerTree == nil {
		t.Errorf("getOrCreateInnerTree did not create a new inner tree")
	}

	// Test retrieval of existing inner tree
	retrievedTree := getOrCreateInnerTree(tree, "key1")
	if retrievedTree != innerTree {
		t.Errorf("getOrCreateInnerTree did not return the existing inner tree")
	}
}
