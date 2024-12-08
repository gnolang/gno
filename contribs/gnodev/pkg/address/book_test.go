package address

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testAddr = crypto.MustAddressFromString("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5")

func TestNewBook(t *testing.T) {
	t.Parallel()

	bk := NewBook()
	assert.Empty(t, bk.addrsToNames)
	assert.Empty(t, bk.namesToAddrs)
}

func TestAddEmptyName(t *testing.T) {
	t.Parallel()

	bk := NewBook()

	// Add address
	bk.Add(testAddr, "")

	names, ok := bk.GetByAddress(testAddr)
	require.True(t, ok)
	require.Equal(t, 0, len(names))
}

func TestAdd(t *testing.T) {
	bk := NewBook()

	// Add address
	bk.Add(testAddr, "testname")

	t.Run("get by address", func(t *testing.T) {
		names, ok := bk.GetByAddress(testAddr)
		require.True(t, ok)
		require.Equal(t, 1, len(names))
		assert.Equal(t, "testname", names[0])
	})

	t.Run("get by name", func(t *testing.T) {
		addrFromName, ok := bk.GetByName("testname")
		assert.True(t, ok)
		assert.True(t, addrFromName.Compare(testAddr) == 0)
	})

	// Add same address with a new name
	bk.Add(testAddr, "testname2")

	t.Run("get two names with same address", func(t *testing.T) {
		// Get by name
		addr1, ok := bk.GetByName("testname")
		require.True(t, ok)
		addr2, ok := bk.GetByName("testname2")
		require.True(t, ok)
		assert.True(t, addr1.Compare(addr2) == 0)
	})
}

func TestList(t *testing.T) {
	t.Parallel()

	bk := NewBook()

	bk.Add(testAddr, "testname")

	entries := bk.List()

	assert.Equal(t, 1, len(entries))
	entry := entries[0]

	assert.True(t, testAddr.Compare(entry.Address) == 0)
	assert.Equal(t, 1, len(entries[0].Names))
	assert.Equal(t, "testname", entries[0].Names[0])
}

func TestGetFromNameOrAddress(t *testing.T) {
	t.Parallel()

	bk := NewBook()

	t.Run("failure", func(t *testing.T) {
		t.Parallel()

		resultAddr, names, ok := bk.GetFromNameOrAddress("unknown_key")
		assert.False(t, ok)
		assert.True(t, resultAddr.IsZero())
		assert.Len(t, names, 0)
	})

	// Add address
	bk.Add(testAddr, "testname")

	for _, addrOrName := range []string{"testname", testAddr.String()} {
		t.Run(addrOrName, func(t *testing.T) {
			resultAddr, names, ok := bk.GetFromNameOrAddress("testname")
			require.True(t, ok)
			require.Len(t, names, 1)
			assert.Equal(t, "testname", names[0])
			assert.True(t, resultAddr.Compare(testAddr) == 0)
		})
	}
}
