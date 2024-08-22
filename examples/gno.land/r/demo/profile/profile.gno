package profile

import (
	"std"

	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/avl"
	"gno.land/p/demo/mux"
)

var (
	fields = avl.NewTree()
	router = mux.NewRouter()
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

func SetStringField(field, value string) bool {
	addr := std.PrevRealm().Addr()
	key := addr.String() + ":" + field
	updated := fields.Set(key, value)

	event := "ProfileFieldCreated"
	if updated {
		event = "ProfileFieldUpdated"
	}

	std.Emit(event, "type", "String", field, value)

	return updated
}

func SetIntField(field string, value int) bool {
	addr := std.PrevRealm().Addr()
	key := addr.String() + ":" + field
	updated := fields.Set(key, value)

	event := "ProfileFieldCreated"
	if updated {
		event = "ProfileFieldUpdated"
	}

	std.Emit(event, "type", "Int", field, string(value))

	return updated
}

func SetBoolField(field string, value bool) bool {
	addr := std.PrevRealm().Addr()
	key := addr.String() + ":" + field
	updated := fields.Set(key, value)

	event := "ProfileFieldCreated"
	if updated {
		event = "ProfileFieldUpdated"
	}

	std.Emit(event, "type", "Bool", field, ufmt.Sprintf("%t", value))

	return updated		
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
