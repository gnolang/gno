// Package omnisearch owns every `$search` URL: the server-rendered results
// page and the JSON the header omnibar consumes.
//
// A query is free text plus any number of `key:value` qualifiers. One
// selector answers it, and names where the answer came from — the chain, or
// a configured indexer. Indexer-backed selectors exist only when one is
// configured. See gno.land/adr/prxxxx_gnoweb_omnisearch_indexer.md.
package omnisearch
