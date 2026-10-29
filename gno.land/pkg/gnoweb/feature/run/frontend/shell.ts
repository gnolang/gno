// Shell-safe pieces of the command the run page prints; shellQuote, which
// $help uses too, is in @gnoweb/js/utils.

// heredocDelimiter returns a heredoc delimiter that no line of code equals,
// so the heredoc carrying code ends where the code does and no line of it
// runs as a command.
export function heredocDelimiter(code: string): string {
	const lines = new Set(code.split("\n"));
	let delim = "GNO_EOF";
	for (let i = 1; lines.has(delim); i++) delim = `GNO_EOF_${i}`;
	return delim;
}
