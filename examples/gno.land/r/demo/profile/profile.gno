package profile

import (
	"bytes"
	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	"gno.land/p/demo/mux"
	"std"
    "strings"
	"errors"
    "strconv"
)

var fields = avl.NewTree()
var router = mux.NewRouter()

func init() {
	router.HandleFunc("", homeHandler)
	router.HandleFunc("u/{addr}", profileHandler)
	router.HandleFunc("f/{addr}/{field}", fieldHandler)
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
var intFields = map[string]bool {
    "Age": true,
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

func Render(path string) string {
    return router.Render(path)
}

func homeHandler(res *mux.ResponseWriter, req *mux.Request) {
	var b bytes.Buffer
	addresses := make(map[string]struct{})

	b.WriteString("<html><body>")
	b.WriteString("<h1>Profiles</h1>")
    b.WriteString("<h2>Setters</h2>")
    for field := range stringFields {
        link := ufmt.Sprintf("/r/demo/profile?help&__func=SetStringField&field=%s", field)
        b.WriteString(ufmt.Sprintf("<p>[<a href='%s'>Set %s Field</a>]</p>", link, field))
    }

    for field := range intFields {
        link := ufmt.Sprintf("/r/demo/profile?help&__func=SetIntField&field=%s", field)
        b.WriteString(ufmt.Sprintf("<p>[<a href='%s'>Set %s Field</a>]</p>", link, field))
    }

    for field := range boolFields {
        link := ufmt.Sprintf("/r/demo/profile?help&__func=SetBoolField&field=%s", field)
        b.WriteString(ufmt.Sprintf("<p>[<a href='%s'>Set %s Field</a>]</p>", link, field))
    }

    b.WriteString("<p>--------------------------</p>")

	fields.Iterate("", "", func(key string, value interface{}) bool {
		parts := strings.SplitN(key, ":", 2)
		if len(parts) > 0 {
			addresses[parts[0]] = struct{}{}
		}
		return true
	})

    b.WriteString("<h2>All Address Profiles</h2>")

    for address := range addresses {
        getStringLink := ufmt.Sprintf("/r/demo/profile?help&__func=GetStringField&addr=%s", address)
        getIntLink := ufmt.Sprintf("/r/demo/profile?help&__func=GetIntField&addr=%s", address)
        getBoolLink := ufmt.Sprintf("/r/demo/profile?help&__func=GetBoolField&addr=%s", address)

        b.WriteString(ufmt.Sprintf(
            "<p>Address: %s </br>" +
            "[<a href='/r/demo/profile:u/%s'>View All Fields</a>] "+
                "[<a href='%s'>Get String Field</a>] "+
                "[<a href='%s'>Get Int Field</a>] "+
                "[<a href='%s'>Get Bool Field</a>]</p>",
            address, address, getStringLink, getIntLink, getBoolLink))
        
        b.WriteString("<h2>View Specific Field:</h2>")
        
        b.WriteString("<h5>Strings Fields:</h5>")
        for field := range stringFields {
            fieldLink := ufmt.Sprintf("/r/demo/profile:f/%s/%s", address, field)
            b.WriteString(ufmt.Sprintf("[<a href='%s'>View %s</a>]<br/> ", fieldLink, field))
        }
        b.WriteString("<h5>Int Fields:</h5>")
        for field := range intFields {
            fieldLink := ufmt.Sprintf("/r/demo/profile:f/%s/%s", address, field)
            b.WriteString(ufmt.Sprintf("[<a href='%s'>View %s</a>] <br/>", fieldLink, field))
        }
        b.WriteString("<h5>Bool Fields:</h5>")
        for field := range boolFields {
            fieldLink := ufmt.Sprintf("/r/demo/profile:f/%s/%s", address, field)
            b.WriteString(ufmt.Sprintf("[<a href='%s'>View %s</a>] ", fieldLink, field))
        }

        b.WriteString("<p>--------------------------</p>")
    }	
    
	b.WriteString("</body></html>")
	res.Write(b.String())
}

func profileHandler(res *mux.ResponseWriter, req *mux.Request) {
	var b bytes.Buffer
	addr := req.GetVar("addr")

	b.WriteString("<html><body>")
	b.WriteString(ufmt.Sprintf("<h1>Profile %s</h1>", addr))

	address := std.Address(addr)

	for field := range stringFields {
		value := GetStringField(address, field, "N/A")
        link := ufmt.Sprintf("/r/demo/profile?help&__func=SetStringField&field=%s", field)
		b.WriteString(ufmt.Sprintf("<p>%s: %s [<a href='%s'>Edit</a>]</p>", field, value, link))
	}

	for field := range intFields {
		value := GetIntField(address, field, 0)
        link := ufmt.Sprintf("/r/demo/profile?help&__func=SetIntField&field=%s", field)
		b.WriteString(ufmt.Sprintf("<p>%s: %d [<a href='%s'>Edit</a>]</p>", field, value, link))
	}

	for field := range boolFields {
		value := GetBoolField(address, field, false)
        link := ufmt.Sprintf("/r/demo/profile?help&__func=SetBoolField&field=%s", field)
		b.WriteString(ufmt.Sprintf("<p>%s: %t [<a href='%s'>Edit</a>]</p>", field, value, link))
	}

	b.WriteString("</body></html>")
	res.Write(b.String())
}

func fieldHandler(res *mux.ResponseWriter, req *mux.Request) {
    var b bytes.Buffer
    addr := req.GetVar("addr")
    field := req.GetVar("field")

    b.WriteString("<html><body>")
    b.WriteString(ufmt.Sprintf("<h1>Field %s for %s</h1>", field, addr))

    address := std.Address(addr)
    value := "N/A"
    var editLink string

    if _, ok := stringFields[field]; ok {
        value = GetStringField(address, field, "N/A")
        editLink = ufmt.Sprintf("/r/demo/profile?help&__func=SetStringField&field=%s&addr=%s", field, addr)
    } else if _, ok := intFields[field]; ok {
        value = ufmt.Sprintf("%d", GetIntField(address, field, 0))
        editLink = ufmt.Sprintf("/r/demo/profile?help&__func=SetIntField&field=%s&addr=%s", field, addr)
    } else if _, ok := boolFields[field]; ok {
        value = ufmt.Sprintf("%t", GetBoolField(address, field, false))
        editLink = ufmt.Sprintf("/r/demo/profile?help&__func=SetBoolField&field=%s&addr=%s", field, addr)
    }

    b.WriteString(ufmt.Sprintf("<p>%s: %s [<a href='%s'>Edit</a>]</p>", field, value, editLink))

    b.WriteString("</body></html>")
    res.Write(b.String())
}

