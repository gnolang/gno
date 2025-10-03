package keystore

import (
	"chain/runtime"
	"strings"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/ufmt"
)

var data avl.Tree

const (
	BaseURL               = "/r/demo/keystore"
	StatusOK              = "ok"
	StatusNoUser          = "user not found"
	StatusNotFound        = "key not found"
	StatusNoWriteAccess   = "no write access"
	StatusCouldNotExecute = "could not execute"
	StatusNoDatabases     = "no databases"
)

func init() {
	data = avl.Tree{} // user -> avl.Tree
}

// KeyStore stores the owner-specific avl.Tree
type KeyStore struct {
	Owner address
	Data  avl.Tree
}

// Set will set a value to a key
// requires write-access (original caller must be caller)
func Set(_ realm, k, v string) string {
	origOwner := runtime.OriginCaller()
	return set(origOwner.String(), k, v)
}

// set (private) will set a key to value
// requires write-access (original caller must be caller)
func set(owner, k, v string) string {
	origOwner := runtime.OriginCaller()
	if origOwner.String() != owner {
		return StatusNoWriteAccess
	}
	var keystore *KeyStore
	keystoreInterface, exists := data.Get(owner)
	if !exists {
		keystore = &KeyStore{
			Owner: origOwner,
			Data:  avl.Tree{},
		}
		data.Set(owner, keystore)
	} else {
		keystore = keystoreInterface.(*KeyStore)
	}
	keystore.Data.Set(k, v)
	return StatusOK
}

// Remove removes a key
// requires write-access (original owner must be caller)
func Remove(_ realm, k string) string {
	origOwner := runtime.OriginCaller()
	return remove(origOwner.String(), k)
}

// remove (private) removes a key
// requires write-access (original owner must be caller)
func remove(owner, k string) string {
	origOwner := runtime.OriginCaller()
	if origOwner.String() != owner {
		return StatusNoWriteAccess
	}
	var keystore *KeyStore
	keystoreInterface, exists := data.Get(owner)
	if !exists {
		keystore = &KeyStore{
			Owner: origOwner,
			Data:  avl.Tree{},
		}
		data.Set(owner, keystore)
	} else {
		keystore = keystoreInterface.(*KeyStore)
	}
	_, removed := keystore.Data.Remove(k)
	if !removed {
		return StatusCouldNotExecute
	}
	return StatusOK
}

// Get returns a value for a key
// read-only
func Get(k string) string {
	origOwner := runtime.OriginCaller()
	return remove(origOwner.String(), k)
}

// get (private) returns a value for a key
// read-only
func get(owner, k string) string {
	keystoreInterface, exists := data.Get(owner)
	if !exists {
		return StatusNoUser
	}
	keystore := keystoreInterface.(*KeyStore)
	val, found := keystore.Data.Get(k)
	if !found {
		return StatusNotFound
	}
	return val.(string)
}

// Size returns size of database
// read-only
func Size() string {
	origOwner := runtime.OriginCaller()
	return size(origOwner.String())
}

func size(owner string) string {
	keystoreInterface, exists := data.Get(owner)
	if !exists {
		return StatusNoUser
	}
	keystore := keystoreInterface.(*KeyStore)
	return ufmt.Sprintf("%d", keystore.Data.Size())
}

// Render provides read-only url access to the functions of the keystore
// "" -> show all keystores listed by owner
// "owner" -> show all keys for that owner's keystore
// "owner:size" -> returns size of owner's keystore
// "owner:get:key" -> show value for that key in owner's keystore
func Render(p string) string {
	var response string
	args := strings.Split(p, ":")
	numArgs := len(args)
	if p == "" {
		numArgs = 0
	}
	switch numArgs {
	case 0:
		if data.Size() == 0 {
			return StatusNoDatabases
		}
		data.Iterate("", "", func(key string, value any) bool {
			ks := value.(*KeyStore)
			response += ufmt.Sprintf("- [%s](%s:%s) (%d keys)\n", ks.Owner, BaseURL, ks.Owner, ks.Data.Size())
			return false
		})
	case 1:
		owner := args[0]
		keystoreInterface, exists := data.Get(owner)
		if !exists {
			return StatusNoUser
		}
		ks := keystoreInterface.(*KeyStore)
		i := 0
		response += ufmt.Sprintf("# %s database\n\n", ks.Owner)
		ks.Data.Iterate("", "", func(key string, value any) bool {
			response += ufmt.Sprintf("- %d [%s](%s:%s:get:%s)\n", i, key, BaseURL, ks.Owner, key)
			i++
			return false
		})
	case 2:
		owner := args[0]
		cmd := args[1]
		if cmd == "size" {
			return size(owner)
		}
	case 3:
		owner := args[0]
		cmd := args[1]
		key := args[2]
		if cmd == "get" {
			return get(owner, key)
		}
	}

	return response
}
