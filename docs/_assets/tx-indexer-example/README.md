# tx-indexer Example Assets

This folder contains self‑contained example snippets referenced in the `docs/resources/indexing-gno.md` guide. They illustrate how to:

1. Start and interact with the official `tx-indexer` (GraphQL + JSON-RPC)
2. Query `BankMsgSend` (GNOT transfer) transactions via GraphQL
3. Parse transaction JSON into typed Go structs (idiomatic `json.Unmarshal` usage)
4. Sort and display largest transfers
5. Subscribe to real‑time transactions over GraphQL WebSocket
6. Expose a minimal HTTP stats endpoint

## Related Docs
- Main guide: `docs/resources/indexing-gno.md`
- Upstream project: https://github.com/gnolang/tx-indexer
