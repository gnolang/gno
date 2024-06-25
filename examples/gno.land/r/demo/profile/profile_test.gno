package profile

import (
    "testing"
    "gno.land/p/demo/testutils"
    "std"
)

func TestStringFields(t *testing.T) {
    addr := testutils.TestAddress("alice")
    std.TestSetRealm(std.NewUserRealm(addr))

    // Get before setting
    name := GetStringField(addr, "DisplayName", "anon")
    assertSameString(t, name, "anon")

    // Set
    if err := SetStringField("DisplayName", "Alice foo"); err != nil {
        t.Errorf("unexpected error: %v", err)
    }
    if err := SetStringField("Homepage", "https://example.com"); err != nil {
        t.Errorf("unexpected error: %v", err)
    }

    // Get after setting
    name = GetStringField(addr, "DisplayName", "anon")
    homepage := GetStringField(addr, "Homepage", "")
    bio := GetStringField(addr, "Bio", "42")

    assertSameString(t, name, "Alice foo")
    assertSameString(t, homepage, "https://example.com")
    assertSameString(t, bio, "42")
}

func TestIntFields(t *testing.T) {
    addr := testutils.TestAddress("bob")
    std.TestSetRealm(std.NewUserRealm(addr))

    // Get before setting
    age := GetIntField(addr, "Age", 25)
    assertSameInt(t, age, 25)

    // Set
    if err := SetIntField("Age", 30); err != nil {
        t.Errorf("unexpected error: %v", err)
    }

    // Get after setting
    age = GetIntField(addr, "Age", 25)
    assertSameInt(t, age, 30)
}

func TestBoolFields(t *testing.T) {
    addr := testutils.TestAddress("charlie")
    std.TestSetRealm(std.NewUserRealm(addr))
    
    // Get before setting
    hiring := GetBoolField(addr, "AvailableForHiring", false)
    assertSameBool(t, hiring, false)

    // Set
    if err := SetBoolField("AvailableForHiring", true); err != nil {
        t.Errorf("unexpected error: %v", err)
    }

    // Get after setting
    hiring = GetBoolField(addr, "AvailableForHiring", false)
    assertSameBool(t, hiring, true)
}

func TestInvalidStringField(t *testing.T) {
    addr := testutils.TestAddress("dave")
    std.TestSetRealm(std.NewUserRealm(addr))

    err := SetStringField("InvalidField", "test")
    if err == nil {
        t.Error("expected error for invalid field, got nil")
    }
}

func TestInvalidIntField(t *testing.T) {
    addr := testutils.TestAddress("eve")
    std.TestSetRealm(std.NewUserRealm(addr))

    err := SetIntField("InvalidField", 123)
    if err == nil {
        t.Error("expected error for invalid field, got nil")
    }
}

func TestInvalidBoolField(t *testing.T) {
    addr := testutils.TestAddress("frank")
    std.TestSetRealm(std.NewUserRealm(addr))

    err := SetBoolField("InvalidField", true)
    if err == nil {
        t.Error("expected error for invalid field, got nil")
    }
}

func TestMultipleProfiles(t *testing.T) {
    addr1 := testutils.TestAddress("user1")
    addr2 := testutils.TestAddress("user2")

    // Set profile for addr1
    std.TestSetRealm(std.NewUserRealm(addr1))
    if err := SetStringField("DisplayName", "User One"); err != nil {
        t.Errorf("unexpected error: %v", err)
    }

    // Set profile for addr2
    std.TestSetRealm(std.NewUserRealm(addr2))
    if err := SetStringField("DisplayName", "User Two"); err != nil {
        t.Errorf("unexpected error: %v", err)
    }

    // Get profiles
    std.TestSetRealm(std.NewUserRealm(addr1)) // Switch back to addr1
    name1 := GetStringField(addr1, "DisplayName", "anon")
    std.TestSetRealm(std.NewUserRealm(addr2)) // Switch back to addr2
    name2 := GetStringField(addr2, "DisplayName", "anon")

    assertSameString(t, name1, "User One")
    assertSameString(t, name2, "User Two")
}


// Assertions helpers
func assertSameString(t *testing.T, a, b string) {
    t.Helper()
    if a != b {
        t.Errorf("expected %q, got %q", a, b)
    }
}

func assertSameInt(t *testing.T, a, b int) {
    t.Helper()
    if a != b {
        t.Errorf("expected %d, got %d", a, b)
    }
}

func assertSameBool(t *testing.T, a, b bool) {
    t.Helper()
    if a != b {
        t.Errorf("expected %t, got %t", a, b)
    }
}

