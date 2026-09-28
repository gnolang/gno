package avlhelpers

import (
	"encoding/hex"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/ufmt/v0"
)

func ExampleTextMatch() {
	tree := avl.NewTree()

	{
		// Empty tree.
		matches := ListByteStringKeysByPrefix(tree, "", 10)
		println(ufmt.Sprintf("# matches: %d", len(matches)))
	}

	tree.Set("alice", "")
	tree.Set("andy", "")
	tree.Set("bob", "")

	{
		// Match only alice.
		matches := ListByteStringKeysByPrefix(tree, "al", 10)
		println(ufmt.Sprintf("# matches: %d", len(matches)))
		println("match: " + matches[0])
	}

	{
		// Match alice and andy.
		matches := ListByteStringKeysByPrefix(tree, "a", 10)
		println(ufmt.Sprintf("# matches: %d", len(matches)))
		println("match: " + matches[0])
		println("match: " + matches[1])
	}

	{
		// Match alice and andy limited to 1.
		matches := ListByteStringKeysByPrefix(tree, "a", 1)
		println(ufmt.Sprintf("# matches: %d", len(matches)))
		println("match: " + matches[0])
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
}

func ExampleBinaryMatch() {
	tree := avl.NewTree()
	tree.Set("a\xff", "")
	tree.Set("a\xff\xff", "")
	tree.Set("b", "")
	tree.Set("\xff\xff\x00", "")

	{
		// Match only "a\xff\xff".
		matches := ListByteStringKeysByPrefix(tree, "a\xff\xff", 10)
		println(ufmt.Sprintf("# matches: %d", len(matches)))
		println(ufmt.Sprintf("match: %s", hex.EncodeToString([]byte(matches[0]))))
	}

	{
		// Match "a\xff" and "a\xff\xff".
		matches := ListByteStringKeysByPrefix(tree, "a\xff", 10)
		println(ufmt.Sprintf("# matches: %d", len(matches)))
		println(ufmt.Sprintf("match: %s", hex.EncodeToString([]byte(matches[0]))))
		println(ufmt.Sprintf("match: %s", hex.EncodeToString([]byte(matches[1]))))
	}

	{
		// Edge case: Match only "\xff\xff\x00".
		matches := ListByteStringKeysByPrefix(tree, "\xff\xff", 10)
		println(ufmt.Sprintf("# matches: %d", len(matches)))
		println(ufmt.Sprintf("match: %s", hex.EncodeToString([]byte(matches[0]))))
	}

	// Output:
	// # matches: 1
	// match: 61ffff
	// # matches: 2
	// match: 61ff
	// match: 61ffff
	// # matches: 1
	// match: ffff00
}
