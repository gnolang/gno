// Package t72b holds a type whose method writes the slice a caller hands it.
package t72b

type W struct{}

func (W) Put(b []byte) { b[0] = 'B' }
