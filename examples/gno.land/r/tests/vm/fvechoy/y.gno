// Package fvechoy hands back the function value it is given.
package fvechoy

func Echo(f func(string)) func(string) { return f }
