// Package chonk provides a simple way to store arbitrarily large strings
// in a linked list across transactions for efficient storage and retrieval.
// A Chonk support three operations: Add, Flush, and Scanner.
// - Add appends a string to the Chonk.
// - Flush clears the Chonk.
// - Scanner is used to iterate over the chunks in the Chonk.
package chonk

// Chonk is a linked list string storage and
// retrieval system for fine bois.
type Chonk struct {
	first *chunk
	last  *chunk
}

// chunk is a linked list node for Chonk
type chunk struct {
	text string
	next *chunk
}

// New creates a reference to a new Chonk
func New() *Chonk {
	return &Chonk{}
}

// Add appends a string to the Chonk. If the Chonk is empty,
// the string will be the first and last chunk. Otherwise,
// the string will be appended to the end of the Chonk.
func (c *Chonk) Add(text string) {
	next := &chunk{text: text}
	if c.first == nil {
		c.first = next
		c.last = next
		return
	}
	c.last.next = next
	c.last = next
}

// Flush clears the Chonk by setting the first and last
// chunks to nil. This will allow the garbage collector to
// free the memory used by the Chonk.
func (c *Chonk) Flush() {
	c.first = nil
	c.last = nil
}

// Scanner returns a new Scanner for the Chonk. The Scanner
// is used to iterate over the chunks in the Chonk.
func (c *Chonk) Scanner() *Scanner {
	return &Scanner{
		next: c.first,
	}
}

// Scanner is a simple string scanner for Chonk. It is used
// to iterate over the chunks in a Chonk from first to last.
type Scanner struct {
	current *chunk
	next    *chunk
}

// Scan advances the scanner to the next chunk. It returns
// true if there is a next chunk, and false if there is not.
func (s *Scanner) Scan() bool {
	if s.next != nil {
		s.current = s.next
		s.next = s.next.next
		return true
	}
	return false
}

// Text returns the current chunk. It is only valid to call
// this method after a call to Scan returns true. Expected usage:
//
//		scanner := chonk.Scanner()
//			for scanner.Scan() {
//	    		fmt.Println(scanner.Text())
//			}
func (s *Scanner) Text() string {
	return s.current.text
}
