package indexer

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
)

// The queries below read a whole band of heights in one request: they back
// aggregates (a week of calls, every importer of a package), which their
// callers build band by band, rather than a page of recent rows.
//
// tx-indexer lets a transaction whose message it could not decode
// (__typename UnexpectedMessage) through any message filter: on gnoland-1 in
// October 2026, 19 of the 20 rows a deploy filter for one package returned
// were such rows. Each query here therefore re-checks its rows against the
// message it filtered on before returning them.

// callFields is what CallsBetween selects: a week of calls is thousands of
// rows, so each carries only what a count needs.
const callFields = `
	block_height
	success
	messages {
		value {
			__typename
			... on MsgCall { caller pkg_path }
		}
	}`

// deployFields is what DeploysQuoting selects: which package each deploy
// added, and nothing of its source.
const deployFields = `
	block_height
	messages {
		value {
			__typename
			... on MsgAddPackage { package { path } }
		}
	}`

// CallsBetween returns every transaction carrying a MsgCall in the heights
// (lower, upper], in one query.
//
// The band is all or nothing: when it holds more rows than the indexer's
// element cap the error wraps ErrTooLarge and no rows are returned, because a
// count over part of a band would read as a complete one. The caller narrows
// the band instead.
func (c *Client) CallsBetween(ctx context.Context, lower, upper int) ([]Tx, error) {
	var out struct {
		Txs []Tx `json:"getTransactions"`
	}
	// `like` is a Go regexp on tx-indexer: "." matches any non-empty path.
	q := fmt.Sprintf(`{ getTransactions(where: {
		%s
		messages: { value: { MsgCall: { pkg_path: { like: "." } } } }
	}) { %s } }`, heightBand(lower, upper), callFields)
	if err := c.Query(ctx, q, &out); err != nil {
		return nil, err
	}
	return keepMatching(out.Txs, func(m Message) bool {
		return m.Type() == "MsgCall" && m.Value.PkgPath != ""
	}), nil
}

// DeploysQuoting returns every deploy in the heights (lower, upper] whose
// source contains text. Unlike SourceContains it does not stop at the newest
// few: a caller asking who imports a package needs every candidate, and walks
// the chain band by band to get them, since a scan of the whole chain outlasts
// the client's request timeout.
//
// A transaction matches when any package it deploys contains text, so a
// batched deploy brings along packages that do not. The caller has to check
// each one.
//
// When the answer exceeds the indexer's element cap, the rows it kept are
// returned along with an error wrapping ErrTooLarge, so the caller can say
// "at least".
func (c *Client) DeploysQuoting(ctx context.Context, text string, lower, upper int) ([]Tx, error) {
	var out struct {
		Txs []Tx `json:"getTransactions"`
	}
	// QuoteMeta for the reason SourceContains gives: `like` is a regexp.
	q := fmt.Sprintf(`{ getTransactions(where: {
		%s
		messages: { value: { MsgAddPackage: { package: { files: { body: { like: %s } } } } } }
	}) { %s } }`, heightBand(lower, upper), gqlString(regexp.QuoteMeta(text)), deployFields)
	err := c.Query(ctx, q, &out)
	if err != nil && !errors.Is(err, ErrTooLarge) {
		return nil, err
	}
	return keepMatching(out.Txs, func(m Message) bool {
		return m.Type() == "MsgAddPackage" && m.Path() != ""
	}), err
}

// heightBand filters on the heights (lower, upper]. A negative lower bound
// drops the lower filter altogether: `gt` excludes its bound, so no `gt` can
// reach the packages a chain was launched with, at height 0.
func heightBand(lower, upper int) string {
	if lower < 0 {
		return fmt.Sprintf(`block_height: { lt: %d }`, upper+1)
	}
	return fmt.Sprintf(`block_height: { gt: %d, lt: %d }`, lower, upper+1)
}

// keepMatching keeps the rows carrying at least one message match accepts.
func keepMatching(rows []Tx, match func(Message) bool) []Tx {
	return slices.DeleteFunc(rows, func(tx Tx) bool { return !slices.ContainsFunc(tx.Messages, match) })
}
