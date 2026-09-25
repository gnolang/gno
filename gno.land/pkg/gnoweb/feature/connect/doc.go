// Package connect owns gnoweb's wallet identity: the client-side session, the
// header identity control, and the /wallets install page. The session itself
// lives entirely in the browser (frontend/session.ts); the Go half here serves
// the install page from the same embedded registry the chooser reads, so the
// two cannot drift.
package connect
