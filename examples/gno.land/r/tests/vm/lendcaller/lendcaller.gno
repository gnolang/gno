// Package lendcaller is another realm that calls what it is handed.
package lendcaller

// Call calls f, entered without crossing.
func Call(f func()) { f() }

// CallX calls f inside a crossing call of this realm.
func CallX(cur realm, f func()) { f() }
