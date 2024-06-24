package profile

import (
	"bytes"
	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/mux"
	"std"
	"errors"
)

var fields = avl.NewTree()
var router = mux.NewRouter()

func init() {
	//router.HandleFunc("", homeHandler)
	//router.HandleFunc("u/{addr}", profileHandler)
	//router.HandleFunc("f/{addr}/{field}", fieldHandler)
}

// list of support string fields
var stringFields = map[string]bool{
	"DisplayName": true,
	"Homepage": true,
	"Bio": true,
	"Location": true,
	"Avatar": true,
	"GravatarEmail": true,
}

// list of support int fields
var intFields = map[int]bool {
    "Age": true
}

// list of support bool fields
var boolFields = map[string]bool{
	"AvailableForHiring": true,
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

//Getters

func GetStringField(addr std.Address, field, def string) string {
	key := addr.String() + ":" + field
	if value, ok := fields.Get(key); ok {
		return value.(string)
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

func GetBoolField(addr std.Address, field string, def bool) bool {
	key := addr.String() + ":" + field
	if value, ok := fields.Get(key); ok {
		return value.(bool)
	}
	
    return def
}

// Add getter and setter for bool and int fields

// render: home => generate links like this: https://gno.land/r/demo/boards?help&__func=CreateReply&bid=1&threadid=2&postid=2&body.type=textarea
// render: home/addr => display all fields for the address

// later: works with r/users transparently

/* read optim for later
type Profile struct {
	displayNamePtr *string
	usernamePtr *string
}

func (p Profile) DisplayName(def string) string {
	if p.displayNamePtr != nil {
		return *p.displayNamePtr
	}
	return def
}
*/


//-----------------------------

/*func AdminAddStringField(field string){
	assertIsDAO()
	stringFields[field] = true}
}
*/
