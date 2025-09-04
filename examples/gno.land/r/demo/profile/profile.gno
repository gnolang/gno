package profile

import (
	"chain"
	"chain/runtime"

	"gno.land/p/nt/avl"
	"gno.land/p/nt/mux"
	"gno.land/p/nt/ufmt"
)

var (
	fields = avl.NewTree()
	router = mux.NewRouter()
)

// Standard fields
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

// Events
const (
	ProfileFieldCreated = "ProfileFieldCreated"
	ProfileFieldUpdated = "ProfileFieldUpdated"
)

// Field types used when emitting event
const FieldType = "FieldType"

const (
	BoolField   = "BoolField"
	StringField = "StringField"
	IntField    = "IntField"
)

func init() {
	router.HandleFunc("", homeHandler)
	router.HandleFunc("u/{addr}", profileHandler)
	router.HandleFunc("f/{addr}/{field}", fieldHandler)
}

// List of supported string fields
var stringFields = map[string]bool{
	DisplayName:   true,
	Homepage:      true,
	Bio:           true,
	Location:      true,
	Avatar:        true,
	GravatarEmail: true,
}

// List of support int fields
var intFields = map[string]bool{
	Age: true,
}

// List of support bool fields
var boolFields = map[string]bool{
	AvailableForHiring: true,
}

// Setters

func SetStringField(cur realm, field, value string) bool {
	addr := runtime.PreviousRealm().Address()
	key := addr.String() + ":" + field
	updated := fields.Set(key, value)

	event := ProfileFieldCreated
	if updated {
		event = ProfileFieldUpdated
	}

	chain.Emit(event, FieldType, StringField, field, value)

	return updated
}

func SetIntField(cur realm, field string, value int) bool {
	addr := runtime.PreviousRealm().Address()
	key := addr.String() + ":" + field
	updated := fields.Set(key, value)

	event := ProfileFieldCreated
	if updated {
		event = ProfileFieldUpdated
	}

	chain.Emit(event, FieldType, IntField, field, string(value))

	return updated
}

func SetBoolField(cur realm, field string, value bool) bool {
	addr := runtime.PreviousRealm().Address()
	key := addr.String() + ":" + field
	updated := fields.Set(key, value)

	event := ProfileFieldCreated
	if updated {
		event = ProfileFieldUpdated
	}

	chain.Emit(event, FieldType, BoolField, field, ufmt.Sprintf("%t", value))

	return updated
}

// Getters

func GetStringField(addr address, field, def string) string {
	key := addr.String() + ":" + field
	if value, ok := fields.Get(key); ok {
		return value.(string)
	}

	return def
}

func GetBoolField(addr address, field string, def bool) bool {
	key := addr.String() + ":" + field
	if value, ok := fields.Get(key); ok {
		return value.(bool)
	}

	return def
}

func GetIntField(addr address, field string, def int) int {
	key := addr.String() + ":" + field
	if value, ok := fields.Get(key); ok {
		return value.(int)
	}

	return def
}
