package rpcserver

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	types "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
)

func TestParseJSONMap(t *testing.T) {
	t.Parallel()

	input := []byte(`{"value":"1234","height":22}`)

	// naive is float,string
	var p1 map[string]interface{}
	err := json.Unmarshal(input, &p1)
	if assert.Nil(t, err) {
		h, ok := p1["height"].(float64)
		if assert.True(t, ok, "%#v", p1["height"]) {
			assert.EqualValues(t, 22, h)
		}
		v, ok := p1["value"].(string)
		if assert.True(t, ok, "%#v", p1["value"]) {
			assert.EqualValues(t, "1234", v)
		}
	}

	// preloading map with values doesn't help
	tmp := 0
	p2 := map[string]interface{}{
		"value":  &[]byte{},
		"height": &tmp,
	}
	err = json.Unmarshal(input, &p2)
	if assert.Nil(t, err) {
		h, ok := p2["height"].(float64)
		if assert.True(t, ok, "%#v", p2["height"]) {
			assert.EqualValues(t, 22, h)
		}
		v, ok := p2["value"].(string)
		if assert.True(t, ok, "%#v", p2["value"]) {
			assert.EqualValues(t, "1234", v)
		}
	}

	// preload here with *pointers* to the desired types
	// struct has unknown types, but hard-coded keys
	tmp = 0
	p3 := struct {
		Value  interface{} `json:"value"`
		Height interface{} `json:"height"`
	}{
		Height: &tmp,
		Value:  &[]byte{},
	}
	err = json.Unmarshal(input, &p3)
	if assert.Nil(t, err) {
		h, ok := p3.Height.(*int)
		if assert.True(t, ok, "%#v", p3.Height) {
			assert.Equal(t, 22, *h)
		}
		v, ok := p3.Value.(*[]byte)
		if assert.True(t, ok, "%#v", p3.Value) {
			// "1234" is interpreted as base64, decodes to the following bytes.
			assert.EqualValues(t, []byte{0xd7, 0x6d, 0xf8}, *v)
		}
	}

	// simplest solution, but hard-coded
	p4 := struct {
		Value  []byte `json:"value"`
		Height int    `json:"height"`
	}{}
	err = json.Unmarshal(input, &p4)
	if assert.Nil(t, err) {
		assert.EqualValues(t, 22, p4.Height)
		assert.EqualValues(t, []byte{0xd7, 0x6d, 0xf8}, p4.Value)
	}

	// so, let's use this trick...
	// dynamic keys on map, and we can deserialize to the desired types
	var p5 map[string]*json.RawMessage
	err = json.Unmarshal(input, &p5)
	if assert.Nil(t, err) {
		var h int
		err = json.Unmarshal(*p5["height"], &h)
		if assert.Nil(t, err) {
			assert.Equal(t, 22, h)
		}

		var v []byte
		err = json.Unmarshal(*p5["value"], &v)
		if assert.Nil(t, err) {
			assert.Equal(t, []byte{0xd7, 0x6d, 0xf8}, v)
		}
	}
}

func TestParseJSONArray(t *testing.T) {
	t.Parallel()

	input := []byte(`["1234",22]`)

	// naive is float,string
	var p1 []interface{}
	err := json.Unmarshal(input, &p1)
	if assert.Nil(t, err) {
		v, ok := p1[0].(string)
		if assert.True(t, ok, "%#v", p1[0]) {
			assert.EqualValues(t, "1234", v)
		}
		h, ok := p1[1].(float64)
		if assert.True(t, ok, "%#v", p1[1]) {
			assert.EqualValues(t, 22, h)
		}
	}

	// preloading map with values helps here (unlike map - p2 above)
	tmp := 0
	p2 := []interface{}{&[]byte{}, &tmp}
	err = json.Unmarshal(input, &p2)
	if assert.Nil(t, err) {
		v, ok := p2[0].(*[]byte)
		if assert.True(t, ok, "%#v", p2[0]) {
			assert.EqualValues(t, []byte{0xd7, 0x6d, 0xf8}, *v)
		}
		h, ok := p2[1].(*int)
		if assert.True(t, ok, "%#v", p2[1]) {
			assert.EqualValues(t, 22, *h)
		}
	}
}

func TestParseJSONRPC(t *testing.T) {
	t.Parallel()

	demo := func(ctx *types.Context, height int, name string) {}
	call := NewRPCFunc(demo, "height,name")

	cases := []struct {
		raw    string
		height int64
		name   string
		fail   bool
	}{
		// should parse
		{`["7", "flew"]`, 7, "flew", false},
		{`{"name": "john", "height": "22"}`, 22, "john", false},
		// defaults
		{`{"name": "solo", "unused": "stuff"}`, 0, "solo", false},
		// should fail - wrong types/length
		{`["flew", 7]`, 0, "", true},
		{`[7,"flew",100]`, 0, "", true},
		{`{"name": -12, "height": "fred"}`, 0, "", true},
	}
	for idx, tc := range cases {
		i := strconv.Itoa(idx)
		data := []byte(tc.raw)
		vals, err := jsonParamsToArgs(call, data)
		if tc.fail {
			assert.NotNil(t, err, i)
		} else {
			assert.Nil(t, err, "%s: %+v", i, err)
			if assert.Equal(t, 2, len(vals), i) {
				assert.Equal(t, tc.height, vals[0].Int(), i)
				assert.Equal(t, tc.name, vals[1].String(), i)
			}
		}
	}
}

func TestParseURINonJSON(t *testing.T) {
	t.Parallel()

	// Define a demo RPC function
	demo := func(ctx *types.Context, height int, name string, hash []byte) {}
	call := NewRPCFunc(demo, "height,name,hash")

	// Helper function to decode input base64 string to []byte
	decodeBase64 := func(input string) []byte {
		decoded, _ := base64.StdEncoding.DecodeString(input)
		return decoded
	}

	// Helper function to decode input hex string to []byte
	decodeHex := func(input string) []byte {
		decoded, _ := hex.DecodeString(input[2:])
		return decoded
	}

	// Test cases for non-JSON encoded parameters
	nonJSONCases := []struct {
		raw    []string
		height int64
		name   string
		hash   []byte
		fail   bool
	}{
		// can parse numbers unquoted and strings quoted
		{[]string{"7", `"flew"`, "rnpVPFlGJlauMNiL43Dmcl1U9loOBlib4L9OQAQ29tI="}, 7, "flew", decodeBase64("rnpVPFlGJlauMNiL43Dmcl1U9loOBlib4L9OQAQ29tI="), false},
		{[]string{"22", `"john"`, "/UztdqgPARnM25rjQ1lBsr3dlaaZuk2C8k4m5+bMvk8="}, 22, "john", decodeBase64("/UztdqgPARnM25rjQ1lBsr3dlaaZuk2C8k4m5+bMvk8="), false},
		{[]string{"-10", `"bob"`, "er/8eAAXG4732x8L8zMfJvgU1UH6b76BiU3NisFHh6E="}, -10, "bob", decodeBase64("er/8eAAXG4732x8L8zMfJvgU1UH6b76BiU3NisFHh6E="), false},
		// can parse numbers quoted, too
		{[]string{`"7"`, `"flew"`, "0x486173682076616c7565"}, 7, "flew", decodeHex("0x486173682076616c7565"), false}, // Testing hex encoded data
		{[]string{`"-10"`, `"bob"`, "0x6578616d706c65"}, -10, "bob", decodeHex("0x6578616d706c65"), false},           // Testing hex encoded data
		// can't parse strings unquoted
		{[]string{`"-10"`, `bob`, "invalid_encoded_data"}, -10, "bob", []byte("invalid_encoded_data"), true}, // Invalid encoded data format
	}

	// Iterate over test cases for non-JSON encoded parameters
	for idx, tc := range nonJSONCases {
		i := strconv.Itoa(idx)
		url := fmt.Sprintf("test.com/method?height=%v&name=%v&hash=%v", tc.raw[0], tc.raw[1], url.QueryEscape(tc.raw[2]))
		req, err := http.NewRequest("GET", url, nil)

		t.Error(req.URL)
		assert.NoError(t, err)

		// Invoke httpParamsToArgs to parse the request and convert to reflect.Values
		vals, err := httpParamsToArgs(call, req)

		// Check for expected errors or successful parsing
		if tc.fail {
			assert.NotNil(t, err, i)
		} else {
			assert.Nil(t, err, "%s: %+v", i, err)
			// Assert the parsed values match the expected height, name, and data

			if assert.Equal(t, 3, len(vals), i) {
				assert.Equal(t, tc.height, vals[0].Int(), i)
				assert.Equal(t, tc.name, vals[1].String(), i)
				assert.Equal(t, len(tc.hash), len(vals[2].Bytes()), i)
				assert.True(t, bytes.Equal(tc.hash, vals[2].Bytes()), i)
			}
		}
	}
}

func TestParseURIJSON(t *testing.T) {
	t.Parallel()

	type Data struct {
		Key string `json:"key"`
	}

	// Define a demo RPC function
	demo := func(ctx *types.Context, data Data) {}
	call := NewRPCFunc(demo, "data")

	// Test cases for JSON encoded parameters
	jsonCases := []struct {
		raw  string
		data Data
		fail bool
	}{
		// Valid JSON encoded values
		{`{"key": "value"}`, Data{Key: "value"}, false},
		{`{"id": 123}`, Data{}, false},         // Invalid field "id" (not in struct)
		{`{"list": [1, 2, 3]}`, Data{}, false}, // Invalid field "list" (not in struct)
		// Invalid JSON encoded values
		{`"string_data"`, Data{}, true},                // Invalid JSON format (not an object)
		{`12345`, Data{}, true},                        // Invalid JSON format (not an object)
		{`{"key": true}`, Data{}, true},                // Invalid field "key" type (expected string)
		{`{"key": {"nested": "value"}}`, Data{}, true}, // Invalid field "key" type (nested object)
	}

	// Iterate over test cases for JSON encoded parameters
	for idx, tc := range jsonCases {
		i := strconv.Itoa(idx)
		url := fmt.Sprintf("test.com/method?data=%v", url.PathEscape(tc.raw))
		req, err := http.NewRequest("GET", url, nil)
		assert.NoError(t, err)

		// Invoke httpParamsToArgs to parse the request and convert to reflect.Values
		vals, err := httpParamsToArgs(call, req)

		// Check for expected errors or successful parsing
		if tc.fail {
			assert.NotNil(t, err, i)
		} else {
			assert.Nil(t, err, "%s: %+v", i, err)
			// Assert the parsed values match the expected data
			if assert.Equal(t, 1, len(vals), i) {
				assert.Equal(t, tc.data, vals[0].Interface(), i)
			}
		}
	}
}
