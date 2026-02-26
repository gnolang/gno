// PKGPATH: gno.land/p/test
package test

import (
	"encoding/hex"

	"gno.land/p/jefft0/avlhelpers"
	"gno.land/p/nt/avl"
	"gno.land/p/nt/ufmt"
)

func main() {
	tree := avl.NewTree()

	{
		// Empty tree.
		matches := avlhelpers.ListByteStringKeysByPrefix(tree, "", 10)
		println(ufmt.Sprintf("# matches: %d", len(matches)))
	}

	tree.Set("alice", "")
	tree.Set("andy", "")
	tree.Set("bob", "")

	{
		// Match only alice.
		matches := avlhelpers.ListByteStringKeysByPrefix(tree, "al", 10)
		println(ufmt.Sprintf("# matches: %d", len(matches)))
		println("match: " + matches[0])
	}

	{
		// Match alice and andy.
		matches := avlhelpers.ListByteStringKeysByPrefix(tree, "a", 10)
		println(ufmt.Sprintf("# matches: %d", len(matches)))
		println("match: " + matches[0])
		println("match: " + matches[1])
	}

	{
		// Match alice and andy limited to 1.
		matches := avlhelpers.ListByteStringKeysByPrefix(tree, "a", 1)
		println(ufmt.Sprintf("# matches: %d", len(matches)))
		println("match: " + matches[0])
	}

	tree = avl.NewTree()
	tree.Set("a\xff", "")
	tree.Set("a\xff\xff", "")
	tree.Set("b", "")
	tree.Set("\xff\xff\x00", "")

	{
		// Match only "a\xff\xff".
		matches := avlhelpers.ListByteStringKeysByPrefix(tree, "a\xff\xff", 10)
		println(ufmt.Sprintf("# matches: %d", len(matches)))
		println(ufmt.Sprintf("match: %s", hex.EncodeToString([]byte(matches[0]))))
	}

	{
		// Match "a\xff" and "a\xff\xff".
		matches := avlhelpers.ListByteStringKeysByPrefix(tree, "a\xff", 10)
		println(ufmt.Sprintf("# matches: %d", len(matches)))
		println(ufmt.Sprintf("match: %s", hex.EncodeToString([]byte(matches[0]))))
		println(ufmt.Sprintf("match: %s", hex.EncodeToString([]byte(matches[1]))))
	}

	{
		// Edge case: Match only "\xff\xff\x00".
		matches := avlhelpers.ListByteStringKeysByPrefix(tree, "\xff\xff", 10)
		println(ufmt.Sprintf("# matches: %d", len(matches)))
		println(ufmt.Sprintf("match: %s", hex.EncodeToString([]byte(matches[0]))))
	}
}

// Output:
// # matches: 0
// # matches: 1
// match: alice
// # matches: 2
// match: alice
// match: andy
// # matches: 1
// match: alice
// # matches: 1
// match: 61ffff
// # matches: 2
// match: 61ff
// match: 61ffff
// # matches: 1
// match: ffff00
