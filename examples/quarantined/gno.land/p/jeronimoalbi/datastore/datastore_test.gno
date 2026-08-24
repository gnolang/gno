package datastore

import (
	"testing"

	"gno.land/p/nt/uassert/v0"
	"gno.land/p/nt/urequire/v0"
)

func TestDatastoreCreateStorage(t *testing.T) {
	cases := []struct {
		name        string
		storageName string
		mustFail    bool
		setup       func(*Datastore)
	}{
		{
			name:        "success",
			storageName: "users",
		},
		{
			name:        "storage exists",
			storageName: "users",
			mustFail:    true,
			setup: func(db *Datastore) {
				db.CreateStorage("users")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var db Datastore
			if tc.setup != nil {
				tc.setup(&db)
			}

			storage := db.CreateStorage(tc.storageName)

			if tc.mustFail {
				uassert.Equal(t, nil, storage)
				return
			}

			urequire.NotEqual(t, nil, storage, "storage created")
			uassert.Equal(t, tc.storageName, storage.Name())
			uassert.True(t, db.HasStorage(tc.storageName))
		})
	}
}

func TestDatastoreHasStorage(t *testing.T) {
	var (
		db   Datastore
		name = "users"
	)

	uassert.False(t, db.HasStorage(name))

	db.CreateStorage(name)
	uassert.True(t, db.HasStorage(name))
}

func TestDatastoreGetStorage(t *testing.T) {
	var (
		db   Datastore
		name = "users"
	)

	storage := db.GetStorage(name)
	uassert.Equal(t, nil, storage)

	db.CreateStorage(name)

	storage = db.GetStorage(name)
	urequire.NotEqual(t, nil, storage, "storage found")
	uassert.Equal(t, name, storage.Name())
}
