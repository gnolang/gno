// Package t55other is another realm that calls what it is handed.
package t55other

// Call calls f, entered without crossing.
func Call(f func()) { f() }

// CallX calls f inside a crossing call of this realm.
func CallX(cur realm, f func()) { f() }
