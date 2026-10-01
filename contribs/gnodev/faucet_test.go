package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
)

func TestFaucet(t *testing.T) {
	const addr = "g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"

	type claim struct {
		form    bool   // form field instead of a JSON body
		address string // sent as is
		advance time.Duration
		sendErr error
		status  int
		sent    bool   // whether the sender was called
		body    string // substring expected in the response
	}

	cases := []struct {
		name   string
		claims []claim
	}{
		{"json claim", []claim{
			{address: addr, status: 200, sent: true, body: `"tx_hash":"CAFE"`},
		}},
		{"form claim", []claim{
			{form: true, address: addr, status: 200, sent: true, body: "tx <code>CAFE</code>"},
		}},
		{"invalid address", []claim{
			{address: "g1nope", status: 400, body: "not a valid g1 address"},
		}},
		{"html in the address is escaped", []claim{
			{form: true, address: "<script>x</script>", status: 400, body: "&lt;script&gt;"},
		}},
		{"cooldown, then claim again", []claim{
			{address: addr, status: 200, sent: true},
			{address: addr, advance: 30 * time.Second, status: 429, body: "try again in 30s"},
			{address: addr, advance: 30 * time.Second, status: 200, sent: true},
		}},
		{"a failed send does not start the cooldown", []claim{
			{address: addr, sendErr: errors.New("node down"), status: 500, sent: true, body: "send failed"},
			{address: addr, status: 200, sent: true},
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			now := time.Unix(1_700_000_000, 0)
			var sendErr error
			var sent bool

			f := newFaucet(discardLogger(), "test-chain", std.MustParseCoins("10000000ugnot"),
				func(_ context.Context, _ crypto.Address) (string, error) {
					sent = true
					if sendErr != nil {
						return "", sendErr
					}
					return "CAFE", nil
				})
			f.now = func() time.Time { return now }

			for i, c := range tc.claims {
				now = now.Add(c.advance)
				sendErr, sent = c.sendErr, false

				var req *http.Request
				if c.form {
					req = httptest.NewRequest("POST", "/faucet", strings.NewReader(url.Values{"address": {c.address}}.Encode()))
					req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				} else {
					req = httptest.NewRequest("POST", "/faucet", strings.NewReader(`{"address":"`+c.address+`"}`))
					req.Header.Set("Content-Type", "application/json")
				}
				rec := httptest.NewRecorder()
				f.ServeHTTP(rec, req)

				assert.Equal(t, c.status, rec.Code, "claim %d: %s", i, rec.Body.String())
				assert.Equal(t, c.sent, sent, "claim %d: sender called", i)
				assert.Contains(t, rec.Body.String(), c.body, "claim %d", i)
			}
		})
	}
}
