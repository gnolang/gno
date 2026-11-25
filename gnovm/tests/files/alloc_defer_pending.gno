// MAXALLOC: 5000000
package main

// Pending defers hold their record and copied arguments until the frame
// returns, so they must count against the allocation limit as they are
// registered. 100k pending defers of f(int) are ~12MB, over the 5MB limit.
// Before they were counted, this loop never touched the allocator.
func f(a int) {}

func main() {
	for i := 0; i < 100000; i++ {
		defer f(i)
	}
}

// Error:
// allocation limit exceeded
