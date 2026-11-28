package run

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDomain(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "gno.land", New(Deps{}).deps.Domain)
	assert.Equal(t, "example.com", New(Deps{Domain: "example.com"}).deps.Domain)
}
