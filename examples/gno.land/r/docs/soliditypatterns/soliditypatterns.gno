package soliditypatterns

func Render(_ string) string {
	content := `# Common Solidity Patterns

This is a list of common Solidity patterns, converted and explained to match Gno best practices.

## Examples

- [Basic Counter](/r/docs/soliditypatterns/counter) - Increment and decrement a counter stored in a contract.
- [Only Owner can change](/r/docs/soliditypatterns/onlyownercanchange) - A contract that only the owner can change data inside.
- [State Lock](/r/docs/soliditypatterns/statelock) - A lock to dissalow transaction if it was locked

## Other resources

- You can also run your raw EVM bytecode from compiled Solidity on [Gno EVM](/r/demo/evm)
- [Official documentation](https://docs.gno.land) <!-- should be /docs with gnoweb embedding the docs/ folder. -->
`
	return content
}
