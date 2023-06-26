/*
Package foo20_lp provides a simple implementation of a fixed price liquidity queue for swapping two types of tokens.

A liquidity queue represents a queue of tokens that can be swapped for tokens of a different type at a fixed price. When a token is swapped, it is removed from its queue and a token of the other type is removed from the other queue. If there are not enough tokens in the other queue to swap with, the token is added to its queue to be swapped later when more tokens are added to the other queue.

Please note that this is a simple implementation and does not handle many practical considerations that would need to be addressed in a real-world application.
*/
package foo20_lp // import "gno.land/r/demo/foo20_lp"
