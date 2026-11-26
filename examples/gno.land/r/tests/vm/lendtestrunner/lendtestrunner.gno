// Package lendtestrunner checks that the test runner's own call of a test or an
// example runs it as the realm's code, while a value the test hands to a
// testing helper does not write.
package lendtestrunner

var count int

func Count() int { return count }
