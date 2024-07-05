package db_test

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db"
)

// Empty iterator for empty db.
func TestPrefixIteratorNoMatchNil(t *testing.T) {
	t.Parallel()

	for _, backend := range db.BackendList() {
		t.Run(fmt.Sprintf("Prefix w/ backend %s", backend), func(t *testing.T) {
			t.Parallel()

			tmpdb := newTempDB(t, backend)
			itr := db.IteratePrefix(tmpdb, []byte("2"))

			checkInvalid(t, itr)
		})
	}
}

// Empty iterator for db populated after iterator created.
func TestPrefixIteratorNoMatch1(t *testing.T) {
	t.Parallel()

	for _, backend := range db.BackendList() {
		if backend == db.BoltDBBackend {
			t.Log("bolt does not support concurrent writes while iterating")
			continue
		}

		t.Run(fmt.Sprintf("Prefix w/ backend %s", backend), func(t *testing.T) {
			t.Parallel()

			tmpdb := newTempDB(t, backend)
			itr := db.IteratePrefix(tmpdb, []byte("2"))
			tmpdb.SetSync(bz("1"), bz("value_1"))

			checkInvalid(t, itr)
		})
	}
}

// Empty iterator for prefix starting after db entry.
func TestPrefixIteratorNoMatch2(t *testing.T) {
	t.Parallel()

	for _, backend := range db.BackendList() {
		t.Run(fmt.Sprintf("Prefix w/ backend %s", backend), func(t *testing.T) {
			t.Parallel()

			tmpdb := newTempDB(t, backend)
			tmpdb.SetSync(bz("3"), bz("value_3"))
			itr := db.IteratePrefix(tmpdb, []byte("4"))

			checkInvalid(t, itr)
		})
	}
}

// Iterator with single val for db with single val, starting from that val.
func TestPrefixIteratorMatch1(t *testing.T) {
	t.Parallel()

	for _, backend := range db.BackendList() {
		t.Run(fmt.Sprintf("Prefix w/ backend %s", backend), func(t *testing.T) {
			t.Parallel()

			tmpdb := newTempDB(t, backend)
			tmpdb.SetSync(bz("2"), bz("value_2"))
			itr := db.IteratePrefix(tmpdb, bz("2"))

			checkValid(t, itr, true)
			checkItem(t, itr, bz("2"), bz("value_2"))
			checkNext(t, itr, false)

			// Once invalid...
			checkInvalid(t, itr)
		})
	}
}

// Iterator with prefix iterates over everything with same prefix.
func TestPrefixIteratorMatches1N(t *testing.T) {
	t.Parallel()

	for _, backend := range db.BackendList() {
		t.Run(fmt.Sprintf("Prefix w/ backend %s", backend), func(t *testing.T) {
			t.Parallel()

			tmpdb := newTempDB(t, backend)

			// prefixed
			tmpdb.SetSync(bz("a/1"), bz("value_1"))
			tmpdb.SetSync(bz("a/3"), bz("value_3"))

			// not
			tmpdb.SetSync(bz("b/3"), bz("value_3"))
			tmpdb.SetSync(bz("a-3"), bz("value_3"))
			tmpdb.SetSync(bz("a.3"), bz("value_3"))
			tmpdb.SetSync(bz("abcdefg"), bz("value_3"))
			itr := db.IteratePrefix(tmpdb, bz("a/"))

			checkValid(t, itr, true)
			checkItem(t, itr, bz("a/1"), bz("value_1"))
			checkNext(t, itr, true)
			checkItem(t, itr, bz("a/3"), bz("value_3"))

			// Bad!
			checkNext(t, itr, false)

			// Once invalid...
			checkInvalid(t, itr)
		})
	}
}
