package dao_registry

import (
	"std"
	"testing"
	"time"

	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/dao_registry"
)

func TestRegistry(t *testing.T) {
	registrations := List("", "", 0, false)
	if len(registrations) != 0 {
		t.Fatal("expected 0 registrations")
	}

	registrationsJSON := ListJSON("", "", 0, false)
	if registrationsJSON != "[]" {
		t.Fatal("expected empty JSON array")
	}

	pkgPath := "gno.land/r/demo/test_dao"
	addr := std.DerivePkgAddr(pkgPath)
	createdAt := time.Now().UnixMilli()
	doRegister(pkgPath, addr, "Test DAO", `This is a test DAO"\n\r\t\`+"\n\r\t", "https://example.com/image.png")

	registrations = List("", "", 0, false)
	if len(registrations) != 1 {
		t.Fatal("expected 1 registration")
	}
	reg := registrations[0]
	if reg.PkgPath != pkgPath {
		t.Fatal("expected pkgPath to match")
	}
	if reg.Addr != addr {
		t.Fatal("expected addr to match")
	}
	if reg.CreatedAt != createdAt {
		t.Fatal("expected createdAt to match")
	}

	registrationsJSON = ListJSON("", "", 0, false)
	expected := ufmt.Sprintf(`[{"pkgPath":"gno.land/r/demo/test_dao","addr":"g1m56y6xlx95sykjchn9h9q4e7k5u2zte4as9rdr","createdAt":%d,"name":"Test DAO","description":"This is a test DAO\"\\n\\r\\t\\\n\r\t","imageURI":"https://example.com/image.png"}]`, createdAt)
	if registrationsJSON != expected {
		t.Fatal("expected JSON to match, got:\n" + registrationsJSON + "\nexpected:\n" + expected)
	}

	registrationJSON := GetJSON(pkgPath)
	expected = ufmt.Sprintf(`{"pkgPath":"gno.land/r/demo/test_dao","addr":"g1m56y6xlx95sykjchn9h9q4e7k5u2zte4as9rdr","createdAt":%d,"name":"Test DAO","description":"This is a test DAO\"\\n\\r\\t\\\n\r\t","imageURI":"https://example.com/image.png"}`, createdAt)
	if registrationJSON != expected {
		t.Fatal("expected JSON to match, got:\n" + registrationJSON + "\nexpected:\n" + expected)
	}
}
