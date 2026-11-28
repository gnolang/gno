package echo

// NOTE: this import names a path on the *deployed* chain, not one in this
// tree. The harness runs `gno test .` here, which resolves gno.land/... from
// the live chain, so the path has to be one the chain actually serves — it is
// deliberately not rewritten when packages move in examples/. Update it when a
// chain ships the new path, not when the tree does.
import "gno.land/p/moul/md"

func Render(path string) string {
	// Fixed: escape attacker-controlled path text before inserting it in markdown.
	return "# Echo\n\n" + md.EscapeText(path)
}
