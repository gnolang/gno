package config

import (
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/urequire"
)

var (
	mainAddr   = address("g1wq2h93ppkf2gkgncz5unayrsmt7pl8npktnznd")
	backupAddr = address("g1hrfvdh7jdvnlxpk2y20tp3scj9jqal3zzu7wjz")

	addr1 = testutils.TestAddress("addr1")
	addr2 = testutils.TestAddress("addr2")
	addr3 = testutils.TestAddress("addr3")
)

func TestInitialOwnership(t *testing.T) {
	uassert.Equal(t, OwnableMain.Owner(), mainAddr)
	uassert.Equal(t, OwnableBackup.Owner(), backupAddr)
}

func TestIsAuthorized(t *testing.T) {
	tests := []struct {
		name string
		addr address
		want bool
	}{
		{"main address is authorized", mainAddr, true},
		{"backup address is authorized", backupAddr, true},
		{"random address not authorized", addr3, false},
		{"empty address not authorized", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsAuthorized(tt.addr)
			uassert.Equal(t, got, tt.want)
		})
	}
}

func TestSetMainAddr(t *testing.T) {
	testing.SetOriginCaller(mainAddr)

	// Test successful transfer
	err := SetMainAddr(addr2)
	urequire.NoError(t, err)
	uassert.Equal(t, OwnableMain.Owner(), addr2)

	// Test unauthorized transfer
	testing.SetOriginCaller(addr3)
	err = SetMainAddr(addr1)
	uassert.ErrorContains(t, err, "ownable: caller is not owner")

	// Test invalid address
	testing.SetOriginCaller(addr2)
	err = SetMainAddr("")
	uassert.ErrorContains(t, err, "ownable: new owner address is invalid")

	// Reset state
	testing.SetOriginCaller(addr2)
	err = SetMainAddr(mainAddr)
	urequire.NoError(t, err)
}

func TestSetBackupAddr(t *testing.T) {
	testing.SetOriginCaller(backupAddr)

	err := SetBackupAddr(addr2)
	urequire.NoError(t, err)
	uassert.Equal(t, OwnableBackup.Owner(), addr2)

	testing.SetOriginCaller(addr3)
	err = SetBackupAddr(addr1)
	uassert.ErrorContains(t, err, "ownable: caller is not owner")

	testing.SetOriginCaller(addr2)
	err = SetBackupAddr("")
	uassert.ErrorContains(t, err, "ownable: new owner address is invalid")

	testing.SetOriginCaller(addr2)
	err = SetBackupAddr(backupAddr)
	urequire.NoError(t, err)
}
