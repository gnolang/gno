package indexer

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
)

// The queries below read a whole band of heights, or the whole chain, in one
// request: they back aggregates (a week of calls, every importer of a
// package) rather than a page of recent rows.
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
		block_height: { gt: %d, lt: %d }
		messages: { value: { MsgCall: { pkg_path: { like: "." } } } }
	}) { %s } }`, lower, upper+1, callFields)
	if err := c.Query(ctx, q, &out); err != nil {
		return nil, err
	}
	return keepMatching(out.Txs, func(m Message) bool {
		return m.Type() == "MsgCall" && m.Value.PkgPath != ""
	}), nil
}

// DeploysQuoting returns every deploy, at any height, whose source contains
// text, in one query over the whole chain, genesis included. Unlike
// SourceContains it is not windowed: a caller asking who imports a package
// needs every candidate, not the newest few.
//
// A transaction matches when any package it deploys contains text, so a
// batched deploy brings along packages that do not. The caller has to check
// each one.
//
// When the answer exceeds the indexer's element cap, the rows it kept are
// returned along with an error wrapping ErrTooLarge, so the caller can say
// "at least".
func (c *Client) DeploysQuoting(ctx context.Context, text string) ([]Tx, error) {
	var out struct {
		Txs []Tx `json:"getTransactions"`
	}
	// QuoteMeta for the reason SourceContains gives: `like` is a regexp.
	q := fmt.Sprintf(`{ getTransactions(where: {
		messages: { value: { MsgAddPackage: { package: { files: { body: { like: %s } } } } } }
	}) { %s } }`, gqlString(regexp.QuoteMeta(text)), deployFields)
	err := c.Query(ctx, q, &out)
	if err != nil && !errors.Is(err, ErrTooLarge) {
		return nil, err
	}
	return keepMatching(out.Txs, func(m Message) bool {
		return m.Type() == "MsgAddPackage" && m.Path() != ""
	}), err
}

// keepMatching returns the rows carrying at least one message match accepts.
func keepMatching(rows []Tx, match func(Message) bool) []Tx {
	out := rows[:0:0]
	for _, tx := range rows {
		if slices.ContainsFunc(tx.Messages, match) {
			out = append(out, tx)
		}
	}
	return out
}
