package keystore

import (
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

var data avl.Tree

var (
	BaseURL               = "/r/demo/keystore"
	StatusOK              = "ok"
	StatusNoUser          = "user not found"
	StatusNotFound        = "key not found"
	StatusNoWriteAccess   = "no write access"
	StatusCouldNotExecute = "could not execute"
)

func init() {
	data = avl.Tree{} // user -> avl.Tree
}

// KeyStore stores the owner-specific avl.Tree
type KeyStore struct {
	Owner std.Address
	Data  avl.Tree
}

// Set will set a value to a key
// requires write-access (original caller must be caller)
func Set(k, v string) string {
	origOwner := std.GetOrigCaller()
	return set(origOwner.String(), k, v)
}

// set (private) will set a key to value
// requires write-access (original caller must be caller)
func set(owner, k, v string) string {
	origOwner := std.GetOrigCaller()
	if origOwner.String() != owner {
		return StatusNoWriteAccess
	}
	keystoreInterface, exists := data.Get(owner)
	if !exists {
		data.Set(owner, &KeyStore{
			Owner: origOwner,
			Data:  avl.Tree{},
		})
		keystoreInterface, _ = data.Get(owner)
	}
	keystore := keystoreInterface.(*KeyStore)
	keystore.Data.Set(k, v)
	return StatusOK
}

// Remove removes a key
// requires write-access (original owner must be caller)
func Remove(k string) string {
	origOwner := std.GetOrigCaller()
	return remove(origOwner.String(), k)
}

// remove (private) removes a key
// requires write-access (original owner must be caller)
func remove(owner, k string) string {
	origOwner := std.GetOrigCaller()
	if origOwner.String() != owner {
		return StatusNoWriteAccess
	}
	keystoreInterface, exists := data.Get(owner)
	if !exists {
		data.Set(owner, &KeyStore{
			Owner: origOwner,
			Data:  avl.Tree{},
		})
		keystoreInterface, _ = data.Get(owner)
	}
	keystore := keystoreInterface.(*KeyStore)
	_, removed := keystore.Data.Remove(k)
	if !removed {
		return StatusCouldNotExecute
	}
	return StatusOK
}

// Get returns a value for a key
// read-only
func Get(owner, k string) string {
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
	origOwner := std.GetOrigCaller()
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

// Render provides url access to the functions of the keystore
// "" -> show all keystores listed by owner
// "owner" -> show all keys for that owner's keystore
// "owner:size" -> returns size of owner's keystore
// "owner:get:key" -> show value for that key in owner's keystore
// "owner:set:key:value" -> sets a key-value pair for owner's keystore (owner must be caller)
// "owner:remove:key" -> removes key from owner keystore (owner must be caller)
func Render(p string) string {
	var response string
	args := strings.Split(p, ":")
	numArgs := len(args)
	if p == "" {
		if data.Size() > 0 {
			data.Iterate("", "", func(key string, value interface{}) bool {
				ks := value.(*KeyStore)
				response += ufmt.Sprintf("- [%s](%s:%s) (%d keys)\n", ks.Owner, BaseURL, ks.Owner, ks.Data.Size())
				return false
			})
		} else {
			response = "no databases"
		}
	} else if numArgs == 1 {
		owner := args[0]
		keystoreInterface, exists := data.Get(owner)
		if !exists {
			return StatusNoUser
		}
		ks := keystoreInterface.(*KeyStore)
		i := 0
		response += ufmt.Sprintf("# %s database\n\n", ks.Owner)
		ks.Data.Iterate("", "", func(key string, value interface{}) bool {
			response += ufmt.Sprintf("- %d [%s](%s:%s:get:%s)\n", i, key, BaseURL, ks.Owner, key)
			i++
			return false
		})
	} else if numArgs == 2 {
		owner := args[0]
		cmd := args[1]
		if cmd == "size" {
			return size(owner)
		}
	} else if numArgs == 3 {
		owner := args[0]
		cmd := args[1]
		key := args[2]
		if cmd == "get" {
			return Get(owner, key)
		} else if cmd == "remove" {
			// remove will only work if caller is owner
			return remove(owner, key)
		}
	} else if numArgs == 4 {
		owner := args[0]
		cmd := args[1]
		key := args[2]
		val := args[3]
		if cmd == "set" {
			// set only works if caller is owner
			return set(owner, key, val)
		}
	}
	return response
}
