package indexer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Tx is the subset this package selects. Never file bodies: a deploy carries
// its whole source, and a page of them is megabytes.
type Tx struct {
	Hash     string    `json:"hash"`
	Height   int       `json:"block_height"`
	Success  bool      `json:"success"`
	GasUsed  int       `json:"gas_used"`
	Messages []Message `json:"messages"`
}

// Message flattens the MessageValue union; Type names the populated group.
type Message struct {
	Value struct {
		Type    string `json:"__typename"`
		Caller  string `json:"caller"`
		PkgPath string `json:"pkg_path"`
		Func    string `json:"func"`
		Creator string `json:"creator"`
		From    string `json:"from_address"`
		To      string `json:"to_address"`
		Amount  string `json:"amount"`
		Package *struct {
			Path string `json:"path"`
		} `json:"package"`
	} `json:"value"`
}

// Type reports the message kind, e.g. "MsgCall".
func (m Message) Type() string { return m.Value.Type }

// Path reports the package a message acts on, when it names one.
func (m Message) Path() string {
	if m.Value.PkgPath != "" {
		return m.Value.PkgPath
	}
	if m.Value.Package != nil {
		return m.Value.Package.Path
	}
	return ""
}

// Signer reports the authoring address, whatever the kind calls it.
func (m Message) Signer() string {
	switch {
	case m.Value.Caller != "":
		return m.Value.Caller
	case m.Value.Creator != "":
		return m.Value.Creator
	default:
		return m.Value.From
	}
}

// txFields is shared so the queries cannot drift, and is free of
// `files { body }`.
const txFields = `
	hash
	success
	block_height
	gas_used
	messages {
		value {
			__typename
			... on MsgCall { caller pkg_path func }
			... on MsgAddPackage { creator package { path } }
			... on MsgRun { caller package { path } }
			... on BankMsgSend { from_address to_address amount }
		}
	}`

// heightTTL: every windowed query needs the tip, and without reuse each one
// costs two round trips.
const heightTTL = 5 * time.Second

// LatestBlockHeight returns the chain tip, cached briefly. It doubles as the
// freshness stamp, so a panel can say how far behind the indexer is.
func (c *Client) LatestBlockHeight(ctx context.Context) (int, error) {
	if height, ok := c.cachedTip(); ok {
		return height, nil
	}

	ch := c.tipGroup.DoChan("tip", func() (any, error) {
		// Re-check under the group: the caller that won the race has already
		// stored a fresh value by the time the others get here.
		if height, ok := c.cachedTip(); ok {
			return height, nil
		}

		// Detached from whichever request started it: singleflight has no
		// context, so the leader's would cancel every follower's answer.
		fetchCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), defaultTimeout)
		defer cancel()

		var out struct {
			LatestBlockHeight int `json:"latestBlockHeight"`
		}
		if err := c.Query(fetchCtx, `{ latestBlockHeight }`, &out); err != nil {
			return 0, err
		}

		c.tipMu.Lock()
		c.tip, c.tipAt = out.LatestBlockHeight, time.Now()
		c.tipMu.Unlock()
		return out.LatestBlockHeight, nil
	})

	// Each caller waits on its own deadline, not the leader's.
	select {
	case res := <-ch:
		if res.Err != nil {
			return 0, res.Err
		}
		return res.Val.(int), nil
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

// cachedTip returns the cached chain tip while it is still fresh.
func (c *Client) cachedTip() (int, bool) {
	c.tipMu.Lock()
	defer c.tipMu.Unlock()
	if c.tip > 0 && time.Since(c.tipAt) < heightTTL {
		return c.tip, true
	}
	return 0, false
}

// TxByHash looks up a single transaction.
func (c *Client) TxByHash(ctx context.Context, hash string) (*Tx, error) {
	var out struct {
		Txs []Tx `json:"getTransactions"`
	}
	q := fmt.Sprintf(`{ getTransactions(where: { hash: { eq: %s } }) { %s } }`,
		gqlString(hash), txFields)
	if err := c.Query(ctx, q, &out); err != nil {
		return nil, err
	}
	if len(out.Txs) == 0 {
		return nil, fmt.Errorf("transaction %w: %s", ErrNotFound, hash)
	}
	return &out.Txs[0], nil
}

// RecentByPackage returns the most recent transactions touching a package,
// either calling into it or deploying it.
func (c *Client) RecentByPackage(ctx context.Context, pkgPath string, limit int) ([]Tx, error) {
	p := gqlString(pkgPath)
	where := fmt.Sprintf(`_or: [
			{ messages: { value: { MsgCall: { pkg_path: { eq: %s } } } } }
			{ messages: { value: { MsgAddPackage: { package: { path: { eq: %s } } } } } }
		]`, p, p)
	return c.recent(ctx, where, limit)
}

// RecentByAddress returns the most recent transactions an address took part
// in, as caller, deployer, sender or recipient.
func (c *Client) RecentByAddress(ctx context.Context, addr string, limit int) ([]Tx, error) {
	a := gqlString(addr)
	where := fmt.Sprintf(`_or: [
			{ messages: { value: { MsgCall: { caller: { eq: %s } } } } }
			{ messages: { value: { MsgAddPackage: { creator: { eq: %s } } } } }
			{ messages: { value: { MsgRun: { caller: { eq: %s } } } } }
			{ messages: { value: { BankMsgSend: { from_address: { eq: %s } } } } }
			{ messages: { value: { BankMsgSend: { to_address: { eq: %s } } } } }
		]`, a, a, a, a, a)
	return c.recent(ctx, where, limit)
}

// Deploys returns the deploy history of a package, newest first.
func (c *Client) Deploys(ctx context.Context, pkgPath string, limit int) ([]Tx, error) {
	where := fmt.Sprintf(`messages: { value: { MsgAddPackage: { package: { path: { eq: %s } } } } }`,
		gqlString(pkgPath))
	return c.recent(ctx, where, limit)
}

// SourceContains matches a substring over every deployed file body — the
// most expensive query tx-indexer answers. Deployed source, not rendered
// output: nothing indexes what Render() prints.
func (c *Client) SourceContains(ctx context.Context, text string, limit int) ([]Tx, error) {
	where := fmt.Sprintf(`messages: { value: { MsgAddPackage: { package: { files: { body: { like: %s } } } } } }`,
		gqlString(text))
	return c.recent(ctx, where, limit)
}

// Block returns one block's header fields.
func (c *Client) Block(ctx context.Context, height int) (*Block, error) {
	var out struct {
		Blocks []Block `json:"getBlocks"`
	}
	q := fmt.Sprintf(`{ getBlocks(where: { height: { eq: %d } }) {
		hash height chain_id time num_txs
	} }`, height)
	if err := c.Query(ctx, q, &out); err != nil {
		return nil, err
	}
	if len(out.Blocks) == 0 {
		return nil, fmt.Errorf("block %w: %d", ErrNotFound, height)
	}
	return &out.Blocks[0], nil
}

// Block is a block header.
type Block struct {
	Hash    string    `json:"hash"`
	Height  int       `json:"height"`
	ChainID string    `json:"chain_id"`
	Time    time.Time `json:"time"`
	NumTxs  int       `json:"num_txs"`
}

// Window sizes. getTransactions takes only `where` and `order`, so height is
// the only way to bound a query. Start narrow: overshooting costs a payload
// that grows with the chain's density, undershooting one more round trip.
const (
	initialWindow = 2000
	windowGrowth  = 8
	// A quiet package has nothing to find at any width, and a reader waits.
	maxWindowSteps = 4

	// minWindowBudget: below it, a further window would answer after the
	// caller gave up.
	minWindowBudget = 900 * time.Millisecond
)

// recent walks DESC over a widening height window until it has `need` rows,
// reaches genesis, or runs out of steps.
func (c *Client) recent(ctx context.Context, where string, need int) ([]Tx, error) {
	if need <= 0 {
		return nil, nil
	}
	head, err := c.LatestBlockHeight(ctx)
	if err != nil {
		return nil, err
	}

	var last []Tx
	window := initialWindow
	for step := 0; step < maxWindowSteps; step, window = step+1, window*windowGrowth {
		// Each step re-scans the previous ones and the last covers a million
		// blocks, so a window nobody will see costs what one they read does.
		if dl, ok := ctx.Deadline(); ok && step > 0 && time.Until(dl) < minWindowBudget {
			break
		}

		from := head - window
		if from < 0 {
			from = 0
		}

		// `gt` excludes its bound, so the bottom window carries none: a chain
		// launched with packages at genesis keeps them at height 0.
		height := fmt.Sprintf(`block_height: { gt: %d }`, from)
		if from == 0 {
			height = ""
		}

		var out struct {
			Txs []Tx `json:"getTransactions"`
		}
		q := fmt.Sprintf(`{ getTransactions(
			where: { %s %s }
			order: { heightAndIndex: DESC }
		) { %s } }`, height, where, txFields)

		err := c.Query(ctx, q, &out)
		switch {
		case err == nil:
		case errors.Is(err, ErrTooLarge):
			// The answer, not a failure: the walk is DESC, so the rows kept
			// before the cap are the newest ones.
			return trim(out.Txs, need), nil
		default:
			return nil, err
		}

		last = out.Txs
		if len(last) >= need || from == 0 {
			break
		}
	}
	return trim(last, need), nil
}

func trim(txs []Tx, need int) []Tx {
	if len(txs) > need {
		return txs[:need]
	}
	return txs
}

// gqlString renders s as a GraphQL string literal. GraphQL string syntax is
// JSON-compatible, so json.Marshal is the escaping — and the only barrier
// between caller-supplied text and the query.
func gqlString(s string) string {
	b, err := json.Marshal(s)
	if err != nil { // unreachable for a string
		return `""`
	}
	return string(b)
}
