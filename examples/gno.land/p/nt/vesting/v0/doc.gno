// Package vesting is the state machine for a cliff-and-vest grant.
//
// A hosting realm owns each Grant, authenticates callers with a live
// cur, and moves coins. This package never calls untrusted code and
// never keeps a realm value. Schedules are pure functions of their
// terms and a unix second. v0 is unaudited.
package vesting
