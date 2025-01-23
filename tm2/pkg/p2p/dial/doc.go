// Package dial contains an implementation of a thread-safe priority dial queue. The queue is sorted by
// dial items, time ascending.
// The behavior of the dial queue is the following:
//
// - Peeking the dial queue will return the most urgent dial item, or nil if the queue is empty.
//
// - Popping the dial queue will return the most urgent dial item or nil if the queue is empty. Popping removes the dial item.
//
// - Push will push a new item to the dial queue, upon which the queue will find an adequate place for it.
package dial
