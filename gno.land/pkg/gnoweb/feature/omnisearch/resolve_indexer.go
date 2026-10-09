package omnisearch

import (
	"context"
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/indexer"
)

// recentLimit matches MaxResults: fetching more than the page shows spends
// the indexer's time on rows nobody reads. One tx can expand to several
// rows, so capResults still applies.
const recentLimit = MaxResults

// indexerSelectors exist only when Deps.Indexer is non-nil: no hint, no
// autocompletion, no empty group without `-indexer-url`.
func indexerSelectors() []*Selector {
	return []*Selector{
		{
			Name:  "tx",
			Hint:  "tx:<hash>",
			Label: "Transaction",
			Scope: ScopeGlobal,
			resolve: func(ctx context.Context, h *Handler, q *Query, term string) ([]Result, error) {
				tx, err := h.deps.Indexer.TxByHash(ctx, term)
				if err != nil {
					return nil, err
				}
				return h.txResults([]indexer.Tx{*tx}), nil
			},
		},
		{
			Name:  "account",
			Hint:  "account:<address>",
			Label: "Account activity",
			Scope: ScopeGlobal,
			resolve: func(ctx context.Context, h *Handler, q *Query, term string) ([]Result, error) {
				// Rows travel with ErrPartial; Search tells it from a failure.
				txs, err := h.deps.Indexer.RecentByAddress(ctx, term, recentLimit)
				return h.txResults(txs), err
			},
		},
		{
			Name:  "block",
			Hint:  "block:<height>",
			Label: "Block",
			Scope: ScopeGlobal,
			resolve: func(ctx context.Context, h *Handler, q *Query, term string) ([]Result, error) {
				height, err := strconv.Atoi(term)
				if err != nil || height < 0 {
					return nil, inputError(fmt.Sprintf("%q is not a block height", term))
				}
				b, err := h.deps.Indexer.Block(ctx, height)
				if err != nil {
					return nil, err
				}
				return []Result{{
					Title:  "block " + strconv.Itoa(b.Height),
					Detail: b.Hash,
					Tags: []string{
						b.ChainID,
						strconv.Itoa(b.NumTxs) + " tx",
						b.Time.Format("2006-01-02 15:04:05 MST"),
					},
				}}, nil
			},
		},
		{
			Name:  "activity",
			Hint:  "activity",
			Label: "Recent activity",
			Scope: ScopePackage,
			Bare:  true,
			resolve: func(ctx context.Context, h *Handler, q *Query, term string) ([]Result, error) {
				// Rows travel with ErrPartial; Search tells it from a failure.
				txs, err := h.deps.Indexer.RecentByPackage(ctx, q.ChainPath, recentLimit)
				return h.txResults(txs), err
			},
		},
		{
			Name:  "deploys",
			Hint:  "deploys",
			Label: "Deploy history",
			Scope: ScopePackage,
			Bare:  true,
			resolve: func(ctx context.Context, h *Handler, q *Query, term string) ([]Result, error) {
				// Rows travel with ErrPartial; Search tells it from a failure.
				txs, err := h.deps.Indexer.Deploys(ctx, q.ChainPath, recentLimit)
				return h.txResults(txs), err
			},
		},
		{
			// Deployed SOURCE, not rendered output — nothing indexes what
			// Render() prints. The label says so, for the reader's sake.
			Name:  "content",
			Hint:  "content:<text>",
			Label: "Source code",
			Scope: ScopeGlobal,
			// A short term scans every deployed file body on the chain. The
			// CPU is the indexer's, but gnoweb is the amplifier.
			MinTerm:  4,
			PageOnly: true,
			resolve: func(ctx context.Context, h *Handler, q *Query, term string) ([]Result, error) {
				return h.resolveContent(ctx, q, term)
			},
		},
		{
			Name:  "importers",
			Hint:  "importers",
			Label: "Imported by",
			Scope: ScopePackage,
			Bare:  true,
			// Same query as content, same reason to keep it off keystrokes.
			PageOnly: true,
			resolve: func(ctx context.Context, h *Handler, q *Query, term string) ([]Result, error) {
				return h.resolveImporters(ctx, q)
			},
		},
	}
}

// resolveImporters lists packages whose source quotes this one's path.
// Labelled "mentions", not "imports": a string literal outside an import
// block counts too.
func (h *Handler) resolveImporters(ctx context.Context, q *Query) ([]Result, error) {
	// Rows travel with ErrPartial; Search tells it from a failure.
	txs, err := h.deps.Indexer.DeploysImporting(ctx, q.ChainPath, recentLimit)

	seen := make(map[string]bool, len(txs))
	out := make([]Result, 0, len(txs))
	for _, tx := range txs {
		batch := deployCount(tx) > 1
		for _, m := range tx.Messages {
			path := m.Path()
			// A package always mentions itself; that is not an importer.
			if m.Type() != "MsgAddPackage" || path == "" || path == q.ChainPath || seen[path] {
				continue
			}
			seen[path] = true
			out = append(out, Result{
				Title:  path,
				Detail: "deployed by " + m.Signer(),
				Href:   chainPathHref(path, h.deps.Domain),
				Tags:   batchTags(batch, "mentions", "block "+strconv.Itoa(tx.Height)),
			})
		}
	}
	return out, newestOnly(txs, err)
}

// newestOnly flags a deploy search that filled its page: older deploys may
// match too, and the page shows the newest only.
func newestOnly(txs []indexer.Tx, err error) error {
	if err == nil && len(txs) >= recentLimit {
		return partialAnswerError(fmt.Sprintf(
			"Showing the %d newest matching deploys: older ones may match too.", recentLimit))
	}
	return err
}

// deployCount counts the packages a transaction deploys. The indexer returns
// a matched transaction whole, every message included.
func deployCount(tx indexer.Tx) int {
	n := 0
	for _, m := range tx.Messages {
		if m.Type() == "MsgAddPackage" {
			n++
		}
	}
	return n
}

// batchTags marks a row from a transaction deploying several packages: the
// indexer matched one of them, and does not say which.
func batchTags(batch bool, tags ...string) []string {
	if batch {
		return append(tags, "batch deploy")
	}
	return tags
}

// resolveContent lists packages whose source contains the term. The indexer
// returns deploys; duplicates collapse to the newest per package.
func (h *Handler) resolveContent(ctx context.Context, q *Query, term string) ([]Result, error) {
	// The author goes into the indexer's filter: applied here alone, it would
	// only thin out the newest matches, and miss the author's older ones.
	author, _ := q.Get(FilterAuthor)
	// Rows travel with ErrPartial; Search tells it from a failure.
	txs, err := h.deps.Indexer.SourceContains(ctx, term, author, recentLimit)

	seen := make(map[string]bool, len(txs))
	out := make([]Result, 0, len(txs))
	for _, tx := range txs {
		batch := deployCount(tx) > 1
		for _, m := range tx.Messages {
			path := m.Path()
			// Only a deploy carries source; a call in the same transaction
			// matched nothing.
			if m.Type() != "MsgAddPackage" || path == "" || seen[path] {
				continue
			}
			// Still checked per message: a transaction that matched may
			// carry other deploys.
			if !matchesAuthor(path, m.Signer(), author, h.deps.Domain) {
				continue
			}
			seen[path] = true
			out = append(out, Result{
				Title:  path,
				Detail: "deployed by " + m.Signer(),
				Href:   chainPathHref(path, h.deps.Domain),
				Tags:   batchTags(batch, "contains "+term, "block "+strconv.Itoa(tx.Height)),
			})
		}
	}
	return out, newestOnly(txs, err)
}

// txResults renders transactions as rows. gnoweb has no transaction page, so
// a row links to the package the transaction touched.
func (h *Handler) txResults(txs []indexer.Tx) []Result {
	out := make([]Result, 0, len(txs))
	for _, tx := range txs {
		r := Result{
			Title:  txTitle(tx),
			Detail: txDetail(tx),
			Tags: []string{
				"block " + strconv.Itoa(tx.Height),
				strconv.Itoa(tx.GasUsed) + " gas",
			},
		}
		if !tx.Success {
			r.Tags = append([]string{"failed"}, r.Tags...)
		}
		if len(tx.Messages) > 0 {
			if p := tx.Messages[0].Path(); p != "" {
				r.Href = chainPathHref(p, h.deps.Domain)
			}
		}
		out = append(out, r)
	}
	return out
}

// txTitle summarises by first message, and says when there are more.
func txTitle(tx indexer.Tx) string {
	if len(tx.Messages) == 0 {
		return tx.Hash
	}
	m := tx.Messages[0]
	var head string
	switch m.Type() {
	case "MsgCall":
		head = m.Value.Func + "() on " + m.Path()
	case "MsgAddPackage":
		head = "deploy " + m.Path()
	case "MsgRun":
		head = "run"
	case "BankMsgSend":
		head = "send " + m.Value.Amount + " to " + m.Value.To
	default:
		head = strings.TrimPrefix(m.Type(), "Msg")
	}
	if n := len(tx.Messages); n > 1 {
		head += fmt.Sprintf(" (+%d more)", n-1)
	}
	return head
}

func txDetail(tx indexer.Tx) string {
	if len(tx.Messages) == 0 {
		return ""
	}
	if s := tx.Messages[0].Signer(); s != "" {
		return s
	}
	return tx.Hash
}

// chainPathHref turns a chain-qualified path back into a gnoweb link, or
// nothing when it would point at a local page that does not exist.
func chainPathHref(chainPath, domain string) template.URL {
	return safePathHref(strings.TrimPrefix(chainPath, domain))
}
