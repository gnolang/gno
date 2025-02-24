// https://github.com/gnolang/gno/issues/3751

package math

import "testing"

func Add(a, b int) int {
    return a + b
}

func TestAdd(t *testing.T) {
    go Add(1, 1)
}

// Error:
// files/goroutine_not_supported.gno:12:5: goroutines are not permitted
