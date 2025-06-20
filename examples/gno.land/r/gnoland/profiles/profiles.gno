package profiles

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/mux"
	"gno.land/p/demo/ufmt"
)

var (
	fields = avl.NewTree()
	router = mux.NewRouter()
)

// Standard fields
const (
	DisplayName        = "DisplayName"
	Bio                = "Bio"
	Age                = "Age"
	Location           = "Location"
	GitHubHandle       = "GitHubHandle"
	XHandle            = "XHandle"
	Website            = "Website"
	Location           = "Location"
	ProfilePicture     = "ProfilePicture"
	AvailableForHiring = "AvailableForHiring"
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
	DisplayName:    true,
	Bio:            true,
	Location:       true,
	ProfilePicture: true,
	GitHubHandle:   true,
	XHandle:        true,
	Website:        true,
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

// SetStringField sets the specified string field value to the previous realm.
func SetStringField(cur realm, field, value string) bool {
	addr := std.PreviousRealm().Address()
	key := addr.String() + ":" + field
	updated := fields.Set(key, value)

	event := ProfileFieldCreated
	if updated {
		event = ProfileFieldUpdated
	}

	std.Emit(event, FieldType, StringField, field, value)

	return updated
}

// SetIntField sets the specified int field value to the previous realm.
func SetIntField(cur realm, field string, value int) bool {
	addr := std.PreviousRealm().Address()
	key := addr.String() + ":" + field
	updated := fields.Set(key, value)

	event := ProfileFieldCreated
	if updated {
		event = ProfileFieldUpdated
	}

	std.Emit(event, FieldType, IntField, field, string(value))

	return updated
}

// SetBoolField sets the specified bool field value to the previous realm.
func SetBoolField(cur realm, field string, value bool) bool {
	addr := std.PreviousRealm().Address()
	key := addr.String() + ":" + field
	updated := fields.Set(key, value)

	event := ProfileFieldCreated
	if updated {
		event = ProfileFieldUpdated
	}

	std.Emit(event, FieldType, BoolField, field, ufmt.Sprintf("%t", value))

	return updated
}

// Getters

func GetStringField(addr std.Address, field string) string {
	key := addr.String() + ":" + field
	if value, ok := fields.Get(key); ok {
		return value.(string)
	}

	return ""
}

func GetBoolField(addr std.Address, field string) bool {
	key := addr.String() + ":" + field
	if value, ok := fields.Get(key); ok {
		return value.(bool)
	}

	return false
}

func GetIntField(addr std.Address, field string, def int) int {
	key := addr.String() + ":" + field
	if value, ok := fields.Get(key); ok {
		return value.(int)
	}

	return def
}
