package core

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPaginationPerPage(t *testing.T) {
	t.Parallel()

	cases := []struct {
		totalCount int
		perPage    int
		newPerPage int
	}{
		{5, 0, defaultPerPage},
		{5, 1, 1},
		{5, 2, 2},
		{5, defaultPerPage, defaultPerPage},
		{5, maxPerPage - 1, maxPerPage - 1},
		{5, maxPerPage, maxPerPage},
		{5, maxPerPage + 1, maxPerPage},
	}

	for _, c := range cases {
		p := validatePerPage(c.perPage)
		assert.Equal(t, c.newPerPage, p, fmt.Sprintf("%v", c))
	}
}
