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
				txs, err := h.deps.Indexer.RecentByAddress(ctx, term, recentLimit)
				if err != nil {
					return nil, err
				}
				return h.txResults(txs), nil
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
					return nil, fmt.Errorf("%q is not a block height", term)
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
				txs, err := h.deps.Indexer.RecentByPackage(ctx, q.ChainPath, recentLimit)
				if err != nil {
					return nil, err
				}
				return h.txResults(txs), nil
			},
		},
		{
			Name:  "deploys",
			Hint:  "deploys",
			Label: "Deploy history",
			Scope: ScopePackage,
			Bare:  true,
			resolve: func(ctx context.Context, h *Handler, q *Query, term string) ([]Result, error) {
				txs, err := h.deps.Indexer.Deploys(ctx, q.ChainPath, recentLimit)
				if err != nil {
					return nil, err
				}
				return h.txResults(txs), nil
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

// resolveImporters lists packages whose source mentions this one. Labelled
// "mentions", not "imports": a substring match counts comments and string
// literals too.
func (h *Handler) resolveImporters(ctx context.Context, q *Query) ([]Result, error) {
	txs, err := h.deps.Indexer.SourceContains(ctx, q.ChainPath, recentLimit)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool, len(txs))
	out := make([]Result, 0, len(txs))
	for _, tx := range txs {
		for _, m := range tx.Messages {
			path := m.Path()
			// A package always mentions itself; that is not an importer.
			if path == "" || path == q.ChainPath || seen[path] {
				continue
			}
			seen[path] = true
			out = append(out, Result{
				Title:  path,
				Detail: "deployed by " + m.Signer(),
				Href:   chainPathHref(path, h.deps.Domain),
				Tags:   []string{"mentions", "block " + strconv.Itoa(tx.Height)},
			})
		}
	}
	return capResults(out), nil
}

// resolveContent lists packages whose source contains the term. The indexer
// returns deploys; duplicates collapse to the newest per package.
func (h *Handler) resolveContent(ctx context.Context, q *Query, term string) ([]Result, error) {
	txs, err := h.deps.Indexer.SourceContains(ctx, term, recentLimit)
	if err != nil {
		return nil, err
	}
	author, _ := q.Get(FilterAuthor)

	seen := make(map[string]bool, len(txs))
	out := make([]Result, 0, len(txs))
	for _, tx := range txs {
		for _, m := range tx.Messages {
			path := m.Path()
			if path == "" || seen[path] {
				continue
			}
			if !matchesAuthor(path, m.Signer(), author, h.deps.Domain) {
				continue
			}
			seen[path] = true
			out = append(out, Result{
				Title:  path,
				Detail: "deployed by " + m.Signer(),
				Href:   chainPathHref(path, h.deps.Domain),
				Tags:   []string{"contains " + term, "block " + strconv.Itoa(tx.Height)},
			})
		}
	}
	return capResults(out), nil
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
	return capResults(out)
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
