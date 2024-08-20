package profile

import (
	"errors"
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/mux"
	"gno.land/p/demo/json"
)

var (
	fields = avl.NewTree()
	router = mux.NewRouter()
	arbitraryFields = avl.NewTree()  // address => *avl.Tree (field => value)
)

const (
	DisplayName        = "DisplayName"
	Homepage           = "Homepage"
	Bio                = "Bio"
	Age                = "Age"
	Location           = "Location"
	Avatar             = "Avatar"
	GravatarEmail      = "GravatarEmail"
	AvailableForHiring = "AvailableForHiring"
	InvalidField       = "InvalidField"
)

func init() {
	router.HandleFunc("", homeHandler)
	router.HandleFunc("u/{addr}", profileHandler)
	router.HandleFunc("f/{addr}/{field}", fieldHandler)
}

// list of supported string fields
var stringFields = map[string]bool{
	DisplayName:   true,
	Homepage:      true,
	Bio:           true,
	Location:      true,
	Avatar:        true,
	GravatarEmail: true,
}

// list of support int fields
var intFields = map[string]bool{
	Age: true,
}

// list of support bool fields
var boolFields = map[string]bool{
	AvailableForHiring: true,
}

// Setters

func SetStringField(field, value string) error {
	addr := std.PrevRealm().Addr()
	if _, ok := stringFields[field]; !ok {
		return errors.New("invalid string field")
	}

	key := addr.String() + ":" + field
	fields.Set(key, value)

	return nil
}

func SetIntField(field string, value int) error {
	addr := std.PrevRealm().Addr()

	if _, ok := intFields[field]; !ok {
		return errors.New("invalid int field")
	}

	key := addr.String() + ":" + field
	fields.Set(key, value)

	return nil
}

func SetBoolField(field string, value bool) error {
	addr := std.PrevRealm().Addr()

	if _, ok := boolFields[field]; !ok {
		return errors.New("invalid bool field")
	}

	key := addr.String() + ":" + field
	fields.Set(key, value)

	return nil
}

func SetField(field string, value interface{}) error {
	addr := std.PrevRealm().Addr()

	if _, ok := boolFields[field]; ok {
		return SetBoolField(field, value.(bool))
	}

	if _, ok := intFields[field]; ok {
		return SetIntField(field, value.(int))
	}

	if _, ok := stringFields[field]; ok {
		return SetStringField(field, value.(string))
	}

	caller := addr.String()
	userFields, exists := arbitraryFields.Get(caller)
	if !exists {
		userFields = avl.NewTree()
		arbitraryFields.Set(caller, userFields)
	}

	userFields.(*avl.Tree).Set(field, value)
	return nil
}

// Getters

func GetStringField(addr std.Address, field, def string) string {
	key := addr.String() + ":" + field
	if value, ok := fields.Get(key); ok {
		return value.(string)
	}

	return def
}

func GetBoolField(addr std.Address, field string, def bool) bool {
	key := addr.String() + ":" + field
	if value, ok := fields.Get(key); ok {
		return value.(bool)
	}

	return def
}

func GetIntField(addr std.Address, field string, def int) int {
	key := addr.String() + ":" + field
	if value, ok := fields.Get(key); ok {
		return value.(int)
	}

	return def
}

func GetField(addr std.Address, field string, def interface{}) interface{} {
	if _, ok := boolFields[field]; ok {
		return GetBoolField(addr, field, def.(bool))
	}

	if _, ok := intFields[field]; ok {
		return GetIntField(addr, field, def.(int))
	}

	if _, ok := stringFields[field]; ok {
		return GetStringField(addr, field, def.(string))
	}

	caller := addr.String()
	userFields, exists := arbitraryFields.Get(caller)
	if !exists {
		return nil
	}

	val, exists :=  userFields.(*avl.Tree).Get(field)
	if !exists {
		return nil
	}

	return val
}