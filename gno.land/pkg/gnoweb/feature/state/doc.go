// Package state implements the gnoweb state-explorer feature.
//
// It exposes ?state* URLs: the initial full HTML page (?state,
// ?state&oid), htmx-driven HTML fragments (?state&frag=node|source),
// and the unchanged JSON API (?state&json, ?state&oid&json, ?state&tid&json).
package state
