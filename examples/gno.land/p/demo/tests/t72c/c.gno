// Package t72c holds a type whose method writes the slice a caller hands it.
package t72c

type C struct{}

func (C) Put(b []byte) { b[0] = 'C' }
