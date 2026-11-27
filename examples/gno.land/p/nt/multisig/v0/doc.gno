// Package multisig is the state machine for an M-of-N treasury.
//
// A hosting realm owns each Safe, authenticates callers with a live
// cur, and moves coins. This package never calls untrusted code and
// never keeps a realm value. Finished proposals past the retention
// cap are dropped; events are the record. v0 is unaudited.
package multisig
