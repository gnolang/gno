// PKGPATH: gno.land/r/test
package test

import (
	"encoding/hex"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/avlhelpers"
	"gno.land/p/demo/ufmt"
)

func main() {
	tree := avl.Tree{}

	{
		// Empty tree.
		matches := avlhelpers.ListKeysByPrefix(tree, "", 10)
		println(ufmt.Sprintf("# matches: %d", len(matches)))
	}

	tree.Set("alice", "")
	tree.Set("andy", "")
	tree.Set("bob", "")

	{
		// Match only alice.
		matches := avlhelpers.ListKeysByPrefix(tree, "al", 10)
		println(ufmt.Sprintf("# matches: %d", len(matches)))
		println("match: " + matches[0])
	}

	{
		// Match alice and andy.
		matches := avlhelpers.ListKeysByPrefix(tree, "a", 10)
		println(ufmt.Sprintf("# matches: %d", len(matches)))
		println("match: " + matches[0])
		println("match: " + matches[1])
	}

	{
		// Match alice and andy limited to 1.
		matches := avlhelpers.ListKeysByPrefix(tree, "a", 1)
		println(ufmt.Sprintf("# matches: %d", len(matches)))
		println("match: " + matches[0])
	}

	tree = avl.Tree{}
	tree.Set("a\xff", "")
	tree.Set("a\xff\xff", "")
	tree.Set("b", "")
	tree.Set("\xff\xff\x00", "")

	{
		// Match only "a\xff\xff".
		matches := avlhelpers.ListKeysByPrefix(tree, "a\xff\xff", 10)
		println(ufmt.Sprintf("# matches: %d", len(matches)))
		println(ufmt.Sprintf("match: %s", hex.EncodeToString([]byte(matches[0]))))
	}

	{
		// Match "a\xff" and "a\xff\xff".
		matches := avlhelpers.ListKeysByPrefix(tree, "a\xff", 10)
		println(ufmt.Sprintf("# matches: %d", len(matches)))
		println(ufmt.Sprintf("match: %s", hex.EncodeToString([]byte(matches[0]))))
		println(ufmt.Sprintf("match: %s", hex.EncodeToString([]byte(matches[1]))))
	}

	{
		// Edge case: Match only "\xff\xff\x00".
		matches := avlhelpers.ListKeysByPrefix(tree, "\xff\xff", 10)
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
