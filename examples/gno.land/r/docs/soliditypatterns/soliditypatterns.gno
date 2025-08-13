package soliditypatterns

func Render(_ string) string {
	content := `# Common Solidity Patterns

This is a list of common Solidity patterns, converted and explained to match Gno best practices.

## Examples

- [Basic Counter](/r/docs/soliditypatterns/counter) - Increment and decrement a counter stored in a contract.
- [Only Owner can change](/r/docs/soliditypatterns/ownable) - Simple example of ownership over a contract.
- [State Lock](/r/docs/soliditypatterns/statelock) - Modifiers.
- More common examples coming soon!

> [!NOTE]
>  If you have experience with Solidity, or a common pattern is missing from this list, please add more examples to the [Monorepo](https://github.com/gnolang/gno/tree/master/examples/gno.land/r/docs/soliditypatterns)!

`
	return content
}
