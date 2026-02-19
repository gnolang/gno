package antisquat

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
)

var (
	testAddr1 = std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")
	testAddr2 = std.Address("g1us8428u2a5satrlxzagqqa5m6vmuze025anjlj")
	testAddr3 = std.Address("g1w4ek2u3sx9047h6lta047h6lta047h6lh0ssfv")
)

func TestValidation(t *testing.T) {
	cfg := DefaultConfig()

	validNames := []string{
		"a",
		"alice",
		"bob123",
		"my_namespace",
		"x1y2z3",
		"abcdefghijklmnopqrstuvwxyz0123456789",
	}
	for _, name := range validNames {
		err := ValidateName(name, cfg)
		uassert.NoError(t, err, "expected valid name: "+name)
	}

	uassert.ErrorIs(t, ValidateName("", cfg), ErrNameEmpty)
	uassert.ErrorIs(t, ValidateName("1abc", cfg), ErrNameInvalidChars)
	uassert.ErrorIs(t, ValidateName("_abc", cfg), ErrNameInvalidChars)
	uassert.ErrorIs(t, ValidateName("Alice", cfg), ErrNameInvalidChars)
	uassert.ErrorIs(t, ValidateName("my-name", cfg), ErrNameInvalidChars)
	uassert.ErrorIs(t, ValidateName("my name", cfg), ErrNameInvalidChars)

	longName := "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklm"
	uassert.True(t, len(longName) == 65, "longName should be 65 chars")
	uassert.ErrorIs(t, ValidateName(longName, cfg), ErrNameTooLong)
}

func TestNormalizeName(t *testing.T) {
	uassert.Equal(t, "alice", NormalizeName("Alice"))
	uassert.Equal(t, "bob", NormalizeName("  BOB  "))
	uassert.Equal(t, "my_name", NormalizeName("My_Name"))
}

func TestRegister(t *testing.T) {
	as := New(nil)

	record, err := as.Register("alice", testAddr1)
	uassert.NoError(t, err, "registration should succeed")
	uassert.True(t, record != nil, "record should not be nil")
	uassert.Equal(t, "alice", record.Name)
	uassert.Equal(t, string(testAddr1), string(record.Owner))
	uassert.True(t, record.RegisteredAt > 0, "registered height should be positive")
	uassert.True(t, record.ExpiresAt > record.RegisteredAt, "expiry should be after registration")
	uassert.True(t, record.RenewalFee > 0, "renewal fee should be positive")
	uassert.Equal(t, 1, as.TotalRegistered())
}

func TestRegisterNormalization(t *testing.T) {
	as := New(nil)

	record, err := as.Register("Alice", testAddr1)
	uassert.NoError(t, err, "registration with mixed case should succeed")
	uassert.Equal(t, "alice", record.Name)

	_, err = as.Register("alice", testAddr2)
	uassert.ErrorIs(t, err, ErrNameTaken)
}

func TestRegisterDuplicate(t *testing.T) {
	as := New(nil)

	_, err := as.Register("bob", testAddr1)
	uassert.NoError(t, err, "first registration should succeed")

	_, err = as.Register("bob", testAddr2)
	uassert.ErrorIs(t, err, ErrNameTaken)

	_, err = as.Register("bob", testAddr1)
	uassert.ErrorIs(t, err, ErrNameTaken)
}

func TestRegisterReserved(t *testing.T) {
	as := New(nil)

	as.AddReservedName("admin", ReservedSystem)

	_, err := as.Register("admin", testAddr1)
	uassert.ErrorIs(t, err, ErrNameReserved)

	err = as.CanRegister("admin", testAddr1)
	uassert.ErrorIs(t, err, ErrNameReserved)
}

func TestRateLimit(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxRegistrationsPerWindow = 3
	cfg.RateLimitWindowBlocks = 100

	as := New(cfg)

	_, err := as.Register("name1", testAddr1)
	uassert.NoError(t, err, "1st registration should succeed")

	_, err = as.Register("name2", testAddr1)
	uassert.NoError(t, err, "2nd registration should succeed")

	_, err = as.Register("name3", testAddr1)
	uassert.NoError(t, err, "3rd registration should succeed")

	_, err = as.Register("name4", testAddr1)
	uassert.ErrorIs(t, err, ErrRateLimited)

	_, err = as.Register("name5", testAddr2)
	uassert.NoError(t, err, "different address should not be rate limited")
}

func TestPricing(t *testing.T) {
	as := New(nil)

	// 1-3 char names: 1,000,000,000 ugnot (1000 GNOT)
	uassert.Equal(t, int(int64(1_000_000_000)), int(as.CalculateFee("a")))
	uassert.Equal(t, int(int64(1_000_000_000)), int(as.CalculateFee("ab")))
	uassert.Equal(t, int(int64(1_000_000_000)), int(as.CalculateFee("abc")))

	// 4 char names: 100,000,000 ugnot (100 GNOT)
	uassert.Equal(t, int(int64(100_000_000)), int(as.CalculateFee("abcd")))

	// 5 char names: 50,000,000 ugnot (50 GNOT)
	uassert.Equal(t, int(int64(50_000_000)), int(as.CalculateFee("abcde")))

	// 6-10 char names: 10,000,000 ugnot (10 GNOT)
	uassert.Equal(t, int(int64(10_000_000)), int(as.CalculateFee("abcdef")))
	uassert.Equal(t, int(int64(10_000_000)), int(as.CalculateFee("abcdefghij")))

	// 11+ char names: 5,000,000 ugnot (5 GNOT)
	uassert.Equal(t, int(int64(5_000_000)), int(as.CalculateFee("abcdefghijk")))
	uassert.Equal(t, int(int64(5_000_000)), int(as.CalculateFee("averylongnamespace")))
}

func TestReservedNames(t *testing.T) {
	as := New(nil)

	uassert.False(t, as.IsReserved("admin"))
	uassert.False(t, as.IsReserved("system"))

	as.AddReservedName("admin", ReservedSystem)
	as.AddReservedName("govdao", ReservedGovernance)
	as.AddReservedName("gnoland", ReservedBrand)

	uassert.True(t, as.IsReserved("admin"))
	uassert.True(t, as.IsReserved("govdao"))
	uassert.True(t, as.IsReserved("gnoland"))
	uassert.False(t, as.IsReserved("alice"))

	uassert.Equal(t, int(ReservedSystem), int(as.GetReservedCategory("admin")))
	uassert.Equal(t, int(ReservedGovernance), int(as.GetReservedCategory("govdao")))
	uassert.Equal(t, int(ReservedBrand), int(as.GetReservedCategory("gnoland")))
	uassert.Equal(t, 0, int(as.GetReservedCategory("nonexistent")))

	as.RemoveReservedName("admin")
	uassert.False(t, as.IsReserved("admin"))

	_, err := as.Register("admin", testAddr1)
	uassert.NoError(t, err, "previously reserved name should be registrable after removal")
}

func TestReservedNameNormalization(t *testing.T) {
	as := New(nil)

	as.AddReservedName("Admin", ReservedSystem)
	uassert.True(t, as.IsReserved("admin"))
	uassert.True(t, as.IsReserved("Admin"))
	uassert.True(t, as.IsReserved("ADMIN"))
}

func TestResolve(t *testing.T) {
	as := New(nil)

	addr := as.Resolve("alice")
	uassert.Equal(t, "", string(addr))

	as.Register("alice", testAddr1)
	addr = as.Resolve("alice")
	uassert.Equal(t, string(testAddr1), string(addr))

	addr = as.Resolve("Alice")
	uassert.Equal(t, string(testAddr1), string(addr))
}

func TestGetNamesForAddress(t *testing.T) {
	as := New(nil)

	names := as.GetNamesForAddress(testAddr1)
	uassert.True(t, names == nil, "should have no names initially")

	as.Register("alice", testAddr1)
	as.Register("bob", testAddr1)
	as.Register("carol", testAddr2)

	names = as.GetNamesForAddress(testAddr1)
	uassert.Equal(t, 2, len(names))
	uassert.Equal(t, "alice", names[0])
	uassert.Equal(t, "bob", names[1])

	names = as.GetNamesForAddress(testAddr2)
	uassert.Equal(t, 1, len(names))
	uassert.Equal(t, "carol", names[0])

	names = as.GetNamesForAddress(testAddr3)
	uassert.True(t, names == nil, "should have no names")
}

func TestGetRecord(t *testing.T) {
	as := New(nil)

	record := as.GetRecord("nonexistent")
	uassert.True(t, record == nil, "record should be nil for unregistered name")

	as.Register("alice", testAddr1)
	record = as.GetRecord("alice")
	uassert.True(t, record != nil, "record should not be nil")
	uassert.Equal(t, "alice", record.Name)
	uassert.Equal(t, string(testAddr1), string(record.Owner))
}

func TestCanRegisterValidation(t *testing.T) {
	as := New(nil)

	err := as.CanRegister("", testAddr1)
	uassert.ErrorIs(t, err, ErrNameEmpty)

	err = as.CanRegister("1invalid", testAddr1)
	uassert.ErrorIs(t, err, ErrNameInvalidChars)

	err = as.CanRegister("validname", testAddr1)
	uassert.NoError(t, err, "valid name should pass CanRegister")
}

func TestNameRecordIsExpired(t *testing.T) {
	record := &NameRecord{
		Name:         "test",
		Owner:        testAddr1,
		RegisteredAt: 100,
		ExpiresAt:    200,
	}

	uassert.False(t, record.IsExpired(150))
	uassert.False(t, record.IsExpired(200))
	uassert.True(t, record.IsExpired(201))
	uassert.True(t, record.IsExpired(300))

	noExpiry := &NameRecord{
		Name:         "permanent",
		Owner:        testAddr1,
		RegisteredAt: 100,
		ExpiresAt:    0,
	}
	uassert.False(t, noExpiry.IsExpired(999999))
}

func TestNameRecordIsPastGrace(t *testing.T) {
	record := &NameRecord{
		Name:         "test",
		Owner:        testAddr1,
		RegisteredAt: 100,
		ExpiresAt:    200,
	}

	gracePeriod := int64(50)

	uassert.False(t, record.IsPastGrace(150, gracePeriod))
	uassert.False(t, record.IsPastGrace(201, gracePeriod))
	uassert.False(t, record.IsPastGrace(250, gracePeriod))
	uassert.True(t, record.IsPastGrace(251, gracePeriod))
	uassert.True(t, record.IsPastGrace(300, gracePeriod))

	noExpiry := &NameRecord{ExpiresAt: 0}
	uassert.False(t, noExpiry.IsPastGrace(999999, gracePeriod))
}

func TestReservedCategoryString(t *testing.T) {
	uassert.Equal(t, "system", ReservedSystem.String())
	uassert.Equal(t, "governance", ReservedGovernance.String())
	uassert.Equal(t, "brand", ReservedBrand.String())
	uassert.Equal(t, "unknown", ReservedCategory(99).String())
}

func TestTotalRegistered(t *testing.T) {
	as := New(nil)

	uassert.Equal(t, 0, as.TotalRegistered())

	as.Register("alice", testAddr1)
	uassert.Equal(t, 1, as.TotalRegistered())

	as.Register("bob", testAddr2)
	uassert.Equal(t, 2, as.TotalRegistered())
}

func TestCalculateFeeNormalization(t *testing.T) {
	as := New(nil)

	uassert.Equal(t, int(as.CalculateFee("alice")), int(as.CalculateFee("Alice")))
	uassert.Equal(t, int(as.CalculateFee("alice")), int(as.CalculateFee("ALICE")))
}
