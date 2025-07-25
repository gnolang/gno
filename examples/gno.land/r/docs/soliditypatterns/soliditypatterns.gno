package soliditypatterns

func Render(_ string) string {
	content := `# Solidity Famous Patterns

Welcome to most common solidity patterns in Solidity, transpiled and explained in go

## Examples

- [Store Hello World](/r/docs/soliditypatterns/helloworld) - Store hello world in a string in the contract.
- [Basic Counter](/r/docs/soliditypatterns/counter) - Increment and decrement a counter stored in a contract.
- [Only Owner can change](/r/docs/soliditypatterns/onlyownercanchange) - A contract that only the owner can change data inside.
- [State Lock](/r/docs/soliditypatterns/statelock) - A lock to dissalow transaction if it was locked

## Other resources

- [Official documentation](https://docs.gno.land) <!-- should be /docs with gnoweb embedding the docs/ folder. -->
`
	return content
}
