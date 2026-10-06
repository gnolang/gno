package indexer

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

// Rows the indexer could not decode, and rows for another message type, must
// not reach a count: they are what tx-indexer lets through a message filter.
func TestCallsBetweenKeepsOnlyDecodedCalls(t *testing.T) {
	t.Parallel()

	var query string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req gqlRequest
		_ = json.Unmarshal(body, &req)
		query = req.Query
		respond(w, `{"data":{"getTransactions":[
			{"block_height":5,"success":true,"messages":[{"value":{"__typename":"UnexpectedMessage"}}]},
			{"block_height":6,"success":true,"messages":[{"value":{"__typename":"BankMsgSend"}}]},
			{"block_height":7,"success":true,"messages":[{"value":{"__typename":"MsgCall","caller":"g1a","pkg_path":"gno.land/r/demo/boards"}}]}
		]}}`)
	})

	txs, err := c.CallsBetween(context.Background(), 4, 9)
	if err != nil {
		t.Fatalf("CallsBetween: %v", err)
	}
	if len(txs) != 1 || txs[0].Height != 7 {
		t.Fatalf("txs = %+v, want only the decoded call", txs)
	}
	if !strings.Contains(query, "gt: 4") || !strings.Contains(query, "lt: 10") {
		t.Errorf("band (4, 9] not encoded as gt: 4, lt: 10:\n%s", query)
	}
	if strings.Contains(query, "body") {
		t.Error("a call count must not select file bodies")
	}
}

// A capped band is refused whole: counting part of it would read as complete.
func TestCallsBetweenRefusesACappedBand(t *testing.T) {
	t.Parallel()

	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		respond(w, `{"data":{"getTransactions":[{"block_height":1,"messages":[{"value":{"__typename":"MsgCall","pkg_path":"gno.land/r/a"}}]}]},
			"errors":[{"message":"max elements per query reached"}]}`)
	})

	txs, err := c.CallsBetween(context.Background(), 0, 100)
	if !errors.Is(err, ErrTooLarge) {
		t.Fatalf("err = %v, want ErrTooLarge", err)
	}
	if len(txs) != 0 {
		t.Fatalf("txs = %d, want none from a capped band", len(txs))
	}
}

// A capped whole-chain search keeps its rows: the caller says "at least".
func TestDeploysQuotingKeepsRowsWhenCapped(t *testing.T) {
	t.Parallel()

	var query string
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req gqlRequest
		_ = json.Unmarshal(body, &req)
		query = req.Query
		respond(w, `{"data":{"getTransactions":[
			{"messages":[{"value":{"__typename":"UnexpectedMessage"}}]},
			{"messages":[{"value":{"__typename":"MsgAddPackage","package":{"path":"gno.land/r/user"}}}]}
		]},"errors":[{"message":"max elements per query reached"}]}`)
	})

	txs, err := c.DeploysQuoting(context.Background(), `"gno.land/p/a.b"`, -1, 1000)
	if !errors.Is(err, ErrTooLarge) {
		t.Fatalf("err = %v, want ErrTooLarge", err)
	}
	if len(txs) != 1 || txs[0].Messages[0].Path() != "gno.land/r/user" {
		t.Fatalf("txs = %+v, want the one decoded deploy", txs)
	}
	if strings.Contains(query, "gt:") || !strings.Contains(query, "lt: 1001") {
		t.Errorf("the bottom band must drop its lower bound to reach height 0:\n%s", query)
	}
	if got := likeValues(t, query); len(got) != 1 || got[0] != `"gno\.land/p/a\.b"` {
		t.Errorf("like = %q, want the quoted path with regexp metacharacters escaped", got)
	}
}
