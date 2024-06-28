package profile

import (
    "bytes"
    "gno.land/p/demo/ufmt"
    "gno.land/p/demo/mux"
    "strings"
    "std"
)

const (
    BaseURL           = "/r/demo/profile"
    SetStringFieldURL = BaseURL + "?help&__func=SetStringField&field=%s"
    SetIntFieldURL    = BaseURL + "?help&__func=SetIntField&field=%s"
    SetBoolFieldURL   = BaseURL + "?help&__func=SetBoolField&field=%s"
    GetStringFieldURL = BaseURL + "?help&__func=GetStringField&addr=%s"
    GetIntFieldURL    = BaseURL + "?help&__func=GetIntField&addr=%s"
    GetBoolFieldURL   = BaseURL + "?help&__func=GetBoolField&addr=%s"
    ViewAllFieldsURL  = BaseURL + ":u/%s"
    ViewFieldURL      = BaseURL + ":f/%s/%s"
)

func homeHandler(res *mux.ResponseWriter, req *mux.Request) {
    var b bytes.Buffer
    addresses := make(map[string]struct{})

    b.WriteString("<html><body>")
    b.WriteString("<h1>Profiles</h1>")
    b.WriteString("<h2>Setters</h2>")
    for field := range stringFields {
        link := ufmt.Sprintf(SetStringFieldURL, field)
        b.WriteString(ufmt.Sprintf("<p>[<a href='%s'>Set %s</a>]</p>", link, field))
    }

    for field := range intFields {
        link := ufmt.Sprintf(SetIntFieldURL, field)
        b.WriteString(ufmt.Sprintf("<p>[<a href='%s'>Set %s Field</a>]</p>", link, field))
    }

    for field := range boolFields {
        link := ufmt.Sprintf(SetBoolFieldURL, field)
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
        getStringLink := ufmt.Sprintf(GetStringFieldURL, address)
        getIntLink := ufmt.Sprintf(GetIntFieldURL, address)
        getBoolLink := ufmt.Sprintf(GetBoolFieldURL, address)

        b.WriteString(ufmt.Sprintf(
            "<p>Address: %s </br>"+
                "[<a href='%s'>View All Fields</a>] "+
                "[<a href='%s'>Get String Field</a>] "+
                "[<a href='%s'>Get Int Field</a>] "+
                "[<a href='%s'>Get Bool Field</a>]</p>",
            address, ufmt.Sprintf(ViewAllFieldsURL, address), getStringLink, getIntLink, getBoolLink))

        b.WriteString("<h2>View Specific Field:</h2>")

        b.WriteString("<h5>String Fields:</h5>")
        for field := range stringFields {
            fieldLink := ufmt.Sprintf(ViewFieldURL, address, field)
            b.WriteString(ufmt.Sprintf("[<a href='%s'>View %s</a>]<br/>", fieldLink, field))
        }
        b.WriteString("<h5>Int Fields:</h5>")
        for field := range intFields {
            fieldLink := ufmt.Sprintf(ViewFieldURL, address, field)
            b.WriteString(ufmt.Sprintf("[<a href='%s'>View %s</a>] <br/>", fieldLink, field))
        }
        b.WriteString("<h5>Bool Fields:</h5>")
        for field := range boolFields {
            fieldLink := ufmt.Sprintf(ViewFieldURL, address, field)
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
        value := GetStringField(address, field, "n/a")
        link := ufmt.Sprintf(SetStringFieldURL, field)
        b.WriteString(ufmt.Sprintf("<p>%s: %s [<a href='%s'>Edit</a>]</p>", field, value, link))
    }

    for field := range intFields {
        value := GetIntField(address, field, 0)
        link := ufmt.Sprintf(SetIntFieldURL, field)
        b.WriteString(ufmt.Sprintf("<p>%s: %d [<a href='%s'>Edit</a>]</p>", field, value, link))
    }

    for field := range boolFields {
        value := GetBoolField(address, field, false)
        link := ufmt.Sprintf(SetBoolFieldURL, field)
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
        editLink = ufmt.Sprintf("%s&addr=%s", SetStringFieldURL, field, addr)
    } else if _, ok := intFields[field]; ok {
        value = ufmt.Sprintf("%d", GetIntField(address, field, 0))
        editLink = ufmt.Sprintf(SetIntFieldURL+"&addr=%s", field, addr)
    } else if _, ok := boolFields[field]; ok {
        value = ufmt.Sprintf("%t", GetBoolField(address, field, false))
        editLink = ufmt.Sprintf(SetBoolFieldURL+"&addr=%s", field, addr)
    }

    b.WriteString(ufmt.Sprintf("<p>%s: %s [<a href='%s'>Edit</a>]</p>", field, value, editLink))

    b.WriteString("</body></html>")
    res.Write(b.String())
}

